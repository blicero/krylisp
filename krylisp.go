// /home/krylon/go/src/krylisp/krylisp.go
// -*- mode: go; coding: utf-8; -*-
// Created on 23. 09. 2017 by Benjamin Walkenhorst
// (c) 2017 Benjamin Walkenhorst
// Time-stamp: <2024-05-23 23:14:25 krylon>

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/blicero/krylisp/interpreter"
	"github.com/blicero/krylisp/lexer"
	"github.com/blicero/krylisp/parser"
	"github.com/blicero/krylisp/repl"
	"github.com/blicero/krylisp/value"
)

func main() {
	var (
		r     *repl.Repl
		inter *interpreter.Interpreter
		err   error
		batch bool
		debug bool
	)

	flag.BoolVar(
		&batch,
		"batch",
		false,
		"If given, runs the interpreter in batch mode instead of starting a REPL",
	)

	flag.BoolVar(
		&debug,
		"debug",
		false,
		"If given, run the interpreter in debug mode",
	)

	flag.Parse()

	inter = interpreter.New(debug)

	for _, filename := range flag.Args() {
		var rawProg interface{}
		var prog value.Program
		var ok bool
		var lex *lexer.Lexer //= lexer.NewLexerFile(filename)
		var pars = parser.NewParser()

		if lex, err = lexer.NewLexerFile(filename); err != nil {
			fmt.Printf("Error creating Lexer for file %s: %s\n",
				filename,
				err.Error())
			os.Exit(1)
		} else if rawProg, err = pars.Parse(lex); err != nil {
			fmt.Printf("Error parsing %s: %s\n",
				filename,
				err.Error())
			os.Exit(1)
		} else if prog, ok = rawProg.([]value.LispValue); !ok {
			fmt.Printf("Cannot convert parse result to value.Program: Wrong type %T\n",
				rawProg)
			os.Exit(1)
		} else if _, err = inter.Eval(prog); err != nil {
			fmt.Printf("Error evaluating file %s: %s\n",
				filename,
				err.Error())
			os.Exit(2)
		}
	}

	if batch {
		os.Exit(0)
	} else if r, err = repl.NewForInterpreter(inter); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	r.Run()
} // func main()
