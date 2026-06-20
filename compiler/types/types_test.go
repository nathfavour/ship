package types

import (
	"testing"

	"github.com/nathfavour/ship/compiler/lexer"
	"github.com/nathfavour/ship/compiler/parser"
)

func TestTypeCheck(t *testing.T) {
	input := `
type Person struct {
	name string
	age int
}

fn add(x int, y int) -> int {
	return x + y;
}

let result = add(5, 10);
let a = 1;
let b = 2;
let c = a + b;
`

	l := lexer.New("test.ship", input)
	p := parser.New(l)
	file := p.ParseFile()
	
	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	checker := NewChecker()
	checker.CheckFile(file)

	if len(checker.Errors()) != 0 {
		t.Fatalf("checker errors: %v", checker.Errors())
	}

	// Verify the struct was registered
	st, ok := checker.Structs["Person"]
	if !ok {
		t.Fatalf("struct Person not found")
	}

	if st.Name() != "Person" {
		t.Errorf("expected struct name Person, got %s", st.Name())
	}

	if len(st.Fields()) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(st.Fields()))
	}
	
	if st.Fields()[0].Name != "name" {
		t.Errorf("expected field name 'name', got %s", st.Fields()[0].Name)
	}
}

func TestTypeCheckErrors(t *testing.T) {
	input := `
let x = 5;
let y = "hello";
let z = x + y; // TYPE_MISMATCH
`

	l := lexer.New("test.ship", input)
	p := parser.New(l)
	file := p.ParseFile()
	
	checker := NewChecker()
	checker.CheckFile(file)

	if len(checker.Errors()) != 1 {
		t.Fatalf("expected 1 error, got %d", len(checker.Errors()))
	}

	if checker.Errors()[0].ErrorCode != "TYPE_MISMATCH" {
		t.Errorf("expected TYPE_MISMATCH, got %s", checker.Errors()[0].ErrorCode)
	}
}
