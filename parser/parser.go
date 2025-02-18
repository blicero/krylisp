// /home/krylon/go/src/github.com/blicero/krylisp/parser/parser.go
// -*- mode: go; coding: utf-8; -*-
// Created on 21. 12. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2025-02-18 14:54:16 krylon>

// Package parser provides the ... parser.
package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/blicero/krylisp/types"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

var lex = lexer.MustSimple([]lexer.SimpleRule{
	{Name: `Symbol`, Pattern: `[-+*/%:a-zA-Z][-+*/%:a-zA-Z\d]*`},
	{Name: `Integer`, Pattern: `\d+`},
	{Name: `String`, Pattern: `"(?:[^\"]*)"`},
	{Name: `OpenParen`, Pattern: `\(`},
	{Name: `CloseParen`, Pattern: `\)`},
	{Name: `Blank`, Pattern: `\s+`},
	{Name: `Quote`, Pattern: `'`},
})

// New creates a new Parser.
func New() *participle.Parser[LispValue] {
	par := participle.MustBuild[LispValue](
		participle.Lexer(lex),
		participle.Unquote("String"),
		participle.Elide("Blank"),
		participle.Upper("Symbol"),
		participle.Union[LispValue](Symbol{}, Integer{}, String{}, List{}),
	)

	return par
} // func New() *participle.Parser[LispValue]

// LispValue is the common interface for types in the Lisp Interpreter
type LispValue interface {
	fmt.Stringer
	Type() types.Type
	Equal(other LispValue) bool
}

// Symbol is a symbol.
type Symbol struct {
	Sym string `parser:"@Symbol"`
}

// Type returns the type of the Symbol.
func (s Symbol) Type() types.Type { return types.Symbol }
func (s Symbol) String() string   { return s.Sym }

// IsKeyword returns true if the receiver is a keyword symbol.
func (s Symbol) IsKeyword() bool { return s.Sym[0] == ':' }

// Equal compares the receiver to another LispValue for equality.
func (s Symbol) Equal(other LispValue) bool {
	// TODO Make sure NIL equals ()
	switch val := other.(type) {
	case Symbol:
		return s.Sym == val.Sym
	default:
		return false
	}
} // func (s Symbol) Equal(other LispValue) bool

// Integer is a signed 64-bit integer
type Integer struct {
	Int int64 `parser:"@Integer"`
}

// Type returns the type of the receiver
func (i Integer) Type() types.Type { return types.Integer }
func (i Integer) String() string   { return strconv.FormatInt(i.Int, 10) }

// Equal compares the receiver to the given LispValue for equality.
func (i Integer) Equal(other LispValue) bool {
	switch o := other.(type) {
	case Integer:
		return i.Int == o.Int
	default:
		return false
	}
} // func (i Integer) Equal(other LispValue) bool

// String is a ... string.
type String struct {
	Str string `parser:"@String"`
}

// Type returns the type of the receiver.
func (s String) Type() types.Type { return types.String }
func (s String) String() string   { return `"` + s.Str + `"` }

// Equal compares the receiver to the given LispValue for equality.
func (s String) Equal(other LispValue) bool {
	switch o := other.(type) {
	case String:
		return s.Str == o.Str
	default:
		return false
	}
} // func (s String) Equal(other LispValue) bool

// FIXME Lists shall be made of ConsCells (which I still need to implement), not slices!!!

// List is a Lisp list.
type List struct {
	Items []LispValue `parser:"OpenParen @@* CloseParen"`
}

// Type returns the type of the receiver.
func (l List) Type() types.Type { return types.List }

func (l List) String() string {
	var (
		sb    strings.Builder
		items = make([]string, len(l.Items))
	)

	for i, v := range l.Items {
		items[i] = v.String()
	}

	sb.WriteString("(")
	sb.WriteString(strings.Join(items, " "))
	sb.WriteString(")")

	return sb.String()
}

// Equal compares the receiver to the given LispValue for equality.
func (l List) Equal(other LispValue) bool {
	switch o := other.(type) {
	case Symbol:
		return len(l.Items) == 0 && o.Sym == "NIL"
	case List:
		if len(l.Items) != len(o.Items) {
			return false
		}

		for idx, val := range l.Items {
			if !val.Equal(o.Items[idx]) {
				return false
			}
		}

		return true
	default:
		return false
	}
} // func (l List) Equal(other LispValue) bool
