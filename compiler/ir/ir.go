package ir

import "fmt"

type OpCode int

const (
	OpStore OpCode = iota
	OpLoad
	OpAdd
	OpSub
	OpMul
	OpDiv
	OpCall
	OpLabel
	OpJumpIfZero
	OpJump
	OpRet
	OpEq
	OpNeq
	OpLt
	OpGt
)

func (op OpCode) String() string {
	switch op {
	case OpStore: return "STORE"
	case OpLoad: return "LOAD"
	case OpAdd: return "ADD"
	case OpSub: return "SUB"
	case OpMul: return "MUL"
	case OpDiv: return "DIV"
	case OpCall: return "CALL"
	case OpLabel: return "LABEL"
	case OpJumpIfZero: return "JMPZ"
	case OpJump: return "JMP"
	case OpRet: return "RET"
	case OpEq: return "EQ"
	case OpNeq: return "NEQ"
	case OpLt: return "LT"
	case OpGt: return "GT"
	}
	return "UNKNOWN"
}

type Operand struct {
	Type  string // e.g. "register", "immediate", "label"
	Value string
}

func (o Operand) String() string {
	return o.Value
}

type Instruction struct {
	Op       OpCode
	Dest     Operand
	Src1     Operand
	Src2     Operand
	Comment  string
}

func (i Instruction) String() string {
	if i.Op == OpLabel {
		return fmt.Sprintf("%s:", i.Dest.Value)
	}
	res := fmt.Sprintf("  %s", i.Op.String())
	if i.Dest.Value != "" {
		res += fmt.Sprintf(" %s", i.Dest.Value)
	}
	if i.Src1.Value != "" {
		res += fmt.Sprintf(", %s", i.Src1.Value)
	}
	if i.Src2.Value != "" {
		res += fmt.Sprintf(", %s", i.Src2.Value)
	}
	if i.Comment != "" {
		res += fmt.Sprintf(" ; %s", i.Comment)
	}
	return res
}

type Program struct {
	Instructions []Instruction
}

func (p *Program) String() string {
	res := ""
	for _, inst := range p.Instructions {
		res += inst.String() + "\n"
	}
	return res
}
