package evaluator

import (
	"fmt"

	"interpreter/ast"

	"interpreter/object"
)

var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {
	case *ast.Program:
		return evalProgram(node.Statements, env)
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.Boolean:
		return nativeBoolToBooleanObject(node.Value)
	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		return evalInfixExpression(node.Operator, left, right)
	case *ast.IfExpression:
		val := Eval(node.Condition, env)
		if isError(val) {
			return val
		}
		cond := isTruthy(val)
		if cond {
			return Eval(node.Consequence, env)
		}
		if node.Alternative != nil {
			return Eval(node.Alternative, env)
		}
		return NULL
	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue, env)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}
	case *ast.BlockStatement:
		return evalBlockStatements(node.Statements, env)
	case *ast.LetStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		env.Set(node.Name.Value, val)
		return val
	case *ast.Identifier:
		return evalIdentifier(node, env)
	}
	return nil
}

func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	val, ok := env.Get(node.Value)
	if !ok {
		return newError("identifier not found %s", node.Value)
	}
	return val
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}

func isTruthy(obj object.Object) bool {
	switch obj {
	case TRUE:
		return true
	case FALSE:
		return false
	case NULL:
		return false
	default:
		return true
	}
}

func evalInfixExpression(op string, left, right object.Object) object.Object {
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(op, left, right)
	case left.Type() == object.BOOLEAN_OBJ && right.Type() == object.BOOLEAN_OBJ:
		return evalBooleanInfixExpression(op, left, right)
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s", left.Type(), op, right.Type())
	default:
		return newError("unknown operator: %s %s %s", left.Type(), op, right.Type())
	}
}
func evalBooleanInfixExpression(op string, left, right object.Object) object.Object {
	lv := left.(*object.Boolean).Value
	rv := right.(*object.Boolean).Value
	switch op {
	case "==":
		return nativeBoolToBooleanObject(lv == rv)
	case "!=":
		return nativeBoolToBooleanObject(lv != rv)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), op, right.Type())
	}
}
func evalIntegerInfixExpression(op string, left object.Object, right object.Object) object.Object {
	lv := left.(*object.Integer).Value
	rv := right.(*object.Integer).Value
	switch op {
	case "-":
		return &object.Integer{Value: lv - rv}
	case "+":
		return &object.Integer{Value: lv + rv}
	case "*":
		return &object.Integer{Value: lv * rv}
	case "/":
		return &object.Integer{Value: lv / rv}
	case ">":
		return nativeBoolToBooleanObject(lv > rv)
	case "<":
		return nativeBoolToBooleanObject(lv < rv)
	case "==":
		return nativeBoolToBooleanObject(lv == rv)
	case "!=":
		return nativeBoolToBooleanObject(lv != rv)
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), op, right.Type())
	}
}
func evalPrefixExpression(op string, right object.Object) object.Object {
	switch op {
	case "!":
		return evalBangOperator(right)
	case "-":
		return evalMinusOperator(right)
	}
	return newError("unknown operator: %s%s", op, right.Type())
}
func evalMinusOperator(right object.Object) object.Object {
	if right.Type() != object.INTEGER_OBJ {
		return newError("unknown operator: -%s", right.Type())
	}
	value := right.(*object.Integer).Value
	return &object.Integer{Value: -value}
}
func evalBangOperator(right object.Object) object.Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}

func evalProgram(statements []ast.Statement, env *object.Environment) object.Object {
	var result object.Object
	for _, statement := range statements {
		result = Eval(statement, env)
		if result.Type() == object.RETURN_VALUE_OBJ {
			return result.(*object.ReturnValue).Value
		}
		if result.Type() == object.ERROR_OBJ {
			return result
		}
	}
	return result
}
func evalBlockStatements(statements []ast.Statement, env *object.Environment) object.Object {
	var result object.Object
	for _, statement := range statements {
		result = Eval(statement, env)
		if result != nil && result.Type() == object.RETURN_VALUE_OBJ {
			return result
		}
		if result.Type() == object.ERROR_OBJ {
			return result
		}
	}
	return result
}
func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}
