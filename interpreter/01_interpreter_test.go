// /home/krylon/go/src/krylisp/interpreter/01_interpreter_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 12. 09. 2017 by Benjamin Walkenhorst
// (c) 2017 Benjamin Walkenhorst
// Time-stamp: <2017-12-13 18:46:35 krylon>
//
// NOTE Most of these tests are arranged in a pattern I have got to know under
//      the name table-driven development.
//      Just in case you have never heard of that, I'll briefly explain.
//      In most Test* functions, you will find a declaration of a data type
//      representing a single test case at the beginning, followed by an array
//      (okay - slice!) of values of said type, followed by a for loop that
//      iterates over the array, takes some test input from the struct,
//      does something - depending on what is being tested, obviously -
//      and compares the outcome to some reference value stored in the
//      test case struct. If the result does not match the expectation,
//      the code signals an error.
//
//      You will also notice that this arrangement requires relatively much
//      boilerplate code per Test* function, and that some Test* functions may
//      only have one or two test cases.
//
//      The point with table-driven development is not that I should have
//      or want a completely comprehensive test suite now, but that adding
//      additional test cases later on becomes trivial.
//
//      It would probably be possible to factor out some of the boilerplate
//      code into one or a few functions, but right now (2017-11-15) I am
//      not unhappy enough about it to do that.
//
//
// Mittwoch, 15. 11. 2017, 17:09
// In an extreme case of copy-paste-programming, I have changed all Test*
// functions so they create a new Interpreter object from scratch.
// If that works, I would not have to use a global variable for this
// purpose.
// This assumes, however, that no test relies on the interpreter being
// in a certain state it was left by a prior test. And right now, I am
// too lazy to check that.
// So far the test breaks in the exact same place it broke before,
// so that seems to have worked out alright.
//
// Remind me to turn the interpreter into a local variable.
//
// Freitag, 17. 11. 2017, 12:41
// Note to self: Several of these tests instantiate a fresh interpreter
// by using a literal. Since native functions are resolved in the same way
// as Lisp functions, they need to be present int the function environment
// of the interpreter.
// I have started writing a method for that, but when I instiate an
// Interpreter as a Literal instead of calling an "constructor" function,
// this method is not called. In the long run, I should factor that part out
// into a helper function.
//
// Samstag, 18. 11. 2017, 18:36
// The Interpreter instance shall not be a global / file scope variable
// any longer. Every test function gets on of their own.

package interpreter

