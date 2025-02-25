// /home/krylon/go/src/github.com/blicero/krylisp/interpreter/environment.go
// -*- mode: go; coding: utf-8; -*-
// Created on 25. 02. 2025 by Benjamin Walkenhorst
// (c) 2025 Benjamin Walkenhorst
// Time-stamp: <2025-02-25 14:36:22 krylon>

package interpreter

import "github.com/blicero/krylisp/parser"

// Environment is a set of bindings of symbols to values.
type Environment struct {
	Parent   *Environment
	Bindings map[parser.Symbol]parser.LispValue
}

// MakeEnvironment creates a fresh Environment with the given parent.
func MakeEnvironment(par *Environment) *Environment {
	var env = &Environment{
		Parent:   par,
		Bindings: make(map[parser.Symbol]parser.LispValue),
	}

	return env
} // func MakeEnvironment(par *Environment) *Environment

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

// Set sets the binding for the given Symbol to the given value. If a binding
// for that symbol already exists, it is replaced.
func (e *Environment) Set(key parser.Symbol, val parser.LispValue) {
	e.Bindings[key] = val
} // func (e *Environment) Set(key parser.Symbol, val parser.LispValue)

// Delete removes the binding for the given symbol from the current scope.
// If no binding for the symbol exists, it is a no-op.
// If a binding for the symbol exists in the Environment's Parent(s), those are
// not affected.
func (e *Environment) Delete(key parser.Symbol) {
	delete(e.Bindings, key)
}
