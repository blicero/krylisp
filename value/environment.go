// /home/krylon/go/src/krylisp/interpreter/environment.go
// -*- mode: go; coding: utf-8; -*-
// Created on 08. 09. 2017 by Benjamin Walkenhorst
// (c) 2017 Benjamin Walkenhorst
// Time-stamp: <2017-12-13 16:41:44 krylon>
//
// Montag, 11. 12. 2017, 18:10
// I am currently attempting to add macros to the language, and I just realize
// that rather than have the Interpreter have multiple environments for data,
// functions, and (soon) macros, I might want to put all of the in the
// environment itself.
// This would also fix the slightly embarrassing bug that currently all functions
// live in a global scope. Nested functions do not work.
//
// Okay, so, if I want to have a unified environment that can hold data, functions,
// and macros, and have uniform lookup semantics, I am going to need a few more
// methods.
//
// Montag, 11. 12. 2017, 21:53
// Okay, I did not think that all the way through, I think.
// If I wanted to make the Functions map map to actual pointers-to-Function,
// then what about GoFunctions?
// You know the answer - make Function an interface with (currently) two
// implementations:
// GoFunction and LispFunction.

package value

import (
	"fmt"
	"io"
)

// Environment represents the set of bindings reachable in the current scope.
// Environments can be nested, which is used in lexical scoping.
type Environment struct {
	Data      map[string]LispValue
	Functions map[string]Function
	Macros    map[string]*Macro
	Parent    *Environment
	Level     int
}

// NewEnvironment create a fresh environment with the given parent
// (which may be nil).
func NewEnvironment(parent *Environment) *Environment {
	var env = &Environment{
		Data:      make(map[string]LispValue),
		Functions: make(map[string]Function),
		Macros:    make(map[string]*Macro),
		Parent:    parent,
	}

	if parent != nil {
		env.Level = parent.Level + 1
	}

	return env
} // func NewEnvironment(parent *Environment) *Environment

// Get looks up a key in the current environment. If the key is not found, it
// works its way up through the parent environments (if there are any), until
// the key is found or the chain of Environments is exhausted.
// The second return value is true, if the given key was found, to distinguish
// the case where a key is found but with a value of nil.
func (env *Environment) Get(key string) (LispValue, bool) {
	var val LispValue
	var ok bool
	for e := env; e != nil; e = e.Parent {
		if val, ok = e.Data[key]; ok {
			return val, ok
		}
	}

	return nil, false
} // func (env *Environment) Get(key string) (LispValue, bool)

// Set sets a key in the current environment.
// If the key is found in the Environment or any of its parent Environments,
// that binding is updated. A new binding in the current scope is created
// only if the binding does not already exist in the lookup chain.
func (env *Environment) Set(key string, val LispValue) *Environment {
	for e := env; e != nil; e = e.Parent {
		if _, ok := e.Data[key]; ok {
			e.Data[key] = val
			return env
		}
	}

	env.Data[key] = val
	return env
} // func (env *Environment) Set(key string, val LispValue) *Environment

// Ins sets the binding for the given key to the given
// If a binding for the given key already exists higher up in the lookup chain,
// that binding is masked for the duration of the current scope (or until this
// binding is deleted from the current scope.
func (env *Environment) Ins(key string, val LispValue) *Environment {
	env.Data[key] = val
	return env
} // func (env *Environment) Ins(key string, val LispValue) *Environment

// Del removes a binding from the current scope.
// It is _not_ an error to delete a binding that does not exist in the
// current scope - the end result is the same, an environment without the
// binding.
// If bindings for the given key exist higher up in the lookup chain, they
// are *not* affected.
func (env *Environment) Del(key string) *Environment {
	delete(env.Data, key)
	return env
} // func (env *Environment) Del(key string) *Environment

// InsMultiple sets multiple bindings at once. Like Ins, it sets the bindings
// in the current environment, masking any bindings that might exist higher up
// in the lookup chain.
func (env *Environment) InsMultiple(data map[string]LispValue) *Environment {
	for key, val := range data {
		env.Data[key] = val
	}

	return env
} // func (env *Environment) InsMultiple(data map[string]LispValue) *Environment

// SetMultiple updates several bindings at once. Like Set, it first looks for
// each binding in the lookup chain and updates the original binding, and it only
// creates a new binding if the key has not been found higher up in the lookup
// chain.
//
// (In fact, this method currently calls Set for each binding in the argument map.)
func (env *Environment) SetMultiple(data map[string]LispValue) *Environment {
	for k, v := range data {
		env.Set(k, v)
	}

	return env
} // func (env *Environment) SetMultiple(data map[string]LispValue) *Environment