import (
	"bytes"
	"fmt"
	"krylisp/lexer"
	"krylisp/parser"
	"krylisp/types"
	"krylisp/value"
	"os"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

const (
	//plus  = value.Symbol("+")
	minus = value.Symbol("-")
	mult  = value.Symbol("*")
	div   = value.Symbol("/")
)

func TestPlus(t *testing.T) {
	type testPlus struct {
		input         string
		expectedValue value.LispValue
		expectError   bool
	}

	var testCases = []testPlus{
		testPlus{
			input:         "(+)",
			expectedValue: value.IntValue(0),
		},
		testPlus{
			input:         "(+ 42)",
			expectedValue: value.IntValue(42),
		},
		testPlus{
			input:         "(+ 64 128)",
			expectedValue: value.IntValue(192),
		},
		testPlus{
			input:         "(+ 42 3.5)",
			expectedValue: value.FloatValue(45.5),
		},
		testPlus{
			input:         "(+ 2.5 5)",
			expectedValue: value.FloatValue(7.5),
		},
		testPlus{
			input:         "(+ 4096b 42)",
			expectedValue: value.IntValue(4138),
		},
	}

	interp := freshInterpreter(false)

	for _, test := range testCases {

		var (
			parsed interface{}
			prog   value.Program
			ok     bool
			p      = parser.NewParser()
			l      = lexer.NewLexer([]byte(test.input))
			err    error
			val    value.LispValue
		)

		if parsed, err = p.Parse(l); err != nil {
			t.Errorf("Error parsing test input %s: %s",
				test.input,
				err.Error())
		} else if prog, ok = parsed.([]value.LispValue); !ok {
			t.Errorf("Parser returned unexpected data type: %T",
				parsed)
		} else if val, err = interp.Eval(prog); err != nil {
			if !test.expectError {
				t.Errorf("Error measuring length of %s: %s",
					test.input,
					err.Error())
			}
		} else if !test.expectedValue.Equal(val) {
			t.Errorf("Unexpected result from ConcatArray: %s (expected %s)",
				test.expectedValue,
				val)
		}

	}
} // func TestPlus(t *testing.T)

func TestMinus(t *testing.T) {
	type testMinus struct {
		input         *value.List
		expectedValue value.LispValue
		expectedError bool
	}

	var testCases = []testMinus{
		testMinus{
			input: &value.List{
				Car: &value.ConsCell{
					Car: minus,
					Cdr: nil,
				},
				Length: 1,
			},
			expectedValue: value.IntValue(0),
			expectedError: false,
		},
		testMinus{
			input: &value.List{
				Car: &value.ConsCell{
					Car: minus,
					Cdr: &value.ConsCell{
						Car: value.IntValue(1),
						Cdr: nil,
					},
				},
				Length: 2,
			},
			expectedValue: value.IntValue(-1),
		},
		testMinus{
			input: &value.List{
				Car: &value.ConsCell{
					Car: minus,
					Cdr: &value.ConsCell{
						Car: value.IntValue(192),
						Cdr: &value.ConsCell{
							Car: value.IntValue(128),
							Cdr: nil,
						},
					},
				},
				Length: 3,
			},
			expectedValue: value.IntValue(64),
		},
		testMinus{
			input: &value.List{
				Car: &value.ConsCell{
					Car: minus,
					Cdr: &value.ConsCell{
						Car: value.IntValue(5),
						Cdr: &value.ConsCell{
							Car: value.FloatValue(2.5),
						},
					},
				},
				Length: 3,
			},
			expectedValue: value.FloatValue(2.5),
		},
		testMinus{
			input: &value.List{
				Car: &value.ConsCell{
					Car: minus,
					Cdr: &value.ConsCell{
						Car: value.FloatValue(3.141592),
					},
				},
				Length: 2,
			},
			expectedValue: value.FloatValue(-3.141592),
		},
		testMinus{
			input: &value.List{
				Car: &value.ConsCell{
					Car: minus,
					Cdr: &value.ConsCell{
						Car: value.IntValue(100),
						Cdr: &value.ConsCell{
							Car: value.IntValue(8),
							Cdr: &value.ConsCell{
								Car: value.FloatValue(8.0),
								Cdr: &value.ConsCell{
									Car: value.IntValue(8),
								},
							},
						},
					},
				},
				Length: 5,
			},
			expectedValue: value.FloatValue(76.0),
		},
	}

	var interp = &Interpreter{
		debug: false,
		env:   value.NewEnvironment(nil),
		//fnEnv:  value.NewEnvironment(nil),
		stdin:  os.Stdin,
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
	interp._initNativeFunctions()

	for idx, test := range testCases {
		var err error
		var result value.LispValue

		if result, err = interp.Eval(test.input); err != nil {
			if !test.expectedError {
				t.Errorf("Unexpected error from test case #%d [ %s ]: %s",
					idx+1,
					test.input.String(),
					err.Error())
			}
		} else if err == nil && test.expectedError {
			t.Errorf("Expected error from test case #%d, but I got a value: %s",
				idx+1,
				result.String())
		} else if result == nil {
			t.Errorf("Test case #%d returned _nil_", idx+1)
		} else if !result.Eq(test.expectedValue) {
			t.Errorf("Test case #%d returned unexpected result: Expected = %s, Actual = %s",
				idx+1,
				test.expectedValue.String(),
				result.String())
		}
	}
} // func TestMinus(t *testing.T)

func TestMultiply(t *testing.T) {
	type testMultiply struct {
		input         *value.List
		expectedValue value.LispValue
		expectedError bool
	}

	var testCases = []testMultiply{
		testMultiply{
			input: &value.List{
				Car: &value.ConsCell{
					Car: mult,
					Cdr: nil,
				},
				Length: 1,
			},
			expectedValue: value.IntValue(1),
		},
		testMultiply{
			input: &value.List{
				Car: &value.ConsCell{
					Car: mult,
					Cdr: &value.ConsCell{
						Car: value.IntValue(23),
						Cdr: nil,
					},
				},
				Length: 2,
			},
			expectedValue: value.IntValue(23),
		},
		testMultiply{
			input: &value.List{
				Car: &value.ConsCell{
					Car: mult,
					Cdr: &value.ConsCell{
						Car: value.IntValue(2),
						Cdr: &value.ConsCell{
							Car: value.IntValue(4),
							Cdr: &value.ConsCell{
								Car: value.IntValue(8),
								Cdr: nil,
							},
						},
					},
				},
				Length: 4,
			},
			expectedValue: value.IntValue(64),
		},
		testMultiply{
			input: &value.List{
				Car: &value.ConsCell{
					Car: mult,
					Cdr: &value.ConsCell{
						Car: value.IntValue(40),
						Cdr: &value.ConsCell{
							Car: value.FloatValue(2.5),
						},
					},
				},
				Length: 3,
			},
			expectedValue: value.FloatValue(100.0),
		},
		testMultiply{
			input: &value.List{
				Car: &value.ConsCell{
					Car: mult,
					Cdr: &value.ConsCell{
						Car: value.IntValue(64),
						Cdr: &value.ConsCell{
							Car: value.FloatValue(0.25),
						},
					},
				},
				Length: 3,
			},
			expectedValue: value.FloatValue(16.0),
		},
	}

	var interp = freshInterpreter(false)

	for idx, test := range testCases {
		var err error
		var result value.LispValue

		if result, err = interp.Eval(test.input); err != nil {
			if !test.expectedError {
				t.Errorf("Unexpected error from test case #%d: %s",
					idx+1,
					err.Error())
			}
		} else if test.expectedError {
			t.Errorf("Expected error from test case #%d, but I got a value: %s",
				idx+1,
				result.String())
		} else if !result.Eq(test.expectedValue) {
			t.Errorf("Unexpected result from test case #%d: %s (Expected %s)",
				idx+1,
				result.String(),
				test.expectedValue.String())
		}
	}
} // func TestMultiply(t *testing.T)

func TestDivide(t *testing.T) {
	// Donnerstag, 14. 09. 2017, 14:48
	// In Common Lisp and in Scheme, passing a single argument x to divide
	// returns 1/x. But currently, we only support integers, so I'll consider
	// fewer than tree arguments an error until I come around to implementing
	// additional numeric types.
	type testDivide struct {
		input         *value.List
		expectedValue value.LispValue
		expectedError bool
	}

	var testCases = []testDivide{
		testDivide{
			input: &value.List{
				Car: &value.ConsCell{
					Car: div,
					Cdr: nil,
				},
				Length: 1,
			},
			expectedValue: value.NIL,
			expectedError: true,
		},
		testDivide{
			input: &value.List{
				Car: &value.ConsCell{
					Car: div,
					Cdr: &value.ConsCell{
						Car: value.IntValue(42),
						Cdr: nil,
					},
				},
				Length: 2,
			},
			expectedValue: value.FloatValue(1.0 / 42.0),
			expectedError: false,
		},
		testDivide{
			input: &value.List{
				Car: &value.ConsCell{
					Car: div,
					Cdr: &value.ConsCell{
						Car: value.IntValue(192),
						Cdr: &value.ConsCell{
							Car: value.IntValue(64),
							Cdr: nil,
						},
					},
				},
				Length: 3,
			},
			expectedValue: value.IntValue(3),
		},
		testDivide{
			input: &value.List{
				Car: &value.ConsCell{
					Car: div,
					Cdr: &value.ConsCell{
						Car: value.IntValue(192),
						Cdr: &value.ConsCell{
							Car: value.IntValue(0),
							Cdr: nil,
						},
					},
				},
				Length: 3,
			},
			expectedValue: value.NIL,
			expectedError: true,
		},
		testDivide{
			input: &value.List{
				Car: &value.ConsCell{
					Car: div,
					Cdr: &value.ConsCell{
						Car: value.IntValue(10),
						Cdr: &value.ConsCell{
							Car: value.IntValue(4),
						},
					},
				},
				Length: 3,
			},
			expectedValue: value.IntValue(2),
		},
		testDivide{
			input: &value.List{
				Car: &value.ConsCell{
					Car: div,
					Cdr: &value.ConsCell{
						Car: value.IntValue(10),
						Cdr: &value.ConsCell{
							Car: value.FloatValue(4.0),
						},
					},
				},
				Length: 3,
			},
			expectedValue: value.FloatValue(2.5),
		},
		testDivide{
			input: &value.List{
				Car: &value.ConsCell{
					Car: div,
					Cdr: &value.ConsCell{
						Car: parseBigInt("590295810358705651584"),
						Cdr: &value.ConsCell{
							Car: parseBigInt("128"),
						},
					},
				},
				Length: 3,
			},
			expectedValue: value.IntValue(4611686018427387903),
		},
	}

	var interp = freshInterpreter(false)

	for idx, test := range testCases {
		var err error
		var result value.LispValue

		if result, err = interp.Eval(test.input); err != nil {
			if !test.expectedError {
				t.Errorf("Unexpected error from test case #%d [ %s ] -- %s",
					idx+1,
					test.input.String(),
					err.Error())
			}
		} else if test.expectedError {
			t.Errorf("Expected test case #%d [ %s ] to return an error, but I got result %s",
				idx+1,
				test.input.String(),
				result.String())
		} else if !result.Eq(test.expectedValue) {
			t.Errorf("Wrong result from test case #%d: Expected %s, Actual %s",
				idx+1,
				test.expectedValue.String(),
				result.String())
		}
	}
} // func TestDivide(t *testing.T)

func TestIf(t *testing.T) {
	type testIf struct {
		input         *value.List
		expectedValue value.LispValue
		expectedError bool
	}

	var testCases = []testIf{
		testIf{
			input: &value.List{
				Car: &value.ConsCell{
					Car: value.Symbol("IF"),
					Cdr: &value.ConsCell{
						Car: value.IntValue(3),
						Cdr: &value.ConsCell{
							Car: value.Symbol(":RICHTIG"),
							Cdr: &value.ConsCell{
								Car: value.Symbol(":FALSCH"),
								Cdr: value.NIL,
							},
						},
					},
				},
				Length: 4,
			},
			expectedValue: value.Symbol(":RICHTIG"),
		},
		testIf{
			input: &value.List{
				Car: &value.ConsCell{
					Car: value.Symbol("IF"),
					Cdr: &value.ConsCell{
						Car: value.NIL,
						Cdr: &value.ConsCell{
							Car: value.Symbol(":FALSCH"),
							Cdr: &value.ConsCell{
								Car: value.Symbol(":RICHTIG"),
								Cdr: nil,
							},
						},
					},
				},
				Length: 4,
			},
			expectedValue: value.Symbol(":RICHTIG"),
		},
	}

	var interp = freshInterpreter(false)

	for idx, test := range testCases {
		var err error
		var result value.LispValue

		if result, err = interp.evalIf(test.input); err != nil {
			if !test.expectedError {
				t.Errorf("Unexpected error from test case #%d [ %s ]: %s",
					idx+1,
					test.input.String(),
					err.Error())
			}
		} else if test.expectedError {
			t.Errorf("Expected error from test case #%d [ %s ], but I got a value: %s",
				idx+1,
				test.input.String(),
				result.String())
		} else if !test.expectedValue.Eq(result) {
			t.Errorf("Wrong result from test case #%d: Expected %s, got %s",
				idx+1,
				test.expectedValue.String(),
				result.String())
		}
	}
} // func TestIf(t *testing.T)

func TestLambda(t *testing.T) {
	// For functions, Eq tests for object identity, so the value returned by
	// evalLambda is going to be different from any value I can supply for
	// expectedValue. :-(
	// I could either compare for structural identity, or I could *call* the
	// resulting function object and see if it returns the expected value.
	type testLambda struct {
		lambda        *value.List
		input         *value.ConsCell
		expectedValue value.LispValue
		expectedError bool
	}

	var testCases = []testLambda{
		testLambda{
			lambda: &value.List{
				// (lambda (x) (* x 2))
				Car: &value.ConsCell{
					Car: value.Symbol("LAMBDA"),
					Cdr: &value.ConsCell{
						Car: &value.List{
							Car: &value.ConsCell{
								Car: value.Symbol("X"),
								Cdr: nil,
							},
							Length: 1,
						},
						Cdr: &value.ConsCell{
							Car: &value.List{
								Car: &value.ConsCell{
									Car: value.Symbol("*"),
									Cdr: &value.ConsCell{
										Car: value.Symbol("X"),
										Cdr: &value.ConsCell{
											Car: value.IntValue(2),
											Cdr: nil,
										},
									},
								},
								Length: 3,
							},
						},
					},
				},
				Length: 5,
			},
			input:         &value.ConsCell{Car: value.IntValue(2), Cdr: nil},
			expectedValue: value.IntValue(4),
		},
	}

	var interp = freshInterpreter(false)

	for _, test := range testCases {
		var fn value.Function
		var err error

		if fn, err = interp.evalLambda(test.lambda); err != nil {
			t.Errorf("Error evaluating lambda list %s: %s",
				test.lambda.String(),
				err.Error())
			continue
		} else if fn == nil {
			t.Errorf("evalLambda of %s did not return an error, but the function object is nil!",
				test.lambda.String())
		}

		var fnCall = &value.List{
			Car: &value.ConsCell{
				Car: fn,
				Cdr: &value.ConsCell{
					Car: value.IntValue(2),
					Cdr: nil,
				},
			},
			Length: 2,
		}

		var result value.LispValue

		if result, err = interp.evalFuncall(fnCall); err != nil {
			if !test.expectedError {
				t.Errorf("Unexpected error from function Call %s: %s",
					//fnCall.String(),
					spew.Sdump(fnCall),
					err.Error())
			}
		} else if test.expectedError {
			t.Errorf("Expected error from function call %s, but I got a result: %s",
				fnCall.String(),
				result.String())
		} else if !result.Eq(test.expectedValue) {
			var expString = test.expectedValue.String()
			t.Errorf("Unexpected return value from function call: Expected %s, got %s",
				expString,
				result.String())
		}
	}
} // func TestLambda(t *testing.T)

func TestDefun(t *testing.T) {
	// I could/should add a function member to check for things that
	// cannot be easily expressed otherwise, e.g. does the function show
	// up in the environment, does the doc string have the expected
	// value, etc.
	type testProgram struct {
		source        string
		name          string
		expectedValue value.LispValue
		expectedError bool
	}
	var testCases = []testProgram{
		testProgram{
			source: `
(defun squared (x) (* x x))

(squared 5)
`,
			expectedValue: value.IntValue(25),
			name:          "SQUARED",
		},
		testProgram{
			source: `
(defun factorial (x) 
    "Returns the factorial of x"
    (if (eq x 1) 1 (* x (factorial (- x 1)))))

(factorial 5)
`,
			name:          "FACTORIAL",
			expectedValue: value.IntValue(120),
		},
	}

	var interp = freshInterpreter(false)

	for idx, test := range testCases {
		fmt.Printf("Running DEFUN-test #%d\n",
			idx+1)
		var result interface{}
		var prog value.Program //[]value.LispValue
		var err error
		var ok bool
		var pars = parser.NewParser()
		var lex = lexer.NewLexer([]byte(test.source))
		var fn value.Function
		var fnVal value.LispValue

		if result, err = pars.Parse(lex); err != nil {
			t.Fatalf("Error parsing test program: %s",
				err.Error())
		}

		/*else {
			fmt.Printf("Parsed test program #%d successfully.\n",
				idx+1)
		} */

		if prog, ok = result.([]value.LispValue); !ok {
			t.Fatalf("Parser did not return a program, but a %T", result)
		} else if result, err = interp.Eval(prog); err != nil {
			t.Fatalf("Error evaluating test program: %s",
				err.Error())
		} else if fnVal = interp.env.GetFn(test.name); fnVal == nil {
			t.Fatalf("Did not find function %s in environment",
				test.name)
		} else if fnVal == nil {
			t.Fatalf("Function %s is nil",
				test.name)
		} else {
			fn = fnVal.(value.Function)
		}

		fmt.Printf("#%s => %s\n",
			test.name,
			fn.String())

		if !test.expectedValue.Eq(result.(value.LispValue)) {
			t.Fatalf("Unexpected return value from program: %s (expected %s)",
				result,
				test.expectedValue)
		}

	}
} // func TestDefun(t *testing.T)

func TestDefunKeyword(t *testing.T) {
	type kwTest struct {
		input            string
		expectedKeywords []value.Symbol
	}

	var testCases = []kwTest{
		kwTest{
			input: `
(defun fopen (path &key (mode :rw) (permissions 0644))
    (print path)
    (print mode)
    (print permissions))

#fopen
`,
			expectedKeywords: []value.Symbol{
				value.Symbol("MODE"),
				value.Symbol("PERMISSIONS"),
			},
		},

		kwTest{
			input: `
(defun f (x &key (name "Peter") (age 42))
   (print x)
   (* age 2))

#f
`,
			expectedKeywords: []value.Symbol{
				value.Symbol("NAME"),
				value.Symbol("AGE"),
			},
		},

		kwTest{
			input: `
(defun tcp-connect (host port &key (blocking nil) (rbuf 4096))
    "TCP connect does not actually connect, it just complains."
    (print "I should really get connected."))

#tcp-connect
`,
			expectedKeywords: []value.Symbol{
				value.Symbol("BLOCKING"),
				value.Symbol("RBUF"),
			},
		},
	}

	var interp = freshInterpreter(false)

	for _, test := range testCases {
		var (
			parsed interface{}
			prog   value.Program
			ok     bool
			p      = parser.NewParser()
			l      = lexer.NewLexer([]byte(test.input))
			err    error
			val    value.LispValue
			fn     *value.LispFunction
		)

		if parsed, err = p.Parse(l); err != nil {
			t.Errorf("Error parsing test input %s: %s",
				test.input,
				err.Error())
		} else if prog, ok = parsed.([]value.LispValue); !ok {
			t.Errorf("Parser returned unexpected data type: %T",
				parsed)
		} else if val, err = interp.Eval(prog); err != nil {
			t.Errorf("Error evaluating test input \"%s\": %s",
				test.input,
				err.Error())
		} else if fn, ok = val.(*value.LispFunction); !ok {
			t.Errorf("Type error: Expected a function object, got a %T",
				val)
		} else {
			for _, sym := range test.expectedKeywords {
				if _, ok = fn.Keywordargs[sym]; !ok {
					t.Errorf("Expected keyword argument %s - not present in function",
						sym)
				}
			}
		}
	}
} // func TestDefunKeyword(t *testing.T)

func TestLT(t *testing.T) {
	type testLT struct {
		source        string
		expectedValue value.LispValue
		expectedError bool
	}

	var testCases = []testLT{
		testLT{
			source:        "(< 1 2 3 4 5)",
			expectedValue: value.T,
		},
		testLT{
			source:        "(< 3 2 1)",
			expectedValue: value.NIL,
		},
		testLT{
			source:        "(< 5 10)",
			expectedValue: value.T,
		},
		testLT{
			source:        "(< 25 10)",
			expectedValue: value.NIL,
		},
		testLT{
			source:        "(< 25 100b)",
			expectedValue: value.T,
		},
		testLT{
			source:        "(< 3.5  99b)",
			expectedValue: value.T,
		},
		testLT{
			source:        "(< 5b 99)",
			expectedValue: value.T,
		},
		testLT{
			source:        "(< 22 7b)",
			expectedValue: value.NIL,
		},
		testLT{
			source:        "(< 17b 23.9)",
			expectedValue: value.T,
		},
	}

	var interp = freshInterpreter(false)

	for _, test := range testCases {
		var tree interface{}
		var res value.LispValue
		var err error
		var prog value.Program
		var ok bool
		var p = parser.NewParser()
		var l = lexer.NewLexer([]byte(test.source))

		if tree, err = p.Parse(l); err != nil {
			t.Errorf("Error parsing expression %s: %s",
				test.source,
				err.Error())
		} else if prog, ok = tree.([]value.LispValue); !ok {
			t.Fatalf("Parser did not return a program, but a %T",
				tree)
		} else if res, err = interp.Eval(prog[0].(*value.List)); err != nil {
			t.Errorf("Error evaluating %s: %s",
				test.source,
				err.Error())
		} else if !res.Eq(test.expectedValue) {
			t.Errorf("Unexpected result from test: Expected %s, got %s",
				test.expectedValue.String(),
				res.String())
		}

	}
} // func TestLT(t *testing.T)

func TestOverflow(t *testing.T) {
	type operation uint8

	const (
		add operation = iota
		sub
		mul
	)

	type overflowTest struct {
		left           value.Number
		right          value.Number
		op             operation
		expectedType   types.ID
		expectedResult string
	}

	var testCases = []overflowTest{
		overflowTest{
			left:           value.IntValue(mostPositive) >> 1,
			right:          value.IntValue(128),
			op:             mul,
			expectedType:   types.BigInt,
			expectedResult: "590295810358705651584",
		},
		overflowTest{
			left:           value.IntValue(1024),
			right:          value.IntValue(2),
			op:             add,
			expectedType:   types.Integer,
			expectedResult: "1026",
		},
	}

	// var interp = freshInterpreter(false)

	x := 2 * 4
	fmt.Printf("%d\n", x)

	for _, test := range testCases {
		var res value.Number
		var err error
		var resString string

		switch test.op {
		case add:
			res, err = evalAddition(test.left, test.right)
		case sub:
			res, err = evalSubtraction(test.left, test.right)
		case mul:
			res, err = evalMultiplication(test.left, test.right)
		default:
			t.Errorf("Unexpected operation in test case: %d", test.op)
			continue
		}

		if err != nil {
			t.Errorf("Error adding %s and %s: %s",
				test.left.String(),
				test.right.String(),
				err.Error())
		} else if res.Type() != test.expectedType {
			t.Errorf("Unexpected type for arithmetic operation: %s (expected %s)",
				res.Type().String(),
				test.expectedType.String())
		} else if resString = res.String(); resString != test.expectedResult {
			t.Errorf("Unexpected result from arithmetic operation: %s (expected %s)",
				resString,
				test.expectedResult)
		}
	}
} // func TestOverflow(t *testing.T)

func TestCons(t *testing.T) {
	type testCons struct {
		input          string
		expectedResult string
		expectedType   types.ID
	}

	// Freitag, 22. 09. 2017, 17:14
	// I keep getting weird errors, I think I should write the input as
	// strings and then parse them.
	// Constructing even mildly complex Lisp data as structures is way
	// too complicated and error-prone. :(

	var testCases = []testCons{
		testCons{
			input:          `(cons 1 "Peter")`,
			expectedResult: `(1 . "Peter")`,
			expectedType:   types.ConsCell,
		},
		testCons{
			input:          `(cons 1 (quote (2 3)))`,
			expectedResult: `(1 2 3)`,
			expectedType:   types.List,
		},
		testCons{
			input:          `(cons 1 nil)`,
			expectedResult: "(1)",
			expectedType:   types.List,
		},
	}

	var interp = freshInterpreter(false)

	for _, test := range testCases {
		var p = parser.NewParser()
		var l = lexer.NewLexer([]byte(test.input))
		var parsed interface{}
		var prog value.Program
		var ok bool
		var consed value.LispValue
		var err error
		var result string

		if parsed, err = p.Parse(l); err != nil {
			t.Errorf("Error parsing input %s: %s",
				test.input,
				err.Error())
		} else if prog, ok = parsed.([]value.LispValue); !ok {
			t.Errorf("Parser returned unexpected type: %T (expected value.Program)",
				parsed)
		} else if consed, err = interp.evalCons(prog[0].(*value.List)); err != nil {
			t.Errorf("Error evaluating cons-expression %s: %s",
				test.input,
				err.Error())
		} else if consed.Type() != test.expectedType {
			t.Errorf("Unexpected type returned from evalCons: %s (expected %s)",
				consed.Type(),
				test.expectedType)
		} else if result = consed.String(); result != test.expectedResult {
			t.Errorf("Unexpected result from %s: %s -- Expected %s",
				test.input,
				result,
				test.expectedResult)
		}
	}
} // func TestCons(t *testing.T)

func TestLet(t *testing.T) {
	type letTest struct {
		outerEnv      map[string]value.LispValue
		input         string
		expectedType  types.ID
		expectedValue string
	}

	// Mittwoch, 13. 12. 2017, 18:28
	// Yeah, now that functions and macros are carried along as part of
	// the environment, and now that many things that used to be special
	// forms are GoFunctions, just creating an environment like this
	// breaks things.
	var testCases = []letTest{
		letTest{
			outerEnv: map[string]value.LispValue{
				"X": value.IntValue(5),
			},
			input:         "(let ((x 1)) (+ x x))",
			expectedType:  types.Integer,
			expectedValue: "2",
		},
	}

	var interp = freshInterpreter(false)

	var oldEnv = interp.env
	defer func() { interp.env = oldEnv }()

	var oldDebug = interp.debug
	interp.debug = true
	defer func() { interp.debug = oldDebug }()

	for _, test := range testCases {
		var p = parser.NewParser()
		var l = lexer.NewLexer([]byte(test.input))
		var parsed interface{}
		var prog value.Program
		var ok bool
		var res value.LispValue
		var resstr string
		var err error

		interp.env.Data = test.outerEnv

		if parsed, err = p.Parse(l); err != nil {
			t.Errorf("Error parsing %s: %s",
				test.input,
				err.Error())
		} else if prog, ok = parsed.([]value.LispValue); !ok {
			t.Errorf("Parser returned unexpected type: %T (expected value.Program)",
				parsed)
		} else if res, err = interp.evalLet(prog[0].(*value.List)); err != nil {
			interp.env.Dump(os.Stdout)
			t.Errorf("Error evaluating let-form %s: %s",
				test.input,
				err.Error())
		} else if res.Type() != test.expectedType {
			t.Errorf("%s evaluates to unexpected type: %s (expected %s)",
				test.input,
				res.Type().String(),
				test.expectedType.String())
		} else if resstr = res.String(); resstr != test.expectedValue {
			t.Errorf("%s evaluates to unexpected value %s (expected %s)",
				test.input,
				resstr,
				test.expectedValue)
		}
	}
} // func TestLet(t *testing.T)

func TestNot(t *testing.T) {
	type notTest struct {
		input          string
		expectedResult value.LispValue
	}

	var testCases = []notTest{
		notTest{
			input:          "(not nil)",
			expectedResult: value.T,
		},
		notTest{
			input:          "(not 3)",
			expectedResult: value.NIL,
		},
		notTest{
			input:          "(not (not nil))",
			expectedResult: value.NIL,
		},
	}

	var interp = freshInterpreter(false)

	for _, test := range testCases {
		var parsed interface{}
		var prog value.Program
		var ok bool
		var p = parser.NewParser()
		var l = lexer.NewLexer([]byte(test.input))
		var val value.LispValue
		var err error

		if parsed, err = p.Parse(l); err != nil {
			t.Errorf("Error parsing input %s: %s",
				test.input,
				err.Error())
		} else if prog, ok = parsed.([]value.LispValue); !ok {
			t.Errorf("Parser returned unexpected type: %T",
				parsed)
		} else if val, err = interp.Eval(prog); err != nil {
			t.Errorf("Error evaluating %s: %s",
				test.input,
				err.Error())
		} else if !test.expectedResult.Eq(val) {
			t.Errorf("Unexpected result from evaluating %s: %s [ expected %s ]",
				test.input,
				val,
				test.expectedResult)
		}
	}
} // func TestNot(t *testing.T)

func TestAnd(t *testing.T) {
	type andTest struct {
		input          string
		expectedResult string
	}

	var testCases = []andTest{
		andTest{
			input:          "(and)",
			expectedResult: "T",
		},
		andTest{
			input:          "(and 1)",
			expectedResult: "1",
		},
		andTest{
			input:          "(and 1 2 3 4 5)",
			expectedResult: "5",
		},
		andTest{
			input:          "(and 1 2 nil 4 5)",
			expectedResult: "NIL",
		},
	}

	var interp = freshInterpreter(false)

	for _, test := range testCases {
		var parsed interface{}
		var prog value.Program
		var ok bool
		var p = parser.NewParser()
		var l = lexer.NewLexer([]byte(test.input))
		var val value.LispValue
		var err error
		var res string

		if parsed, err = p.Parse(l); err != nil {
			t.Errorf("Error parsing test input %s: %s",
				test.input,
				err.Error())
		} else if prog, ok = parsed.([]value.LispValue); !ok {
			t.Errorf("Parser returned unexpected data type: %T",
				parsed)
		} else if val, err = interp.evalAnd(prog[0].(*value.List)); err != nil {
			t.Errorf("Error evaluating AND-form %s: %s",
				test.input,
				err.Error())
		} else if val == nil {
			t.Errorf("%s was unexpectedly evaluated as nil",
				test.input)
		} else if res = val.String(); res != test.expectedResult {
			t.Errorf("Unexpected result from %s: Expected %s / Actual %s",
				test.input,
				test.expectedResult,
				res)
		}
	}
} // func TestAnd(t *testing.T)

func TestOr(t *testing.T) {
	type orTest struct {
		input          string
		expectedResult string
	}

	var testCases = []orTest{
		orTest{
			input:          "(or)",
			expectedResult: "NIL",
		},
		orTest{
			input:          "(or 1)",
			expectedResult: "1",
		},
		orTest{
			input:          "(or 1 2 3 4 5)",
			expectedResult: "1",
		},
		orTest{
			input:          "(or 1 2 nil 4 5)",
			expectedResult: "1",
		},
		orTest{
			input:          "(or nil (not (not nil)))",
			expectedResult: "NIL",
		},
		orTest{
			input:          "(or nil (not (not (not nil))))",
			expectedResult: "T",
		},
	}

	var interp = freshInterpreter(false)

	for _, test := range testCases {
		var parsed interface{}
		var prog value.Program
		var ok bool
		var p = parser.NewParser()
		var l = lexer.NewLexer([]byte(test.input))
		var val value.LispValue
		var err error
		var res string

		if parsed, err = p.Parse(l); err != nil {
			t.Errorf("Error parsing test input %s: %s",
				test.input,
				err.Error())
		} else if prog, ok = parsed.([]value.LispValue); !ok {
			t.Errorf("Parser returned unexpected data type: %T",
				parsed)
		} else if val, err = interp.evalOr(prog[0].(*value.List)); err != nil {
			t.Errorf("Error evaluating OR-form %s: %s",
				test.input,
				err.Error())
		} else if val == nil {
			t.Errorf("%s was unexpectedly evaluated as nil",
				test.input)
		} else if res = val.String(); res != test.expectedResult {
			t.Errorf("Unexpected result from %s: Expected %s / Actual %s",
				test.input,
				test.expectedResult,
				res)
		}
	}
} // func TestOr(t *testing.T)

func TestDefine(t *testing.T) {
	type defineTest struct {
		input         string
		key           string
		expectedValue string
	}

	var testCases = []defineTest{
		defineTest{
			input:         `(define x 10)`,
			key:           "X",
			expectedValue: "10",
		},
		defineTest{
			input: `
(define x "Peter")

(let ((x 42))
  (* x 2))
`,
			key:           "X",
			expectedValue: `"Peter"`,
		},
	}

	var interp = freshInterpreter(false)

	for _, test := range testCases {
		var parsed interface{}
		var prog value.Program
		var ok bool
		var p = parser.NewParser()
		var l = lexer.NewLexer([]byte(test.input))
		var val value.LispValue
		var err error
		var res string

		if parsed, err = p.Parse(l); err != nil {
			t.Errorf("Error parsing test input %s: %s",
				test.input,
				err.Error())
		} else if prog, ok = parsed.([]value.LispValue); !ok {
			t.Errorf("Parser returned unexpected data type: %T",
				parsed)
		} else if _, err = interp.Eval(prog); err != nil {
			t.Errorf("Error evaluating DEFINE-form %s: %s",
				test.input,
				err.Error())
		} else if val, ok = interp.env.Get(test.key); !ok {
			t.Errorf("Did not find key %s in environment",
				test.key)
		} else if val == nil {
			t.Errorf("Did find key %s in environment, but its value is nil",
				test.key)
		} else if res = val.String(); res != test.expectedValue {
			t.Errorf("Wrong value found for key %s: Expected = %s, Actual = %s",
				test.key,
				test.expectedValue,
				res)
		}
	}
} // func TestDefine(t *testing.T)

func TestSet(t *testing.T) {
	type setTest struct {
		input         string
		key           string
		expectedValue string
	}

	var testCases = []setTest{
		setTest{
			input:         `(set! x 42)`,
			key:           "X",
			expectedValue: "42",
		},
		setTest{
			input: `
(define x 17)
(set! x 42)
`,
			key:           "X",
			expectedValue: "42",
		},
		setTest{
			input: `
(define x 42)

(let ((x 0))
    (defun test-set (v)
        (set! x v)))

(test-set (* 23 3))
`,
			key:           "X",
			expectedValue: "42",
		},
		setTest{
			input: `
(define x 0)

(defun test-set ()
    (set! x 42))

(test-set)
`,
			key:           "X",
			expectedValue: "42",
		},
	}

	var interp = freshInterpreter(true)

	for _, test := range testCases {
		var parsed interface{}
		var prog value.Program
		var ok bool
		var p = parser.NewParser()
		var l = lexer.NewLexer([]byte(test.input))
		var val value.LispValue
		var err error
		var res string

		fmt.Println("!!!\n" + test.input + "!!!\n")

		if parsed, err = p.Parse(l); err != nil {
			t.Errorf("Error parsing test input %s: %s",
				test.input,
				err.Error())
		} else if prog, ok = parsed.([]value.LispValue); !ok {
			t.Errorf("Parser returned unexpected data type: %T",
				parsed)
		} else if _, err = interp.Eval(prog); err != nil {
			t.Errorf("Error evaluating SET!-form %s: %s",
				test.input,
				err.Error())
		} else if val, ok = interp.env.Get(test.key); !ok {
			t.Errorf("Did not find key %s in environment",
				test.key)
		} else if val == nil {
			t.Errorf("Did find key %s in environment, but its value is nil",
				test.key)
		} else if res = val.String(); res != test.expectedValue {
			t.Errorf("Wrong value found for key %s: Expected = %s, Actual = %s",
				test.key,
				test.expectedValue,
				res)
		}
	}
} // func TestSet(t *testing.T)

func TestPrint(t *testing.T) {
	type printTest struct {
		input          string
		expectedOutput string
	}

	var testCases = []printTest{
		printTest{
			input:          "(print 1 2 3)",
			expectedOutput: "1\n2\n3\n",
		},
		printTest{
			input:          "(print)",
			expectedOutput: "",
		},
	}

	var interp = freshInterpreter(false)

	var oldOut = interp.stdout
	defer func() { interp.stdout = oldOut }()

	for _, test := range testCases {
		var parsed interface{}
		var prog value.Program
		var ok bool
		var p = parser.NewParser()
		var l = lexer.NewLexer([]byte(test.input))
		var err error
		var buf bytes.Buffer
		var res string

		interp.stdout = &buf

		if parsed, err = p.Parse(l); err != nil {
			t.Errorf("Error parsing test input %s: %s",
				test.input,
				err.Error())
		} else if prog, ok = parsed.([]value.LispValue); !ok {
			t.Errorf("Parser returned unexpected data type: %T",
				parsed)
		} else if _, err = interp.Eval(prog); err != nil {
			t.Errorf("Error evaluating PRINT!-form %s: %s",
				test.input,
				err.Error())
		} else if res = buf.String(); res != test.expectedOutput {
			t.Errorf(`Unexpected output for program %s\n
Expected: %q
Actual:   %q
`,
				test.input,
				test.expectedOutput,
				res)
		}
	}
} // func TestPrint(t *testing.T)

func TestApply(t *testing.T) {
	type applyTest struct {
		input          string
		expectedOutput string
	}

	var testCases = []applyTest{
		applyTest{
			input: `
(defun inc (x) (+ x 1))
(print (apply #inc '(1)))
`,
			expectedOutput: "2\n",
		},
		applyTest{
			input: `
(defun twice (x) (+ x x))
(print (apply #twice (list 5))) 
`,
			expectedOutput: "10\n",
		},
		applyTest{
			input: `
(defun threesome (x y z) (+ x y z))
(print (apply #threesome (list (+ 4 5) (* 2 3) (/ 9 3))))
`,
			expectedOutput: "18\n",
		},
	}

	var interp = freshInterpreter(false)

	var oldOut = interp.stdout
	defer func() { interp.stdout = oldOut }()

	for _, test := range testCases {
		var parsed interface{}
		var prog value.Program
		var ok bool
		var p = parser.NewParser()
		var l = lexer.NewLexer([]byte(test.input))
		var err error
		var buf bytes.Buffer
		var res string

		interp.stdout = &buf

		if parsed, err = p.Parse(l); err != nil {
			t.Errorf("Error parsing test input %s: %s",
				test.input,
				err.Error())
		} else if prog, ok = parsed.([]value.LispValue); !ok {
			t.Errorf("Parser returned unexpected data type: %T",
				parsed)
		} else if _, err = interp.Eval(prog); err != nil {
			t.Errorf("Error evaluating APPLY!-form %s: %s",
				test.input,
				err.Error())
		} else if res = buf.String(); res != test.expectedOutput {
			t.Errorf(`Unexpected output for program %s\n
Expected: %q
Actual:   %q
`,
				test.input,
				test.expectedOutput,
				res)
		}
	}

} // func TestApply(t *testing.T)

func TestList(t *testing.T) {
	type listTest struct {
		input          string
		expectedResult string
		expectedError  bool
	}

	var testCases = []listTest{
		listTest{
			input:          "(list)",
			expectedResult: "NIL",
		},
		listTest{
			input:          "(list 1)",
			expectedResult: "(1)",
		},
		listTest{
			input:          "(list 1 (list 2 3) 4)",
			expectedResult: "(1 (2 3) 4)",
		},
	}

	var interp = freshInterpreter(false)

	for _, test := range testCases {
		var (
			parsed interface{}
			prog   value.Program
			ok     bool
			p      = parser.NewParser()
			l      = lexer.NewLexer([]byte(test.input))
			err    error
			val    value.LispValue
			res    string
		)

		if parsed, err = p.Parse(l); err != nil {
			t.Errorf("Error parsing test input %s: %s",
				test.input,
				err.Error())
		} else if prog, ok = parsed.([]value.LispValue); !ok {
			t.Errorf("Parser returned unexpected data type: %T",
				parsed)
		} else if val, err = interp.Eval(prog); err != nil {
			if !test.expectedError {
				t.Errorf("Error evaluating APPLY!-form %s: %s",
					test.input,
					err.Error())
			}
		} else if res = val.String(); res != test.expectedResult {
			t.Errorf("Unexepcted result from list-form:\n\tExpected %s\nActual %s\n",
				test.expectedResult,
				res)
		}
	}
} // func TestList(t *testing.T)

func TestMakeArray(t *testing.T) {
	type arrTest struct {
		input          string
		expectedResult value.LispValue
		expectedError  bool
	}

	var testCases = []arrTest{
		arrTest{
			input: "(make-array '(1 2 3))",
			expectedResult: value.Array{
				value.IntValue(1),
				value.IntValue(2),
				value.IntValue(3),
			},
		},
		arrTest{
			input:          "(make-array)",
			expectedResult: value.Array{},
		},
	}

	var interp = freshInterpreter(false)

	for _, test := range testCases {
		var (
			parsed interface{}
			prog   value.Program
			ok     bool
			p      = parser.NewParser()
			l      = lexer.NewLexer([]byte(test.input))
			err    error
			val    value.LispValue
		)

		if parsed, err = p.Parse(l); err != nil {
			t.Errorf("Error parsing test input %s: %s",
				test.input,
				err.Error())
		} else if prog, ok = parsed.([]value.LispValue); !ok {
			t.Errorf("Parser returned unexpected data type: %T",
				parsed)
		} else if val, err = interp.Eval(prog); err != nil {
			if !test.expectedError {
				t.Errorf("Error evaluating MAKE-ARRAY form %s: %s",
					test.input,
					err.Error())
			}
		} else if value.IsNil(val) {
			t.Errorf("Input %s evaluated to NIL",
				test.input)
		} else if val.Type() != types.Array {
			t.Errorf("Unexpected type returned from MAKE-ARRAY %s: %T - %s",
				test.input,
				val,
				val.String())
		} else if !val.Eq(test.expectedResult) {
			t.Errorf("Expected and actual results are not equal: Expected %s, actual %s",
				test.expectedResult.String(),
				val.String())
		}
	}
} // func TestMakeArray(t *testing.T)

func TestAPush(t *testing.T) {
	type pushTest struct {
		input          string
		expectedResult value.Array
	}

	var testCases = []pushTest{
		pushTest{
			input: "(apush [1 2 3] 4 5 6)",
			expectedResult: value.Array{
				value.IntValue(1),
				value.IntValue(2),
				value.IntValue(3),
				value.IntValue(4),
				value.IntValue(5),
				value.IntValue(6),
			},
		},
	}

	var interp = freshInterpreter(false)

	for _, test := range testCases {
		var (
			parsed interface{}
			prog   value.Program
			ok     bool
			p      = parser.NewParser()
			l      = lexer.NewLexer([]byte(test.input))
			err    error
			val    value.LispValue
		)

		if parsed, err = p.Parse(l); err != nil {
			t.Errorf("Error parsing test input %s: %s",
				test.input,
				err.Error())
		} else if prog, ok = parsed.([]value.LispValue); !ok {
			t.Errorf("Parser returned unexpected data type: %T",
				parsed)
		} else if val, err = interp.Eval(prog); err != nil {
			t.Errorf("Error evaluating MAKE-ARRAY form %s: %s",
				test.input,
				err.Error())
		} else if value.IsNil(val) {
			t.Errorf("Input %s evaluated to NIL",
				test.input)
		} else if val.Type() != types.Array {
			t.Errorf("Unexpected type returned from MAKE-ARRAY %s: %T - %s",
				test.input,
				val,
				val.String())
		} else if !val.Eq(test.expectedResult) {
			t.Errorf("Expected and actual results are not equal: Expected %s, actual %s",
				test.expectedResult.String(),
				val.String())
		}
	}
} // func TestAPush(t *testing.T)

func TestARef(t *testing.T) {
	type arefTest struct {
		input          string
		expectedResult value.LispValue
		expectError    bool
	}

	var testCases = []arefTest{
		arefTest{
			input: `
(define arr [1 2 3 4])

(aref arr 2)
`,
			expectedResult: value.IntValue(3),
		},
		arefTest{
			input: `
(define arr [1 2 3 4])

(aref arr 6)
`,
			expectError: true,
		},
	}

	var interp = freshInterpreter(false)

	for _, test := range testCases {
		var (
			parsed interface{}
			prog   value.Program
			ok     bool
			p      = parser.NewParser()
			l      = lexer.NewLexer([]byte(test.input))
			err    error
			val    value.LispValue
		)

		if parsed, err = p.Parse(l); err != nil {
			t.Errorf("Error parsing test input %s: %s",
				test.input,
				err.Error())
		} else if prog, ok = parsed.([]value.LispValue); !ok {
			t.Errorf("Parser returned unexpected data type: %T",
				parsed)
		} else if val, err = interp.Eval(prog); err != nil {
			if !test.expectError {
				t.Errorf("Error evaluating MAKE-ARRAY form %s: %s",
					test.input,
					err.Error())
			}
		} else if test.expectError {
			t.Errorf("Test program should have raised an error: %s",
				test.input)
		} else if value.IsNil(val) {
			t.Errorf("Input %s evaluated to NIL",
				test.input)
		} else if !test.expectedResult.Eq(val) {
			t.Errorf("Unexpected return value from AREF: Expected %s, got %s",
				test.expectedResult.String(),
				val.String())
		}
	}
} // func TestARef(t *testing.T)

func TestHashCreate(t *testing.T) {
	type hashTest struct {
		input          string
		expectedResult value.LispValue
		expectError    bool
	}

	var testCases = []hashTest{
		hashTest{
			input:          "(make-hash)",
			expectedResult: make(value.Hashtable),
		},
		hashTest{
			input: `{ Karl : 3.5, "Peter" : Wilhelm, 7 : 7 }`,
			expectedResult: value.Hashtable{
				value.Symbol("KARL"):       value.FloatValue(3.5),
				value.StringValue("Peter"): value.Symbol("WILHELM"),
				value.IntValue(7):          value.IntValue(7),
			},
		},
	}

	var interp = freshInterpreter(false)

	for _, test := range testCases {
		var (
			parsed interface{}
			prog   value.Program
			ok     bool
			p      = parser.NewParser()
			l      = lexer.NewLexer([]byte(test.input))
			err    error
			val    value.LispValue
		)

		if parsed, err = p.Parse(l); err != nil {
			t.Errorf("Error parsing test input %s: %s",
				test.input,
				err.Error())
		} else if prog, ok = parsed.([]value.LispValue); !ok {
			t.Errorf("Parser returned unexpected data type: %T",
				parsed)
		} else if val, err = interp.Eval(prog); err != nil {
			t.Errorf("Error evaluating Hash creation form %s: %s",
				test.input,
				err.Error())
		} else if value.IsNil(val) {
			t.Errorf("Input %s evaluated to NIL",
				test.input)
		} else if val.Type() != types.Hashtable {
			t.Errorf("Unexpected type (expected Hashtable) returned from %s: %T - %s",
				test.input,
				val,
				val.String())
		} else if !val.Eq(test.expectedResult) {
			t.Errorf("Expected and actual results are not equal: Expected %s, actual %s",
				test.expectedResult.String(),
				val.String())
		}
	}
} // func TestHashCreate(t *testing.T)

func TestHashLookup(t *testing.T) {
	type lookupTest struct {
		input          string
		key            value.LispValue
		expectedResult value.LispValue
	}

	var testCases = []lookupTest{
		lookupTest{
			input: `(define tbl { 42 : "The Answer", Pi : 3.141592, 23 : 666 })

(hashref tbl 42)
`,
			key:            value.IntValue(42),
			expectedResult: value.StringValue("The Answer"),
		},
		lookupTest{
			input: `(define tbl { "Hans" : 3.74, "Peter" : 5.29, "Karl" : 4.46 })

(hashref tbl "Ludwig")
`,
			key:            value.Symbol("Ludwig"),
			expectedResult: value.NIL,
		},
	}

	for _, test := range testCases {
		var (
			parsed interface{}
			prog   value.Program
			ok     bool
			err    error
			val    value.LispValue
			p      = parser.NewParser()
			l      = lexer.NewLexer([]byte(test.input))
			interp = freshInterpreter(false)
		)

		if parsed, err = p.Parse(l); err != nil {
			t.Errorf("Error parsing test input %s: %s",
				test.input,
				err.Error())
		} else if prog, ok = parsed.([]value.LispValue); !ok {
			t.Errorf("Parser returned unexpected data type: %T",
				parsed)
		} else if val, err = interp.Eval(prog); err != nil {
			t.Errorf("Error evaluating Hash creation form %s: %s",
				test.input,
				err.Error())
		} else if val == nil {
			t.Errorf("Input %s evaluated to nil",
				test.input)
		} else if !val.Eq(test.expectedResult) {
			t.Errorf("Expected and actual results are not equal: Expected %s, actual %s",
				test.expectedResult.String(),
				val.String())
		}
	}
} // func TestHashLookup(t *testing.T)

func TestHashSet(t *testing.T) {
	type setTest struct {
		input         string
		name          string
		expectedKey   value.LispValue
		expectedValue value.LispValue
	}

	var testCases = []setTest{
		setTest{
			input: `(define tbl (make-hash))

(hash-set tbl "Peter" 23)
`,
			expectedKey:   value.StringValue("Peter"),
			expectedValue: value.IntValue(23),
			name:          "TBL",
		},
	}

	var interp = freshInterpreter(false)

	for _, test := range testCases {
		var (
			parsed interface{}
			prog   value.Program
			ok     bool
			p      = parser.NewParser()
			l      = lexer.NewLexer([]byte(test.input))
			err    error
			val    value.LispValue
		)

		if parsed, err = p.Parse(l); err != nil {
			t.Errorf("Error parsing test input %s: %s",
				test.input,
				err.Error())
		} else if prog, ok = parsed.([]value.LispValue); !ok {
			t.Errorf("Parser returned unexpected data type: %T",
				parsed)
		} else if val, err = interp.Eval(prog); err != nil {
			t.Errorf("Error evaluating Hash creation form %s: %s",
				test.input,
				err.Error())
		}

		var tbl value.Hashtable
		if val, ok = interp.env.Data["TBL"]; !ok {
			t.Error("Did not find hash table named tbl in Environment")
		} else if val == nil || val.Type() != types.Hashtable {
			t.Error("TBL is not a hash table")
		}

		tbl, _ = val.(value.Hashtable)

		if val, ok = tbl[test.expectedKey]; !ok {
			t.Errorf("Did not find key %s in hash table",
				test.expectedKey)
		} else if !test.expectedValue.Eq(val) {
			t.Errorf("Unexpected value found in hash table: %s (expected %s)",
				val.String(),
				test.expectedValue.String())
		}
	}
} // func TestHashSet(t *testing.T)

func TestHashDelete(t *testing.T) {
	type delTest struct {
		input      string
		deletedKey value.LispValue
	}

	var testCases = []delTest{
		delTest{
			input: `
(define tbl { 1 : Peter, 2 : Karl, 3 : Johannes, 4 : Ludwig })

(hash-delete tbl 2)
`,
			deletedKey: value.IntValue(2),
		},
		delTest{
			input: `
(define tbl { Karl : 3, Ludwig : 5, "Ludwig": 15 })

(hash-delete tbl 'Ludwig)
`,
			deletedKey: value.Symbol("LUDWIG"),
		},
	}

	var interp = freshInterpreter(false)

	for _, test := range testCases {
		var (
			parsed interface{}
			prog   value.Program
			ok     bool
			p      = parser.NewParser()
			l      = lexer.NewLexer([]byte(test.input))
			err    error
			val    value.LispValue
			tbl    value.Hashtable
		)

		if parsed, err = p.Parse(l); err != nil {
			t.Errorf("Error parsing test input %s: %s",
				test.input,
				err.Error())
		} else if prog, ok = parsed.([]value.LispValue); !ok {
			t.Errorf("Parser returned unexpected data type: %T",
				parsed)
		} else if val, err = interp.Eval(prog); err != nil {
			t.Errorf("Error evaluating Hash creation form %s: %s",
				test.input,
				err.Error())
		} else if val, ok = interp.env.Data["TBL"]; !ok {
			t.Error("Cannot find symbol TBL in current environment")
		} else if val.Type() != types.Hashtable {
			t.Errorf("First argument to HASH-DELETE is not a hash table, but a %T",
				val)
		} else if tbl, ok = val.(value.Hashtable); !ok {
			t.Errorf("First argument to HASH-DELETE is not a hash table, but a %T",
				val)
		} else if val, ok = tbl[test.deletedKey]; ok {
			t.Errorf("Key %s should not have been present in that hash table",
				test.deletedKey)
		}
	}
} // func TestHashDelete(t *testing.T)

func TestRegexpCompile(t *testing.T) {
	type regexTest struct {
		input          string
		expectedString string
		expectError    bool
	}

	// var olddbg = interp.debug
	// interp.debug = true
	// defer func() { interp.debug = olddbg }()

	var testCases = []regexTest{
		regexTest{
			input: `
(regexp-compile "^\d+\s+\w+$")
`,
			expectedString: `^\d+\s+\w+$`,
		},
		regexTest{
			input: `
(regexp-compile "(\d{1,3})[.](\d{1,3})[.](\d{1,3})[.](\d{1,3})")
`,
			expectedString: `(\d{1,3})[.](\d{1,3})[.](\d{1,3})[.](\d{1,3})`,
		},
	}

	var interp = freshInterpreter(false)

	for _, test := range testCases {
		var (
			parsed interface{}
			prog   value.Program
			ok     bool
			p      = parser.NewParser()
			l      = lexer.NewLexer([]byte(test.input))
			err    error
			val    value.LispValue
		)

		if parsed, err = p.Parse(l); err != nil {
			t.Errorf("Error parsing test input %s: %s",
				test.input,
				err.Error())
		} else if prog, ok = parsed.([]value.LispValue); !ok {
			t.Errorf("Parser returned unexpected data type: %T",
				parsed)
		} else if val, err = interp.Eval(prog); err != nil {
			if !test.expectError {
				t.Errorf("Error compiling regexp %s: %s",
					test.input,
					err.Error())
			}
		} else if val.Type() != types.Regexp {
			t.Errorf("REGEXP-COMPILE did not return a regexp, but a %T: %v",
				val,
				val)
		}
	}
} // func TestRegexpCompile(t *testing.T)

func TestRegexpMatch(t *testing.T) {
	type regexTest struct {
		input          string
		expectedResult value.LispValue
		expectError    bool
	}

	var testCases = []regexTest{
		regexTest{
			input: `
(define p (regexp-compile "^\d+\s+\w+\s+\d+"))

(define single-line "42    Horst    23")

(regexp-match p single-line)
`,
			expectedResult: value.Array{
				value.Array{
					value.StringValue("42    Horst    23"),
				},
			},
		},

		regexTest{
			input: `
(define pat (regexp-compile "\d{2}(\w+)::"))
(define str "23Peter::     45Horst:: 3Karl:: 73Hugo:: 21Siegfried::")

(regexp-match pat str)
`,
			expectedResult: value.Array{
				value.Array{
					value.StringValue("23Peter::"),
					value.StringValue("Peter"),
				},
				value.Array{
					value.StringValue("45Horst::"),
					value.StringValue("Horst"),
				},
				value.Array{
					value.StringValue("73Hugo::"),
					value.StringValue("Hugo"),
				},
				value.Array{
					value.StringValue("21Siegfried::"),
					value.StringValue("Siegfried"),
				},
			},
		},

		regexTest{
			input: `
(define pat (regexp-compile "\d+ooo\d+"))
(define str "123oooooo3423")

(regexp-match pat str)
`,
			expectedResult: value.NIL,
		},
	}

	var interp = freshInterpreter(false)

	for _, test := range testCases {
		var (
			parsed interface{}
			prog   value.Program
			ok     bool
			p      = parser.NewParser()
			l      = lexer.NewLexer([]byte(test.input))
			err    error
			val    value.LispValue
		)

		if parsed, err = p.Parse(l); err != nil {
			t.Errorf("Error parsing test input %s: %s",
				test.input,
				err.Error())
		} else if prog, ok = parsed.([]value.LispValue); !ok {
			t.Errorf("Parser returned unexpected data type: %T",
				parsed)
		} else if val, err = interp.Eval(prog); err != nil {
			if !test.expectError {
				t.Errorf("Error compiling regexp %s: %s",
					test.input,
					err.Error())
			}
		} else if val.Type() != types.Array && val.Type() != types.Nil {
			t.Errorf("Unexpected return value from REGEXP-MATCH: %T - %v",
				val,
				val)
		} else if !val.Eq(test.expectedResult) {
			t.Errorf(`Regexp match is not what we expected: 
Expected: %T -> %v
Actual:   %T -> %v`,
				test.expectedResult,
				test.expectedResult,
				val,
				val)
		}
	}
} // func TestRegexpMatch(t *testing.T)

func TestDoLoop(t *testing.T) {
	type doTest struct {
		input          string
		expectedResult value.LispValue
		expectError    bool
	}

	var testCases = []doTest{
		doTest{
			// Beispiel in Common Lisp:
			// (defun factorial (x)
			// (do ((i 1 (1+ i))
			//      (j 1 (* j i)))
			//      ((> i x) j)))
			input: `
(do ((i 1 (+ i 1))
     (j 1 (* j i)))
     (> i 9)
     j)
`,
			expectedResult: value.IntValue(3628800),
		},
	}

	var interp = freshInterpreter(false)

	interp._initNativeFunctions()

	for _, test := range testCases {
		var (
			parsed interface{}
			prog   value.Program
			ok     bool
			p      = parser.NewParser()
			l      = lexer.NewLexer([]byte(test.input))
			err    error
			val    value.LispValue
		)

		if parsed, err = p.Parse(l); err != nil {
			t.Errorf("Error parsing test input %s: %s",
				test.input,
				err.Error())
		} else if prog, ok = parsed.([]value.LispValue); !ok {
			t.Errorf("Parser returned unexpected data type: %T",
				parsed)
		} else if val, err = interp.Eval(prog); err != nil {
			if !test.expectError {
				t.Errorf("Error running DO loop %s: %s",
					test.input,
					err.Error())
			}
		} else if !test.expectedResult.Eq(val) {
			t.Errorf("Unexpected result from DO loop: Expected %s, actual %s",
				test.expectedResult,
				val)
		}
	}
} // func TestDoLoop(t *testing.T)

func TestLength(t *testing.T) {
	type lengthTest struct {
		input          string
		expectedResult value.IntValue
		expectError    bool
	}

	var testCases = []lengthTest{
		lengthTest{
			input:          "(length '(1 2 3))",
			expectedResult: 3,
		},
		lengthTest{
			input:          "(length [1 2 3 4 5])",
			expectedResult: 5,
		},
		lengthTest{
			input:          "(length { karl : 3, peter : 7, horst : 2, willi : 11 })",
			expectedResult: 4,
		},
		lengthTest{
			input:          `(length "Hallo, Welt!")`,
			expectedResult: 12,
		},
	}

	var interp = freshInterpreter(false)

	for _, test := range testCases {
		var (
			parsed interface{}
			prog   value.Program
			ok     bool
			p      = parser.NewParser()
			l      = lexer.NewLexer([]byte(test.input))
			err    error
			val    value.LispValue
		)

		if parsed, err = p.Parse(l); err != nil {
			t.Errorf("Error parsing test input %s: %s",
				test.input,
				err.Error())
		} else if prog, ok = parsed.([]value.LispValue); !ok {
			t.Errorf("Parser returned unexpected data type: %T",
				parsed)
		} else if val, err = interp.Eval(prog); err != nil {
			if !test.expectError {
				t.Errorf("Error measuring length of %s: %s",
					test.input,
					err.Error())
			}
		} else if !test.expectedResult.Eq(val) {
			t.Errorf("Unexpected result from LENGTH: Expected %s, actual %s",
				test.expectedResult,
				val)
		}
	}
} // func TestLength(t *testing.T)

func TestConcatList(t *testing.T) {
	type concatTest struct {
		input          string
		expectedResult value.LispValue
		expectError    bool
	}

	var testCases = []concatTest{
		concatTest{
			input: "(concat '(1 2 3) '(4 5 6))",
			expectedResult: value.MakeList(
				value.IntValue(1),
				value.IntValue(2),
				value.IntValue(3),
				value.IntValue(4),
				value.IntValue(5),
				value.IntValue(6),
			),
		},
		concatTest{
			input: "(concat '(1 2 3) [4 5 6])",
			expectedResult: value.MakeList(
				value.IntValue(1),
				value.IntValue(2),
				value.IntValue(3),
				value.IntValue(4),
				value.IntValue(5),
				value.IntValue(6),
			),
		},
	}

	var interp = freshInterpreter(false)

	for _, test := range testCases {
		var (
			parsed interface{}
			prog   value.Program
			ok     bool
			p      = parser.NewParser()
			l      = lexer.NewLexer([]byte(test.input))
			err    error
			val    value.LispValue
		)

		if parsed, err = p.Parse(l); err != nil {
			t.Errorf("Error parsing test input %s: %s",
				test.input,
				err.Error())
		} else if prog, ok = parsed.([]value.LispValue); !ok {
			t.Errorf("Parser returned unexpected data type: %T",
				parsed)
		} else if val, err = interp.Eval(prog); err != nil {
			if !test.expectError {
				t.Errorf("Error measuring length of %s: %s",
					test.input,
					err.Error())
			}
		} else if !test.expectedResult.Equal(val) {
			t.Errorf("Unexpected result from ConcatList: %s (expected %s)",
				test.expectedResult,
				val)
		}
	}
} // func TestConcatList(t *testing.T)

func TestConcatArray(t *testing.T) {
	type concatArrayTest struct {
		input          string
		expectedResult value.LispValue
		expectError    bool
	}

	var testCases = []concatArrayTest{
		concatArrayTest{
			input: "(concat [1 2 3] '(4 5 6))",
			expectedResult: value.Array{
				value.IntValue(1),
				value.IntValue(2),
				value.IntValue(3),
				value.IntValue(4),
				value.IntValue(5),
				value.IntValue(6),
			},
		},
		concatArrayTest{
			input: `(concat [1 2 3] "Harald")`,
			expectedResult: value.Array{
				value.IntValue(1),
				value.IntValue(2),
				value.IntValue(3),
				value.StringValue("Harald"),
			},
		},
	}

	var interp = freshInterpreter(false)

	for _, test := range testCases {
		var (
			parsed interface{}
			prog   value.Program
			ok     bool
			p      = parser.NewParser()
			l      = lexer.NewLexer([]byte(test.input))
			err    error
			val    value.LispValue
		)

		if parsed, err = p.Parse(l); err != nil {
			t.Errorf("Error parsing test input %s: %s",
				test.input,
				err.Error())
		} else if prog, ok = parsed.([]value.LispValue); !ok {
			t.Errorf("Parser returned unexpected data type: %T",
				parsed)
		} else if val, err = interp.Eval(prog); err != nil {
			if !test.expectError {
				t.Errorf("Error measuring length of %s: %s",
					test.input,
					err.Error())
			}
		} else if !test.expectedResult.Equal(val) {
			t.Errorf("Unexpected result from ConcatArray: %s (expected %s)",
				test.expectedResult,
				val)
		}
	}
} // func TestConcatArray(t *testing.T)

func TestConcatString(t *testing.T) {
	type concatStringTest struct {
		input          string
		expectedResult value.StringValue
		expectError    bool
	}

	var testCases = []concatStringTest{
		concatStringTest{
			input:          `(concat "Peter" " " "Lustig")`,
			expectedResult: value.StringValue("Peter Lustig"),
		},
	}

	var interp = freshInterpreter(false)

	for _, test := range testCases {
		var (
			parsed interface{}
			prog   value.Program
			ok     bool
			p      = parser.NewParser()
			l      = lexer.NewLexer([]byte(test.input))
			err    error
			val    value.LispValue
		)

		if parsed, err = p.Parse(l); err != nil {
			t.Errorf("Error parsing test input %s: %s",
				test.input,
				err.Error())
		} else if prog, ok = parsed.([]value.LispValue); !ok {
			t.Errorf("Parser returned unexpected data type: %T",
				parsed)
		} else if val, err = interp.Eval(prog); err != nil {
			if !test.expectError {
				t.Errorf("Error measuring length of %s: %s",
					test.input,
					err.Error())
			}
		} else if !test.expectedResult.Equal(val) {
			t.Errorf("Unexpected result from ConcatArray: %s (expected %s)",
				test.expectedResult,
				val)
		}
	}
} // func TestConcatString(t *testing.T)

// I don't want to rely on specific environment variables existing
// or having a specific value. This is going to be awkward.
// I COULD manually set a few environment variables "manually",
// then call GETENV in Lisp and see if the correct result comes
// back.
func TestGetenv(t *testing.T) {
	type envTest struct {
		name string
		val  string
	}

	var testCases = []envTest{
		envTest{
			name: "TEST123",
			val:  "TEST456",
		},
	}

	var interp = freshInterpreter(false)

	for _, test := range testCases {
		var (
			err      error
			oldValue string
			res      value.LispValue
			call     = value.MakeList(
				value.Symbol("GETENV"),
				value.StringValue(test.name),
			)
		)

		oldValue = os.Getenv(test.name)

		if err = os.Setenv(test.name, test.val); err != nil {
			t.Errorf("Error setting environment variable %s to %s: %s",
				test.name,
				test.val,
				err.Error())
		} else if res, err = interp.Eval(call); err != nil {
			t.Errorf("Error calling GETENV: %s", err.Error())
		} else if !res.Equal(value.StringValue(test.val)) {
			t.Errorf("Unexpected return value from (GETENV %s): %s (expected %s)",
				test.name,
				res,
				test.val)
		}

		if oldValue != "" {
			_ = os.Setenv(test.name, oldValue)
		}
	}
} // func TestGetenv(t *testing.T)

func TestSetEnv(t *testing.T) {
	type envTest struct {
		name string
		val  string
	}

	var testCases = []envTest{
		envTest{
			name: "TEST123",
			val:  "HALLOWELT",
		},
		envTest{
			name: "TESTUMLAUT",
			val:  "äöüÄÖÜ",
		},
	}

	var interp = freshInterpreter(false)

	for _, test := range testCases {
		var (
			err             error
			oldValue, check string
			call            = value.MakeList(
				value.Symbol("SETENV"),
				value.StringValue(test.name),
				value.StringValue(test.val),
			)
		)

		oldValue = os.Getenv(test.name)

		if _, err = interp.Eval(call); err != nil {
			t.Errorf("Error setting environment variable %s to %s: %s",
				test.name,
				test.val,
				err.Error())
		} else if check = os.Getenv(test.name); check != test.val {
			t.Errorf("Unexpected value found in environment variable %s: %s (expected %s)",
				test.name,
				check,
				test.val)
		}

		if oldValue != "" {
			_ = os.Setenv(test.name, oldValue)
		}
	}
} // func TestSetEnv(t *testing.T)

func TestKeywordArgs(t *testing.T) {
	type kwTest struct {
		input          string
		expectedResult value.LispValue
		expectError    bool
	}

	var testCases = []kwTest{
		kwTest{
			input: `
(defun f (x &key (name "Peter") (age 42))
   (print x)
   (* age 2))

(f 10 :age 64)
`,
			expectedResult: value.IntValue(128),
		},
	}

	var interp = freshInterpreter(false)

	var oldDebug = interp.debug
	interp.debug = true
	defer func() { interp.debug = oldDebug }()

	for _, test := range testCases {
		var (
			parsed interface{}
			prog   value.Program
			ok     bool
			p      = parser.NewParser()
			l      = lexer.NewLexer([]byte(test.input))
			err    error
			val    value.LispValue
		)

		if parsed, err = p.Parse(l); err != nil {
			t.Errorf("Error parsing test input %s: %s",
				test.input,
				err.Error())
		} else if prog, ok = parsed.([]value.LispValue); !ok {
			t.Errorf("Parser returned unexpected data type: %T",
				parsed)
		} else if val, err = interp.Eval(prog); err != nil {
			if !test.expectError {
				t.Errorf("Error evaluating keyword test script %s: %s",
					test.input,
					err.Error())
			}
		} else if !test.expectedResult.Equal(val) {
			t.Errorf("Unexpected result from ConcatArray: %s (expected %s)",
				test.expectedResult,
				val)
		}
	}
} // func TestKeywordArgs(t *testing.T)

func TestReadFromString(t *testing.T) {
	type readTest struct {
		input          string
		expectedResult value.LispValue
		expectError    bool
	}

	var testCases = []readTest{
		readTest{
			input:          "3",
			expectedResult: value.IntValue(3),
		},
		readTest{
			input:          "peter",
			expectedResult: value.Symbol("PETER"),
		},
	}

	var interp = freshInterpreter(false)

	for _, test := range testCases {
		var (
			res value.LispValue
			err error
			arg = value.Arguments{
				Positional: []value.LispValue{value.StringValue(test.input)},
			}
		)

		if res, err = interp.evalReadString(&arg); err != nil {
			if !test.expectError {
				t.Errorf("Error reading test input \"%s\": %s",
					test.input,
					err.Error())
			}
		} else if !test.expectedResult.Equal(res) {
			t.Errorf("Unexpected result from READ-FROM-STRING: %s (expected %s)",
				res,
				test.expectedResult)
		}
	}
} // func TestReadFromString(t *testing.T)

///////////////////////////////////////////////////////////
// Utility functions //////////////////////////////////////
///////////////////////////////////////////////////////////

func parseBigInt(s string) *value.BigInt {
	if b, e := value.BigIntFromString(s); e != nil {
		panic(e)
	} else {
		return b
	}
} // func parseBigInt(s string) *BigInt

func freshInterpreter(dbg bool) *Interpreter {
	var inter = &Interpreter{
		debug:  dbg,
		env:    value.NewEnvironment(nil),
		stdin:  os.Stdin,
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
	inter._initNativeFunctions()
	return inter
}

///////////////////////////////////////////////////////////
// main ///////////////////////////////////////////////////
///////////////////////////////////////////////////////////

func TestMain(m *testing.M) {
	spew.Config.DisableMethods = true
	spew.Config.Indent = "\t"
	spew.Config.SortKeys = true

	os.Exit(m.Run())
}
