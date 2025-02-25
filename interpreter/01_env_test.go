// /home/krylon/go/src/github.com/blicero/krylisp/interpreter/01_env_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 15. 02. 2025 by Benjamin Walkenhorst
// (c) 2025 Benjamin Walkenhorst
// Time-stamp: <2025-02-25 15:27:31 krylon>

package interpreter

import (
	"testing"

	"github.com/blicero/krylisp/parser"
)

func TestLookupSimple(t *testing.T) {
	type testCase struct {
		key           parser.Symbol
		expectOK      bool
		expectedValue parser.LispValue
	}

	var env = environment{
		scope: &scope{
			bindings: map[parser.Symbol]parser.LispValue{
				sym("zero"): parser.Integer{Int: 0},
				sym("name"): parser.String{Str: "Odysseus"},
				sym("age"):  parser.Integer{Int: 42},
			},
		},
	}

	var cases = []testCase{
		{key: sym("zappelwurst"), expectOK: false},
		{key: sym("zero"), expectOK: true, expectedValue: parser.Integer{Int: 0}},
		{key: sym("name"), expectOK: true, expectedValue: parser.String{Str: "Odysseus"}},
	}

	for _, c := range cases {
		var (
			val parser.LispValue
			ok  bool
		)

		if val, ok = env.Lookup(c.key); !ok {
			if c.expectOK {
				t.Errorf("Lookup of Symbol %s in test environment failed",
					c.key)
				continue
			}
		} else if !val.Equal(c.expectedValue) {
			t.Errorf("Lookup returned unexpected value %s (expected %s)",
				val,
				c.expectedValue)
		}
	}

} // func TestLookupSimple(t *testing.T)

func TestLookupScoped(t *testing.T) {
	var env = environment{
		scope: &scope{
			bindings: map[parser.Symbol]parser.LispValue{
				sym("y"):    parser.Integer{Int: 23},
				sym("temp"): parser.Integer{Int: 109},
			},
			parent: &scope{
				bindings: map[parser.Symbol]parser.LispValue{
					sym("x"):      parser.Integer{Int: 10},
					sym("temp"):   parser.Integer{Int: 42},
					sym("lambda"): parser.Integer{Int: 83},
				},
			},
		},
	}

	type testCase struct {
		key           parser.Symbol
		expectOK      bool
		expectedValue parser.LispValue
	}

	var cases = []testCase{
		{
			key:           sym("y"),
			expectOK:      true,
			expectedValue: parser.Integer{Int: 23},
		},
		{
			key:           sym("x"),
			expectOK:      true,
			expectedValue: parser.Integer{Int: 10},
		},
		{
			key:           sym("temp"),
			expectOK:      true,
			expectedValue: parser.Integer{Int: 109},
		},
		{
			key: sym("nothing"),
		},
	}

	for _, c := range cases {
		var (
			ok  bool
			val parser.LispValue
		)

		if val, ok = env.Lookup(c.key); !ok {
			if c.expectOK {
				t.Errorf("Failed to lookup %s in environment",
					c.key)
				continue
			}
		} else if !val.Equal(c.expectedValue) {
			t.Errorf("Unexpected value for %s: %s (expected %s)",
				c.key,
				val,
				c.expectedValue)
		}
	}
} // func TestLookupScoped(t *testing.T)
