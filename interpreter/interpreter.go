// /home/krylon/go/src/github.com/blicero/krylisp/interpreter/interpreter.go
// -*- mode: go; coding: utf-8; -*-
// Created on 15. 02. 2025 by Benjamin Walkenhorst
// (c) 2025 Benjamin Walkenhorst
// Time-stamp: <2025-02-15 15:23:52 krylon>

// Package interpreter implements the traversal and evaluation of ASTs.
package interpreter

import "github.com/blicero/krylisp/parser"

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
