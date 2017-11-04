// /home/krylon/go/src/krylisp/value/04_convert_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 22. 10. 2017 by Benjamin Walkenhorst
// (c) 2017 Benjamin Walkenhorst
// Time-stamp: <2017-11-04 00:40:00 krylon>
//
// This file contains the tests for type conversion

package value

import (
	"krylisp/types"
	"math"
	"math/big"
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

func parseBigInt(s string) *BigInt {
	if b, e := BigIntFromString(s); e != nil {
		panic(e)
	} else {
		return b
	}
} // func parseBigInt(s string) *BigInt

func TestBigInt(t *testing.T) {
	var cases = []convertTest{
		convertTest{
			input:          &BigInt{Value: big.NewInt(32)},
			destination:    types.Integer,
			expectedResult: IntValue(32),
		},
		convertTest{
			input:       parseBigInt("295842908756029385409213854092745689237409823490127835873895723098754"),
			destination: types.Integer,
			expectError: true,
		},
		convertTest{
			input:          parseBigInt("295842908756029385409213854092745689237409823490127835873895723098754"),
			destination:    types.String,
			expectedResult: StringValue("295842908756029385409213854092745689237409823490127835873895723098754"),
		},
	}

	for _, test := range cases {
		var res LispValue
		var err error

		res, err = test.input.Convert(test.destination)

		if (err != nil) != test.expectError {
			t.Errorf("Error converting BigInt to %s: %s",
				test.destination.String(),
				err.Error())
		} else if !test.expectError && !res.Eq(test.expectedResult) {
			t.Errorf("Error converting BigInt to %s: Expected %s, got %s",
				test.destination.String(),
				test.expectedResult.String(),
				res.String())
		}
	}
} // func TestBigInt(t *testing.T)

func TestArray(t *testing.T) {
	type arrayTest struct {
		input          Array
		destination    types.ID
		expectedResult string
		expectError    bool
	}

	var testCases = []arrayTest{
		arrayTest{
			input:          Array{IntValue(1), IntValue(2), IntValue(3)},
			destination:    types.List,
			expectedResult: "(1 2 3)",
		},
	}

	for _, test := range testCases {
		var res LispValue
		var err error
		var str string

		if res, err = test.input.Convert(test.destination); err != nil {
			if !test.expectError {
				t.Errorf("Error converting Array to %s: %s",
					test.destination.String(),
					err.Error())
			}
		} else if str = res.String(); str != test.expectedResult {
			t.Errorf("Unexpected result from conversion of Array to %s: %s (expected %s)",
				test.destination.String(),
				str,
				test.expectedResult)
		}
	}
} // func TestArray(t *testing.T)

func TestList(t *testing.T) {
	type listTest struct {
		input          LispValue
		destination    types.ID
		expectedResult string
		expectedType   types.ID
		expectedError  bool
	}

	var testCases = []listTest{
		listTest{
			input: &List{
				Length: 3,
				Car: &ConsCell{
					Car: IntValue(1),
					Cdr: &ConsCell{
						Car: IntValue(2),
						Cdr: &ConsCell{
							Car: IntValue(3),
						},
					},
				},
			},
			destination:    types.Array,
			expectedResult: "[1 2 3]",
			expectedType:   types.Array,
		},
	}

	for _, test := range testCases {
		var (
			err error
			val LispValue
			res string
		)

		if val, err = test.input.Convert(test.destination); err != nil {
			if !test.expectedError {
				t.Errorf("Error evaluating APPLY!-form %s: %s",
					test.input,
					err.Error())
			}
		} else if val == nil {
			t.Error("Result from conversion of list is nil!")
		} else if val.Type() != test.expectedType {
			t.Errorf("Unexpected type for result: %s (expected %s)",
				val.Type().String(),
				test.expectedType.String())
		} else if res = val.String(); res != test.expectedResult {
			t.Errorf("Unexpected result from converting List to %s: %s",
				test.destination,
				res)
		}
	}
} // func TestList(t *testing.T)
