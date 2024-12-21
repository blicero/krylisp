// /home/krylon/go/src/github.com/blicero/krylisp/types/types.go
// -*- mode: go; coding: utf-8; -*-
// Created on 21. 12. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2024-12-21 21:59:42 krylon>

// Package types provides symbolic constants to identify the types used in kryLisp
package types

// Type identifies the type of a Lisp value
type Type uint8

const (
	Symbol Type = iota
	String
	Integer
	Float
	ConsCell
)