// /home/krylon/go/src/krylisp/value/util.go
// -*- mode: go; coding: utf-8; -*-
// Created on 23. 09. 2017 by Benjamin Walkenhorst
// (c) 2017 Benjamin Walkenhorst
// Time-stamp: <2017-10-25 20:11:08 krylon>

package value

import "krylisp/types"

// IsNil returns true if the given argument is considered a nil value.
// (Yeah, sorry, there's a confusing multitude of values that are considered
// nil...)
func IsNil(v LispValue) bool {
	if v == nil {
		return true
	} else if v == NIL {
		return true
	} else if v.Type() == types.Nil {
		return true
	} else if sym, ok := v.(Symbol); ok && sym == "NIL" {
		return true
	} else if lst, ok := v.(*List); ok && (lst.Length == 0 || lst.Car == nil) {
		return true
	}

	return false
} // func IsNil(v LispValue) bool

// IsNumber checks if the given value belongs to one of the numeric types.
func IsNumber(v LispValue) bool {
	if v == nil {
		return false
	}

	switch v.Type() {
	case types.Integer, types.Float:
		return true
	default:
		return false
	}
} // func IsNumber(v LispValue) bool
