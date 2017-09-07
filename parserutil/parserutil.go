// /home/krylon/go/src/krylisp/parserutil/parserutil.go
// -*- mode: go; coding: utf-8; -*-
// Created on 07. 09. 2017 by Benjamin Walkenhorst
// (c) 2017 Benjamin Walkenhorst
// Time-stamp: <2017-09-07 17:55:07 krylon>
//
// Donnerstag, 07. 09. 2017, 17:51
// I am going to need some way of handling errors properly.
// For now, I just call panic, but I think gocc provides some feature
// for dealing with errors that is hopefully a little more elegant.

// Package parserutil defines utility functions used by the parser.
package parserutil

import (
	"krylisp/value"
	"strconv"
)

// IntValue attempts to parse a string into an integer value.
func IntValue(s string) value.IntValue {
	if n, err := strconv.Atoi(s); err != nil {
		panic(err)
	} else {
		return value.IntValue(n)
	}
} // func IntValue(s string) int
