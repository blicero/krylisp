// /home/krylon/go/src/github.com/blicero/krylisp/interpreter/02_interpreter_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 17. 02. 2025 by Benjamin Walkenhorst
// (c) 2025 Benjamin Walkenhorst
// Time-stamp: <2025-02-17 15:52:52 krylon>

package interpreter

import (
	"testing"

	"github.com/blicero/krylisp/parser"
)

var in = Interpreter{
	Env: &Environment{
		Bindings: map[parser.Symbol]parser.LispValue{
			sym("karl"):       parser.String{Str: "Otto"},
			sym("the-answer"): parser.Integer{Int: 42},
		},
	},
	Debug: true,
}

func TestEvalSimple(t *testing.T) {
	type testCase struct {
		input          parser.LispValue
		expectError    bool
		expectedResult parser.LispValue
	}

	var cases = []testCase{
		{
			input:          parser.Integer{Int: 262144},
			expectedResult: parser.Integer{Int: 262144},
		},
		{
			input:          parser.String{Str: "Wer das liest, ist doof."},
			expectedResult: parser.String{Str: "Wer das liest, ist doof."},
		},
	}

	for _, c := range cases {
		var (
			err error
			res parser.LispValue
		)

		if res, err = in.Eval(c.input); err != nil {
			if !c.expectError {
				t.Errorf("Unexpected error evaluating %s: %s",
					c.input,
					err.Error())
			}
		} else if !res.Equal(c.expectedResult) {
			t.Errorf("Unexpected result evaluating %s: %s (expected %s)",
				c.input,
				res,
				c.expectedResult)

		}
	}
} // func TestEvalSimple(t *testing.T)

func TestEvalSymbol(t *testing.T) {
	type testCase struct {
		input          parser.LispValue
		expectError    bool
		expectedResult parser.LispValue
	}

	var cases = []testCase{
		{
			input:          sym("t"),
			expectedResult: sym("t"),
		},
		{
			input:          sym("nil"),
			expectedResult: sym("nil"),
		},
		{
			input:          sym(":hello"),
			expectedResult: sym(":hello"),
		},
		{
			input:          sym("the-answer"),
			expectedResult: parser.Integer{Int: 42},
		},
		{
			input:       sym("horst"),
			expectError: true,
		},
	}

	for _, c := range cases {
		var (
			err error
			res parser.LispValue
		)

		if res, err = in.Eval(c.input); err != nil {
			if !c.expectError {
				t.Errorf("Unexpected error evaluating %s: %s",
					c.input,
					err.Error())
			}
		} else if !res.Equal(c.expectedResult) {
			t.Errorf("Unexpected result evaluating %s: %s (expected %s)",
				c.input,
				res,
				c.expectedResult)

		}
	}
} // func TestEvalSymbol(t *testing.T)
