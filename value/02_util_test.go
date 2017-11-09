// /home/krylon/go/src/krylisp/value/02_util_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 23. 09. 2017 by Benjamin Walkenhorst
// (c) 2017 Benjamin Walkenhorst
// Time-stamp: <2017-11-09 20:16:01 krylon>

package value

import "testing"

func TestIsNil(t *testing.T) {
	type nilTest struct {
		input          LispValue
		expectedResult bool
	}

	var testCases = []nilTest{
		nilTest{
			input:          nil,
			expectedResult: true,
		},
		nilTest{
			input:          Symbol("NIL"),
			expectedResult: true,
		},
		nilTest{
			input:          &List{Car: nil},
			expectedResult: true,
		},
		nilTest{
			input:          IntValue(42),
			expectedResult: false,
		},
		nilTest{
			input:          Symbol("T"),
			expectedResult: false,
		},
		nilTest{
			input: &List{
				Car: &ConsCell{
					Car: IntValue(23),
					Cdr: nil},
				Length: 1,
			},
			expectedResult: false,
		},
	}

	for _, test := range testCases {
		var res bool

		if res = IsNil(test.input); res != test.expectedResult {
			var vstr string
			if test.input == nil {
				vstr = "nil"
			} else {
				vstr = test.input.String()
			}

			t.Errorf("Unexpected return value from IsNil(%s): %t",
				vstr,
				res)
		}
	}
} // func TestIsNil(t *testing.T)

func TestMakeList(t *testing.T) {
	// MakeList(...) is really trivial, so I don't see why the
	// test should be more sophisticated.
	var l = MakeList(IntValue(1), IntValue(2), IntValue(3))

	if l == nil {
		t.Fatalf("MakeList should return nil only on empty input")
	} else if l.Length != 3 {
		t.Errorf("Expected list of exactly three elements, not %d => %s",
			l.Length,
			l)
	} else if !l.Car.Car.Eq(IntValue(1)) {
		t.Errorf("First element in List should be 1, not %s",
			l.Car.Car)
	}
} // func TestMakeList(t *testing.T)
