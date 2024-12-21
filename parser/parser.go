// /home/krylon/go/src/github.com/blicero/krylisp/parser/parser.go
// -*- mode: go; coding: utf-8; -*-
// Created on 21. 12. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2024-12-21 23:03:09 krylon>

// Package parser provides the ... parser.
package parser

import "github.com/blicero/krylisp/types"

type LispValue interface {
	Type() types.Type
}
