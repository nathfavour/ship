package wasm

import (
	"bytes"
	"encoding/binary"
	"github.com/nathfavour/ship/compiler/ir"
)

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
	// WebAssembly Magic & Version
	magic := []byte{0x00, 0x61, 0x73, 0x6d}
	version := uint32(1)

	e.buf.Write(magic)
	binary.Write(e.buf, binary.LittleEndian, version)

	// Placeholder indicator for WebAssembly format
	e.buf.Write([]byte("WASM skeleton output"))

	return e.buf.Bytes(), nil
}