// FIXME Now that we keep data, functions AND macros in the Environment, I need
//       to dump all three!

// Dump writes a string representation of the environment to the given io.Writer
func (env *Environment) Dump(out io.Writer) {
	if out == nil {
		fmt.Println("Attempt to Dump Environment to nil Writer")
		return
	} else if env == nil {
		fmt.Println("Attempt to dump nil environment")
		return
	}

	fmt.Fprintf(out, "Environment Level %d {\n\tData: {\n", env.Level)

	for sym, val := range env.Data {
		var vstr string
		if val == nil {
			vstr = "[nil]"
		} else {
			vstr = val.String()
		}
		fmt.Fprintf(out, "\n\t%s => %s",
			sym,
			vstr)
	}

	fmt.Fprintf(out, "}\n\n\tFunctions: {\n")

	for sym, val := range env.Functions {
		var vstr string
		if val == nil {
			vstr = "[nil]"
		} else {
			vstr = val.String()
		}
		fmt.Fprintf(out, "\n\t%s => %s",
			sym,
			vstr)
	}

	fmt.Fprintf(out, "}\n\n\tMacros: {\n")

	for sym, val := range env.Macros {
		var vstr string
		if val == nil {
			vstr = "[nil]"
		} else {
			vstr = val.String()
		}

		fmt.Fprintf(out, "\n\t%s => %s",
			sym,
			vstr)
	}

	fmt.Fprintln(out, "\n}")

	if env.Parent != nil {
		env.Parent.Dump(out)
	}
} // func (env *Environment) Dump(out io.Writer)

// In order to store data, functions, and macros in a single environment, I need
// to effectively replicate the above API for functions and macros.
// Except I can remove the second return value in the lookup methods. That was
// meant to allow one to differentiate between a nonexisting binding and an
// existing binding that happens to have a value of nil.

// GetFn looks up a function in the environment. If it is not found in the
// current envrionment, it works it way up through the entire chain of
// parent environments until either a binding is found or the chain is exhausted.
func (env *Environment) GetFn(key string) Function {
	var (
		fn Function
		ok bool
	)

	for e := env; e != nil; e = e.Parent {
		if fn, ok = e.Functions[key]; ok {
			return fn
		}
	}

	return nil
} // func (env *Environment) GetFn(key string) Function

// InsFn insert a function into the current environment. If a function of the
// same name already exists in an outer environment, it is masked temporarily.
//
// This method returns the receiver.
func (env *Environment) InsFn(key string, fn Function) *Environment {
	env.Functions[key] = fn
	return env
} // func (env *Environment) InsFn(key string, fn Function) *Environment

// SetFn sets a function in the current environment. If a binding for the function
// does not exist in the current environment, this method works its way up through
// all the parent environments. If it finds a binding for the given key, it replaces
// it.
// Otherwise, a new binding is created in the current environment.
//
// This method returns the receiver.
func (env *Environment) SetFn(key string, fn Function) *Environment {
	for e := env; e != nil; e = e.Parent {
		if _, ok := e.Functions[key]; ok {
			e.Functions[key] = fn
			return env
		}
	}

	env.Functions[key] = fn
	return env
} // func (env *Environment) SetFn(key string, fn Function) *Environment

// GetMacro is like GetFn for Macros
func (env *Environment) GetMacro(key string) *Macro {
	var (
		m *Macro
	)

	for e := env; e != nil; e = e.Parent {
		if m = e.Macros[key]; m != nil {
			return m
		}
	}

	return nil
} // func (env *Environment) GetMacro(key String) *Macro

// InsMacro inserts a macro into the current lexical envrinoment,
// temporarily masking any macros of the name in outer scope.
func (env *Environment) InsMacro(key string, m *Macro) *Environment {
	env.Macros[key] = m
	return env
} // func (env *Environment) InsMacro(key string, m *Macro) *Environment

// SetMacro sets a macro in the current environment. If a binding for the macro
// does not exist in the current environment, this method works its way up through
// all the parent environments. If it finds a binding for the given key, it replaces
// it.
// Otherwise, a new binding is created in the current environment.
//
// This method returns the receiver.
func (env *Environment) SetMacro(key string, m *Macro) *Environment {
	var ok bool

	for e := env; e != nil; e = e.Parent {
		if _, ok = e.Macros[key]; ok {
			e.Macros[key] = m
			return env
		}
	}

	env.Macros[key] = m
	return env
} // func (env *Environment) SetMacro(key string, m *Macro) *Environment
