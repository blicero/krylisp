// /home/krylon/go/src/github.com/blicero/krylisp/helpers.go
// -*- mode: go; coding: utf-8; -*-
// Created on 22. 02. 2025 by Benjamin Walkenhorst
// (c) 2025 Benjamin Walkenhorst
// Time-stamp: <2025-02-23 21:34:13 krylon>

package interpreter

import (
	"fmt"
	"strings"

	"github.com/blicero/krylisp/common"
	"github.com/blicero/krylisp/parser"
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
if
lambda
>
<
=
+
eq
eql
cond
and
or
not
car
cdr
cons
apply
let
while
defun
defmacro
set!
quote
var
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
