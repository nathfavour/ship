package ast

import "github.com/nathfavour/ship/compiler/token"

type Node interface {
	Pos() token.Token
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

type File struct {
	Name       string
	Statements []Statement
}

type StructField struct {
	Name token.Token // Identifier
	Type token.Token // Type Identifier
}

type StructDecl struct {
	Token  token.Token // token.TYPE
	Name   token.Token // Name of the struct
	Fields []StructField
}

func (sd *StructDecl) statementNode() {}
func (sd *StructDecl) Pos() token.Token { return sd.Token }

type ContractBlock struct {
	Token    token.Token // token.CONTRACT
	Requires []Expression
	Ensures  []Expression
}

type FuncDecl struct {
	Token      token.Token // token.FN
	Name       token.Token
	Params     []StructField
	ReturnType token.Token
	Contract   *ContractBlock
	Body       *BlockStatement
}

func (fd *FuncDecl) statementNode() {}
func (fd *FuncDecl) Pos() token.Token { return fd.Token }

type LetStatement struct {
	Token token.Token // token.LET
	Name  *Identifier
	Value Expression
}

func (ls *LetStatement) statementNode() {}
func (ls *LetStatement) Pos() token.Token { return ls.Token }

type ReturnStatement struct {
	Token       token.Token // token.RETURN
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode() {}
func (rs *ReturnStatement) Pos() token.Token { return rs.Token }

type ExpressionStatement struct {
	Token      token.Token // the first token of the expression
	Expression Expression
}

func (es *ExpressionStatement) statementNode() {}
func (es *ExpressionStatement) Pos() token.Token { return es.Token }

type BlockStatement struct {
	Token      token.Token // token.LBRACE
	Statements []Statement
}

func (bs *BlockStatement) statementNode() {}
func (bs *BlockStatement) Pos() token.Token { return bs.Token }

type Identifier struct {
	Token token.Token // token.IDENT
	Value string
}

func (i *Identifier) expressionNode() {}
func (i *Identifier) Pos() token.Token { return i.Token }

type IntegerLiteral struct {
	Token token.Token // token.INT
	Value int64
}

func (il *IntegerLiteral) expressionNode() {}
func (il *IntegerLiteral) Pos() token.Token { return il.Token }

type StringLiteral struct {
	Token token.Token // token.STRING
	Value string
}

func (sl *StringLiteral) expressionNode() {}
func (sl *StringLiteral) Pos() token.Token { return sl.Token }

type Boolean struct {
	Token token.Token
	Value bool
}

func (b *Boolean) expressionNode() {}
func (b *Boolean) Pos() token.Token { return b.Token }

type PrefixExpression struct {
	Token    token.Token // The prefix token, e.g. !
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode() {}
func (pe *PrefixExpression) Pos() token.Token { return pe.Token }

type InfixExpression struct {
	Token    token.Token // The operator token, e.g. +
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode() {}
func (ie *InfixExpression) Pos() token.Token { return ie.Token }

type IfExpression struct {
	Token       token.Token // token.IF
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (ie *IfExpression) expressionNode() {}
func (ie *IfExpression) Pos() token.Token { return ie.Token }

type CallExpression struct {
	Token     token.Token // The '(' token
	Function  Expression  // Identifier or FunctionLiteral
	Arguments []Expression
}

func (ce *CallExpression) expressionNode() {}
func (ce *CallExpression) Pos() token.Token { return ce.Token }
