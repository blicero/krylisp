// /home/krylon/go/src/krylisp/interpreter/01_interpreter_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 12. 09. 2017 by Benjamin Walkenhorst
// (c) 2017 Benjamin Walkenhorst
// Time-stamp: <2017-09-13 20:54:06 krylon>

package interpreter

import (
	"krylisp/value"
	"os"
	"testing"
)

const (
	plus  = value.Symbol("+")
	minus = value.Symbol("-")
	mult  = value.Symbol("*")
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

///////////////////////////////////////////////////////////

func TestMain(m *testing.M) {
	interp = &Interpreter{
		debug: true,
		env:   value.NewEnvironment(nil),
		fnEnv: value.NewEnvironment(nil),
	}

	os.Exit(m.Run())
}
