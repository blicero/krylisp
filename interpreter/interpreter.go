// /home/krylon/go/src/github.com/blicero/krylisp/interpreter/interpreter.go
// -*- mode: go; coding: utf-8; -*-
// Created on 15. 02. 2025 by Benjamin Walkenhorst
// (c) 2025 Benjamin Walkenhorst
// Time-stamp: <2025-03-01 15:17:38 krylon>

// Package interpreter implements the traversal and evaluation of ASTs.
package interpreter

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/blicero/krylib"
	"github.com/blicero/krylisp/common"
	"github.com/blicero/krylisp/logdomain"
	"github.com/blicero/krylisp/parser"
	"github.com/blicero/krylisp/types"
	"github.com/davecgh/go-spew/spew"
)

// ErrEval indicates some non-specified problem during evaluation
var ErrEval = errors.New("Error evaluating expression")

// ErrType indicates an invalid/unexpected type was encountered in an expression.
var ErrType = errors.New("Invalid type in expression")

// Interpreter implements the evaluation of Lisp expressions.
type Interpreter struct {
	Env           *environment
	Debug         bool
	GensymCounter int
	log           *log.Logger
}

// MakeInterpreter creates a fresh Interpreter. If the given Environment is nil,
// a fresh one is created as well.
func MakeInterpreter(env *environment, dbg bool) (*Interpreter, error) {
	var (
		err error
		in  = &Interpreter{
			Debug: dbg,
		}
	)

	if env != nil {
		in.Env = env
	} else {
		in.Env = makeEnv()
	}

	if in.log, err = common.GetLogger(logdomain.Interpreter); err != nil {
		return nil, err
	}

	return in, nil
} // func MakeInterpreter(env *Environment, dbg bool) (*Interpreter, error)

// TODO I think I need to pass in the environment as an argument to Eval, lest
//      handling function calls and so forth becomes very tedious.
//      Unless I make the Environment itself handle that - instead of a Pointer
//      to a parent Environment it could have a stack of Binding maps.

// Eval is the heart of the interpreter.
func (in *Interpreter) Eval(v parser.LispValue) (parser.LispValue, error) {
	in.log.Printf("[DEBUG] Eval %T\n%s\n",
		v,
		spew.Sdump(v))

	switch real := v.(type) {
	case parser.Symbol:
		switch real.Sym {
		case "T":
			return real, nil
		case "NIL":
			return real, nil
		default:
			if real.IsKeyword() {
				return real, nil
			}

			if val, ok := in.Env.Lookup(real); ok {
				return val, nil
			}
		}
	case parser.Integer:
		return real, nil
	case parser.String:
		return real, nil
	case parser.List:
		in.log.Printf("[DEBUG] Head of list to be evaluated is %T %s, length of List is %d\n",
			real.Car,
			real.Car,
			real.Length())
		if real.Car == nil && real.Cdr == nil {
			return sym("nil"), nil
		} else if real.Car.Type() == types.Symbol {
			if isSpecial(real.Car) {
				return in.evalSpecial(real)
			}

			in.log.Printf("[DEBUG] %s is not special.\n",
				real.Car)

			return in.evalList(real)
		}
		return nil, fmt.Errorf("Unexpected type for head of list (expected symbol): %s",
			real.Car.Type())
	default:
		return nil, fmt.Errorf("Unsupported type %t", real)
	}

	return nil, krylib.ErrNotImplemented
} // func (in *Interpreter) Eval(v parser.LispValue) (parser.LispValue, error)

