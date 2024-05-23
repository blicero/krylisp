// /home/krylon/go/src/krylisp/parserutil/parserutil.go
// -*- mode: go; coding: utf-8; -*-
// Created on 07. 09. 2017 by Benjamin Walkenhorst
// (c) 2017 Benjamin Walkenhorst
// Time-stamp: <2024-05-23 09:52:32 krylon>
//
// Donnerstag, 07. 09. 2017, 17:51
// I am going to need some way of handling errors properly.
// For now, I just call panic, but I think gocc provides some feature
// for dealing with errors that is hopefully a little more elegant.

// Package parserutil defines utility functions used by the parser.
package parserutil

import (
	"math/big"
	"regexp"
	"strconv"

	"github.com/blicero/krylisp/value"
)

var erangePat = regexp.MustCompile("value out of range")

// IntValue attempts to parse a string into an integer value.
// If the s contains a valid number that happens to be outside the range
// int64 can represent, a BigInt is returned instead.
func IntValue(s string) value.Number {
	if n, err := strconv.ParseInt(s, 10, 64); err != nil {
		if erangePat.MatchString(err.Error()) {
			return Bignum(s)
		}

		panic(err)
	} else {
		return value.IntValue(n)
	}
} // func IntValue(s string) value.Number

// FloatValue attempts to parse a string into a floating point value
func FloatValue(s string) value.Number {
	if f, err := strconv.ParseFloat(s, 64); err != nil {
		panic(err)
	} else {
		return value.FloatValue(f)
	}
} // func FloatValue(s string) value.Number

// Bignum attempts to parse a string into a BigInt value
func Bignum(s string) value.Number {
	if b, e := value.BigIntFromString(s[:len(s)-1]); e != nil {
		panic(e)
	} else {
		return b
	}
} // func Bignum(s string) value.Number

// OctalNumber parses a number in octal notation from a string.
func OctalNumber(s string) value.Number {
	if n, err := strconv.ParseInt(s, 8, 64); err != nil {
		if erangePat.MatchString(err.Error()) {
			var ok bool
			var biggie = new(value.BigInt)

			biggie.Value = big.NewInt(0)
			if _, ok = biggie.Value.SetString(s, 8); !ok {
				panic(s)
			} else {
				return biggie
			}
		} else {
			panic(err)
		}
	} else {
		return value.IntValue(n)
	}
} // func OctalNumber(s string) value.Number

// HashAdd adds a key-value-pair to a Hashtable and returns the table, multiple calls
// can be chained.
func HashAdd(tbl value.Hashtable, key, val value.LispValue) value.Hashtable {
	tbl[key] = val
	return tbl
} // func HashAdd(tbl map[value.LispValue]value.LispValue, key, val value.LispValue) map[value.LispValue]value.LispValue

// StringValue takes a string token and return a Lisp string,
// minus the double quotes.
func StringValue(s []byte) value.StringValue {
	return value.StringValue(s[1 : len(s)-1])
} // func StringValue(s []byte) value.StringValue
