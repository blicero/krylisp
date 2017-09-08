// /home/krylon/go/src/krylisp/interpreter/interpreter.go
// -*- mode: go; coding: utf-8; -*-
// Created on 08. 09. 2017 by Benjamin Walkenhorst
// (c) 2017 Benjamin Walkenhorst
// Time-stamp: <2017-09-08 21:32:43 krylon>

// Package interpreter implements the actual interpreter.
// The first time 'round, the interpreter is simply going to walk the parse tree
// recursively and evaluate each node.
//
// Later on, I might make a try at performing some basic optimizations or
// generating byte code?
package interpreter

import "krylisp/value"

// Interpreter is my first shot at a tree-walking interpreter for my
// toy Lisp dialect.
type Interpreter struct {
	program []value.LispValue
}
