package compiler

import (
	"fmt"
	"interpreter/ast"
	"interpreter/code"
	"interpreter/lexer"
	"interpreter/object"
	"interpreter/parser"
)

type EmittedInsturctions struct {
	Opcode   code.Opcode
	Position int
}

type Compiler struct {
	instructions        code.Instructions
	constants           []object.Object
	lastInstruction     EmittedInsturctions
	previousInstruction EmittedInsturctions
	symbolTable         *SymbolTable
}

func New() *Compiler {
	return &Compiler{
		instructions:        code.Instructions{},
		constants:           []object.Object{},
		lastInstruction:     EmittedInsturctions{},
		previousInstruction: EmittedInsturctions{},
		symbolTable:         NewSymbolTable(),
	}
}

func NewWithState(s *SymbolTable, constants []object.Object) *Compiler {
	return &Compiler{
		instructions:        code.Instructions{},
		constants:           constants,
		lastInstruction:     EmittedInsturctions{},
		previousInstruction: EmittedInsturctions{},
		symbolTable:         s,
	}
}

func (c *Compiler) Compile(node ast.Node) error {
	switch node := node.(type) {
	case *ast.Program:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}
	case *ast.BlockStatement:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}
	case *ast.ExpressionStatement:
		err := c.Compile(node.Expression)
		if err != nil {
			return err
		}
		c.emit(code.OpPop)
	case *ast.InfixExpression:
		// switch order
		if node.Operator == "<" {
			err := c.Compile(node.Right)
			if err != nil {
				return err
			}
			err = c.Compile(node.Left)
			if err != nil {
				return err
			}
		} else {
			err := c.Compile(node.Left)
			if err != nil {
				return err
			}
			err = c.Compile(node.Right)
			if err != nil {
				return err
			}
		}
		switch node.Operator {
		case "+":
			c.emit(code.OpAdd)
		case "-":
			c.emit(code.OpSub)
		case "/":
			c.emit(code.OpDiv)
		case "*":
			c.emit(code.OpMul)
		case "<":
			c.emit(code.OpGreaterThan)
		case ">":
			c.emit(code.OpGreaterThan)
		case "==":
			c.emit(code.OpEqual)
		case "!=":
			c.emit(code.OpNotEqual)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}
	case *ast.PrefixExpression:
		err := c.Compile(node.Right)
		if err != nil {
			return err
		}
		switch node.Operator {
		case "-":
			c.emit(code.OpMinus)
		case "!":
			c.emit(code.OpBang)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}
	case *ast.IntegerLiteral:
		integer := &object.Integer{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(integer))
	case *ast.Boolean:
		if node.Value {
			c.emit(code.OpTrue)
		} else {
			c.emit(code.OpFalse)
		}
	case *ast.IfExpression:
		err := c.Compile(node.Condition)
		if err != nil {
			return err
		}
		jumpNotTruthy := c.emit(code.OpJumpNotTruthy, 9999)

		err = c.Compile(node.Consequence)
		if err != nil {
			return err
		}
		if c.lastInstructionIsPop() {
			c.removeLastPop()
		}

		jumpPos := c.emit(code.OpJump, 9999)
		afterConsequencePos := len(c.instructions)
		c.changeOperand(jumpNotTruthy, afterConsequencePos)

		if node.Alternative == nil {
			c.emit(code.OpNull)
		} else {
			err := c.Compile(node.Alternative)
			if err != nil {
				return err
			}
			if c.lastInstructionIsPop() {
				c.removeLastPop()
			}
		}
		afterAlternativePos := len(c.instructions)
		c.changeOperand(jumpPos, afterAlternativePos)
	case *ast.LetStatement:
		err := c.Compile(node.Value)
		if err != nil {
			return err
		}
		symbol := c.symbolTable.Define(node.Name.Value)
		c.emit(code.OpSetGlobal, symbol.Index)
	case *ast.Identifier:
		symbol, ok := c.symbolTable.Resolve(node.Value)
		if !ok {
			return fmt.Errorf("undefined variable %s", node.Value)
		}
		c.emit(code.OpGetGlobal, symbol.Index)
	case *ast.StringLiteral:
       	str := &object.String{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(str))
	}
	return nil
}

func (c *Compiler) removeLastPop() {
	c.instructions = c.instructions[:c.lastInstruction.Position]
	c.lastInstruction = c.previousInstruction
}

func (c *Compiler) changeOperand(opPos int, operand int) {
	op := code.Opcode(c.instructions[opPos])
	ins := code.Make(op, operand)
	c.replaceInstruction(opPos, ins)
}

func (c *Compiler) replaceInstruction(opPos int, ins []byte) {
	for i := 0; i < len(ins); i++ {
		c.instructions[opPos+i] = ins[i]
	}
}

func (c *Compiler) lastInstructionIsPop() bool {
	return c.lastInstruction.Opcode == code.OpPop
}

func (c *Compiler) emit(op code.Opcode, operands ...int) int {
	ins := code.Make(op, operands...)
	pos := c.addInstruction(ins)
	c.setLastInstruction(op, pos)
	return pos
}

func (c *Compiler) setLastInstruction(op code.Opcode, pos int) {
	previous := c.lastInstruction
	last := EmittedInsturctions{Opcode: op, Position: pos}
	c.previousInstruction = previous
	c.lastInstruction = last
}

func (c *Compiler) addInstruction(ins []byte) int {
	pos := len(c.instructions)
	c.instructions = append(c.instructions, ins...)
	return pos
}

func (c *Compiler) addConstant(integer object.Object) int {
	c.constants = append(c.constants, integer)
	return len(c.constants) - 1
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.instructions,
		Constants:    c.constants,
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