func (in *Interpreter) evalSpecial(l parser.List) (parser.LispValue, error) {
	var (
		err error
		ok  bool
	)

	in.log.Printf("[DEBUG] Evaluate special form %s\n%s\n",
		l,
		spew.Sdump(l))

	switch form := strings.ToUpper(l.Car.String()); form {
	case "IF":
		in.log.Println("[TRACE] Eval IF clause")
		if x := l.Length(); x != 4 {
			return nil, fmt.Errorf("if-clause needs 4 elements, not %d",
				x)
		}

		var (
			cond, ifBranch, elseBranch parser.LispValue
			val, branch                parser.LispValue
		)

		cond, _ = l.At(1)
		ifBranch, _ = l.At(2)
		elseBranch, _ = l.At(3)

		if val, err = in.Eval(cond); err != nil {
			return nil, err
		} else if asBool(val) {
			branch = ifBranch
		} else {
			branch = elseBranch
		}

		return in.Eval(branch)
	case "+":
		var (
			cons       = l.Cdr
			acc  int64 = 0
		)

		for cons != nil {
			if cons.Car == nil {
				in.log.Println("[ERROR] cons.Car is nil")
				return nil, ErrEval
			} else if cons.Car.Type() != types.Integer {
				return nil, ErrType
			}

			acc += cons.Car.(parser.Integer).Int
			cons = cons.Cdr
		}

		return parser.Integer{Int: acc}, nil
	case "*":
		var (
			cons       = l.Cdr
			cnt        = 1
			acc  int64 = 1
		)

		for cons != nil {
			var res parser.LispValue

			if cons.Car == nil {
				in.log.Println("[ERROR] cons.Car is nil")
				return nil, ErrEval
			} else if res, err = in.Eval(cons.Car); err != nil {
				return nil, fmt.Errorf("Error evaluating argument #%d (%s): %s",
					cnt,
					cons.Car,
					err.Error())
			} else if res.Type() != types.Integer {
				return nil, fmt.Errorf("Unexpected type: %s (expect Integer)",
					res.Type())
			}

			acc *= res.(parser.Integer).Int
			cons = cons.Cdr
		}

		return parser.Integer{Int: acc}, nil
	case "<":
		var (
			res  parser.LispValue
			n    parser.Integer
			cons = l.Cdr
		)

		in.log.Printf("[TRACE] Evaluate less-than (<): %s\n",
			spew.Sdump(l.Cdr))

		if res, err = in.Eval(cons.Cdr.Car); err != nil {
			var msg = fmt.Sprintf("Error evaluating %q: %s",
				cons.Cdr.Car,
				err.Error())
			in.log.Printf("[ERROR] %s\n", msg)
			return nil, errors.New(msg)
		} else if n, ok = res.(parser.Integer); !ok {
			return nil, fmt.Errorf("Invalid type for <: %s (expected Integer)",
				cons.Car.Type())
		}

		cons = cons.Cdr

		for cons != nil {
			var num parser.Integer
			if res, err = in.Eval(cons.Car); err != nil {
				var msg = fmt.Sprintf("Error evaluating %q: %s",
					cons.Cdr.Car,
					err.Error())
				in.log.Printf("[ERROR] %s\n", msg)
				return nil, errors.New(msg)
			} else if num, ok = res.(parser.Integer); !ok {
				return nil, fmt.Errorf("Unexpected type for <: %s (expected Integer)",
					cons.Car.Type())
			}

			if num.Int < n.Int {
				return sym("nil"), nil
			}

			cons = cons.Cdr
		}

		return sym("t"), nil
	case "NULL":
		if cnt := l.Length(); cnt != 2 {
			return nil, fmt.Errorf("Wrong number of arguments for NULL: %d (expect 0)",
				cnt)
		}

		var arg = l.Cdr.Car

		if asBool(arg) {
			return sym("nil"), nil
		}

		return sym("t"), nil
	case "DEFUN":
		if cnt := l.Length(); cnt < 3 {
			return nil, fmt.Errorf("Wrong number of arguments to DEFUN: %d (expect >= 3)",
				cnt)
		}

		var (
			v, name   parser.LispValue
			argList   parser.List
			args      []parser.LispValue
			docString string
			body      *parser.ConsCell
		)

		if argList, ok = l.Cdr.Cdr.Car.(parser.List); !ok {
			return nil, fmt.Errorf("Second argument to defun must be a List of arguments, not a %T",
				l.Cdr.Cdr.Car)
		}

		v, _ = l.At(2)

		if v.Type() == types.String {
			docString = v.(parser.String).Str
			body = l.Cdr.Cdr.Cdr
		} else {
			body = l.Cdr.Cdr
		}

		name = l.Cdr.Car

		if t := name.Type(); t != types.Symbol {
			return nil, fmt.Errorf("First argument to DEFUN must be a symbol, not a %s",
				t)
		}

		args = make([]parser.LispValue, 0, 4)

		if argList.Car != nil {
			args = append(args, argList.Car)
			var a = argList.Cdr

			for a != nil {
				args = append(args, a.Car)
				a = a.Cdr
			}
		}

		var fn = &Function{
			name:      name.(parser.Symbol).Sym,
			docString: docString,
			argList:   args,
			body:      body,
		}

		in.Env.Set(name.(parser.Symbol), fn)

		return name, nil
	default:
		var msg = fmt.Sprintf("Special form %s is not implemented, yet",
			form)
		in.log.Printf("[ERROR] %s\n", msg)
		return nil, errors.New(msg)
	}
} // func (in *Interpreter) evalSpecial(l parser.List) (parser.LispValue, error)

