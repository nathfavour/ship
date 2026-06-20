package types

import (
	"fmt"

	"github.com/nathfavour/ship/compiler/ast"
	"github.com/nathfavour/ship/compiler/token"
)

// Type represents a resolved type in the Ship language.
type Type interface {
	Name() string
	SizeBytes() int
}

type BaseType struct {
	name string
	size int
}

func (b *BaseType) Name() string   { return b.name }
func (b *BaseType) SizeBytes() int { return b.size }

var (
	IntType    = &BaseType{name: "int", size: 8}
	StringType = &BaseType{name: "string", size: 16} // Assuming 16 bytes for string header
	BoolType   = &BaseType{name: "bool", size: 1}
	VoidType   = &BaseType{name: "void", size: 0}
)

type StructFieldInfo struct {
	Name   string
	Type   Type
	Offset int
}

type StructType struct {
	name   string
	size   int
	fields []StructFieldInfo
}

func (s *StructType) Name() string   { return s.name }
func (s *StructType) SizeBytes() int { return s.size }
func (s *StructType) Fields() []StructFieldInfo { return s.fields }

type CheckerError struct {
	Phase      string
	ErrorCode  string
	TargetFile string
	Function   string
	Line       int
	Col        int
	Context    map[string]interface{}
}

func (e *CheckerError) Error() string {
	return fmt.Sprintf("type error [%s] at %s:%d:%d - context: %v", e.ErrorCode, e.TargetFile, e.Line, e.Col, e.Context)
}

type Env struct {
	vars  map[string]Type
	outer *Env
}

func NewEnv(outer *Env) *Env {
	return &Env{vars: make(map[string]Type), outer: outer}
}

func (e *Env) Set(name string, t Type) {
	e.vars[name] = t
}

func (e *Env) Get(name string) (Type, bool) {
	t, ok := e.vars[name]
	if !ok && e.outer != nil {
		return e.outer.Get(name)
	}
	return t, ok
}

type Checker struct {
	globalEnv    *Env
	currentEnv   *Env
	structs      map[string]*StructType
	errors       []*CheckerError
	currentFile  string
	currentFunc  string
	VarTypes     map[string]string // Expose var types
}

func NewChecker() *Checker {
	env := NewEnv(nil)
	env.Set("true", BoolType)
	env.Set("false", BoolType)
	
	return &Checker{
		globalEnv:  env,
		currentEnv: env,
		structs:    make(map[string]*StructType),
		errors:     []*CheckerError{},
		VarTypes:   make(map[string]string),
	}
}

func (c *Checker) Errors() []*CheckerError {
	return c.errors
}

func (c *Checker) CheckFile(file *ast.File) {
	c.currentFile = file.Name
	for _, stmt := range file.Statements {
		c.checkStatement(stmt)
	}
}

func (c *Checker) reportError(code string, pos token.Token, context map[string]interface{}) {
	c.errors = append(c.errors, &CheckerError{
		Phase:      "TYPE_CHECKER",
		ErrorCode:  code,
		TargetFile: pos.File,
		Function:   c.currentFunc,
		Line:       pos.Line,
		Col:        pos.Col,
		Context:    context,
	})
}

func (c *Checker) checkStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.StructDecl:
		c.checkStructDecl(s)
	case *ast.FuncDecl:
		c.checkFuncDecl(s)
	case *ast.LetStatement:
		c.checkLetStatement(s)
	case *ast.ReturnStatement:
		c.checkReturnStatement(s)
	case *ast.ExpressionStatement:
		c.checkExpression(s.Expression)
	case *ast.BlockStatement:
		c.checkBlockStatement(s)
	}
}

func (c *Checker) checkStructDecl(s *ast.StructDecl) {
	st := &StructType{
		name:   s.Name.Literal,
		fields: []StructFieldInfo{},
		size:   0,
	}

	offset := 0
	for _, f := range s.Fields {
		fieldType := c.resolveTypeName(f.Type)
		if fieldType == nil {
			c.reportError("UNKNOWN_TYPE", f.Type, map[string]interface{}{"type": f.Type.Literal})
			return
		}
		st.fields = append(st.fields, StructFieldInfo{
			Name:   f.Name.Literal,
			Type:   fieldType,
			Offset: offset,
		})
		offset += fieldType.SizeBytes()
	}
	st.size = offset
	c.structs[s.Name.Literal] = st
}

func (c *Checker) resolveTypeName(tok token.Token) Type {
	switch tok.Literal {
	case "int":
		return IntType
	case "string":
		return StringType
	case "bool":
		return BoolType
	}
	if st, ok := c.structs[tok.Literal]; ok {
		return st
	}
	return nil
}

