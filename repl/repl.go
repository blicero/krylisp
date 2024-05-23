// /home/krylon/go/src/krylisp/repl/repl.go
// -*- mode: go; coding: utf-8; -*-
// Created on 23. 09. 2017 by Benjamin Walkenhorst
// (c) 2017 Benjamin Walkenhorst
// Time-stamp: <2024-05-23 09:46:17 krylon>

// Package repl implements the REPL. Consider it a work in progress.
package repl

import (
	"bytes"
	"fmt"

	"github.com/blicero/krylisp/interpreter"
	"github.com/blicero/krylisp/lexer"
	"github.com/blicero/krylisp/parser"
	"github.com/blicero/krylisp/value"

	"github.com/chzyer/readline"
)

// Debug, if true, makes the interpreter emit additional messages
// Prompt is the string the repl emits before waiting for user input.
// Greeting is the message that is emitted once, at startup.
const (
	Debug    = true
	Prompt   = "> "
	Greeting = `kryLisp - a primitive Lisp written in Go
(c) 2017 Benjamin Walkenhorst <walkenhorst.benjamin@gmail.com>

Type :h for hep
`
)

// Repl rocks!
type Repl struct {
	eval *interpreter.Interpreter
	// pars *parser.Parser
	// lex  *lexer.Lexer
	rl   *readline.Instance
	rbuf bytes.Buffer
}

// New creates a new Repl.
func New() (*Repl, error) {
	var err error
	var r = &Repl{
		eval: interpreter.New(Debug),
		//rl:   readline.New(Prompt),
	}

	if r.rl, err = readline.New(Prompt); err != nil {
		return nil, err
	}
	// r.lex = lexer.NewLexer(r.rbuf)
	// r.pars = parser.NewParser(r.lex)

	return r, nil
} // func New() (*Repl, error)

// NewForInterpreter creates a new REPL with an existing interpreter
// instance.
func NewForInterpreter(i *interpreter.Interpreter) (*Repl, error) {
	var err error
	var r = &Repl{
		eval: i,
	}

	if r.rl, err = readline.New(Prompt); err != nil {
		return nil, err
	}

	return r, nil
} // func NewForInterpreter(i *interpreter.Interpreter) (*Repl, error)

// Run executes the REPL's main loop.
func (r *Repl) Run() {
	var input []byte
	var err error
	var prog value.Program
	var ok bool
	var result value.LispValue
	var l *lexer.Lexer
	var p *parser.Parser
	var tree interface{}

INPUT:
	for {
		if input, err = r.rl.ReadSlice(); err != nil {
			fmt.Printf("Error reading input: %s\n",
				err.Error())
			continue INPUT
		}

		r.rbuf.Write(input)
		l = lexer.NewLexer(r.rbuf.Bytes())
		p = parser.NewParser()

		if tree, err = p.Parse(l); err != nil {
			fmt.Printf("Error parsing input: %s\n",
				err.Error())
			continue
		} else if prog, ok = tree.([]value.LispValue); !ok {
			fmt.Printf("CANTHAPPEN: Parser returned a %T\n",
				tree)
			continue INPUT
		} else {
		EXPRESSION:
			for _, exp := range prog {
				if result, err = r.eval.Eval(exp); err != nil {
					fmt.Printf("Error evaluating %s: %s\n",
						exp.String(),
						err.Error())
					break EXPRESSION
				} else {
					fmt.Println(result.String())
				}
			}
		}
	}
} // func (r *Repl) Run()
