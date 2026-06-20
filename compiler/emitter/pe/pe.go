package pe

import (
	"bytes"
	"encoding/binary"
	"github.com/nathfavour/ship/compiler/ir"
)

// PE format structures for Windows x86_64
type DosHeader struct {
	Magic    uint16   // "MZ"
	Reserved [29]uint16
	PeOffset uint32
}

type PeHeader struct {
	Magic        uint32 // "PE\0\0"
	Machine      uint16 // AMD64
	NumberOfSecs uint16
	TimeDate     uint32
	SymTable     uint32
	NumberOfSyms uint32
	OptHeaderSz  uint16
	Characts     uint16
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
	var dos DosHeader
	dos.Magic = 0x5a4d // "MZ"
	dos.PeOffset = 64  // Directly after DOS header

	var pe PeHeader
	pe.Magic = 0x00004550 // "PE\0\0"
	pe.Machine = 0x8664   // IMAGE_FILE_MACHINE_AMD64

	binary.Write(e.buf, binary.LittleEndian, &dos)
	binary.Write(e.buf, binary.LittleEndian, &pe)

	// Placeholder indicator for PE format
	e.buf.Write([]byte("PE skeleton output"))

	return e.buf.Bytes(), nil
}
