// /home/krylon/go/src/krylisp/interpreter/02_script_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 04. 10. 2017 by Benjamin Walkenhorst
// (c) 2017 Benjamin Walkenhorst
// Time-stamp: <2017-11-07 21:24:26 krylon>

package interpreter

import (
	"krylisp/lexer"
	"krylisp/parser"
	"krylisp/types"
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

	// interp.debug = true
	// defer func() { interp.debug = false }()

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

func TestFactorial(t *testing.T) {
	const scriptPath = "testdata/test002.lisp"
	var (
		l         *lexer.Lexer
		p         = parser.NewParser()
		err       error
		val       value.LispValue
		parseTree interface{}
		program   value.Program
		ok        bool
	)

	if l, err = lexer.NewLexerFile(scriptPath); err != nil {
		t.Errorf("Error creating Lexer for %s: %s",
			scriptPath,
			err.Error())
	} else if parseTree, err = p.Parse(l); err != nil {
		t.Errorf("Error parsing %s: %s",
			scriptPath,
			err.Error())
	} else if program, ok = parseTree.([]value.LispValue); !ok {
		t.Errorf("Unexpected type returned from Parser: %T",
			parseTree)
	} else if val, err = interp.Eval(program); err != nil {
		t.Errorf("Error running %s: %s",
			scriptPath,
			err.Error())
	} else if val.Type() != types.Integer {
		t.Errorf("Unexpected type returned from %s: Expected Number, got %s",
			scriptPath,
			val.Type().String())
	} else if int64(val.(value.IntValue)) != 3628800 {
		t.Errorf("Unexpected result returned from %s: %s",
			scriptPath,
			val.String())
	}

	// 3628800
} // func TestFactorial(t *testing.T)

func TestFactorialBignum(t *testing.T) {
	const scriptPath = "testdata/test003.lisp"
	var (
		l         *lexer.Lexer
		p         = parser.NewParser()
		err       error
		val       value.LispValue
		parseTree interface{}
		program   value.Program
		ok        bool
	)

	// interp.debug = true
	// defer func() { interp.debug = false }()

	if l, err = lexer.NewLexerFile(scriptPath); err != nil {
		t.Errorf("Error creating Lexer for %s: %s",
			scriptPath,
			err.Error())
	} else if parseTree, err = p.Parse(l); err != nil {
		t.Errorf("Error parsing %s: %s",
			scriptPath,
			err.Error())
	} else if program, ok = parseTree.([]value.LispValue); !ok {
		t.Errorf("Unexpected type returned from Parser: %T",
			parseTree)
	} else if val, err = interp.Eval(program); err != nil {
		t.Errorf("Error running %s: %s",
			scriptPath,
			err.Error())
	} else if val.Type() != types.BigInt && val.Type() != types.Integer {
		t.Errorf("Unexpected type returned from %s: Expected Number, got %s",
			scriptPath,
			val.Type().String())
	} else if val.String() != "3628800" {
		t.Errorf("Unexpected result returned from %s: %s",
			scriptPath,
			val.String())
	}

	// 3628800
} // func TestFactorial(t *testing.T)

func TestRegexp(t *testing.T) {
	const scriptPath = "testdata/test004.lisp"
	var (
		l         *lexer.Lexer
		p         = parser.NewParser()
		err       error
		val       value.LispValue
		parseTree interface{}
		program   value.Program
		ok        bool
	)

	// interp.debug = true
	// defer func() { interp.debug = false }()

	if l, err = lexer.NewLexerFile(scriptPath); err != nil {
		t.Errorf("Error creating Lexer for %s: %s",
			scriptPath,
			err.Error())
	} else if parseTree, err = p.Parse(l); err != nil {
		t.Errorf("Error parsing %s: %s",
			scriptPath,
			err.Error())
	} else if program, ok = parseTree.([]value.LispValue); !ok {
		t.Errorf("Unexpected type returned from Parser: %T",
			parseTree)
	} else if val, err = interp.Eval(program); err != nil {
		t.Errorf("Error running %s: %s",
			scriptPath,
			err.Error())
	} else if val.Type() != types.Array {
		t.Errorf("Expected an array, got a %T",
			val)
	}
}