func (in *Interpreter) evalList(l parser.List) (parser.LispValue, error) {
	var (
		ok        bool
		head, val parser.LispValue
		fn        *Function
		cnt       int
	)

	if cnt = l.Length(); cnt < 1 {
		return sym("nil"), nil
	}

	head = l.Car

	switch v := head.(type) {
	case parser.Symbol:
		if val, ok = in.Env.Lookup(v); !ok {
			return nil, fmt.Errorf("No binding was found for %s",
				v)
		} else if fn, ok = val.(*Function); !ok {
			return nil, fmt.Errorf("Type error: Binding for %s is not a function, but a %s (%s)",
				v,
				val.Type(),
				val)
		}

		in.log.Printf("[TRACE] Evaluating call to %s\n",
			fn.name)
	case *Function:
		fn = v
	default:
		return nil, fmt.Errorf("Head of list must be a Symbol that resolves to a function or a Function object, not a %T", v)
	}

	// Next, we need to evaluate the arguments to the function call, bind
	// them to the argument list of the function, push those to the
	// Environment stack, and evaluate the function body.

	if cnt != len(fn.argList)+1 {
		return nil, fmt.Errorf("Incorrect number of arguments in function call: want %d, got %d",
			cnt-1,
			len(fn.argList))
	}

	var (
		cell = l.Cdr
		args = make([]parser.LispValue, 0, len(fn.argList))
	)

	for cell != nil {
		var (
			err error
			res parser.LispValue
		)

		in.log.Printf("[TRACE] Evaluate argument: %s\n",
			cell.Car)

		if res, err = in.Eval(cell.Car); err != nil {
			in.log.Printf("[ERROR] Error evaluating %q: %s\n",
				cell.Car,
				err.Error())
			return nil, err
		}

		args = append(args, res)
		cell = cell.Cdr
	}

	in.Env.Push()

	for i, s := range fn.argList {
		in.Env.Set(s.(parser.Symbol), args[i])
	}

	defer in.Env.Pop()

	in.log.Printf("[TRACE] Evaluate function body:\n%s\n",
		spew.Sdump(fn.body))

	var (
		err  error
		res  parser.LispValue
		body = fn.body
	)

	for body != nil {
		if res, err = in.Eval(body.Car); err != nil {
			in.log.Printf("[ERROR] Error evaluating expression %s: %s\n",
				body.Car,
				err.Error())
			return nil, err
		}

		body = body.Cdr
	}

	return res, nil
} // func (in *Interpreter) evalList(l parser.List) (parser.LispValue, error)
