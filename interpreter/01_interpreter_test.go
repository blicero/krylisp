// /home/krylon/go/src/krylisp/interpreter/01_interpreter_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 12. 09. 2017 by Benjamin Walkenhorst
// (c) 2017 Benjamin Walkenhorst
// Time-stamp: <2017-09-15 19:02:49 krylon>

package interpreter

import (
	"fmt"
	"krylisp/lexer"
	"krylisp/parser"
	"krylisp/value"
	"os"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

const (
	plus  = value.Symbol("+")
	minus = value.Symbol("-")
	mult  = value.Symbol("*")
	div   = value.Symbol("/")
)

var interp *Interpreter

func TestPlus(t *testing.T) {
	type testPlus struct {
		input         *value.List
		expectedValue value.LispValue
		expectedError bool
	}

	var testCases = []testPlus{
		testPlus{
			input: &value.List{
				Car: &value.ConsCell{
					Car: plus,
					Cdr: nil,
				},
				Length: 1,
			},
			expectedValue: value.IntValue(0),
		},
		testPlus{
			input: &value.List{
				Car: &value.ConsCell{
					Car: plus,
					Cdr: &value.ConsCell{
						Car: value.IntValue(42),
						Cdr: nil,
					},
				},
				Length: 2,
			},
			expectedValue: value.IntValue(42),
		},
		testPlus{
			input: &value.List{
				Car: &value.ConsCell{
					Car: plus,
					Cdr: &value.ConsCell{
						Car: value.IntValue(64),
						Cdr: &value.ConsCell{
							Car: value.IntValue(128),
							Cdr: nil,
						},
					},
				},
				Length: 3,
			},
			expectedValue: value.IntValue(192),
		},
	}

	for idx, test := range testCases {
		var result value.LispValue
		var err error

		if result, err = interp.evalPlus(test.input); err != nil {
			if !test.expectedError {
				t.Errorf("Unexpected error from test case #%d: %s",
					idx+1,
					err.Error())
			}
		} else if err == nil && test.expectedError {
			t.Errorf("Test case #%d returned value %s, but expected an error",
				idx+1,
				result.String())
		} else if !result.Eq(test.expectedValue) {
			t.Errorf("Unexpected return value from test case #%d: %s (expected %s)",
				idx+1,
				result.String(),
				test.expectedValue.String())
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
			expectedValue: value.NIL,
			expectedError: true,
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
	}

	for idx, test := range testCases {
		var err error
		var result value.LispValue

		if result, err = interp.evalMinus(test.input); err != nil {
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
	}

	for idx, test := range testCases {
		var err error
		var result value.LispValue

		if result, err = interp.evalMultiply(test.input); err != nil {
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
			expectedValue: value.NIL,
			expectedError: true,
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
	}

	for idx, test := range testCases {
		var err error
		var result value.LispValue

		if result, err = interp.evalDivide(test.input); err != nil {
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

	for _, test := range testCases {
		var fn *value.Function
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
	type testProgram struct {
		source        string
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
		},
	}

	for idx, test := range testCases {
		fmt.Printf("Running DEFUN-test #%d\n",
			idx+1)
		var result interface{}
		var prog value.Program //[]value.LispValue
		var err error
		var ok bool
		var pars = parser.NewParser()
		var lex = lexer.NewLexer([]byte(test.source))

		if result, err = pars.Parse(lex); err != nil {
			t.Fatalf("Error parsing test program: %s",
				err.Error())
		} else {
			fmt.Printf("Parsed test program #%d successfully.\n",
				idx+1)
		}

		if prog, ok = result.([]value.LispValue); !ok {
			t.Fatalf("Parser did not return a program, but a %T", result)
		} else if result, err = interp.Eval(prog); err != nil {
			t.Fatalf("Error evaluating test program: %s",
				err.Error())
		} else if !test.expectedValue.Eq(result.(value.LispValue)) {
			t.Fatalf("Unexpected return value from program: %s (expected %s)",
				result,
				test.expectedValue)
		}

	}
} // func TestDefun(t *testing.T)

///////////////////////////////////////////////////////////

func TestMain(m *testing.M) {
	interp = &Interpreter{
		debug: true,
		env:   value.NewEnvironment(nil),
		fnEnv: value.NewEnvironment(nil),
	}

	os.Exit(m.Run())
}
