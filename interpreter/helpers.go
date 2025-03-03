// /home/krylon/go/src/github.com/blicero/krylisp/helpers.go
// -*- mode: go; coding: utf-8; -*-
// Created on 22. 02. 2025 by Benjamin Walkenhorst
// (c) 2025 Benjamin Walkenhorst
// Time-stamp: <2025-03-02 16:50:51 krylon>

package interpreter

import (
	"fmt"
	"strings"

	"github.com/blicero/krylisp/common"
	"github.com/blicero/krylisp/parser"
	"github.com/blicero/krylisp/types"
)

func list(args ...parser.LispValue) parser.LispValue {
	fmt.Printf("Make List of %d elements\n", len(args))

	if len(args) == 0 {
		return sym("nil")
	}

	var (
		lst = parser.List{
			Car: args[0],
		}
		head, cons *parser.ConsCell
	)

	if len(args) == 1 {
		return lst
	}

	head = new(parser.ConsCell)
	cons = head

	// for idx, val := range args[1:] {
	// 	cons.Car = val
	// 	if idx < len(args)-1 {
	// 		cons.Cdr = new(parser.ConsCell)
	// 	}

	// 	cons = cons.Cdr
	// }

	for i := 1; i < len(args)-1; i++ {
		cons.Car = args[i]
		cons.Cdr = new(parser.ConsCell)
		cons = cons.Cdr
	}

	cons.Car = args[len(args)-1]

	lst.Cdr = head

	return lst
} // func list(args ...parser.LispValue) parser.List

func sym(s string) parser.Symbol {
	return parser.Symbol{Sym: strings.ToUpper(s)}
} // func sym(s string) parser.Symbol

const specialFormList = `
%
*
+
-
/
<
=
>
and
apply
car
cdr
cond
cons
defmacro
defun
eq
eql
if
lambda
let
list
not
null
or
quote
set!
var
while
`

var specialForms map[string]bool

func init() {
	var symbols = common.WhiteSpace.Split(specialFormList, -1)

	specialForms = make(map[string]bool, len(symbols))

	for _, s := range symbols {
		specialForms[strings.ToUpper(s)] = true
	}
} // func init()

func isSpecial(sym fmt.Stringer) bool {
	var s = sym.String()
	return specialForms[s]
} // func isSpecial(sym fmt.Stringer) bool

func asBool(val parser.LispValue) bool {
	if val == nil {
		return false
	}

	switch act := val.(type) {
	case parser.Symbol:
		return act.Sym != "NIL"
	case parser.List:
		return act.Length() != 0
	default:
		return true
	}
} // func asBool(val parser.LispValue) bool

// Function represents a function. I'm using a special type for these,
// because I'll be handling those a lot, and I want to make that a bit less
// painful.
type Function struct {
	name      string
	docString string
	argList   []parser.LispValue
	body      *parser.ConsCell
}

func (f *Function) String() string {
	var (
		sb   strings.Builder
		args = make([]string, len(f.argList))
	)

	for i, v := range f.argList {
		args[i] = v.String()
	}

	sb.WriteString("(LAMBDA (")
	sb.WriteString(strings.Join(args, " "))
	sb.WriteString(")")

	var c = f.body

	for c != nil {
		sb.WriteString("\n\t")
		sb.WriteString(c.Car.String())
		c = c.Cdr
	}

	sb.WriteString(")")

	return sb.String()
} // func (f *Function) String() string

// Type returns the type of the receiver, i.e. types.Function
func (f *Function) Type() types.Type { return types.Function }

// Equal compares the receiver to another LispValue for equality.
func (f *Function) Equal(other parser.LispValue) bool {
	var (
		fn *Function
		ok bool
	)

	if other == nil {
		return f == nil
	} else if fn, ok = other.(*Function); !ok {
		return false
	} else if fn == f {
		return true
	} else if len(f.argList) != len(fn.argList) {
		return false
	}

	for i, s := range f.argList {
		if !fn.argList[i].Equal(s) {
			return false
		}
	}

	return f.name == fn.name &&
		f.body.Equal(fn.body)
} // func (f *Function) Equal(other parser.LispValue) bool
