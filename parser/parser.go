// /home/krylon/go/src/github.com/blicero/krylisp/parser/parser.go
// -*- mode: go; coding: utf-8; -*-
// Created on 21. 12. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2025-02-13 18:39:02 krylon>

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
	{Name: `Symbol`, Pattern: `[-+a-zA-Z][\w\d]*`},
	{Name: `Integer`, Pattern: `\d+`},
	{Name: `String`, Pattern: `"(?:[^\"]*)"`},
	{Name: `OpenParen`, Pattern: `\(`},
	{Name: `CloseParen`, Pattern: `\)`},
	{Name: `Blank`, Pattern: `\s+`},
})

func New() *participle.Parser[LispValue] {
	par := participle.MustBuild[LispValue](
		participle.Lexer(lex),
		participle.Unquote("String"),
		participle.Elide("Blank"),
		participle.Union[LispValue](Symbol{}, Integer{}, String{}, List{}),
	)

	return par
} // func New() *participle.Parser[LispValue]

type LispValue interface {
	fmt.Stringer
	Type() types.Type
}

type Symbol struct {
	Sym string `parser:"@Symbol"`
}

func (s Symbol) Type() types.Type { return types.Symbol }

func (s Symbol) String() string {
	return strings.ToUpper(s.Sym)
}

type Integer struct {
	Int int64 `parser:"@Integer"`
}

func (i Integer) Type() types.Type { return types.Integer }

func (i Integer) String() string {
	return strconv.FormatInt(i.Int, 10)
}

type String struct {
	Str string `parser:"@String"`
}

func (s String) Type() types.Type { return types.String }

func (s String) String() string {
	return `"` + s.Str + `"`
}

type List struct {
	Items []LispValue `parser:"OpenParen @@* CloseParen"`
}

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