func (c *Checker) checkFuncDecl(s *ast.FuncDecl) {
	c.currentFunc = s.Name.Literal
	funcEnv := NewEnv(c.currentEnv)
	c.currentEnv = funcEnv
	defer func() { c.currentEnv = funcEnv.outer; c.currentFunc = "" }()

	if s.Receiver != nil {
		receiverType := c.resolveTypeName(s.Receiver.Type)
		if receiverType == nil {
			c.reportError("UNKNOWN_RECEIVER_TYPE", s.Receiver.Type, map[string]interface{}{"type": s.Receiver.Type.Literal})
		} else {
			c.currentEnv.Set(s.Receiver.Name.Literal, receiverType)
			c.VarTypes[s.Receiver.Name.Literal] = receiverType.Name()
		}
	}

	for _, param := range s.Params {
		paramType := c.resolveTypeName(param.Type)
		if paramType == nil {
			c.reportError("UNKNOWN_PARAM_TYPE", param.Type, map[string]interface{}{"type": param.Type.Literal})
			continue
		}
		c.currentEnv.Set(param.Name.Literal, paramType)
		c.VarTypes[param.Name.Literal] = paramType.Name()
	}

	if s.Contract != nil {
		c.checkContract(s.Contract)
	}

	if s.Body != nil {
		c.checkBlockStatement(s.Body)
	}
}

func (c *Checker) checkContract(s *ast.ContractBlock) {
	for _, req := range s.Requires {
		t := c.checkExpression(req)
		if t != BoolType && t != nil {
			c.reportError("CONTRACT_VIOLATION_REQUIRE", req.Pos(), map[string]interface{}{
				"ast_node":            "ContractBlock",
				"violated_expression": "require clause must be boolean",
				"inferred_type":       t.Name(),
			})
		}
	}
	for _, ens := range s.Ensures {
		t := c.checkExpression(ens)
		if t != BoolType && t != nil {
			c.reportError("CONTRACT_VIOLATION_ENSURE", ens.Pos(), map[string]interface{}{
				"ast_node":            "ContractBlock",
				"violated_expression": "ensure clause must be boolean",
				"inferred_type":       t.Name(),
			})
		}
	}
}

func (c *Checker) checkLetStatement(s *ast.LetStatement) {
	valType := c.checkExpression(s.Value)
	if valType != nil {
		c.currentEnv.Set(s.Name.Value, valType)
		c.VarTypes[s.Name.Value] = valType.Name()
	}
}

func (c *Checker) checkReturnStatement(s *ast.ReturnStatement) {
	c.checkExpression(s.ReturnValue)
	// Add proper return type matching when function signatures fully specify return types
}

func (c *Checker) checkBlockStatement(s *ast.BlockStatement) {
	blockEnv := NewEnv(c.currentEnv)
	c.currentEnv = blockEnv
	defer func() { c.currentEnv = blockEnv.outer }()

	for _, stmt := range s.Statements {
		c.checkStatement(stmt)
	}
}

func (c *Checker) checkExpression(exp ast.Expression) Type {
	switch e := exp.(type) {
	case *ast.IntegerLiteral:
		return IntType
	case *ast.StringLiteral:
		return StringType
	case *ast.Boolean:
		return BoolType
	case *ast.Identifier:
		t, ok := c.currentEnv.Get(e.Value)
		if !ok {
			c.reportError("UNDEFINED_IDENTIFIER", e.Token, map[string]interface{}{"identifier": e.Value})
			return nil
		}
		return t
	case *ast.PrefixExpression:
		return c.checkExpression(e.Right)
	case *ast.InfixExpression:
		left := c.checkExpression(e.Left)
		right := c.checkExpression(e.Right)
		if left != right {
			leftName := "unknown"
			if left != nil {
				leftName = left.Name()
			}
			rightName := "unknown"
			if right != nil {
				rightName = right.Name()
			}
			c.reportError("TYPE_MISMATCH", e.Token, map[string]interface{}{
				"left":  leftName,
				"right": rightName,
			})
			return nil
		}
		// For relational operators, result is bool
		if e.Operator == "==" || e.Operator == "!=" || e.Operator == "<" || e.Operator == ">" {
			return BoolType
		}
		return left
	case *ast.CallExpression:
		// very naive right now, ideally we should lookup function signature
		for _, arg := range e.Arguments {
			c.checkExpression(arg)
		}
		if ident, ok := e.Function.(*ast.Identifier); ok {
			if ident.Value == "read_file" || ident.Value == "read_str" || ident.Value == "input" || ident.Value == "net_read" {
				return StringType
			}
		}
		return IntType // default placeholder
	case *ast.SelectorExpression:
		leftType := c.checkExpression(e.Left)
		if leftType == nil {
			return nil
		}
		if st, ok := leftType.(*StructType); ok {
			for _, field := range st.fields {
				if field.Name == e.Right.Value {
					return field.Type
				}
			}
		}
		return IntType
	}
	return nil
}
