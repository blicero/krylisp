// /home/krylon/go/src/github.com/blicero/krylisp/interpreter/environment.go
// -*- mode: go; coding: utf-8; -*-
// Created on 25. 02. 2025 by Benjamin Walkenhorst
// (c) 2025 Benjamin Walkenhorst
// Time-stamp: <2025-02-25 15:28:58 krylon>

package interpreter

import "github.com/blicero/krylisp/parser"

type scope struct {
	bindings map[parser.Symbol]parser.LispValue
	parent   *scope
}

// environment is a set of bindings of symbols to values.
type environment struct {
	scope *scope
}

// makeEnv creates a fresh Environment with the given parent.
func makeEnv() *environment {
	var env = &environment{
		scope: &scope{
			bindings: make(map[parser.Symbol]parser.LispValue),
		},
	}

	return env
} // func MakeEnvironment() *Environment

func (e *environment) Push() {
	var s = &scope{
		bindings: make(map[parser.Symbol]parser.LispValue),
		parent:   e.scope,
	}

	e.scope = s
} // func (e *environment) Push()

func (e *environment) Pop() {
	if e.scope.parent == nil {
		panic("Cannot pop Scope, no parent!")
	}

	e.scope = e.scope.parent
} // func (e *environment) Pop()

// Lookup attempts to look up the binding to a Symbol. If the Symbol is
// not found in the current Environment, it recursively tries the parent
// Environments until a binding is found or the chain of environments
// is exhausted.
func (e *environment) Lookup(key parser.Symbol) (parser.LispValue, bool) {
	var (
		ok  bool
		val parser.LispValue
	)

	if val, ok = e.scope.bindings[key]; ok {
		return val, ok
	}

	var s = e.scope.parent

	for s != nil {
		if val, ok = s.bindings[key]; ok {
			return val, ok
		}

		s = s.parent
	}

	return nil, false
} // func (e *Environment) Lookup(key parser.Symbol) (parser.LispValue, error)

// Set sets the binding for the given Symbol to the given value. If a binding
// for that symbol already exists, it is replaced.
func (e *environment) Set(key parser.Symbol, val parser.LispValue) {
	// FIXME When setting a key, I need to check first if the binding
	//       appears in one of the parent scopes.
	e.scope.bindings[key] = val
} // func (e *Environment) Set(key parser.Symbol, val parser.LispValue)

// Delete removes the binding for the given symbol from the current scope.
// If no binding for the symbol exists, it is a no-op.
// If a binding for the symbol exists in the Environment's Parent(s), those are
// not affected.
func (e *environment) Delete(key parser.Symbol) {
	delete(e.scope.bindings, key)
}
