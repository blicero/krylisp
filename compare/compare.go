// /home/krylon/go/src/krylisp/compare/compare.go
// -*- mode: go; coding: utf-8; -*-
// Created on 31. 10. 2017 by Benjamin Walkenhorst
// (c) 2017 Benjamin Walkenhorst
// Time-stamp: <2017-10-31 21:36:04 krylon>

// Package compare contains the results of a comparison between two values.
package compare

//go:generate stringer -type=Result

// Result is the result of a comparison of two values of a fully
// ordered type.
type Result int8

// Undefined is placeholder value that will (hopefully) be mostly used in
// debugging.
// The other constants should be self-explanatory.
const (
	Undefined Result = iota
	LessThan
	Equal
	GreaterThan
)
