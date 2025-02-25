// /home/krylon/go/src/github.com/blicero/krylisp/interpreter/interpreter.go
// -*- mode: go; coding: utf-8; -*-
// Created on 15. 02. 2025 by Benjamin Walkenhorst
// (c) 2025 Benjamin Walkenhorst
// Time-stamp: <2025-02-25 14:37:02 krylon>

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
)

// ErrEval indicates some non-specified problem during evaluation
var ErrEval = errors.New("Error evaluating expression")

// ErrType indicates an invalid/unexpected type was encountered in an expression.
var ErrType = errors.New("Invalid type in expression")

// Interpreter implements the evaluation of Lisp expressions.
type Interpreter struct {
	Env           *Environment
	Debug         bool
	GensymCounter int
	log           *log.Logger
}

// MakeInterpreter creates a fresh Interpreter. If the given Environment is nil,
// a fresh one is created as well.
func MakeInterpreter(env *Environment, dbg bool) (*Interpreter, error) {
	var (
		err error
		in  = &Interpreter{
			Debug: dbg,
		}
	)

	if env != nil {
		in.Env = env
	} else {
		in.Env = MakeEnvironment(nil)
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
	in.log.Printf("[DEBUG] Eval %T %s\n",
		v,
		v)

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

	in.log.Printf("[DEBUG] Evaluate List %s\n",
		l)

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

		var fn = &Function{
			name:      name.(parser.Symbol).Sym,
			docString: docString,
			argList:   argList,
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
		err  error
		ok   bool
		head parser.LispValue
		fn   *Function
	)

	if cnt := l.Length(); cnt < 1 {
		return sym("nil"), nil
	}

	head = l.Car

	switch v := head.(type) {
	case parser.Symbol:

	}
} // func (in *Interpreter) evalList(l parser.List) (parser.LispValue, error)
