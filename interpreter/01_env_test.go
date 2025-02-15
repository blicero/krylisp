// /home/krylon/go/src/github.com/blicero/krylisp/interpreter/01_env_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 15. 02. 2025 by Benjamin Walkenhorst
// (c) 2025 Benjamin Walkenhorst
// Time-stamp: <2025-02-15 15:56:51 krylon>

package interpreter

import (
	"strings"
	"testing"

	"github.com/blicero/krylisp/parser"
)

func sym(s string) parser.Symbol {
	return parser.Symbol{Sym: strings.ToUpper(s)}
}

func TestLookupSimple(t *testing.T) {
	type testCase struct {
		key           parser.Symbol
		expectOK      bool
		expectedValue parser.LispValue
	}

	var env = Environment{
		Bindings: map[parser.Symbol]parser.LispValue{
			sym("zero"): parser.Integer{Int: 0},
		},
	}

	var cases = []testCase{
		{key: sym("ZAPPELWURST"), expectOK: false},
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

}
