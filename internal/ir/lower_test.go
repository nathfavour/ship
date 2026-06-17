package ir

import (
	"strings"
	"testing"

	"github.com/nathfavour/ship/internal/lexer"
	"github.com/nathfavour/ship/internal/parser"
)

func TestIRLowering(t *testing.T) {
	input := `
fn add(x int, y int) -> int {
	return x + y;
}
let a = 5;
let b = 10;
let c = a + b;
`

	l := lexer.New("test.ship", input)
	p := parser.New(l)
	file := p.ParseFile()

	lowerer := NewLowerer()
	program := lowerer.LowerFile(file)

	out := program.String()
	
	if !strings.Contains(out, "add:") {
		t.Errorf("expected 'add:' label in IR, got:\n%s", out)
	}
	
	if !strings.Contains(out, "ADD") {
		t.Errorf("expected 'ADD' opcode in IR, got:\n%s", out)
	}

	if !strings.Contains(out, "STORE") {
		t.Errorf("expected 'STORE' opcode in IR, got:\n%s", out)
	}
}
