// /home/krylon/go/src/krylisp/interpreter/01_interpreter_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 12. 09. 2017 by Benjamin Walkenhorst
// (c) 2017 Benjamin Walkenhorst
// Time-stamp: <2017-09-13 20:19:32 krylon>

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

		if result, err = interp.evalPlus(test.input); err != nil && !test.expectedError {
			t.Errorf("Unexpected error from test case #%d: %s",
				idx+1,
				err.Error())
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
} // func (t *testing.T)

///////////////////////////////////////////////////////////

func TestMain(m *testing.M) {
	interp = &Interpreter{
		debug: true,
		env:   value.NewEnvironment(nil),
		fnEnv: value.NewEnvironment(nil),
	}

	os.Exit(m.Run())
}
