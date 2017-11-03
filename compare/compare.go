// /home/krylon/go/src/krylisp/compare/compare.go
// -*- mode: go; coding: utf-8; -*-
// Created on 31. 10. 2017 by Benjamin Walkenhorst
// (c) 2017 Benjamin Walkenhorst
// Time-stamp: <2017-11-01 21:13:44 krylon>

// Package compare contains the results of a comparison between two values.
// I am not entirely sure if I am not pushing this a little too far, but
// I kind of liked the idea, so here we are.
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
