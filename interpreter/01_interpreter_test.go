// /home/krylon/go/src/krylisp/interpreter/01_interpreter_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 12. 09. 2017 by Benjamin Walkenhorst
// (c) 2017 Benjamin Walkenhorst
// Time-stamp: <2017-10-20 19:07:48 krylon>

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

	for idx, test := range testCases {
		fmt.Printf("Running DEFUN-test #%d\n",
			idx+1)
		var result interface{}
		var prog value.Program //[]value.LispValue
		var err error
		var ok bool
		var pars = parser.NewParser()
		var lex = lexer.NewLexer([]byte(test.source))
		var fn *value.Function
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
		} else if fnVal, ok = interp.fnEnv.Get(test.name); !ok {
			t.Fatalf("Did not find function %s in environment",
				test.name)
		} else if fnVal == nil {
			t.Fatalf("Function %s is nil",
				test.name)
		} else {
			fn = fnVal.(*value.Function)
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
	}

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
		} else if res, err = interp.evalLessThan(prog[0].(*value.List)); err != nil {
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
		outerEnv      *value.Environment
		input         string
		expectedType  types.ID
		expectedValue string
	}

	var testCases = []letTest{
		letTest{
			outerEnv: &value.Environment{
				Data: map[string]value.LispValue{
					"X": value.IntValue(5),
				},
			},
			input:         "(let ((x 1)) (+ x x))",
			expectedType:  types.Integer,
			expectedValue: "2",
		},
	}

	var oldEnv = interp.env
	defer func() { interp.env = oldEnv }()

	for _, test := range testCases {
		var p = parser.NewParser()
		var l = lexer.NewLexer([]byte(test.input))
		var parsed interface{}
		var prog value.Program
		var ok bool
		var res value.LispValue
		var resstr string
		var err error

		interp.env = test.outerEnv

		if parsed, err = p.Parse(l); err != nil {
			t.Errorf("Error parsing %s: %s",
				test.input,
				err.Error())
		} else if prog, ok = parsed.([]value.LispValue); !ok {
			t.Errorf("Parser returned unexpected type: %T (expected value.Program)",
				parsed)
		} else if res, err = interp.evalLet(prog[0].(*value.List)); err != nil {
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
		} else if val, err = interp.evalNot(prog[0].(*value.List)); err != nil {
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

	// var oldDebug = interp.debug
	// interp.debug = true
	// defer func() { interp.debug = oldDebug }()

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

///////////////////////////////////////////////////////////
// main ///////////////////////////////////////////////////
///////////////////////////////////////////////////////////

func TestMain(m *testing.M) {
	spew.Config.DisableMethods = true
	spew.Config.Indent = "\t"
	spew.Config.SortKeys = true

	interp = &Interpreter{
		debug: false,
		env:   value.NewEnvironment(nil),
		fnEnv: value.NewEnvironment(nil),
	}

	os.Exit(m.Run())
}
