// /home/krylon/go/src/krylisp/types/types.go
// -*- mode: go; coding: utf-8; -*-
// Created on 06. 09. 2017 by Benjamin Walkenhorst
// (c) 2017 Benjamin Walkenhorst
// Time-stamp: <2017-09-15 18:46:00 krylon>

package types

//go:generate stringer -type=ID

// ID is the numeric identifier for the type of a Lisp value.
type ID uint8

// The constants' names should be self-explanatory.
const (
	Nil ID = iota
	Number
	String
	Symbol
	KeySym
	ConsCell
	List
	Function
	Program
)
