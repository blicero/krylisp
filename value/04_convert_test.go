// /home/krylon/go/src/krylisp/value/04_convert_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 22. 10. 2017 by Benjamin Walkenhorst
// (c) 2017 Benjamin Walkenhorst
// Time-stamp: <2017-10-25 14:27:32 krylon>
//
// This file contains the tests for type conversion

package value

import (
	"krylisp/types"
	"math"
	"testing"
)

type convertTest struct {
	input          LispValue
	destination    types.ID
	expectError    bool
	expectedResult LispValue
}

func TestNil(t *testing.T) {
	var cases = []convertTest{
		convertTest{
			input:       NIL,
			destination: types.Integer,
			expectError: true,
		},
		convertTest{
			input:          NIL,
			destination:    types.Nil,
			expectedResult: NIL,
		},
		convertTest{
			input:          NIL,
			destination:    types.String,
			expectedResult: StringValue("NIL"),
		},
		convertTest{
			input:       NIL,
			destination: types.Float,
			expectError: true,
		},
		convertTest{
			input:          NIL,
			destination:    types.List,
			expectedResult: NIL,
		},
		convertTest{
			input:          NIL,
			destination:    types.Symbol,
			expectedResult: Symbol("NIL"),
		},
	}

	for _, test := range cases {
		var res LispValue
		var err error

		res, err = test.input.Convert(test.destination)

		if (err != nil) != test.expectError {
			t.Errorf("Error converting NIL to %s: %s",
				test.destination.String(),
				err.Error())
		} else if !test.expectError && !res.Eq(test.expectedResult) {
			t.Errorf("Error converting NIL to %s: Expected %s, got %s",
				test.destination.String(),
				test.expectedResult.String(),
				res.String())
		}
	}
} // func TestNil(t *testing.T)

func TestInt(t *testing.T) {
	var cases = []convertTest{
		convertTest{
			input:          IntValue(42),
			destination:    types.Integer,
			expectedResult: IntValue(42),
		},
		convertTest{
			input:          IntValue(42),
			destination:    types.Float,
			expectedResult: FloatValue(42.0),
		},
		convertTest{
			input:          IntValue(42),
			destination:    types.String,
			expectedResult: StringValue("42"),
		},
		convertTest{
			input:       IntValue(42),
			destination: types.List,
			expectError: true,
		},
		convertTest{
			input:       IntValue(42),
			destination: types.Nil,
			expectError: true,
		},
	}

	for _, test := range cases {
		var res LispValue
		var err error

		res, err = test.input.Convert(test.destination)

		if (err != nil) != test.expectError {
			t.Errorf("Error converting Integer to %s: %s",
				test.destination.String(),
				err.Error())
		} else if !test.expectError && !res.Eq(test.expectedResult) {
			t.Errorf("Error converting Integer to %s: Expected %s, got %s",
				test.destination.String(),
				test.expectedResult.String(),
				res.String())
		}
	}
} // func TestInt(t *testing.T)

func TestFloat(t *testing.T) {
	var cases = []convertTest{
		convertTest{
			input:          FloatValue(math.Pi),
			destination:    types.Float,
			expectedResult: FloatValue(math.Pi),
		},
		convertTest{
			input:          FloatValue(math.Pi),
			destination:    types.Integer,
			expectedResult: IntValue(3),
		},
		convertTest{
			input:          FloatValue(math.Pi),
			destination:    types.String,
			expectedResult: StringValue("3.141592653589793"),
		},
		convertTest{
			input:       FloatValue(math.Pi),
			destination: types.List,
			expectError: true,
		},
	}

	for _, test := range cases {
		var res LispValue
		var err error

		res, err = test.input.Convert(test.destination)

		if (err != nil) != test.expectError {
			t.Errorf("Error converting Float to %s: %s",
				test.destination.String(),
				err.Error())
		} else if !test.expectError && !res.Eq(test.expectedResult) {
			t.Errorf("Error converting Float to %s: Expected %s, got %s",
				test.destination.String(),
				test.expectedResult.String(),
				res.String())
		}
	}
} // func TestFloat(t *testing.T)
