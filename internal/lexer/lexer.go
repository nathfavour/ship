package lexer

import (
	"github.com/nathfavour/ship/internal/token"
)

type Lexer struct {
	input        string
	file         string
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	ch           byte // current char under examination
	line         int
	col          int
}

func New(file, input string) *Lexer {
	l := &Lexer{
		input: input,
		file:  file,
		line:  1,
		col:   0,
	}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition += 1
	l.col += 1
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespace()

	startLine := l.line
	startCol := l.col

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.EQ, Literal: string(ch) + string(l.ch), File: l.file, Line: startLine, Col: startCol}
		} else {
			tok = newToken(token.ASSIGN, l.ch, l.file, startLine, startCol)
		}
	case '+':
		tok = newToken(token.PLUS, l.ch, l.file, startLine, startCol)
	case '-':
		// potentially handle '->'
		if l.peekChar() == '>' {
			ch := l.ch
			l.readChar()
			// For now, we don't have an ARROW token, but we should add one. Let's just emit MINUS and GT if needed, or add ARROW.
			// Let's add ARROW to token.go later, for now just emit ILLEGAL or we can add it here as "->"
			// Wait, the syntax in ARCHITECTURE uses `->`
			// I will update token.go later. Let's assume token.Type("->")
			tok = token.Token{Type: token.ARROW, Literal: string(ch) + string(l.ch), File: l.file, Line: startLine, Col: startCol}
		} else {
			tok = newToken(token.MINUS, l.ch, l.file, startLine, startCol)
		}
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.NOT_EQ, Literal: string(ch) + string(l.ch), File: l.file, Line: startLine, Col: startCol}
		} else {
			tok = newToken(token.BANG, l.ch, l.file, startLine, startCol)
		}
	case '/':
		// Handle comments
		if l.peekChar() == '/' {
			l.skipSingleLineComment()
			return l.NextToken()
		}
		tok = newToken(token.SLASH, l.ch, l.file, startLine, startCol)
	case '*':
		tok = newToken(token.ASTERISK, l.ch, l.file, startLine, startCol)
	case '<':
		tok = newToken(token.LT, l.ch, l.file, startLine, startCol)
	case '>':
		tok = newToken(token.GT, l.ch, l.file, startLine, startCol)
	case ';':
		tok = newToken(token.SEMICOLON, l.ch, l.file, startLine, startCol)
	case ':':
		tok = newToken(token.COLON, l.ch, l.file, startLine, startCol)
	case ',':
		tok = newToken(token.COMMA, l.ch, l.file, startLine, startCol)
	case '.':
		tok = newToken(token.DOT, l.ch, l.file, startLine, startCol)
	case '(':
		tok = newToken(token.LPAREN, l.ch, l.file, startLine, startCol)
	case ')':
		tok = newToken(token.RPAREN, l.ch, l.file, startLine, startCol)
	case '{':
		tok = newToken(token.LBRACE, l.ch, l.file, startLine, startCol)
	case '}':
		tok = newToken(token.RBRACE, l.ch, l.file, startLine, startCol)
	case '[':
		tok = newToken(token.LBRACK, l.ch, l.file, startLine, startCol)
	case ']':
		tok = newToken(token.RBRACK, l.ch, l.file, startLine, startCol)
	case '"':
		tok.Type = token.STRING
		tok.Literal = l.readString()
		tok.File = l.file
		tok.Line = startLine
		tok.Col = startCol
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
		tok.File = l.file
		tok.Line = startLine
		tok.Col = startCol
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			tok.File = l.file
			tok.Line = startLine
			tok.Col = startCol
			return tok
		} else if isDigit(l.ch) {
			tok.Type = token.INT
			tok.Literal = l.readNumber()
			tok.File = l.file
			tok.Line = startLine
			tok.Col = startCol
			return tok
		} else {
			tok = newToken(token.ILLEGAL, l.ch, l.file, startLine, startCol)
		}
	}

	l.readChar()
	return tok
}

func newToken(tokenType token.Type, ch byte, file string, line, col int) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch), File: file, Line: line, Col: col}
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readNumber() string {
	position := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readString() string {
	position := l.position + 1
	for {
		l.readChar()
		if l.ch == '"' || l.ch == 0 {
			break
		}
	}
	return l.input[position:l.position]
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		if l.ch == '\n' {
			l.line += 1
			l.col = 0
		}
		l.readChar()
	}
}

func (l *Lexer) skipSingleLineComment() {
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
	if l.ch == '\n' {
		l.line += 1
		l.col = 0
		l.readChar()
	}
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}
