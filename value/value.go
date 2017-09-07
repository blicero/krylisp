// /home/krylon/go/src/krylisp/value.go
// -*- mode: go; coding: utf-8; -*-
// Created on 06. 09. 2017 by Benjamin Walkenhorst
// (c) 2017 Benjamin Walkenhorst
// Time-stamp: <2017-09-07 17:38:46 krylon>
//
// Donnerstag, 07. 09. 2017, 17:33
// Aus ... Gründen, werden im Paket types nur die symbolischen Konstanten
// für die verschiedenen Typen definiert, die mein Lisp-Interpreter später
// verstehen soll.
// Hier werden die eigentlichen Datentypen definiert, die mein Interpreter
// dann verwendet um Lisp-Daten darzustellen.

package value

import (
	"fmt"
	"krylisp/types"
	"strconv"
	"strings"
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
	return `"` + string(s) + `"`
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

// List is ... well, a singly-linked list, the kind that is so common in Lisp
// they named the language after it.
// We define a separate type for lists because there are some distinctions
// between a simple CONS cell and a proper list.
// If Go was an object-oriented language, we could implement this as a subclass
// of ConsCell. Whatever.
// While we're at it, we can stuff some other fields into the struct that might
// come in handy. Length, for example. Although I am not 100% certain if it is a
// good idea to rely on that. ???
type List struct {
	Car    *ConsCell
	Length int
}

// Type returns the type ID of the receiver.
// (In this case, types.List
func (l *List) Type() types.ID {
	return types.List
} // func (l *List) Type() types.ID

// String returns a string representation of the Lisp value.
func (l *List) String() string {
	if l == nil {
		return "nil"
	} else if l.Car == nil {
		return "nil"
	}

	var elements = make([]string, l.Length)
	var idx = 0
	var cell = l.Car

	for ; cell != nil; idx++ {
		if cell.Car != nil {
			elements[idx] = cell.Car.String()
		} else {
			elements[idx] = "nil"
		}

		if cell.Cdr != nil {
			cell = cell.Cdr.(*ConsCell)
		} else {
			cell = nil
		}
	}

	return "(" + strings.Join(elements, " ") + ")"
} // func (l *List) String() string

// Cdr returns a new List instance that is the CDR of the receiver.
func (l *List) Cdr() *List {
	return &List{
		l.Car.Cdr.(*ConsCell),
		l.Length - 1,
	}
} // func (l *List) Cdr() *List

// Push adds a LispValue to the front of the List.
// Semantically, this method should work like CONS in Common Lisp.
func (l *List) Push(v LispValue) *List {
	var cdr = l.Car
	l.Car = &ConsCell{
		v,
		cdr,
	}
	l.Length++

	return l
} // func (l *List) Push(v LispValue) *List

// Pop removes the first element of the List destructively and
// returns it.
func (l *List) Pop() LispValue {
	// XXX I should check for null-ness and such.
	if l.Car == nil {
		return nil
	}

	var car = l.Car.Car

	if l.Car.Cdr != nil {
		l.Car = l.Car.Cdr.(*ConsCell)
		l.Length--
	} else {
		l.Length = 0
		l.Car = nil
	}

	return car
} // func (l *List) Pop() LispValue

// Eq return true if the receiver and the argument are the same, i.e. if both
// lists' Car member points to the same ConsCell.
func (l *List) Eq(other *List) bool {
	return l.Car == other.Car
} // func (l *List) Eq(other *List) bool

// I am not sure if should represent symbols as plain strings.
// But for now I cannot think of a good reason not to.
// If I were to intern symbols so I only need to compare
// hash codes or something, it might make sense. For now,
// I have no clue, yet, how well or badly my Lisp interpreter
// is going to perform, so I will use a plain string.

// Symbol represents a Lisp symbol.
type Symbol string

func (s Symbol) Type() types.ID {
	return types.Symbol
} // func (s Symbol) Type() types.ID

func (s Symbol) String() string {
	return string(s)
} // func (s Symbol) String() string
