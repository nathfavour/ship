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
	program  *ir.Program
	buf      *bytes.Buffer
	labelPCs map[string]int
}

func New(prog *ir.Program) *Emitter {
	return &Emitter{
		program:  prog,
		buf:      new(bytes.Buffer),
		labelPCs: make(map[string]int),
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

	mainOffset := e.labelPCs["main"]
	entryPoint := uint64(0x400000) + 64 + 56 + uint64(mainOffset)
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
	phdr.Flags = PF_R | PF_W | PF_X
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
	labelPCs := e.labelPCs
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
			} else if inst.Src1.Type == "string" {
				strVal := inst.Src1.Value
				// jmp short over string bytes: 0xeb, byte(len + 1)
				code = append(code, 0xeb, byte(len(strVal)+1))
				// embed string + null byte
				code = append(code, []byte(strVal)...)
				code = append(code, 0x00)
				// lea rax, [rip + offset] -> 0x48, 0x8d, 0x05, 4-byte offset
				code = append(code, 0x48, 0x8d, 0x05)
				// offset from instruction following lea to string start:
				// rip points to (lea_start + 7)
				// string starts at (lea_start - (len + 1))
				// offset = - (len + 1 + 7)
				offset := -int32(len(strVal) + 1 + 7)
				code = append32(code, offset)
			} else if inst.Src1.Type == "deref" {
				parts := strings.Split(inst.Src1.Value, ",")
				varName := parts[0]
				fieldOffset, _ := strconv.Atoi(parts[1])

				ptrOffset := sc.getOffset(ir.Operand{Type: "variable", Value: varName})
				// mov rbx, [rbp + ptrOffset]
				code = append(code, 0x48, 0x8b, 0x9d)
				code = append32(code, int32(ptrOffset))

				// mov rax, [rbx + fieldOffset]
				code = append(code, 0x48, 0x8b, 0x83)
				code = append32(code, int32(fieldOffset))
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
			if inst.Dest.Type == "deref" {
				parts := strings.Split(inst.Dest.Value, ",")
				varName := parts[0]
				fieldOffset, _ := strconv.Atoi(parts[1])

				srcOffset := sc.getOffset(inst.Src1)
				// mov rax, [rbp + srcOffset]
				code = append(code, 0x48, 0x8b, 0x85)
				code = append32(code, int32(srcOffset))

				ptrOffset := sc.getOffset(ir.Operand{Type: "variable", Value: varName})
				// mov rbx, [rbp + ptrOffset]
				code = append(code, 0x48, 0x8b, 0x9d)
				code = append32(code, int32(ptrOffset))

				// mov [rbx + fieldOffset], rax
				code = append(code, 0x48, 0x89, 0x83)
				code = append32(code, int32(fieldOffset))
			} else {
				srcOffset := sc.getOffset(inst.Src1)
				destOffset := sc.getOffset(inst.Dest)
				// mov rax, [rbp + srcOffset]
				code = append(code, 0x48, 0x8b, 0x85)
				code = append32(code, int32(srcOffset))
				// mov [rbp + destOffset], rax
				code = append(code, 0x48, 0x89, 0x85)
				code = append32(code, int32(destOffset))
			}

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
			} else if fnName == "print_str" || fnName == "println_str" {
				argOffset := sc.getOffset(inst.Src2)
				// mov rax, [rbp + argOffset]
				code = append(code, 0x48, 0x8b, 0x85)
				code = append32(code, int32(argOffset))
				// call print_str helper
				code = append(code, 0xe8)
				refs = append(refs, labelRef{
					placeholderOffset: len(code),
					targetLabel:       "__print_str",
					isCall:            true,
				})
				code = append32(code, 0)
			} else if fnName == "write_file" {
				arg1Offset := sc.getOffset(inst.Src2)
				code = append(code, 0x48, 0x8b, 0xbd)
				code = append32(code, int32(arg1Offset))

				arg2Offset := sc.getOffset(ir.Operand{Type: "variable", Value: "__write_file_arg2"})
				code = append(code, 0x48, 0x8b, 0xb5)
				code = append32(code, int32(arg2Offset))

				code = append(code, 0xe8)
				refs = append(refs, labelRef{
					placeholderOffset: len(code),
					targetLabel:       "write_file",
					isCall:            true,
				})
				code = append32(code, 0)

				destOffset := sc.getOffset(inst.Dest)
				code = append(code, 0x48, 0x89, 0x85)
				code = append32(code, int32(destOffset))
			} else if fnName == "read_file" {
				argOffset := sc.getOffset(inst.Src2)
				code = append(code, 0x48, 0x8b, 0xbd)
				code = append32(code, int32(argOffset))

				code = append(code, 0xe8)
				refs = append(refs, labelRef{
					placeholderOffset: len(code),
					targetLabel:       "read_file",
					isCall:            true,
				})
				code = append32(code, 0)

				destOffset := sc.getOffset(inst.Dest)
				code = append(code, 0x48, 0x89, 0x85)
				code = append32(code, int32(destOffset))
			} else if fnName == "read_str" || fnName == "input" {
				code = append(code, 0xe8)
				refs = append(refs, labelRef{
					placeholderOffset: len(code),
					targetLabel:       "read_str",
					isCall:            true,
				})
				code = append32(code, 0)

				destOffset := sc.getOffset(inst.Dest)
				code = append(code, 0x48, 0x89, 0x85)
				code = append32(code, int32(destOffset))
			} else if fnName == "net_write" {
				arg1Offset := sc.getOffset(inst.Src2)
				code = append(code, 0x48, 0x8b, 0xbd)
				code = append32(code, int32(arg1Offset))

				arg2Offset := sc.getOffset(ir.Operand{Type: "variable", Value: "__net_write_arg2"})
				code = append(code, 0x48, 0x8b, 0xb5)
				code = append32(code, int32(arg2Offset))

				code = append(code, 0xe8)
				refs = append(refs, labelRef{
					placeholderOffset: len(code),
					targetLabel:       "net_write",
					isCall:            true,
				})
				code = append32(code, 0)

				destOffset := sc.getOffset(inst.Dest)
				code = append(code, 0x48, 0x89, 0x85)
				code = append32(code, int32(destOffset))
			} else if fnName == "sys_io_uring_setup" {
				arg1Offset := sc.getOffset(inst.Src2)
				code = append(code, 0x48, 0x8b, 0xbd) // mov rdi, [rbp + arg1Offset]
				code = append32(code, int32(arg1Offset))

				arg2Offset := sc.getOffset(ir.Operand{Type: "variable", Value: "__uring_setup_arg2"})
				code = append(code, 0x48, 0x8b, 0xb5) // mov rsi, [rbp + arg2Offset]
				code = append32(code, int32(arg2Offset))

				code = append(code, 0xe8)
				refs = append(refs, labelRef{
					placeholderOffset: len(code),
					targetLabel:       "sys_io_uring_setup",
					isCall:            true,
				})
				code = append32(code, 0)

				destOffset := sc.getOffset(inst.Dest)
				code = append(code, 0x48, 0x89, 0x85)
				code = append32(code, int32(destOffset))
			} else if fnName == "sys_io_uring_enter" {
				arg1Offset := sc.getOffset(inst.Src2)
				code = append(code, 0x48, 0x8b, 0xbd) // mov rdi, [rbp + arg1Offset]
				code = append32(code, int32(arg1Offset))

				arg2Offset := sc.getOffset(ir.Operand{Type: "variable", Value: "__uring_enter_arg2"})
				code = append(code, 0x48, 0x8b, 0xb5) // mov rsi, [rbp + arg2Offset]
				code = append32(code, int32(arg2Offset))

				arg3Offset := sc.getOffset(ir.Operand{Type: "variable", Value: "__uring_enter_arg3"})
				code = append(code, 0x48, 0x8b, 0x95) // mov rdx, [rbp + arg3Offset]
				code = append32(code, int32(arg3Offset))

				arg4Offset := sc.getOffset(ir.Operand{Type: "variable", Value: "__uring_enter_arg4"})
				code = append(code, 0x48, 0x8b, 0x8d) // mov rcx, [rbp + arg4Offset]
				code = append32(code, int32(arg4Offset))

				code = append(code, 0xe8)
				refs = append(refs, labelRef{
					placeholderOffset: len(code),
					targetLabel:       "sys_io_uring_enter",
					isCall:            true,
				})
				code = append32(code, 0)

				destOffset := sc.getOffset(inst.Dest)
				code = append(code, 0x48, 0x89, 0x85)
				code = append32(code, int32(destOffset))
			} else if fnName == "sys_io_uring_register" {
				arg1Offset := sc.getOffset(inst.Src2)
				code = append(code, 0x48, 0x8b, 0xbd) // mov rdi, [rbp + arg1Offset]
				code = append32(code, int32(arg1Offset))

				arg2Offset := sc.getOffset(ir.Operand{Type: "variable", Value: "__uring_reg_arg2"})
				code = append(code, 0x48, 0x8b, 0xb5) // mov rsi, [rbp + arg2Offset]
				code = append32(code, int32(arg2Offset))

				arg3Offset := sc.getOffset(ir.Operand{Type: "variable", Value: "__uring_reg_arg3"})
				code = append(code, 0x48, 0x8b, 0x95) // mov rdx, [rbp + arg3Offset]
				code = append32(code, int32(arg3Offset))

				arg4Offset := sc.getOffset(ir.Operand{Type: "variable", Value: "__uring_reg_arg4"})
				code = append(code, 0x48, 0x8b, 0x8d) // mov rcx, [rbp + arg4Offset]
				code = append32(code, int32(arg4Offset))

				code = append(code, 0xe8)
				refs = append(refs, labelRef{
					placeholderOffset: len(code),
					targetLabel:       "sys_io_uring_register",
					isCall:            true,
				})
				code = append32(code, 0)

				destOffset := sc.getOffset(inst.Dest)
				code = append(code, 0x48, 0x89, 0x85)
				code = append32(code, int32(destOffset))
			} else if fnName == "sys_mmap" {
				arg1Offset := sc.getOffset(inst.Src2)
				code = append(code, 0x48, 0x8b, 0xbd) // mov rdi, [rbp + arg1Offset]
				code = append32(code, int32(arg1Offset))

				arg2Offset := sc.getOffset(ir.Operand{Type: "variable", Value: "__mmap_arg2"})
				code = append(code, 0x48, 0x8b, 0xb5) // mov rsi, [rbp + arg2Offset]
				code = append32(code, int32(arg2Offset))

				arg3Offset := sc.getOffset(ir.Operand{Type: "variable", Value: "__mmap_arg3"})
				code = append(code, 0x48, 0x8b, 0x95) // mov rdx, [rbp + arg3Offset]
				code = append32(code, int32(arg3Offset))

				arg4Offset := sc.getOffset(ir.Operand{Type: "variable", Value: "__mmap_arg4"})
				code = append(code, 0x48, 0x8b, 0x8d) // mov rcx, [rbp + arg4Offset]
				code = append32(code, int32(arg4Offset))

				arg5Offset := sc.getOffset(ir.Operand{Type: "variable", Value: "__mmap_arg5"})
				code = append(code, 0x4c, 0x8b, 0x85) // mov r8, [rbp + arg5Offset]
				code = append32(code, int32(arg5Offset))

				arg6Offset := sc.getOffset(ir.Operand{Type: "variable", Value: "__mmap_arg6"})
				code = append(code, 0x4c, 0x8b, 0x8d) // mov r9, [rbp + arg6Offset]
				code = append32(code, int32(arg6Offset))

				code = append(code, 0xe8)
				refs = append(refs, labelRef{
					placeholderOffset: len(code),
					targetLabel:       "sys_mmap",
					isCall:            true,
				})
				code = append32(code, 0)

				destOffset := sc.getOffset(inst.Dest)
				code = append(code, 0x48, 0x89, 0x85)
				code = append32(code, int32(destOffset))
			} else {
				// Load argument into RAX if present
				if inst.Src2.Type != "" {
					argOffset := sc.getOffset(inst.Src2)
					code = append(code, 0x48, 0x8b, 0x85)
					code = append32(code, int32(argOffset))
				}
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
		0x79, 0x23,             // jns pos
		0x48, 0xf7, 0xd8,       // neg rax
		0x90,                   // nop (align negative block)
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
		0x48, 0x31, 0xd2,       // xor rdx, rdx
		0x48, 0xf7, 0xf1,       // div rcx
		0x48, 0x83, 0xc2, 0x30, // add rdx, 48
		0x52,                   // push rdx
		0x41, 0xff, 0xc0,       // inc r8d
		0x48, 0x85, 0xc0,       // test rax, rax
		0x75, 0xee,             // jnz loop
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

	// Append __print_str helper assembly routine to print null-terminated strings
	labelPCs["__print_str"] = len(code)
	printStrHelper := []byte{
		0x57,                   // push rdi
		0x56,                   // push rsi
		0x52,                   // push rdx
		0x51,                   // push rcx
		0x48, 0x89, 0xc6,       // mov rsi, rax (string address)
		0x48, 0xc7, 0xc2, 0x00, 0x00, 0x00, 0x00, // mov rdx, 0 (len)
		// len_loop:
		0x80, 0x3c, 0x16, 0x00, // cmp byte ptr [rsi + rdx], 0
		0x74, 0x05,             // je len_done
		0x48, 0xff, 0xc2,       // inc rdx
		0xeb, 0xf5,             // jmp len_loop
		// len_done:
		0x48, 0x85, 0xd2,       // test rdx, rdx
		0x74, 0x09,             // jz print_done
		0xbf, 0x01, 0x00, 0x00, 0x00, // mov edi, 1 (stdout)
		0xb8, 0x01, 0x00, 0x00, 0x00, // mov eax, 1 (sys_write)
		0x0f, 0x05,             // syscall
		// print_done:
		0x48, 0xc7, 0xc0, 0x0a, 0x00, 0x00, 0x00, // mov rax, 10 ('\n')
		0x50,                   // push rax
		0xbf, 0x01, 0x00, 0x00, 0x00, // mov edi, 1
		0x48, 0x89, 0xe6,       // mov rsi, rsp
		0xba, 0x01, 0x00, 0x00, 0x00, // mov edx, 1
		0xb8, 0x01, 0x00, 0x00, 0x00, // mov eax, 1
		0x0f, 0x05,             // syscall
		0x58,                   // pop rax
		0x59,                   // pop rcx
		0x5a,                   // pop rdx
		0x5e,                   // pop rsi
		0x5f,                   // pop rdi
		0xc3,                   // ret
	}
	code = append(code, printStrHelper...)

	// Append write_file helper
	labelPCs["write_file"] = len(code)
	writeHelper := []byte{
		0x56,                                     // push rsi
		0x48, 0xc7, 0xc6, 0xa4, 0x01, 0x00, 0x00, // mov rsi, 0x1a4
		0x48, 0xc7, 0xc0, 0x55, 0x00, 0x00, 0x00, // mov rax, 85
		0x0f, 0x05,                               // syscall
		0x5e,                                     // pop rsi
		0x48, 0x85, 0xc0,                         // test rax, rax
		0x78, 0x2b,                               // js write_error (+43)
		0x50,                                     // push rax
		0x48, 0x31, 0xc9,                         // xor rcx, rcx
		// len_loop:
		0x80, 0x3c, 0x0e, 0x00,                   // cmp byte ptr [rsi + rcx], 0
		0x74, 0x05,                               // je len_done
		0x48, 0xff, 0xc1,                         // inc rcx
		0xeb, 0xf5,                               // jmp len_loop
		// len_done:
		0x48, 0x89, 0xca,                         // mov rdx, rcx
		0x5f,                                     // pop rdi
		0x57,                                     // push rdi
		0x48, 0xc7, 0xc0, 0x01, 0x00, 0x00, 0x00, // mov rax, 1
		0x0f, 0x05,                               // syscall
		0x5f,                                     // pop rdi
		0x48, 0xc7, 0xc0, 0x03, 0x00, 0x00, 0x00, // mov rax, 3
		0x0f, 0x05,                               // syscall
		0x48, 0x31, 0xc0,                         // xor rax, rax
		0xc3,                                     // ret
		// write_error:
		0x48, 0xc7, 0xc0, 0xff, 0xff, 0xff, 0xff, // mov rax, -1
		0xc3,                                     // ret
	}
	code = append(code, writeHelper...)

	// Append read_file helper
	labelPCs["read_file"] = len(code)
	readFileStart := len(code)
	readFileHelper := []byte{
		0x48, 0xc7, 0xc6, 0x00, 0x00, 0x00, 0x00, // mov rsi, 0
		0x48, 0xc7, 0xc2, 0x00, 0x00, 0x00, 0x00, // mov rdx, 0
		0x48, 0xc7, 0xc0, 0x02, 0x00, 0x00, 0x00, // mov rax, 2
		0x0f, 0x05,                               // syscall
		0x48, 0x85, 0xc0,                         // test rax, rax
		0x78, 0x3f,                               // js read_error (+63)
		0x50,                                     // push rax (save fd)
		0x5f,                                     // pop rdi (fd)
		0x57,                                     // push rdi (save fd again)
		// lea rsi, [rip + __read_file_buf]
		0x48, 0x8d, 0x35, 0x00, 0x00, 0x00, 0x00, // lea rsi, [rip + __read_file_buf] (placeholder at 35)
		0x48, 0xc7, 0xc2, 0x00, 0x00, 0x01, 0x00, // mov rdx, 65536
		0x48, 0xc7, 0xc0, 0x00, 0x00, 0x00, 0x00, // mov rax, 0
		0x0f, 0x05,                               // syscall
		0x48, 0x85, 0xc0,                         // test rax, rax
		0x79, 0x03,                               // jns read_ok (+3)
		0x48, 0x31, 0xc0,                         // xor rax, rax
		// read_ok:
		0x48, 0x8d, 0x35, 0x00, 0x00, 0x00, 0x00, // lea rsi, [rip + __read_file_buf] (placeholder at 66)
		0xc6, 0x04, 0x06, 0x00,                   // mov byte ptr [rsi + rax], 0
		0x5f,                                     // pop rdi
		0x48, 0xc7, 0xc0, 0x03, 0x00, 0x00, 0x00, // mov rax, 3
		0x0f, 0x05,                               // syscall
		0x48, 0x8d, 0x05, 0x00, 0x00, 0x00, 0x00, // lea rax, [rip + __read_file_buf] (placeholder at 90)
		0xc3,                                     // ret
		// read_error:
		0x48, 0x8d, 0x05, 0x00, 0x00, 0x00, 0x00, // lea rax, [rip + __empty_str] (placeholder at 98)
		0xc3,                                     // ret
	}
	code = append(code, readFileHelper...)

	// Append read_str helper
	labelPCs["read_str"] = len(code)
	readStrStart := len(code)
	readStrHelper := []byte{
		0x48, 0xc7, 0xc7, 0x00, 0x00, 0x00, 0x00, // mov rdi, 0
		0x48, 0x8d, 0x35, 0x00, 0x00, 0x00, 0x00, // lea rsi, [rip + __input_buf] (placeholder at 10)
		0x48, 0xc7, 0xc2, 0x00, 0x10, 0x00, 0x00, // mov rdx, 4096
		0x48, 0xc7, 0xc0, 0x00, 0x00, 0x00, 0x00, // mov rax, 0
		0x0f, 0x05,                               // syscall
		0x48, 0x85, 0xc0,                         // test rax, rax
		0x7e, 0x30,                               // jle read_str_empty (+48)
		0x50,                                     // push rax
		0x48, 0x8d, 0x35, 0x00, 0x00, 0x00, 0x00, // lea rsi, [rip + __input_buf] (placeholder at 39)
		0x58,                                     // pop rax
		0xc6, 0x04, 0x06, 0x00,                   // mov byte ptr [rsi + rax], 0
		0x48, 0xff, 0xc8,                         // dec rax
		0x80, 0x3c, 0x06, 0x0a,                   // cmp byte ptr [rsi + rax], 10
		0x75, 0x04,                               // jne no_nl (+4)
		0xc6, 0x04, 0x06, 0x00,                   // mov byte ptr [rsi + rax], 0
		// no_nl:
		0x48, 0x85, 0xc0,                         // test rax, rax
		0x74, 0x0d,                               // jz no_nl2 (+13)
		0x48, 0xff, 0xc8,                         // dec rax
		0x80, 0x3c, 0x06, 0x0d,                   // cmp byte ptr [rsi + rax], 13
		0x75, 0x04,                               // jne no_nl2 (+4)
		0xc6, 0x04, 0x06, 0x00,                   // mov byte ptr [rsi + rax], 0
		// no_nl2:
		0x48, 0x8d, 0x05, 0x00, 0x00, 0x00, 0x00, // lea rax, [rip + __input_buf] (placeholder at 79)
		0xc3,                                     // ret
		// read_str_empty:
		0x48, 0x8d, 0x05, 0x00, 0x00, 0x00, 0x00, // lea rax, [rip + __empty_str] (placeholder at 85)
		0xc3,                                     // ret
	}
	code = append(code, readStrHelper...)

	// Append net_listen helper
	labelPCs["net_listen"] = len(code)
	netListenHelperParts := [][]byte{
		/* 0: entry & socket call */
		{0x50}, // push rax (save port)
		{0x48, 0xc7, 0xc7, 0x02, 0x00, 0x00, 0x00}, // mov rdi, 2 (AF_INET)
		{0x48, 0xc7, 0xc6, 0x01, 0x00, 0x00, 0x00}, // mov rsi, 1 (SOCK_STREAM)
		{0x48, 0xc7, 0xc2, 0x00, 0x00, 0x00, 0x00}, // mov rdx, 0 (IPPROTO_IP)
		{0x48, 0xc7, 0xc0, 0x29, 0x00, 0x00, 0x00}, // mov rax, 41 (sys_socket)
		{0x0f, 0x05},                               // syscall
		{0x48, 0x85, 0xc0},                         // test rax, rax
		{0x78, 0x00},                               // js socket_error (placeholder offset at idx 7)
		
		/* 8: bind preparation */
		{0x50},                                     // push rax (save socket fd)
		{0x48, 0x8b, 0x44, 0x24, 0x08},             // mov rax, [rsp + 8] (get port)
		{0x66, 0xc1, 0xc0, 0x08},                   // rol ax, 8
		{0x48, 0xc1, 0xe0, 0x10},                   // shl rax, 16
		{0x48, 0x83, 0xc8, 0x02},                   // or rax, 2
		{0x48, 0x31, 0xd2},                         // xor rdx, rdx
		{0x52},                                     // push rdx (sin_zero)
		{0x50},                                     // push rax (family, port, addr)
		
		/* 16: bind call */
		{0x48, 0x8b, 0x7c, 0x24, 0x10},             // mov rdi, [rsp + 16] (socket fd)
		{0x48, 0x89, 0xe6},                         // mov rsi, rsp
		{0x48, 0xc7, 0xc2, 0x10, 0x00, 0x00, 0x00}, // mov rdx, 16
		{0x48, 0xc7, 0xc0, 0x31, 0x00, 0x00, 0x00}, // mov rax, 49 (sys_bind)
		{0x0f, 0x05},                               // syscall
		{0x48, 0x85, 0xc0},                         // test rax, rax
		{0x78, 0x00},                               // js bind_error (placeholder offset at idx 22)
		
		/* 23: listen preparation */
		{0x48, 0x83, 0xc4, 0x10},                   // add rsp, 16
		{0x48, 0x8b, 0x3c, 0x24},                   // mov rdi, [rsp]
		{0x48, 0xc7, 0xc6, 0x80, 0x00, 0x00, 0x00}, // mov rsi, 128
		{0x48, 0xc7, 0xc0, 0x32, 0x00, 0x00, 0x00}, // mov rax, 50 (sys_listen)
		{0x0f, 0x05},                               // syscall
		{0x48, 0x85, 0xc0},                         // test rax, rax
		{0x78, 0x00},                               // js listen_error (placeholder offset at idx 29)
		
		/* 30: listen success */
		{0x58},                                     // pop rax (return fd)
		{0x5e},                                     // pop rsi (clean port)
		{0xc3},                                     // ret
		
		/* 33: bind_error block */
		{0x48, 0x83, 0xc4, 0x10},                   // add rsp, 16 (bind_error target)
		
		/* 34: listen_error block */
		{0x5f},                                     // pop rdi (listen_error target)
		{0x5e},                                     // pop rsi
		{0x57},                                     // push rdi
		{0x48, 0xc7, 0xc0, 0x03, 0x00, 0x00, 0x00}, // mov rax, 3
		{0x0f, 0x05},                               // syscall
		{0x58},                                     // pop rax
		{0x48, 0xc7, 0xc0, 0xff, 0xff, 0xff, 0xff}, // mov rax, -1
		{0xc3},                                     // ret
		
		/* 42: socket_error block */
		{0x5e},                                     // pop rsi (socket_error target)
		{0x48, 0xc7, 0xc0, 0xff, 0xff, 0xff, 0xff}, // mov rax, -1
		{0xc3},                                     // ret
	}
	listenOffsets := make([]int, len(netListenHelperParts))
	currentListenOffset := 0
	for i, part := range netListenHelperParts {
		listenOffsets[i] = currentListenOffset
		currentListenOffset += len(part)
	}
	netListenHelperParts[7][1] = byte(listenOffsets[42] - listenOffsets[8])
	netListenHelperParts[22][1] = byte(listenOffsets[33] - listenOffsets[23])
	netListenHelperParts[29][1] = byte(listenOffsets[34] - listenOffsets[30])
	for _, part := range netListenHelperParts {
		code = append(code, part...)
	}

	// Append net_accept helper
	labelPCs["net_accept"] = len(code)
	netAcceptHelper := []byte{
		0x48, 0x89, 0xc7,                         // mov rdi, rax
		0x48, 0x31, 0xf6,                         // xor rsi, rsi
		0x48, 0x31, 0xd2,                         // xor rdx, rdx
		0x48, 0xc7, 0xc0, 0x2b, 0x00, 0x00, 0x00, // mov rax, 43 (sys_accept)
		0x0f, 0x05,                               // syscall
		0xc3,                                     // ret
	}
	code = append(code, netAcceptHelper...)

	// Append net_read helper
	labelPCs["net_read"] = len(code)
	netReadStart := len(code)
	netReadHelperParts := [][]byte{
		/* 0: read syscall */
		{0x48, 0x89, 0xc7},                         // mov rdi, rax
		{0x48, 0x8d, 0x35, 0x00, 0x00, 0x00, 0x00}, // lea rsi, [rip + __net_buf] (placeholder)
		{0x48, 0xc7, 0xc2, 0x00, 0x10, 0x00, 0x00}, // mov rdx, 4096
		{0x48, 0xc7, 0xc0, 0x00, 0x00, 0x00, 0x00}, // mov rax, 0
		{0x0f, 0x05},                               // syscall
		{0x48, 0x85, 0xc0},                         // test rax, rax
		{0x7e, 0x00},                               // jle net_read_empty (placeholder offset at idx 6)
		
		/* 7: success path */
		{0xc6, 0x04, 0x06, 0x00},                   // mov byte ptr [rsi + rax], 0
		{0x48, 0x89, 0xf0},                         // mov rax, rsi
		{0xc3},                                     // ret
		
		/* 10: empty path */
		{0x48, 0x8d, 0x05, 0x00, 0x00, 0x00, 0x00}, // lea rax, [rip + __empty_str] (placeholder)
		{0xc3},                                     // ret
	}
	readOffsets := make([]int, len(netReadHelperParts))
	currentReadOffset := 0
	for i, part := range netReadHelperParts {
		readOffsets[i] = currentReadOffset
		currentReadOffset += len(part)
	}
	netReadHelperParts[6][1] = byte(readOffsets[10] - readOffsets[7])
	for _, part := range netReadHelperParts {
		code = append(code, part...)
	}

	// Append net_write helper
	labelPCs["net_write"] = len(code)
	netWriteHelper := []byte{
		0x57,                                     // push rdi
		0x56,                                     // push rsi
		0x48, 0x31, 0xd2,                         // xor rdx, rdx
		// net_write_len_loop:
		0x80, 0x3c, 0x16, 0x00,                   // cmp byte ptr [rsi + rdx], 0
		0x74, 0x05,                               // je net_write_len_done
		0x48, 0xff, 0xc2,                         // inc rdx
		0xeb, 0xf5,                               // jmp net_write_len_loop
		// net_write_len_done:
		0x5e,                                     // pop rsi
		0x5f,                                     // pop rdi
		0x48, 0xc7, 0xc0, 0x01, 0x00, 0x00, 0x00, // mov rax, 1 (sys_write)
		0x0f, 0x05,                               // syscall
		0xc3,                                     // ret
	}
	code = append(code, netWriteHelper...)

	// Append net_close helper
	labelPCs["net_close"] = len(code)
	netCloseHelper := []byte{
		0x48, 0x89, 0xc7,                         // mov rdi, rax
		0x48, 0xc7, 0xc0, 0x03, 0x00, 0x00, 0x00, // mov rax, 3 (sys_close)
		0x0f, 0x05,                               // syscall
		0xc3,                                     // ret
	}
	code = append(code, netCloseHelper...)

	// Append sys_io_uring_setup helper
	labelPCs["sys_io_uring_setup"] = len(code)
	sysIouringSetupHelper := []byte{
		0x48, 0xc7, 0xc0, 0xa9, 0x01, 0x00, 0x00, // mov rax, 425 (sys_io_uring_setup)
		0x0f, 0x05,                               // syscall
		0xc3,                                     // ret
	}
	code = append(code, sysIouringSetupHelper...)

	// Append sys_io_uring_enter helper
	labelPCs["sys_io_uring_enter"] = len(code)
	sysIouringEnterHelper := []byte{
		0x49, 0x89, 0xca,                         // mov r10, rcx
		0x48, 0xc7, 0xc0, 0xaa, 0x01, 0x00, 0x00, // mov rax, 426 (sys_io_uring_enter)
		0x0f, 0x05,                               // syscall
		0xc3,                                     // ret
	}
	code = append(code, sysIouringEnterHelper...)

	// Append sys_io_uring_register helper
	labelPCs["sys_io_uring_register"] = len(code)
	sysIouringRegisterHelper := []byte{
		0x49, 0x89, 0xca,                         // mov r10, rcx
		0x48, 0xc7, 0xc0, 0xab, 0x01, 0x00, 0x00, // mov rax, 427 (sys_io_uring_register)
		0x0f, 0x05,                               // syscall
		0xc3,                                     // ret
	}
	code = append(code, sysIouringRegisterHelper...)

	// Append sys_mmap helper
	labelPCs["sys_mmap"] = len(code)
	sysMmapHelper := []byte{
		0x49, 0x89, 0xca,                         // mov r10, rcx
		0x48, 0xc7, 0xc0, 0x09, 0x00, 0x00, 0x00, // mov rax, 9 (sys_mmap)
		0x0f, 0x05,                               // syscall
		0xc3,                                     // ret
	}
	code = append(code, sysMmapHelper...)

	// Append static data buffers
	emptyStrPC := len(code)
	code = append(code, 0x00)

	inputBufPC := len(code)
	inputBufBytes := make([]byte, 4096)
	code = append(code, inputBufBytes...)

	readFileBufPC := len(code)
	readFileBufBytes := make([]byte, 65536)
	code = append(code, readFileBufBytes...)

	netBufPC := len(code)
	netBufBytes := make([]byte, 4096)
	code = append(code, netBufBytes...)

	// Relocate readFileHelper placeholders
	binary.LittleEndian.PutUint32(code[readFileStart+34:readFileStart+38], uint32(readFileBufPC-(readFileStart+38)))
	binary.LittleEndian.PutUint32(code[readFileStart+65:readFileStart+69], uint32(readFileBufPC-(readFileStart+69)))
	binary.LittleEndian.PutUint32(code[readFileStart+86:readFileStart+90], uint32(readFileBufPC-(readFileStart+90)))
	binary.LittleEndian.PutUint32(code[readFileStart+94:readFileStart+98], uint32(emptyStrPC-(readFileStart+98)))

	// Relocate readStrHelper placeholders
	binary.LittleEndian.PutUint32(code[readStrStart+10:readStrStart+14], uint32(inputBufPC-(readStrStart+14)))
	binary.LittleEndian.PutUint32(code[readStrStart+39:readStrStart+43], uint32(inputBufPC-(readStrStart+43)))
	binary.LittleEndian.PutUint32(code[readStrStart+82:readStrStart+86], uint32(inputBufPC-(readStrStart+86)))
	binary.LittleEndian.PutUint32(code[readStrStart+90:readStrStart+94], uint32(emptyStrPC-(readStrStart+94)))

	// Relocate net_read placeholders
	// lea rsi, [rip + __net_buf] at netReadStart + readOffsets[1]
	binary.LittleEndian.PutUint32(code[netReadStart+readOffsets[1]+3:netReadStart+readOffsets[1]+7], uint32(netBufPC-(netReadStart+readOffsets[2])))
	// lea rax, [rip + __empty_str] at netReadStart + readOffsets[10]
	binary.LittleEndian.PutUint32(code[netReadStart+readOffsets[10]+3:netReadStart+readOffsets[10]+7], uint32(emptyStrPC-(netReadStart+readOffsets[11])))

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
