package parser

import (
	"testing"

	"github.com/nathfavour/ship/internal/ast"
	"github.com/nathfavour/ship/internal/lexer"
)

func TestLetStatements(t *testing.T) {
	input := `
let x = 5;
let y = 10;
let foobar = 838383;
`

	l := lexer.New("test.ship", input)
	p := New(l)

	file := p.ParseFile()
	checkParserErrors(t, p)

	if file == nil {
		t.Fatalf("ParseFile() returned nil")
	}
	if len(file.Statements) != 3 {
		t.Fatalf("file.Statements does not contain 3 statements. got=%d", len(file.Statements))
	}

	tests := []struct {
		expectedIdentifier string
	}{
		{"x"},
		{"y"},
		{"foobar"},
	}

	for i, tt := range tests {
		stmt := file.Statements[i]
		if !testLetStatement(t, stmt, tt.expectedIdentifier) {
			return
		}
	}
}

func testLetStatement(t *testing.T, s ast.Statement, name string) bool {
	if s.Pos().Literal != "let" {
		t.Errorf("s.Pos().Literal not 'let'. got=%q", s.Pos().Literal)
		return false
	}

	letStmt, ok := s.(*ast.LetStatement)
	if !ok {
		t.Errorf("s not *ast.LetStatement. got=%T", s)
		return false
	}

	if letStmt.Name.Value != name {
		t.Errorf("letStmt.Name.Value not '%s'. got=%s", name, letStmt.Name.Value)
		return false
	}

	if letStmt.Name.Token.Literal != name {
		t.Errorf("letStmt.Name.Token.Literal not '%s'. got=%s", name, letStmt.Name.Token.Literal)
		return false
	}

	return true
}

func TestReturnStatements(t *testing.T) {
	input := `
return 5;
return 10;
return 993322;
`

	l := lexer.New("test.ship", input)
	p := New(l)

	file := p.ParseFile()
	checkParserErrors(t, p)

	if len(file.Statements) != 3 {
		t.Fatalf("file.Statements does not contain 3 statements. got=%d", len(file.Statements))
	}

	for _, stmt := range file.Statements {
		returnStmt, ok := stmt.(*ast.ReturnStatement)
		if !ok {
			t.Errorf("stmt not *ast.ReturnStatement. got=%T", stmt)
			continue
		}
		if returnStmt.Token.Literal != "return" {
			t.Errorf("returnStmt.Token.Literal not 'return', got %q", returnStmt.Token.Literal)
		}
	}
}

func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}
