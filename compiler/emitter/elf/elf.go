package elf

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"

	"github.com/nathfavour/ship/compiler/ir"
)

// ELF64 Headers and Definitions
const (
	EI_NIDENT = 16
	PT_LOAD   = 1
	PF_X      = 1
	PF_W      = 2
	PF_R      = 4
)

type Elf64Ehdr struct {
	Ident     [EI_NIDENT]byte
	Type      uint16
	Machine   uint16
	Version   uint32
	Entry     uint64
	Phoff     uint64
	Shoff     uint64
	Flags     uint32
	Ehsize    uint16
	Phentsize uint16
	Phnum     uint16
	Shentsize uint16
	Shnum     uint16
	Shstrndx  uint16
}

type Elf64Phdr struct {
	Type   uint32
	Flags  uint32
	Offset uint64
	Vaddr  uint64
	Paddr  uint64
	Filesz uint64
	Memsz  uint64
	Align  uint64
}

type Emitter struct {
	program *ir.Program
	buf     *bytes.Buffer
}

func New(prog *ir.Program) *Emitter {
	return &Emitter{
		program: prog,
		buf:     new(bytes.Buffer),
	}
}

func (e *Emitter) Emit() ([]byte, error) {
	// 1. Generate text segment (machine code)
	textSegment := e.generateTextSegment()
	
	// 2. Setup Headers
	var ehdr Elf64Ehdr
	copy(ehdr.Ident[:], []byte{0x7F, 'E', 'L', 'F', 2, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	ehdr.Type = 2    // ET_EXEC
	ehdr.Machine = 62 // EM_X86_64
	ehdr.Version = 1

	entryPoint := uint64(0x400000) + 64 + 56 // Base + Ehdr + Phdr
	ehdr.Entry = entryPoint
	ehdr.Phoff = 64
	ehdr.Shoff = 0
	ehdr.Ehsize = 64
	ehdr.Phentsize = 56
	ehdr.Phnum = 1
	ehdr.Shentsize = 0
	ehdr.Shnum = 0
	ehdr.Shstrndx = 0

	var phdr Elf64Phdr
	phdr.Type = PT_LOAD
	phdr.Flags = PF_R | PF_X
	phdr.Offset = 0
	phdr.Vaddr = 0x400000
	phdr.Paddr = 0x400000
	phdr.Filesz = uint64(120 + len(textSegment)) // Ehdr + Phdr + text
	phdr.Memsz = phdr.Filesz
	phdr.Align = 0x1000

	// 3. Write
	binary.Write(e.buf, binary.LittleEndian, &ehdr)
	binary.Write(e.buf, binary.LittleEndian, &phdr)
	e.buf.Write(textSegment)

	return e.buf.Bytes(), nil
}

type labelRef struct {
	placeholderOffset int
	targetLabel       string
	isCall            bool
}

type stackContext struct {
	offsets    map[string]int
	nextOffset int
}

func (sc *stackContext) getOffset(op ir.Operand) int {
	if op.Type == "register" || op.Type == "variable" {
		if offset, ok := sc.offsets[op.Value]; ok {
			return offset
		}
		sc.nextOffset += 8
		sc.offsets[op.Value] = -sc.nextOffset
		return -sc.nextOffset
	}
	return 0
}

func append32(code []byte, val int32) []byte {
	return append(code, byte(val), byte(val>>8), byte(val>>16), byte(val>>24))
}

func (e *Emitter) generateTextSegment() []byte {
	var code []byte
	labelPCs := make(map[string]int)
	refs := []labelRef{}

	var currentFunc string
	sc := &stackContext{
		offsets:    make(map[string]int),
		nextOffset: 0,
	}

	for _, inst := range e.program.Instructions {
		switch inst.Op {
		case ir.OpLabel:
			lbl := inst.Dest.Value
			labelPCs[lbl] = len(code)
			if !strings.HasPrefix(lbl, ".") {
				currentFunc = lbl
				// Reset stack context for a new function
				sc = &stackContext{
					offsets:    make(map[string]int),
					nextOffset: 0,
				}
				// Emit prologue: push rbp; mov rbp, rsp; sub rsp, 256
				code = append(code, 0x55)
				code = append(code, 0x48, 0x89, 0xe5)
				code = append(code, 0x48, 0x81, 0xec)
				code = append32(code, 256)
			}

		case ir.OpLoad:
			destOffset := sc.getOffset(inst.Dest)
			if inst.Src1.Type == "immediate" {
				val, _ := strconv.Atoi(inst.Src1.Value)
				// mov rax, imm32
				code = append(code, 0x48, 0xc7, 0xc0)
				code = append32(code, int32(val))
			} else {
				srcOffset := sc.getOffset(inst.Src1)
				// mov rax, [rbp + srcOffset]
				code = append(code, 0x48, 0x8b, 0x85)
				code = append32(code, int32(srcOffset))
			}
			// mov [rbp + destOffset], rax
			code = append(code, 0x48, 0x89, 0x85)
			code = append32(code, int32(destOffset))

		case ir.OpStore:
			srcOffset := sc.getOffset(inst.Src1)
			destOffset := sc.getOffset(inst.Dest)
			// mov rax, [rbp + srcOffset]
			code = append(code, 0x48, 0x8b, 0x85)
			code = append32(code, int32(srcOffset))
			// mov [rbp + destOffset], rax
			code = append(code, 0x48, 0x89, 0x85)
			code = append32(code, int32(destOffset))

		case ir.OpAdd, ir.OpSub, ir.OpMul, ir.OpDiv, ir.OpEq, ir.OpNeq, ir.OpLt, ir.OpGt:
			src1Offset := sc.getOffset(inst.Src1)
			src2Offset := sc.getOffset(inst.Src2)
			destOffset := sc.getOffset(inst.Dest)

			// mov rax, [rbp + src1Offset]
			code = append(code, 0x48, 0x8b, 0x85)
			code = append32(code, int32(src1Offset))
			// mov rcx, [rbp + src2Offset]
			code = append(code, 0x48, 0x8b, 0x8d)
			code = append32(code, int32(src2Offset))

			switch inst.Op {
			case ir.OpAdd:
				// add rax, rcx
				code = append(code, 0x48, 0x01, 0xc8)
			case ir.OpSub:
				// sub rax, rcx
				code = append(code, 0x48, 0x29, 0xc8)
			case ir.OpMul:
				// imul rax, rcx
				code = append(code, 0x48, 0x0f, 0xaf, 0xc1)
			case ir.OpDiv:
				// cqo; idiv rcx
				code = append(code, 0x48, 0x99)
				code = append(code, 0x48, 0xf7, 0xf9)
			case ir.OpEq, ir.OpNeq, ir.OpLt, ir.OpGt:
				// cmp rax, rcx
				code = append(code, 0x48, 0x39, 0xc8)
				var setOp byte
				switch inst.Op {
				case ir.OpEq: setOp = 0x94
				case ir.OpNeq: setOp = 0x95
				case ir.OpLt: setOp = 0x9c
				case ir.OpGt: setOp = 0x9f
				}
				// setX al; movzx rax, al
				code = append(code, 0x0f, setOp, 0xc0)
				code = append(code, 0x48, 0x0f, 0xb6, 0xc0)
			}
			// mov [rbp + destOffset], rax
			code = append(code, 0x48, 0x89, 0x85)
			code = append32(code, int32(destOffset))

		case ir.OpJump:
			// jmp offset
			code = append(code, 0xe9)
			refs = append(refs, labelRef{
				placeholderOffset: len(code),
				targetLabel:       inst.Dest.Value,
				isCall:            false,
			})
			code = append32(code, 0)

		case ir.OpJumpIfZero:
			condOffset := sc.getOffset(inst.Src1)
			// mov rax, [rbp + condOffset]
			code = append(code, 0x48, 0x8b, 0x85)
			code = append32(code, int32(condOffset))
			// test rax, rax
			code = append(code, 0x48, 0x85, 0xc0)
			// jz offset
			code = append(code, 0x0f, 0x84)
			refs = append(refs, labelRef{
				placeholderOffset: len(code),
				targetLabel:       inst.Dest.Value,
				isCall:            false,
			})
			code = append32(code, 0)

		case ir.OpCall:
			fnName := inst.Src1.Value
			if fnName == "print" || fnName == "println" {
				argOffset := sc.getOffset(inst.Src2)
				// mov rax, [rbp + argOffset]
				code = append(code, 0x48, 0x8b, 0x85)
				code = append32(code, int32(argOffset))
				// call print_int helper
				code = append(code, 0xe8)
				refs = append(refs, labelRef{
					placeholderOffset: len(code),
					targetLabel:       "__print_int",
					isCall:            true,
				})
				code = append32(code, 0)
			} else {
				// call fn
				code = append(code, 0xe8)
				refs = append(refs, labelRef{
					placeholderOffset: len(code),
					targetLabel:       fnName,
					isCall:            true,
				})
				code = append32(code, 0)
				// store return value
				destOffset := sc.getOffset(inst.Dest)
				// mov [rbp + destOffset], rax
				code = append(code, 0x48, 0x89, 0x85)
				code = append32(code, int32(destOffset))
			}

		case ir.OpRet:
			retOffset := sc.getOffset(inst.Src1)
			// mov rax, [rbp + retOffset]
			code = append(code, 0x48, 0x8b, 0x85)
			code = append32(code, int32(retOffset))
			if currentFunc == "main" {
				// Exit syscall: mov rdi, rax; mov rax, 60; syscall
				code = append(code, 0x48, 0x89, 0xc7)
				code = append(code, 0x48, 0xc7, 0xc0, 0x3c, 0x00, 0x00, 0x00)
				code = append(code, 0x0f, 0x05)
			} else {
				// mov rsp, rbp; pop rbp; ret
				code = append(code, 0x48, 0x89, 0xec)
				code = append(code, 0x5d)
				code = append(code, 0xc3)
			}
		}
	}

	// Append __print_int helper assembly routine to print numbers
	labelPCs["__print_int"] = len(code)
	printHelper := []byte{
		0x57,                   // push rdi
		0x56,                   // push rsi
		0x52,                   // push rdx
		0x51,                   // push rcx
		0x41, 0x50,             // push r8
		0x48, 0x85, 0xc0,       // test rax, rax
		0x79, 0x22,             // jns pos
		0x48, 0xf7, 0xd8,       // neg rax
		0x50,                   // push rax (absolute value)
		0x48, 0xc7, 0xc0, 0x2d, 0x00, 0x00, 0x00, // mov rax, '-'
		0x50,                   // push rax
		0xbf, 0x01, 0x00, 0x00, 0x00, // mov edi, 1
		0x48, 0x89, 0xe6,       // mov rsi, rsp
		0xba, 0x01, 0x00, 0x00, 0x00, // mov edx, 1
		0xb8, 0x01, 0x00, 0x00, 0x00, // mov eax, 1 (sys_write)
		0x0f, 0x05,             // syscall
		0x58,                   // pop rax (clears '-')
		0x58,                   // pop rax (clears absolute value)
		// pos:
		0x48, 0xc7, 0xc1, 0x0a, 0x00, 0x00, 0x00, // mov rcx, 10
		0x41, 0xb8, 0x00, 0x00, 0x00, 0x00,       // mov r8d, 0
		// loop:
		0x48, 0x99,             // cqo
		0x48, 0xf7, 0xf9,       // idiv rcx
		0x48, 0x83, 0xc2, 0x30, // add rdx, 48
		0x52,                   // push rdx
		0x41, 0xff, 0xc0,       // inc r8d
		0x48, 0x85, 0xc0,       // test rax, rax
		0x75, 0xf0,             // jnz loop
		// print_loop:
		0x45, 0x85, 0xc0,       // test r8d, r8d
		0x74, 0x1a,             // jz done
		0x41, 0xff, 0xc8,       // dec r8d
		0xbf, 0x01, 0x00, 0x00, 0x00, // mov edi, 1 (stdout)
		0x48, 0x89, 0xe6,       // mov rsi, rsp
		0xba, 0x01, 0x00, 0x00, 0x00, // mov edx, 1 (len)
		0xb8, 0x01, 0x00, 0x00, 0x00, // mov eax, 1 (sys_write)
		0x0f, 0x05,             // syscall
		0x5a,                   // pop rdx
		0xeb, 0xe1,             // jmp print_loop
		// done:
		0x48, 0xc7, 0xc0, 0x0a, 0x00, 0x00, 0x00, // mov rax, 10 ('\n')
		0x50,                   // push rax
		0xbf, 0x01, 0x00, 0x00, 0x00, // mov edi, 1
		0x48, 0x89, 0xe6,       // mov rsi, rsp
		0xba, 0x01, 0x00, 0x00, 0x00, // mov edx, 1
		0xb8, 0x01, 0x00, 0x00, 0x00, // mov eax, 1
		0x0f, 0x05,             // syscall
		0x58,                   // pop rax
		0x41, 0x58,             // pop r8
		0x59,                   // pop rcx
		0x5a,                   // pop rdx
		0x5e,                   // pop rsi
		0x5f,                   // pop rdi
		0xc3,                   // ret
	}
	code = append(code, printHelper...)

	// Pass 2: label offsets resolution
	for _, ref := range refs {
		targetPC, ok := labelPCs[ref.targetLabel]
		if !ok {
			panic(fmt.Sprintf("unresolved label reference: %s", ref.targetLabel))
		}
		// relative offset = targetPC - (placeholderOffset + 4)
		rel := int32(targetPC - (ref.placeholderOffset + 4))
		binary.LittleEndian.PutUint32(code[ref.placeholderOffset:ref.placeholderOffset+4], uint32(rel))
	}

	return code
}
