package compiler

import (
	"interpreter/ast"
	"interpreter/code"
	"interpreter/lexer"
	"interpreter/object"
	"interpreter/parser"
)

type Compiler struct {
	instructions code.Instructions
	constants    []object.Object
}

func New() *Compiler {
	return &Compiler{
		instructions: code.Instructions{},
		constants:    []object.Object{},
	}
}

func (c *Compiler) Compile(node ast.Node) error {
	return nil
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: code.Instructions{},
		Constants:    []object.Object{},
	}
}

type Bytecode struct {
	Instructions code.Instructions
	Constants    []object.Object
}

func parse(input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)
	return p.ParseProgram()
}
