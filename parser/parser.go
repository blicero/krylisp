// /home/krylon/go/src/github.com/blicero/krylisp/parser/parser.go
// -*- mode: go; coding: utf-8; -*-
// Created on 21. 12. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2024-12-22 01:29:29 krylon>

// Package parser provides the ... parser.
package parser

import "github.com/blicero/krylisp/types"

type LispValue interface {
	Type() types.Type
}

type Symbol struct {
	Symbol string `parser:"@Symbol"`
}

func (s *Symbol) Type() types.Type { return types.Symbol }

type String struct {
	String string `parser:"@String"`
}

func (s *String) Type() types.Type { return types.String }

type Integer struct {
	Integer int64 `parser:"@Integer"`
}

func (i *Integer) Type() types.Type { return types.Integer }

type Float struct {
	Float float64 `parse:"@Float"`
}

func (f *Float) Type() types.Type { return types.Float }

type ConsCell struct {
	Car LispValue `parse:"@Car"`
	Cdr LispValue `parse:"@Cdr"`
}
