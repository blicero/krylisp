// /home/krylon/go/src/krylisp/value/errors.go
// -*- mode: go; coding: utf-8; -*-
// Created on 22. 10. 2017 by Benjamin Walkenhorst
// (c) 2017 Benjamin Walkenhorst
// Time-stamp: <2017-10-22 15:16:34 krylon>

package value

import (
	"fmt"
	"krylisp/types"
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
