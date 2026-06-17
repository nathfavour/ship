package token

type Type string

const (
	ILLEGAL Type = "ILLEGAL"
	EOF     Type = "EOF"
	IDENT   Type = "IDENT"
	INT     Type = "INT"
	STRING  Type = "STRING"

	// Operators & Delimiters
	ASSIGN   Type = "="
	PLUS     Type = "+"
	MINUS    Type = "-"
	BANG     Type = "!"
	ASTERISK Type = "*"
	SLASH    Type = "/"

	LT Type = "<"
	GT Type = ">"

	EQ     Type = "=="
	NOT_EQ Type = "!="

	COMMA     Type = ","
	SEMICOLON Type = ";"
	COLON     Type = ":"
	DOT       Type = "."

	LPAREN Type = "("
	RPAREN Type = ")"
	LBRACE Type = "{"
	RBRACE Type = "}"
	LBRACK Type = "["
	RBRACK Type = "]"

	ARROW Type = "->"

	// Keywords
	TYPE     Type = "type"
	STRUCT   Type = "struct"
	FN       Type = "fn"
	CONTRACT Type = "contract"
	REQUIRE  Type = "require"
	ENSURE   Type = "ensure"
	LET      Type = "let"
	DEFER    Type = "defer"
	ELSE     Type = "else"
	RETURN   Type = "return"
	IF       Type = "if"
	NEW      Type = "new" // Included for completeness, though ARCHITECTURE says "new does not exist" as implicit.
	TRUE     Type = "true"
	FALSE    Type = "false"
)

type Token struct {
	Type    Type
	Literal string
	File    string
	Line    int
	Col     int
}

var keywords = map[string]Type{
	"type":     TYPE,
	"struct":   STRUCT,
	"fn":       FN,
	"contract": CONTRACT,
	"require":  REQUIRE,
	"ensure":   ENSURE,
	"let":      LET,
	"defer":    DEFER,
	"else":     ELSE,
	"return":   RETURN,
	"if":       IF,
	"new":      NEW,
	"true":     TRUE,
	"false":    FALSE,
}

func LookupIdent(ident string) Type {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
