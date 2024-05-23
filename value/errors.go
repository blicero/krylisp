// /home/krylon/go/src/krylisp/value/errors.go
// -*- mode: go; coding: utf-8; -*-
// Created on 22. 10. 2017 by Benjamin Walkenhorst
// (c) 2017 Benjamin Walkenhorst
// Time-stamp: <2024-05-23 09:55:05 krylon>

package value

import (
	"fmt"

	"github.com/blicero/krylisp/types"
)

// TypeConversionError indicates a failed attempt to convert a LispValue to a
// different type.
type TypeConversionError struct {
	source      types.ID
	destination types.ID
}

// Error returns the error message.
func (tce *TypeConversionError) Error() string {
	return fmt.Sprintf("Invalid type conversion: %s -> %s",
		tce.source,
		tce.destination)
}

// TypeError indicates a type mismatch between expected and provided values.
type TypeError struct {
	Expected string
	Actual   string
}

// Error returns the error message.
func (te *TypeError) Error() string {
	return fmt.Sprintf("Type %s is not allowed (expected %s)",
		te.Actual,
		te.Expected)
} // func (te *TypeError) Error() string
