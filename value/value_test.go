// /home/krylon/go/src/krylisp/value/value_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 07. 09. 2017 by Benjamin Walkenhorst
// (c) 2017 Benjamin Walkenhorst
// Time-stamp: <2017-09-07 14:08:16 krylon>

package value

import (
	"krylisp/types"
	"testing"
)

func TestTypeID(t *testing.T) {
	type testValue struct {
		input        LispValue
		expectedType types.ID
	}

	var values = []testValue{
		testValue{
			input:        IntValue(42),
			expectedType: types.Number,
		},
		testValue{
			input:        StringValue("Wer das liest, ist doof."),
			expectedType: types.String,
		},
		testValue{
			input:        &ConsCell{nil, nil},
			expectedType: types.ConsCell,
		},
	}

	for idx, v := range values {
		if v.input.Type() != v.expectedType {
			t.Errorf("Wrong type ID for testValue %d: %s (expected %s)",
				idx,
				v.input.Type().String(),
				v.expectedType.String())

		}
	}
} // func TestTypeID(t *testing.T)

func TestString(t *testing.T) {
	type testValue struct {
		input    LispValue
		expected string
	}

	var testCases = []testValue{
		testValue{
			IntValue(42),
			"42",
		},
		testValue{
			StringValue("Wer das liest, ist doof."),
			"Wer das liest, ist doof.",
		},
		testValue{
			&ConsCell{IntValue(23), IntValue(42)},
			"(23 . 42)",
		},
		testValue{
			&ConsCell{IntValue(64), nil},
			"(64 . nil)",
		},
	}

	for idx, val := range testCases {
		var s = val.input.String()

		if s != val.expected {
			t.Errorf("Unexpected string value for input %d: Expected \"%s\", got \"%s\"",
				idx,
				val.expected,
				s)
		}
	}
} // func TestString(t *testing.T)

func TestIsList(t *testing.T) {
	type testValue struct {
		input    *ConsCell
		expected bool
	}

	var testData = []testValue{
		testValue{
			&ConsCell{nil, nil},
			true,
		},
		testValue{ // ???
			nil,
			true,
		},
		testValue{
			&ConsCell{
				Car: IntValue(42),
				Cdr: &ConsCell{
					Car: IntValue(64),
					Cdr: IntValue(128),
				},
			},
			false,
		},
	}

	for idx, val := range testData {
		var res = val.input.IsList()

		if res != val.expected {
			t.Errorf("Error in test data #%d: Expected %t, got %t",
				idx,
				val.expected,
				res)
		}
	}
} // func TestIsList(t *testing.T)
