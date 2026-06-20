package macho

import (
	"bytes"
	"encoding/binary"
	"github.com/nathfavour/ship/compiler/ir"
)

// Mach-O Constants for x86_64
const (
	MH_MAGIC_64  = 0xfeedfacf
	CPU_TYPE_X86_64 = 0x01000007
	CPU_SUBTYPE_LIB64 = 0x80000003
	MH_EXECUTE   = 0x2
)

type MachHeader64 struct {
	Magic      uint32
	CpuType    uint32
	CpuSubtype uint32
	FileType   uint32
	NCmds      uint32
	SizeOfCmds uint32
	Flags      uint32
	Reserved   uint32
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
	var hdr MachHeader64
	hdr.Magic = MH_MAGIC_64
	hdr.CpuType = CPU_TYPE_X86_64
	hdr.CpuSubtype = CPU_SUBTYPE_LIB64
	hdr.FileType = MH_EXECUTE
	hdr.NCmds = 0
	hdr.SizeOfCmds = 0
	hdr.Flags = 0x1 // MH_NOUNDEFS

	binary.Write(e.buf, binary.LittleEndian, &hdr)
	
	// Appending a simple notice byte or dummy body to complete Mach-O skeleton layout
	e.buf.Write([]byte("Mach-O skeleton output"))

	return e.buf.Bytes(), nil
}
