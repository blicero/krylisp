// /home/krylon/go/src/github.com/blicero/krylisp/interpreter/interpreter.go
// -*- mode: go; coding: utf-8; -*-
// Created on 15. 02. 2025 by Benjamin Walkenhorst
// (c) 2025 Benjamin Walkenhorst
// Time-stamp: <2025-02-17 18:02:32 krylon>

// Package interpreter implements the traversal and evaluation of ASTs.
package interpreter

import (
	"github.com/blicero/krylib"
	"github.com/blicero/krylisp/parser"
)

// Environment is a set of bindings of symbols to values.
type Environment struct {
	Parent   *Environment
	Bindings map[parser.Symbol]parser.LispValue
}

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

type Interpreter struct {
	Env           *Environment
	Debug         bool
	GensymCounter int
}

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
	}

	return nil, krylib.ErrNotImplemented
} // func (in *Interpreter) Eval(v parser.LispValue) (parser.LispValue, error)
