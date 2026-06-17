package elf

import (
	"bytes"
	"encoding/binary"

	"github.com/nathfavour/ship/internal/ir"
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
	// A highly simplified skeleton for a valid 64-bit ELF executable.
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

func (e *Emitter) generateTextSegment() []byte {
	var code []byte
	
	// Very naive naive compilation: for every Instruction, we'd emit real bytes.
	// For this test phase, let's just emit a simple exit syscall (exit(0)).
	
	// Map ir instructions to x86_64 code
	// Since we are mocking the skeleton, we will just return a generic shellcode if no instructions.
	
	// mov rax, 60 (sys_exit)
	// mov rdi, 0  (status code 0)
	// syscall
	
	// 48 c7 c0 3c 00 00 00    mov rax, 60
	// 48 c7 c7 00 00 00 00    mov rdi, 0
	// 0f 05                   syscall
	
	code = []byte{
		0x48, 0xc7, 0xc0, 0x3c, 0x00, 0x00, 0x00,
		0x48, 0xc7, 0xc7, 0x00, 0x00, 0x00, 0x00,
		0x0f, 0x05,
	}

	return code
}
