package vm

import (
	"encoding/binary"
	"fmt"
	"interpreter/code"
	"interpreter/compiler"
	"interpreter/object"
)

const StackSize = 2048
const GlobalsSize = 65536

var True = &object.Boolean{Value: true}
var False = &object.Boolean{Value: false}
var Null = &object.Null{}

type VM struct {
	constants    []object.Object
	instructions code.Instructions
	stack        []object.Object
	sp           int // points on the next value. Top of the stack is stack[sp-1]
	globals      []object.Object
}

func New(bytecode *compiler.Bytecode) *VM {
	return &VM{
		instructions: bytecode.Instructions,
		constants:    bytecode.Constants,
		stack:        make([]object.Object, StackSize),
		sp:           0,
		globals:      make([]object.Object, GlobalsSize),
	}
}

func NewWithGlobalsStore(bytecode *compiler.Bytecode, s []object.Object) *VM {
	vm := New(bytecode)
	vm.globals = s
	return vm
}

func (vm *VM) Run() error {
	for ip := 0; ip < len(vm.instructions); ip++ {
		op := code.Opcode(vm.instructions[ip])
		switch op {
		case code.OpConstant:
			constIdx := binary.BigEndian.Uint16(vm.instructions[ip+1:])
			ip += 2
			err := vm.push(vm.constants[constIdx])
			if err != nil {
				return err
			}
		case code.OpPop:
			vm.pop()
		case code.OpAdd:
			vm.executeBinaryOperation(op)
		case code.OpSub:
			vm.executeBinaryOperation(op)
		case code.OpDiv:
			vm.executeBinaryOperation(op)
		case code.OpMul:
			vm.executeBinaryOperation(op)
		case code.OpEqual:
			vm.executeBinaryOperation(op)
		case code.OpNotEqual:
			vm.executeBinaryOperation(op)
		case code.OpGreaterThan:
			vm.executeBinaryOperation(op)
		case code.OpTrue:
			err := vm.push(True)
			if err != nil {
				return err
			}
		case code.OpFalse:
			err := vm.push(False)
			if err != nil {
				return err
			}
		case code.OpBang:
			err := vm.executeBangOperator()
			if err != nil {
				return err
			}
		case code.OpMinus:
			err := vm.executeMinusOperator()
			if err != nil {
				return err
			}
		case code.OpNull:
			err := vm.push(Null)
			if err != nil {
				return err
			}
		case code.OpJump:
			pos := int(binary.BigEndian.Uint16(vm.instructions[ip+1:]))
			ip = pos - 1
		case code.OpJumpNotTruthy:
			pos := int(binary.BigEndian.Uint16(vm.instructions[ip+1:]))
			ip += 2

			condition := vm.pop()

			if !isTruthy(condition) {
				ip = pos - 1
			}
		case code.OpSetGlobal:
			index := binary.BigEndian.Uint16(vm.instructions[ip+1:])
			ip += 2
			vm.globals[index] = vm.pop()
		case code.OpGetGlobal:
			index := binary.BigEndian.Uint16(vm.instructions[ip+1:])
			ip += 2
			err := vm.push(vm.globals[index])
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func isTruthy(condition object.Object) bool {
	switch condition {
	case True:
		return true
	case False:
		return false
	case Null:
		return false
	default:
		return true
	}
}

func (vm *VM) executeMinusOperator() error {
	operand := vm.pop()
	if operand.Type() == object.INTEGER_OBJ {
		rv := operand.(*object.Integer).Value
		return vm.push(&object.Integer{Value: -rv})
	}
	return fmt.Errorf("unsupported type for minus operation: %s", operand.Type())
}

func (vm *VM) executeBangOperator() error {
	operand := vm.pop()
	switch operand {
	case True:
		return vm.push(False)
	case False:
		return vm.push(True)
	case Null:
		return vm.push(True)
	default:
		return vm.push(False)
	}
}

func (vm *VM) executeBinaryOperation(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	if left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ {
		return vm.executeBinaryIntegerOperation(left, right, op)
	}
	if left.Type() == object.BOOLEAN_OBJ && right.Type() == object.BOOLEAN_OBJ {
		return vm.executeBinaryBooleanOperation(left, right, op)
	}
	if left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ {
		return vm.executeBinaryStringOperation(left, right, op)
	}
	return fmt.Errorf("unsupported types for binary operation: %s %s", left.Type(), right.Type())
}

func (vm *VM) executeBinaryIntegerOperation(left, right object.Object, op code.Opcode) error {
	lv := left.(*object.Integer).Value
	rv := right.(*object.Integer).Value
	switch op {
	case code.OpAdd:
		return vm.push(&object.Integer{Value: lv + rv})
	case code.OpSub:
		return vm.push(&object.Integer{Value: lv - rv})
	case code.OpMul:
		return vm.push(&object.Integer{Value: lv * rv})
	case code.OpDiv:
		return vm.push(&object.Integer{Value: lv / rv})
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObject(lv == rv))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(lv != rv))
	case code.OpGreaterThan:
		return vm.push(nativeBoolToBooleanObject(lv > rv))
	default:
		return fmt.Errorf("unknown integer operator: %d", op)
	}
}

func (vm *VM) executeBinaryBooleanOperation(left, right object.Object, op code.Opcode) error {
	lv := left.(*object.Boolean).Value
	rv := right.(*object.Boolean).Value
	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObject(lv == rv))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(lv != rv))
	default:
		return fmt.Errorf("unknown boolean operator: %d", op)
	}
}

func (vm *VM) executeBinaryStringOperation(left, right object.Object, op code.Opcode) error {
	lv := left.(*object.String).Value
	rv := right.(*object.String).Value
	switch op {
	case code.OpAdd:
		return vm.push(&object.String{Value: lv + rv})
	default:
		return fmt.Errorf("unknown string operator: %d", op)
	}
}

func nativeBoolToBooleanObject(v bool) *object.Boolean {
	if v {
		return True
	}
	return False
}

func (vm *VM) push(c object.Object) error {
	if vm.sp > StackSize {
		return fmt.Errorf("stack oveflow")
	}
	vm.stack[vm.sp] = c
	vm.sp++
	return nil
}

func (vm *VM) pop() object.Object {
	if vm.sp == 0 {
		return nil
	}
	vm.sp--
	return vm.stack[vm.sp]
}

func (vm *VM) StackTop() object.Object {
	if vm.sp == 0 {
		return nil
	}
	return vm.stack[vm.sp-1]
}

func (vm *VM) LastPoppedStackElem() object.Object {
	return vm.stack[vm.sp]
}
