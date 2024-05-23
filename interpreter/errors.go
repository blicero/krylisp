// /home/krylon/go/src/krylisp/interpreter/errors.go
// -*- mode: go; coding: utf-8; -*-
// Created on 09. 09. 2017 by Benjamin Walkenhorst
// (c) 2017 Benjamin Walkenhorst
// Time-stamp: <2024-05-23 09:55:24 krylon>

package interpreter

import (
	"fmt"

	"github.com/blicero/krylisp/types"
	"github.com/blicero/krylisp/value"
)

// NoBindingError indicates that no binding was found for a symbol.
type NoBindingError struct {
	sym value.Symbol
}

// Error returns the error message.
func (b *NoBindingError) Error() string {
	return fmt.Sprintf("No binding was found for symbol %s",
		b.sym)
} // func (b *NoBindingError) Error() string

// MissingFunctionError signals that a function is being called that the
// interpreter does not know. Since I decided to have separate namespaces
// for functions and everything else, This is separate from NoBindingError.
type MissingFunctionError value.Symbol

// Error returns the error message.
func (sym MissingFunctionError) Error() string {
	return fmt.Sprintf("No function definition wss found for %s", string(sym))
} // var (sym MissingFunctionError) Error string

// ValueError indicates that a value was of the correct type, but that value
// still was not permitted in that context. Think if division by zero, for
// example.
// Or, say, the height of a human, measured in centimeters - this could obviously
// never be negative.
//
// Montag, 06. 11. 2017, 17:46
// I added a second field to hold a string. The value by itself might not be
// very helpful.
type ValueError struct {
	val value.LispValue
	msg string
}

// Error returns the error message.
func (ve *ValueError) Error() string {
	if ve.msg == "" {
		return fmt.Sprintf("Value is not permitted in this context: %v",
			ve.val)
	}

	return fmt.Sprintf("Value %v is not permitted in this context: %s",
		ve.val,
		ve.msg)
} // func (ve *ValueError) Error() string

// SyntaxError indicates invalid use of special forms.
type SyntaxError string

// SyntaxErrorf creates a formatted string out of the given arguments
// and returns that value as a SyntaxError.
func SyntaxErrorf(format string, args ...interface{}) SyntaxError {
	var msg = fmt.Sprintf(format, args...)
	return SyntaxError(msg)
} // func SyntaxErrorf(fmt string, ...args interface{}) SyntaxError

// Error returns the error message.
func (se SyntaxError) Error() string {
	return "Invalid syntax: " + string(se)
} // func (se SyntaxError) Error() string

// TypePromotionError indicates an unresolvable situation with automatic
// type promotion.
type TypePromotionError struct {
	inputLeft  types.ID
	inputRight types.ID
}

// Error returns the error message.
func (tpe *TypePromotionError) Error() string {
	return fmt.Sprintf("No type promotion rule was found for the combination of %s and %s",
		tpe.inputLeft.String(),
		tpe.inputRight.String())
} // func (tpe *TypePromotionError) Error() string
