package object

import (
	"fmt"
)

var Builtins = []struct {
	Name    string
	Builtin *Builtin
}{
	{
		"len",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			switch arg := args[0].(type) {
			case *Array:
				return &Integer{Value: int64(len(arg.Elements))}
			case *String:
				return &Integer{Value: int64(len(arg.Value))}
			default:
				return newError("argument to `len` not supported, got %s",
					args[0].Type())
			}
		},
		},
	},
	{
		"puts",
		&Builtin{
			Fn: func(args ...Object) Object {
				for _, arg := range args {
					fmt.Println(arg.Inspect())
				}
				return nil
			},
		},
	},
	{
		"first",
		&Builtin{
			Fn: func(args ...Object) Object {
				if l := len(args); l != 1 {
					return newError("wrong number of arguments. got=%d, want=%d", l, 1)
				}
				switch arg := args[0].(type) {
				case *Array:
					if len(arg.Elements) > 1 {
						return arg.Elements[0]
					}
					return nil
				}
				return newError("argument to `first` must be %s, got %s", ARRAY_OBJ, args[0].Type())
			},
		},
	},
	{
		"last",
		&Builtin{
			Fn: func(args ...Object) Object {
				if l := len(args); l != 1 {
					return newError("wrong number of arguments. got=%d, want=%d", l, 1)
				}
				switch arg := args[0].(type) {
				case *Array:
					if len(arg.Elements) > 1 {
						return arg.Elements[len(arg.Elements)-1]
					}
					return nil
				}
				return newError("argument to `last` must be %s, got %s", ARRAY_OBJ, args[0].Type())
			},
		},
	},
	{
		"rest",
		&Builtin{
			Fn: func(args ...Object) Object {
				if l := len(args); l != 1 {
					return newError("wrong number of arguments. got=%d, want=%d", l, 1)
				}
				switch arg := args[0].(type) {
				case *Array:
					if length := len(arg.Elements); length > 1 {
						elms := make([]Object, length-1, length-1)
						copy(elms, arg.Elements[1:])
						return &Array{Elements: elms}
					}
					return nil
				}
				return newError("argument to `rest` not supported, got %s", args[0].Type())
			},
		},
	},
	{
		"push",
		&Builtin{
			Fn: func(args ...Object) Object {
				if l := len(args); l < 2 {
					return newError("wrong number of arguments. got=%d, want > 1", l)
				}
				switch arg := args[0].(type) {
				case *Array:
					fnArgs := args[1:]
					fnArgsLen := len(fnArgs)
					length := len(arg.Elements)
					newElements := make([]Object, length+fnArgsLen, length+fnArgsLen)
					copy(newElements, arg.Elements)
					for i, a := range fnArgs {
						newElements[length+i] = a
					}
					return &Array{Elements: newElements}
				}
				return newError("argument to `push` must be %s, got %s", ARRAY_OBJ, args[0].Type())
			},
		},
	},
}

func newError(format string, a ...interface{}) *Error {
	return &Error{Message: fmt.Sprintf(format, a...)}
}

func GetBuiltinByName(name string) *Builtin {
	for _, def := range Builtins {
		if def.Name == name {
			return def.Builtin
		}
	}
	return nil
}
