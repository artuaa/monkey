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

var builtins = map[string]*object.Builtin{
	"puts": {
		Fn: func(args ...object.Object) object.Object {
			for _, arg := range args {
				fmt.Println(arg.Inspect())
			}
			return NULL
		},
	},
	"len": {
		Fn: func(args ...object.Object) object.Object {
			if l := len(args); l != 1 {
				return newError("wrong number of arguments. got=%d, want=%d", l, 1)
			}
			switch arg := args[0].(type) {
			case *object.String:
				return &object.Integer{Value: int64(len(arg.Value))}
			case *object.Array:
				return &object.Integer{Value: int64(len(arg.Elements))}
			}
			return newError("argument to `len` not supported, got %s", args[0].Type())
		},
	},
	"first": {
		Fn: func(args ...object.Object) object.Object {
			if l := len(args); l != 1 {
				return newError("wrong number of arguments. got=%d, want=%d", l, 1)
			}
			switch arg := args[0].(type) {
			case *object.Array:
				if len(arg.Elements) > 1 {
					return arg.Elements[0]
				}
				return NULL
			}
			return newError("argument to `first` not supported, got %s", args[0].Type())
		},
	},
	"last": {
		Fn: func(args ...object.Object) object.Object {
			if l := len(args); l != 1 {
				return newError("wrong number of arguments. got=%d, want=%d", l, 1)
			}
			switch arg := args[0].(type) {
			case *object.Array:
				if len(arg.Elements) > 1 {
					return arg.Elements[len(arg.Elements)-1]
				}
				return NULL
			}
			return newError("argument to `last` not supported, got %s", args[0].Type())
		},
	},
	"rest": {
		Fn: func(args ...object.Object) object.Object {
			if l := len(args); l != 1 {
				return newError("wrong number of arguments. got=%d, want=%d", l, 1)
			}
			switch arg := args[0].(type) {
			case *object.Array:
				if length := len(arg.Elements); length > 1 {
					elms := make([]object.Object, length-1, length-1)
					copy(elms, arg.Elements[1:])
					return &object.Array{Elements: elms}
				}
				return &object.Array{Elements: []object.Object{}}
			}
			return newError("argument to `rest` not supported, got %s", args[0].Type())
		},
	},
	"push": {
		Fn: func(args ...object.Object) object.Object {
			if l := len(args); l != 2 {
				return newError("wrong number of arguments. got=%d, want=%d", l, 2)
			}
			switch arg := args[0].(type) {
			case *object.Array:
				length := len(arg.Elements)
				newElements := make([]object.Object, length+1, length+1)
				copy(newElements, arg.Elements)
				newElements[length] = args[1]
				return &object.Array{Elements: newElements}
			}
			return newError("argument to `push` not supported, got %s", args[0].Type())
		},
	},
}

func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {
	case *ast.Program:
		return evalProgram(node.Statements, env)
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}
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
	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		return &object.Function{Parameters: params, Body: body, Env: env}
	case *ast.CallExpression:
		fun := Eval(node.Function, env)
		if isError(fun) {
			return fun
		}
		args := evalExpressions(node.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}
		return applyFunction(fun, args)
	case *ast.ArrayLiteral:
		elms := evalExpressions(node.Elements, env)
		if len(elms) == 1 && isError(elms[0]) {
			return elms[0]
		}
		return &object.Array{Elements: elms}
	case *ast.IndexExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		index := Eval(node.Index, env)
		if isError(index) {
			return index
		}
		return evalIndexExpression(left, index)
	case *ast.HashLiteral:
		result := &object.Hash{}
		result.Pairs = make(map[object.HashKey]object.HashPair)
		for key, val := range node.Pairs {
			k := Eval(key, env)
			if isError(k) {
				return k
			}
			v := Eval(val, env)
			if isError(v) {
				return v
			}
			hashKey, ok := k.(object.Hashable)
			if !ok {
				return newError("unusable as hash key: %s", k.Type())
			}
			result.Pairs[hashKey.HashKey()] = object.HashPair{Key: k, Value: v}
		}
		return result
	}
	return nil
}

func evalIndexExpression(left, index object.Object) object.Object {
	switch left.Type() {
	case object.ARRAY_OBJ:
		return evalArrayIndexExpression(left, index)
	case object.HASH_OBJ:
		return evalHashIndexExpression(left, index)
	}
	return newError("index operator not supported %s", left.Type())
}

func evalHashIndexExpression(left, index object.Object) object.Object {
	hash := left.(*object.Hash)
	key, ok := index.(object.Hashable)
	if !ok {
		return newError("unusable as hash key: %s", index.Type())
	}
	result, ok := hash.Pairs[key.HashKey()]
	if !ok {
		return NULL
	}
	return result.Value
}

func evalArrayIndexExpression(array, index object.Object) object.Object {
	if array.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ {
		idx := index.(*object.Integer).Value
		elms := array.(*object.Array).Elements
		if v := int(idx); v >= len(elms) || v < 0 {
			return NULL
		}
		return elms[idx]
	}
	return newError("index operator not supported %s", array.Type())
}

func applyFunction(fn object.Object, args []object.Object) object.Object {
	switch fn := fn.(type) {
	case *object.Function:
		extendedEnv := extendFunctionEnv(fn, args)
		evaluated := Eval(fn.Body, extendedEnv)
		return unwrapReturnValue(evaluated)
	case *object.Builtin:
		return fn.Fn(args...)
	default:
		return newError("not a function %s", fn.Type())
	}
}

func unwrapReturnValue(obj object.Object) object.Object {
	if rv, ok := obj.(*object.ReturnValue); ok {
		return rv.Value
	}
	return obj
}

func extendFunctionEnv(function *object.Function, args []object.Object) *object.Environment {
	env := object.NewEnclosedEnvironment(function.Env)
	for i, param := range function.Parameters {
		env.Set(param.Value, args[i])
	}
	return env
}

func evalExpressions(expression []ast.Expression, env *object.Environment) []object.Object {
	var result []object.Object
	for _, e := range expression {
		evaluated := Eval(e, env)
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}
	return result
}

func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	if val, ok := env.Get(node.Value); ok {
		return val
	}
	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}

	return newError("identifier not found: %s", node.Value)
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
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalStringInfixExpression(op, left, right)
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s", left.Type(), op, right.Type())
	default:
		return newError("unknown operator: %s %s %s", left.Type(), op, right.Type())
	}
}

func evalStringInfixExpression(op string, left, right object.Object) object.Object {
	lv := left.(*object.String).Value
	rv := right.(*object.String).Value
	switch op {
	case "+":
		return &object.String{Value: lv + rv}
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), op, right.Type())
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
