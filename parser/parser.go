// /home/krylon/go/src/github.com/blicero/krylisp/parser/parser.go
// -*- mode: go; coding: utf-8; -*-
// Created on 21. 12. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2025-02-19 19:37:53 krylon>

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

// ConsCell is the basic building block of Lisp Lists.
type ConsCell struct {
	Car LispValue `parser:"@"`
	Cdr *ConsCell `parser:"@@*"`
}

func (c ConsCell) String() string {
	if c.Cdr != nil {
		return fmt.Sprintf("(%s . %s)",
			c.Car,
			c.Cdr)
	}

	return fmt.Sprintf("(%s . NIL)",
		c.Car)
} // func (c ConsCell) String() string

// Type returns the type of the receiver.
func (c ConsCell) Type() types.Type {
	return types.ConsCell
} // func (c ConsCell) Type() types.Type

// Equal compares the receiver to the given LispValue for equality.
func (c ConsCell) Equal(other LispValue) bool {
	switch o := other.(type) {
	case ConsCell:
		if c.Car.Equal(o.Car) {
			if c.Cdr == nil {
				return o.Cdr == nil
			}
			return c.Cdr.Equal(o.Cdr)
		}
		return false
	default:
		return false
	}
} // func (c ConsCell) Equal(other LispValue) bool

// List is a Lisp list.
type List struct {
	//Items []LispValue `parser:"OpenParen @@* CloseParen"`
	Items ConsCell `parser:"( @@LispValue @@ConsCell* )"`
}

// Type returns the type of the receiver.
func (l List) Type() types.Type { return types.List }

func (l List) String() string {
	var (
		sb   strings.Builder
		cons *ConsCell = l.Items.Cdr
	)

	sb.WriteString("(")
	sb.WriteString(l.Items.String())

	for cons != nil {
		sb.WriteString(" ")
		sb.WriteString(cons.Car.String())
		cons = cons.Cdr
	}

	sb.WriteString(")")

	// for i, v := range l.Items {
	// 	items[i] = v.String()
	// }

	// sb.WriteString("(")
	// sb.WriteString(strings.Join(items, " "))
	// sb.WriteString(")")

	return sb.String()
} // func (l List) String() string

// Length returns the length of the receiver.
func (l List) Length() int {
	var (
		cnt  = 1
		cons = l.Items.Cdr
	)

	for cons != nil {
		cnt++
		cons = cons.Cdr
	}

	return cnt
} // func (l List) Length() int

// Equal compares the receiver to the given LispValue for equality.
func (l List) Equal(other LispValue) bool {
	switch o := other.(type) {
	// case Symbol:
	// 	return len(l.Items) == 0 && o.Sym == "NIL"
	case List:
		if l.Length() != o.Length() {
			return false
		} else if !l.Items.Car.Equal(o.Items.Car) {
			return false
		}

		var c1, c2 = l.Items.Cdr, o.Items.Cdr

		for c1 != nil {
			if !c1.Car.Equal(c2.Car) {
				return false
			}
			c1 = c1.Cdr
			c2 = c2.Cdr
		}

		return c1 == nil && c2 == nil

		// for idx, val := range l.Items {
		// 	if !val.Equal(o.Items[idx]) {
		// 		return false
		// 	}
		// }

		// return true
	default:
		return false
	}
} // func (l List) Equal(other LispValue) bool
