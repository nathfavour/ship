package ir

import (
	"fmt"

	"github.com/nathfavour/ship/compiler/ast"
)

type Lowerer struct {
	program      *Program
	regCounter   int
	labelCounter int
	varTypes     map[string]string
}

func NewLowerer(varTypes map[string]string) *Lowerer {
	return &Lowerer{
		program:  &Program{Instructions: []Instruction{}},
		varTypes: varTypes,
	}
}

func (l *Lowerer) newReg() Operand {
	r := fmt.Sprintf("r%d", l.regCounter)
	l.regCounter++
	return Operand{Type: "register", Value: r}
}

func (l *Lowerer) newLabel() string {
	lbl := fmt.Sprintf(".L%d", l.labelCounter)
	l.labelCounter++
	return lbl
}

func (l *Lowerer) emit(op OpCode, dest, src1, src2 Operand, comment string) {
	l.program.Instructions = append(l.program.Instructions, Instruction{
		Op:      op,
		Dest:    dest,
		Src1:    src1,
		Src2:    src2,
		Comment: comment,
	})
}

func (l *Lowerer) emitLabel(name string) {
	l.program.Instructions = append(l.program.Instructions, Instruction{
		Op:   OpLabel,
		Dest: Operand{Type: "label", Value: name},
	})
}

func (l *Lowerer) LowerFile(file *ast.File) *Program {
	for _, stmt := range file.Statements {
		l.lowerStatement(stmt)
	}
	return l.program
}

func (l *Lowerer) lowerStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.LetStatement:
		valReg := l.lowerExpression(s.Value)
		destReg := Operand{Type: "variable", Value: s.Name.Value}
		l.emit(OpStore, destReg, valReg, Operand{}, "")
	case *ast.ReturnStatement:
		valReg := l.lowerExpression(s.ReturnValue)
		l.emit(OpRet, Operand{}, valReg, Operand{}, "")
	case *ast.ExpressionStatement:
		l.lowerExpression(s.Expression)
	case *ast.FuncDecl:
		l.emitLabel(s.Name.Literal)
		// Parameters would go here
		if s.Body != nil {
			for _, bs := range s.Body.Statements {
				l.lowerStatement(bs)
			}
		}
	}
}

func (l *Lowerer) lowerExpression(exp ast.Expression) Operand {
	switch e := exp.(type) {
	case *ast.IntegerLiteral:
		reg := l.newReg()
		imm := Operand{Type: "immediate", Value: fmt.Sprintf("%d", e.Value)}
		l.emit(OpLoad, reg, imm, Operand{}, "")
		return reg
	case *ast.StringLiteral:
		reg := l.newReg()
		imm := Operand{Type: "string", Value: e.Value}
		l.emit(OpLoad, reg, imm, Operand{}, "")
		return reg
	case *ast.Identifier:
		reg := l.newReg()
		varOp := Operand{Type: "variable", Value: e.Value}
		l.emit(OpLoad, reg, varOp, Operand{}, "")
		return reg
	case *ast.InfixExpression:
		left := l.lowerExpression(e.Left)
		right := l.lowerExpression(e.Right)
		dest := l.newReg()

		switch e.Operator {
		case "+":
			l.emit(OpAdd, dest, left, right, "")
		case "-":
			l.emit(OpSub, dest, left, right, "")
		case "*":
			l.emit(OpMul, dest, left, right, "")
		case "/":
			l.emit(OpDiv, dest, left, right, "")
		case "==":
			l.emit(OpEq, dest, left, right, "")
		case "!=":
			l.emit(OpNeq, dest, left, right, "")
		case "<":
			l.emit(OpLt, dest, left, right, "")
		case ">":
			l.emit(OpGt, dest, left, right, "")
		}
		return dest
	case *ast.PrefixExpression:
		right := l.lowerExpression(e.Right)
		dest := l.newReg()
		if e.Operator == "-" {
			zeroReg := l.newReg()
			l.emit(OpLoad, zeroReg, Operand{Type: "immediate", Value: "0"}, Operand{}, "")
			l.emit(OpSub, dest, zeroReg, right, "")
		} else if e.Operator == "!" {
			oneReg := l.newReg()
			l.emit(OpLoad, oneReg, Operand{Type: "immediate", Value: "1"}, Operand{}, "")
			l.emit(OpSub, dest, oneReg, right, "")
		}
		return dest
	case *ast.IfExpression:
		// simple lower
		condReg := l.lowerExpression(e.Condition)
		falseLabel := l.newLabel()
		endLabel := l.newLabel()

		l.emit(OpJumpIfZero, Operand{Type: "label", Value: falseLabel}, condReg, Operand{}, "if cond is false jump to else")

		for _, s := range e.Consequence.Statements {
			l.lowerStatement(s)
		}
		
		l.emit(OpJump, Operand{Type: "label", Value: endLabel}, Operand{}, Operand{}, "")
		l.emitLabel(falseLabel)
		
		if e.Alternative != nil {
			for _, s := range e.Alternative.Statements {
				l.lowerStatement(s)
			}
		}
		
		l.emitLabel(endLabel)
		return condReg // normally if statements don't return values, but for now we return condReg
	case *ast.CallExpression:
		fnName := e.Function.(*ast.Identifier).Value
		if fnName == "write_file" && len(e.Arguments) == 2 {
			arg1 := l.lowerExpression(e.Arguments[0])
			arg2 := l.lowerExpression(e.Arguments[1])
			l.emit(OpStore, Operand{Type: "variable", Value: "__write_file_arg2"}, arg2, Operand{}, "")
			dest := l.newReg()
			l.emit(OpCall, dest, Operand{Type: "label", Value: "write_file"}, arg1, "")
			return dest
		}

		var argReg Operand
		if len(e.Arguments) > 0 {
			argReg = l.lowerExpression(e.Arguments[0])
			isString := false
			if _, ok := e.Arguments[0].(*ast.StringLiteral); ok {
				isString = true
			} else if ident, ok := e.Arguments[0].(*ast.Identifier); ok {
				if l.varTypes[ident.Value] == "string" {
					isString = true
				}
			} else if call, ok := e.Arguments[0].(*ast.CallExpression); ok {
				if callIdent, ok := call.Function.(*ast.Identifier); ok {
					if callIdent.Value == "read_file" || callIdent.Value == "read_str" || callIdent.Value == "input" {
						isString = true
					}
				}
			}

			if isString {
				if fnName == "print" {
					fnName = "print_str"
				} else if fnName == "println" {
					fnName = "println_str"
				}
			}
		}
		dest := l.newReg()
		l.emit(OpCall, dest, Operand{Type: "label", Value: fnName}, argReg, "")
		return dest
	}
	return Operand{}
}
