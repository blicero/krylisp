// /home/krylon/go/src/krylisp/interpreter/errors.go
// -*- mode: go; coding: utf-8; -*-
// Created on 09. 09. 2017 by Benjamin Walkenhorst
// (c) 2017 Benjamin Walkenhorst
// Time-stamp: <2017-09-12 17:20:07 krylon>

package interpreter

import (
	"fmt"
	"krylisp/value"
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
	return fmt.Sprintf("No function definition wss found for %s", sym)
} // var (sym MissingFunctionError) Error string

// TypeError indicates a type mismatch between expected and provided values.
type TypeError struct {
	expected string
	actual   string
}

// Error returns the error message.
func (te *TypeError) Error() string {
	return fmt.Sprintf("Type %s it not allowed (expected %s)",
		te.actual,
		te.expected)
} // func (te *TypeError) Error() string

// ValueError indicates that a value was of the correct type, but that value
// still was not permitted in that context. Think if division by zero, for
// example.
// Or, say, the height of a human, measured in centimeters - this could obviously
// never be negative.
type ValueError struct {
	val value.LispValue
}

// Error returns the error message.
func (ve *ValueError) Error() string {
	return fmt.Sprintf("Value is not permitted in this context: %v",
		ve.val)
} // func (ve *ValueError) Error() string

// SyntaxError indicates invalid use of special forms.
type SyntaxError string

// Error returns the error message.
func (se SyntaxError) Error() string {
	return fmt.Sprintf("Invalid Syntax: %s", se)
} // func (se SyntaxError) Error() string
