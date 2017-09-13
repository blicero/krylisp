// /home/krylon/go/src/krylisp/ast/ast_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 08. 09. 2017 by Benjamin Walkenhorst
// (c) 2017 Benjamin Walkenhorst
// Time-stamp: <2017-09-09 15:36:55 krylon>

package ast

import (
	"fmt"
	"krylisp/lexer"
	"krylisp/parser"
	"krylisp/types"
	"krylisp/value"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

const separator = "##########################################################################"

func printSep() {
	fmt.Println(separator)
} // function printSep()

var p *parser.Parser

func TestCreateParser(t *testing.T) {
	if p = parser.NewParser(); p == nil {
		t.Error("NewParser returned nil")
	}
} // func TestCreateParser(t *testing.T)

func TestParseAtom(t *testing.T) {
	type parseAtomTest struct {
		input          string
		expectedType   types.ID
		expectedString string
	}

	if p == nil {
		t.Skip()
	}

	var testData = []parseAtomTest{
		parseAtomTest{
			input:          "+",
			expectedType:   types.Symbol,
			expectedString: "+",
		},
		parseAtomTest{
			input:          "3",
			expectedType:   types.Number,
			expectedString: "3",
		},
		parseAtomTest{
			input:          "peter",
			expectedType:   types.Symbol,
			expectedString: "PETER",
		},
		parseAtomTest{
			input:          `"Peter"`,
			expectedType:   types.String,
			expectedString: `"Peter"`,
		},
		parseAtomTest{
			input:          `"Wer das liest, ist doof."`,
			expectedType:   types.String,
			expectedString: `"Wer das liest, ist doof."`,
		},
	}

	for _, test := range testData {
		//p = parser.NewParser()
		lex := lexer.NewLexer([]byte(test.input))
		result, err := p.Parse(lex)

		if err != nil {
			t.Errorf("Error parsing input (%s): %s",
				test.input,
				err.Error())
			printSep()
		} else {
			spew.Dump(result)
		}

		var prog []value.LispValue
		var val value.LispValue
		var dumpString string
		var ok bool

		if prog, ok = result.([]value.LispValue); !ok {
			t.Errorf("Invalid type for parse result: %T (expected value.LispValue)",
				result)
		} else if len(prog) != 1 {
			t.Errorf("Wrong length for program: Expected one element, got %d",
				len(prog))
		} else if val = prog[0]; val.Type() != test.expectedType {
			t.Errorf("Unexpected type for Atom [%s]: %s (expected %s)",
				test.input,
				val.Type(),
				test.expectedType)
		} else if dumpString = val.String(); dumpString != test.expectedString {
			t.Errorf("String representation for parsed input [%s] is wrong: %s (expected %s)",
				test.input,
				dumpString,
				test.expectedString)
		}

		printSep()
	}
} // func TestParseAtom(t *testing.T)

func TestParseList(t *testing.T) {
	type parseListTest struct {
		input          string
		expectedType   types.ID
		expectedLength int
		expectedString string
	}

	var testCases = []parseListTest{
		parseListTest{
			input:          "(1 2 3)",
			expectedType:   types.List,
			expectedLength: 3,
			expectedString: "(1 2 3)",
		},
		// Ich sollte vielleicht wirklich drüber nachdenken, nil als ausdrücklichen
		// Wert zu behandeln.
		parseListTest{
			input:          "()",
			expectedType:   types.List,
			expectedLength: 0,
			expectedString: "nil",
		},
		parseListTest{
			input:          "( 2    5     17     29    )   ",
			expectedType:   types.List,
			expectedLength: 4,
			expectedString: "(2 5 17 29)",
		},
		parseListTest{
			input:          "(2 3 (4 5 (6 7)) 8 9)",
			expectedType:   types.List,
			expectedLength: 5,
			expectedString: "(2 3 (4 5 (6 7)) 8 9)",
		},
	}
	var err error

	printSep()
	printSep()

	for _, test := range testCases {
		var lex = lexer.NewLexer([]byte(test.input))
		var result interface{}
		var prog []value.LispValue
		var ok bool
		var sval string

		if result, err = p.Parse(lex); err != nil {
			t.Errorf("Error parsing input [%s]: %s",
				test.input,
				err.Error())
		} else if prog, ok = result.([]value.LispValue); !ok {
			t.Errorf("Parsing input [%s] yielded unexpected type: %T (%s)",
				test.input,
				result,
				spew.Sdump(result))
		} else if len(prog) != 1 {
			t.Errorf("Unexpected program length: %d elements (expected 1)",
				len(prog))
		} else if prog[0].Type() != test.expectedType {
			t.Errorf("Unexpected type for input [%s]: %s (expected %s)",
				test.input,
				prog[0].Type(),
				test.expectedType)
		} else if sval = prog[0].String(); sval != test.expectedString {
			t.Errorf("Unexpected string representation for input [%s]: [%s] (expected [%s])",
				test.input,
				sval,
				test.expectedString)
		}

		printSep()
	}
} // func TestParseList(t *testing.T)
