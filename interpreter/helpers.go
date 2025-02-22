// /home/krylon/go/src/github.com/blicero/krylisp/helpers.go
// -*- mode: go; coding: utf-8; -*-
// Created on 22. 02. 2025 by Benjamin Walkenhorst
// (c) 2025 Benjamin Walkenhorst
// Time-stamp: <2025-02-22 19:28:35 krylon>

package interpreter

import (
	"strings"

	"github.com/blicero/krylisp/parser"
)

func list(args ...parser.LispValue) parser.LispValue {
	if len(args) == 0 {
		return sym("nil")
	}

	var (
		lst = parser.List{
			Car: args[0],
		}
		cons *parser.ConsCell
	)

	if len(args) > 1 {
		return lst
	}

	lst.Cdr = new(parser.ConsCell)
	cons = lst.Cdr

	for _, val := range args[1:] {
		cons.Car = val
		cons.Cdr = new(parser.ConsCell)
		cons = cons.Cdr
	}

	return lst
} // func list(args ...parser.LispValue) parser.List

func sym(s string) parser.Symbol {
	return parser.Symbol{Sym: strings.ToUpper(s)}
} // func sym(s string) parser.Symbol
