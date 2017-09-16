// /home/krylon/go/src/krylisp/interpreter/environment.go
// -*- mode: go; coding: utf-8; -*-
// Created on 08. 09. 2017 by Benjamin Walkenhorst
// (c) 2017 Benjamin Walkenhorst
// Time-stamp: <2017-09-15 21:05:18 krylon>

package value

// Environment represents the set of bindings reachable in the current scope.
// And environment has an (optional) parent which is used for lexical scoping.
type Environment struct {
	Data   map[string]LispValue
	Parent *Environment
	Level  int
}

// NewEnvironment create a fresh environment with the given parent
// (which may be nil).
func NewEnvironment(parent *Environment) *Environment {
	var env = &Environment{
		Data:   make(map[string]LispValue),
		Parent: parent,
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
