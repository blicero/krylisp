// /home/krylon/go/src/krylisp/ast/ast_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 08. 09. 2017 by Benjamin Walkenhorst
// (c) 2017 Benjamin Walkenhorst
// Time-stamp: <2017-11-04 19:38:23 krylon>

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
			expectedType:   types.Integer,
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
		parseAtomTest{
			input:          "3.141592",
			expectedType:   types.Float,
			expectedString: "3.141592",
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
			expectedString: "NIL",
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

func TestNumberType(t *testing.T) {
	type numTest struct {
		input          string
		shouldBeNumber bool
		expectedType   types.ID
	}

	var testCases = []numTest{
		numTest{
			input:          "2",
			shouldBeNumber: true,
			expectedType:   types.Integer,
		},
		numTest{
			input:          "3.141592",
			shouldBeNumber: true,
			expectedType:   types.Float,
		},
		numTest{
			input:          "\"Hallo\"",
			shouldBeNumber: false,
			expectedType:   types.String,
		},
		numTest{
			input:          "    +   ",
			shouldBeNumber: false,
			expectedType:   types.Symbol,
		},
		// Donnerstag, 26. 10. 2017, 13:32
		// New tests for BigInt
		numTest{
			input:          "64b",
			shouldBeNumber: true,
			expectedType:   types.BigInt,
		},
		numTest{
			input:          "23059823905838769287450938450960928409384502758273049283509283",
			shouldBeNumber: true,
			expectedType:   types.BigInt,
		},
	}

	for _, test := range testCases {
		var tree interface{}
		var err error
		var prog value.Program
		var ok bool
		var p = parser.NewParser()
		var l = lexer.NewLexer([]byte(test.input))
		//var n Number

		if tree, err = p.Parse(l); err != nil {
			t.Errorf("Error parsing test case %q: %s",
				test.input,
				err.Error())
		} else if prog, ok = tree.([]value.LispValue); !ok {
			t.Fatalf("Parser did not return a Program, but a %T",
				tree)
		} else if _, ok = prog[0].(value.Number); test.shouldBeNumber && !ok {
			t.Errorf("Parser did not return a Number, but a %T (calls itself %s)",
				prog[0],
				prog[0].Type().String())
		} else if test.expectedType != prog[0].Type() {
			t.Errorf("Parser returned an unexpected type: %s (expected %s)",
				test.expectedType.String(),
				prog[0].Type().String())
		}
	}
} // func TestNumberType(t *testing.T)

func TestArray(t *testing.T) {
	type arrayTest struct {
		input          string
		expectedResult []value.LispValue
	}

	var testCases = []arrayTest{
		arrayTest{
			input: "[ 1 2 3 ]",
			expectedResult: []value.LispValue{
				value.IntValue(1),
				value.IntValue(2),
				value.IntValue(3),
			},
		},
		arrayTest{
			input:          "[]",
			expectedResult: []value.LispValue{},
		},
	}

	for _, test := range testCases {
		var tree interface{}
		var err error
		var prog value.Program
		var ok bool
		var p = parser.NewParser()
		var l = lexer.NewLexer([]byte(test.input))
		var arr value.Array

		if tree, err = p.Parse(l); err != nil {
			t.Errorf("Error parsing test case %q: %s",
				test.input,
				err.Error())
		} else if prog, ok = tree.([]value.LispValue); !ok {
			t.Fatalf("Parser did not return a Program, but a %T",
				tree)
		} else if arr, ok = prog[0].(value.Array); !ok {
			t.Errorf("Parser did not return an Array, but a %T",
				prog[0])
		} else if len(arr) != len(test.expectedResult) {
			t.Errorf("Array has unexpected length: %d (expected %d)",
				len(arr),
				len(test.expectedResult))
		} else {
			for idx, val := range arr {
				if !val.Eq(test.expectedResult[idx]) {
					t.Errorf("Unexpected value at position #%d: %s (expected %s)",
						idx+1,
						val.String(),
						test.expectedResult[idx].String())
				}
			}
		}
	}
} // func TestArray(t *testing.T)

func TestHash(t *testing.T) {
	type hashTest struct {
		input          string
		expectedResult value.LispValue
		expectedError  bool
	}

	var testCases = []hashTest{
		hashTest{
			input: `{ 1 : "Peter", 2 : "Karl", 3: 3.141592 }`,
			expectedResult: value.Hashtable{
				value.IntValue(1): value.StringValue("Peter"),
				value.IntValue(2): value.StringValue("Karl"),
				value.IntValue(3): value.FloatValue(3.141592),
			},
			expectedError: false,
		},
		// I would like to support trailing commas in Hashtables, but
		// right now (Samstag, 04. 11. 2017, 03:01) I cannot get the
		// grammar right so gocc plays along.
		hashTest{
			input: `{ 1: Peter, 2: Karl, 3: Pi, }`,
			expectedResult: value.Hashtable{
				value.IntValue(1): value.Symbol("PETER"),
				value.IntValue(2): value.Symbol("KARL"),
				value.IntValue(3): value.Symbol("PI"),
			},
			expectedError: true,
		},
		hashTest{
			input: `{ Peter : Karl, Willi : 2, Horst : "Lieber nicht" }`,
			expectedResult: value.Hashtable{
				value.Symbol("PETER"): value.Symbol("KARL"),
				value.Symbol("WILLI"): value.IntValue(2),
				value.Symbol("HORST"): value.StringValue("Lieber nicht"),
			},
		},
	}

	for idx, test := range testCases {
		var (
			tree interface{}
			err  error
			prog value.Program
			ok   bool
			p    = parser.NewParser()
			l    = lexer.NewLexer([]byte(test.input))
			ht   value.Hashtable
		)

		if tree, err = p.Parse(l); err != nil {
			if !test.expectedError {
				t.Errorf("Hashtable: Error parsing test #%d (%s): %s",
					idx,
					test.input,
					err.Error())
			}
		} else if prog, ok = tree.([]value.LispValue); !ok {
			t.Errorf("Parser did not return a Program, but a %T",
				tree)
		} else if len(prog) != 1 {
			t.Errorf("Parsed program has unexpected length: %d (expected 1)",
				len(prog))
		} else if ht, ok = prog[0].(value.Hashtable); !ok {
			t.Errorf("Parsed value shoud be a Hashtable, not a %T",
				prog[0])
		} else if !test.expectedResult.Eq(ht) {
			t.Errorf("Parsed hashtable is not what we expected: %s (expected %s)",
				ht.String(),
				test.expectedResult.String())
		}
	}
} // func TestHash(t *testing.T)
