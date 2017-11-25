// /home/krylon/go/src/krylisp/interpreter/02_script_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 04. 10. 2017 by Benjamin Walkenhorst
// (c) 2017 Benjamin Walkenhorst
// Time-stamp: <2017-11-25 14:54:03 krylon>

package interpreter

import (
	"fmt"
	"io/ioutil"
	"krylisp/lexer"
	"krylisp/parser"
	"krylisp/types"
	"krylisp/value"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
)

// This files contains tests that are a little more elaborate

func TestRunScript001(t *testing.T) {
	type scriptTest struct {
		path          string
		expectedError bool
		expectedValue string
	}

	var interp = freshInterpreter()

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
		interp    = freshInterpreter()
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
		interp    = freshInterpreter()
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
		interp    = freshInterpreter()
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

func TestEnvVariable(t *testing.T) {
	const scriptPath = "testdata/test006.lisp"
	var (
		l                   *lexer.Lexer
		p                   = parser.NewParser()
		err                 error
		val, expectedResult value.LispValue
		parseTree           interface{}
		program             value.Program
		ok                  bool
		interp              = freshInterpreter()
	)

	expectedResult = value.StringValue(os.Getenv("HOME") + "test.file")

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
	} else if val.Type() != types.String {
		t.Errorf("Unexpected type returned from Script: %s (expected String)",
			val.Type().String())
	} else if !val.Equal(expectedResult) {
		t.Errorf("Unexpected result from Script: %s (expected %s)",
			val,
			expectedResult,
		)
	}
} // func TestEnvVariable(t *testing.T)

func TestFileIO(t *testing.T) {
	const scriptPath = "testdata/test007.lisp"
	// Montag, 13. 11. 2017, 18:24
	// In order to test our file wrapper, we first create a file, by hand,
	// write a bunch of numbers in it while computing the sum of those numbers,
	// then we run a Lisp script to read that file, line by line,
	// parse the numbers and compute the sum.
	// If all goes well, we should end up with a number equal to the one
	// we got when writing the file.

	var (
		tmpDir    = os.TempDir()
		prefix    = time.Now().Format("kryLisp-test-io-20060102-150405")
		path      string
		fh        *os.File
		err       error
		counter   int64
		rng       *rand.Rand
		status    bool
		l         *lexer.Lexer
		p         = parser.NewParser()
		parseTree interface{}
		program   value.Program
		ok        bool
		val       value.LispValue
		check     value.IntValue
		interp    = freshInterpreter()
	)

	if fh, err = ioutil.TempFile(tmpDir, prefix); err != nil {
		t.Fatalf("Error creating temp file: %s",
			err.Error())
	} else {
		path = fh.Name()
		defer func() {
			if status {
				os.Remove(path)
			}
		}()
	}

	rng = rand.New(rand.NewSource(time.Now().Unix()))

	for i := 0; i < 100; i++ {
		var n = rng.Int63n(65536)
		fmt.Fprintf(fh, "%d\n", n)
		counter += n
	}

	fh.Close()

	//interp.debug = true
	interp.env.Ins("FILENAME", value.StringValue(path))

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
	} else if check, ok = val.(value.IntValue); !ok {
		if val == nil {
			val = value.NIL
		}
		t.Errorf("Unexpected type returned from test script %s: %s => %s",
			scriptPath,
			val.Type(),
			val)
	} else if int64(check) != counter {
		t.Errorf("Unexpected value returned from test script: %d (expected %d)",
			check,
			counter)
	}
} // func TestFileIO(t *testing.T)
