// /home/krylon/go/src/github.com/blicero/krylisp/interpreter/02_interpreter_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 17. 02. 2025 by Benjamin Walkenhorst
// (c) 2025 Benjamin Walkenhorst
// Time-stamp: <2025-02-25 15:08:38 krylon>

package interpreter

import (
	"fmt"
	"testing"

	"github.com/blicero/krylisp/common"
	"github.com/blicero/krylisp/logdomain"
	"github.com/blicero/krylisp/parser"
)

var in = Interpreter{
	Env: &environment{
		scope: &scope{
			bindings: map[parser.Symbol]parser.LispValue{
				sym("karl"):       parser.String{Str: "Otto"},
				sym("the-answer"): parser.Integer{Int: 42},
			},
		},
	},
	Debug: true,
}

func init() {
	var err error

	if in.log, err = common.GetLogger(logdomain.Interpreter); err != nil {
		panic(fmt.Errorf("Failed to create Logger for Interpreter: %s",
			err.Error()))
	}
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

func TestEvalList(t *testing.T) {
	type testCase struct {
		input       parser.LispValue
		result      parser.LispValue
		expectError bool
	}

	var cases = []testCase{
		{
			input:  parser.List{},
			result: sym("nil"),
		},
		{
			input:  list(sym("+"), parser.Integer{Int: 1}, parser.Integer{Int: 2}, parser.Integer{Int: 3}),
			result: parser.Integer{Int: 6},
		},
		{
			input:  list(sym("+")),
			result: parser.Integer{Int: 0},
		},
		{
			input: list(
				sym("if"),
				list(sym("null"), sym("the-answer")),
				sym(":falsch"),
				sym(":richtig")),
			result: sym(":richtig"),
		},
		{
			input: list(
				sym("defun"),
				sym("zerop"),
				list(sym("x")),
				list(sym("="),
					sym("x"),
					parser.Integer{Int: 0})),
			result: sym("zerop"),
		},
	}

	for _, c := range cases {
		var (
			err error
			res parser.LispValue
		)

		if res, err = in.Eval(c.input); err != nil {
			if !c.expectError {
				t.Errorf("Failed to evaluate input %T %q: %s",
					c.input,
					c.input,
					err.Error())
			}
		} else if !res.Equal(c.result) {
			t.Errorf(`Evaluating input %q produced unexpected result:
Expected: %s
Got:      %s`,
				c.input,
				c.result,
				res)
		}
	}
} // func TestEvalList(t *testing.T)
