// /home/krylon/go/src/krylisp/types/types.go
// -*- mode: go; coding: utf-8; -*-
// Created on 06. 09. 2017 by Benjamin Walkenhorst
// (c) 2017 Benjamin Walkenhorst
// Time-stamp: <2017-12-11 18:51:25 krylon>

// Package types implements symbolic constants for kryLisp's types.
package types

//go:generate stringer -type=ID

// ID is the numeric identifier for the type of a Lisp value.
type ID uint8

// The constants' names should be self-explanatory.
const (
	Nil ID = iota
	Number
	Integer
	Float
	BigInt
	String
	Regexp
	Symbol
	KeySym
	ConsCell
	List
	Array
	Hashtable
	Function
	GoFunction
	Program
	FileHandle
	Error
	Macro
)
