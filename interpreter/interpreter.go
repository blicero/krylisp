// /home/krylon/go/src/github.com/blicero/krylisp/interpreter/interpreter.go
// -*- mode: go; coding: utf-8; -*-
// Created on 15. 02. 2025 by Benjamin Walkenhorst
// (c) 2025 Benjamin Walkenhorst
// Time-stamp: <2025-02-23 15:53:00 krylon>

// Package interpreter implements the traversal and evaluation of ASTs.
package interpreter

import (
	"fmt"

	"github.com/blicero/krylib"
	"github.com/blicero/krylisp/parser"
	"github.com/blicero/krylisp/types"
)

// Environment is a set of bindings of symbols to values.
type Environment struct {
	Parent   *Environment
	Bindings map[parser.Symbol]parser.LispValue
}

// Lookup attempts to look up the binding to a Symbol. If the Symbol is
// not found in the current Environment, it recursively tries the parent
// Environments until a binding is found or the chain of environments
// is exhausted.
func (e *Environment) Lookup(key parser.Symbol) (parser.LispValue, bool) {
	var (
		ok  bool
		val parser.LispValue
	)

	if val, ok = e.Bindings[key]; ok {
		return val, true
	}

	if e.Parent != nil {
		return e.Parent.Lookup(key)
	}

	return nil, false
} // func (e *Environment) Lookup(key parser.Symbol) (parser.LispValue, error)

// Interpreter implements the evaluation of Lisp expressions.
type Interpreter struct {
	Env           *Environment
	Debug         bool
	GensymCounter int
}

// Eval is the heart of the interpreter.
func (in *Interpreter) Eval(v parser.LispValue) (parser.LispValue, error) {
	// var (
	// 	err    error
	// 	result parser.LispValue
	// )

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
		if real.Car == nil && real.Cdr == nil {
			return sym("nil"), nil
		} else if real.Car.Type() == types.Symbol {
			if isSpecial(real.Car) {
				return in.evalSpecial(real)
			}
		}
	default:
		return nil, fmt.Errorf("Unsupported type %t", real)
	}

	return nil, krylib.ErrNotImplemented
} // func (in *Interpreter) Eval(v parser.LispValue) (parser.LispValue, error)

func (in *Interpreter) evalSpecial(l parser.List) (parser.LispValue, error) {
	var (
		err error
		res parser.LispValue
	)

	switch l.Car.String() {
	case "if":
	}
} // func (in *Interpreter) evalSpecial(l parser.List) (parser.LispValue, error)
