// /home/krylon/go/src/krylisp/value.go
// -*- mode: go; coding: utf-8; -*-
// Created on 06. 09. 2017 by Benjamin Walkenhorst
// (c) 2017 Benjamin Walkenhorst
// Time-stamp: <2017-09-07 14:12:23 krylon>

package value

import (
	"fmt"
	"krylisp/types"
	"strconv"
)

// LispValue is the "abstract base class", so to speak, for Lisp data.
// All values usable in kryLisp internally implement this interface.
type LispValue interface {
	Type() types.ID
	String() string
}

// IntValue is an integer.
// At this point, a signed, 64-bit integer is the only numeric type supported.
// Eventually, I want to add support for bignums and floating point numbers,
// possibly more (rational numbers would be nice).
type IntValue int64

// Type returns the type ID of the Lisp value, in this case
// types.Number
func (i IntValue) Type() types.ID {
	return types.Number
} // function (i IntValue) Type() types.ID

// String returns a string representation of the Lisp value.
// Since this is a number: We use base 10
// I think in Common Lisp, it is possible to change - at runtime! - the
// default base for displaying numbers.
// While it would be cool to implement such a feature, I really do not see
// the use case for this.
func (i IntValue) String() string {
	//return strconv.Itoa(i)
	return strconv.FormatInt(int64(i), 10)
} // func (i IntValue) String() string

// StringValue is a string. Strings are implemented in terms of Go strings, so
// the same rules and restrictions apply: Strings are encoded in UTF-8 and
// immutable.
type StringValue string

// Type returns returns the type ID of the Lisp Value
// (in this case, types.String)
func (s StringValue) Type() types.ID {
	return types.String
} // func (s StringValue) Type() types.ID

// String returns a string representation of the Lisp value.
func (s StringValue) String() string {
	return string(s)
} // func (s StringValue) String() string

// ConsCell is a pair of two Lisp values, used mainly for constructing lists.
type ConsCell struct {
	Car LispValue
	Cdr LispValue
}

// Type returns the type ID of the Lisp value.
// (In this case types.ConsCell)
func (s *ConsCell) Type() types.ID {
	return types.ConsCell
} // func (s ConsCell) Type() types.ID

// String returns a string representation of the Lisp value.
func (s *ConsCell) String() string {
	var s1, s2 string

	if s.Car == nil {
		s1 = "nil"
	} else {
		s1 = s.Car.String()
	}

	if s.Cdr == nil {
		s2 = "nil"
	} else {
		s2 = s.Cdr.String()
	}

	return fmt.Sprintf("(%s . %s)",
		s1,
		s2)
} // func (s ConsCell) String() string

// IsList returns true if the receiver is a proper list.
// Semantically, this should work the same as listp in Common Lisp.
func (s *ConsCell) IsList() bool {
	var cell = s

	for cell != nil {
		// I wonder what is more efficient - calling the Type method
		// or using a type switch.
		if cell.Cdr == nil {
			return true
		} else if cell.Cdr.Type() != types.ConsCell {
			return false
		} else {
			// fmt.Println(spew.Sdump(cell.Cdr))
			// fmt.Printf("Type of CDR is %s\n", cell.Cdr.Type().String())
			cell = cell.Cdr.(*ConsCell)
		}
	}

	// Hier sollten wir nur ankommen, wenn s selbst schon nil ist.
	// Ich weiß gar nicht, ob das in Go erlaubt ist. In C++ ist das im
	// Prinzip erlaubt, meine ich, aber ... das ist auch einer der Gründe,
	// aus denen ich um C++ nach Kräften einen Bogen mache.
	return true
} // func (s *ConsCell) IsList() bool
