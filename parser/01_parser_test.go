// /home/krylon/go/src/github.com/blicero/krylisp/parser/01_test_parser.go
// -*- mode: go; coding: utf-8; -*-
// Created on 11. 02. 2025 by Benjamin Walkenhorst
// (c) 2025 Benjamin Walkenhorst
// Time-stamp: <2025-02-13 15:49:43 krylon>

package parser

import (
	"testing"

	"github.com/alecthomas/participle/v2"
)

var par *participle.Parser[LispValue]

func TestCreateParser(t *testing.T) {
	var err error

	if par, err = participle.Build[LispValue](
		participle.Lexer(lex),
		participle.Unquote("String"),
		participle.Elide("Blank"),
		participle.Union[LispValue](Symbol{}, Integer{}, String{}, List{}),
	); err != nil {
		par = nil
		t.Fatalf("Failed to create Parser: %s", err.Error())
	} else if par == nil {
		t.Fatal("Parser is nil!")
	}

} // func TestCreateParser(t *testing.T)

func TestParse(t *testing.T) {
	if par == nil {
		t.SkipNow()
	}

	type testCase struct {
		filename    string
		expr        string
		expectError bool
	}

	var samples = []testCase{
		{filename: "symbol", expr: "T"},
		{filename: "integer", expr: "42"},
		{filename: "string", expr: `"Wer das liest, ist doof."`},
		{filename: "list", expr: `(zebu alpha 23 69 "Hulululu")`},
	}

	for _, s := range samples {
		var (
			err error
			val *LispValue
		)

		if val, err = par.ParseString(s.filename, s.expr); err != nil {
			if !s.expectError {
				t.Errorf("Failed to parse %s: %s",
					s.filename,
					err.Error())
			}
		} else if val == nil {
			t.Errorf("Parsed value of %s is nil", s.filename)
		} else {
			t.Logf("Parsing %s yielded a %s: %s",
				s.filename,
				(*val).Type(),
				*val)
		}
	}
} // func TestParse(t *testing.T)
