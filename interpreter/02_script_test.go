// /home/krylon/go/src/krylisp/interpreter/02_script_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 04. 10. 2017 by Benjamin Walkenhorst
// (c) 2017 Benjamin Walkenhorst
// Time-stamp: <2017-10-10 18:42:10 krylon>

package interpreter

import (
	"krylisp/lexer"
	"krylisp/parser"
	"krylisp/value"
	"os"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

// This files contains tests that are a little more elaborate

func TestRunScript001(t *testing.T) {
	type scriptTest struct {
		path          string
		expectedError bool
		expectedValue string
	}

	interp.stdout = os.Stdout
	interp.stderr = os.Stderr
	interp.debug = true

	defer func() { interp.debug = false }()

	var testCases = []scriptTest{
		scriptTest{
			path:          "testdata/test001.lisp",
			expectedValue: "(2 4 6)",
		},
	}

	for idx, test := range testCases {
		var (
			parsed interface{}
			prog   value.Program
			ok     bool
			p      = parser.NewParser()
			l      *lexer.Lexer //= lexer.NewLexerFile(test.path)
			err    error
			val    value.LispValue
			res    string
		)

		if l, err = lexer.NewLexerFile(test.path); err != nil {
			t.Errorf("Error creating Lexer for %s: %s",
				test.path,
				err.Error())
		} else if parsed, err = p.Parse(l); err != nil {
			t.Errorf("Error parsing test input %s: %s",
				test.path,
				err.Error())
		} else if prog, ok = parsed.([]value.LispValue); !ok {
			t.Errorf("Parser returned unexpected data type: %T",
				parsed)
		}

		spew.Printf("Run test program #%d: %#v\n",
			idx+1,
			prog)

		if val, err = interp.Eval(prog); err != nil {
			if !test.expectedError {
				t.Errorf("Error evaluating program %s: %s",
					test.path,
					err.Error())
			}
		} else if val == nil {
			t.Errorf("Eval(%s) returned nil",
				test.path)
		} else if res = val.String(); res != test.expectedValue {
			t.Errorf(`Eval(%s) returned unexpected result:
Expected: %s
Actual:   %s
`,
				test.path,
				test.expectedValue,
				res)
		}
	}
} // func TestRunScript001(t *testing.T)
