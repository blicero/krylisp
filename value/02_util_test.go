// /home/krylon/go/src/krylisp/value/02_util_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 23. 09. 2017 by Benjamin Walkenhorst
// (c) 2017 Benjamin Walkenhorst
// Time-stamp: <2017-09-23 13:17:40 krylon>

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
