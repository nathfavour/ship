package elf

import (
	"bytes"
	"testing"

	"github.com/nathfavour/ship/compiler/ir"
)

func TestElfEmission(t *testing.T) {
	prog := &ir.Program{}
	emitter := New(prog)
	
	out, err := emitter.Emit()
	if err != nil {
		t.Fatalf("Emit failed: %v", err)
	}
	
	if len(out) < 120 {
		t.Fatalf("ELF file too small: %d bytes", len(out))
	}
	
	if !bytes.HasPrefix(out, []byte{0x7F, 'E', 'L', 'F'}) {
		t.Fatalf("Missing ELF magic bytes")
	}
}
