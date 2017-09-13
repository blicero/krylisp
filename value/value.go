// /home/krylon/go/src/krylisp/value.go
// -*- mode: go; coding: utf-8; -*-
// Created on 06. 09. 2017 by Benjamin Walkenhorst
// (c) 2017 Benjamin Walkenhorst
// Time-stamp: <2017-09-13 19:25:29 krylon>
//
// Donnerstag, 07. 09. 2017, 17:33
// Aus ... Gründen, werden im Paket types nur die symbolischen Konstanten
// für die verschiedenen Typen definiert, die mein Lisp-Interpreter später
// verstehen soll.
// Hier werden die eigentlichen Datentypen definiert, die mein Interpreter
// dann verwendet um Lisp-Daten darzustellen.

package value

import (
	"bytes"
	"errors"
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
	Bool() bool
	Eq(other LispValue) bool
}

// A NilValue represents nil, the strange list-symbol duality.
type NilValue int

// Type returns the type ID of the Lisp value, in this case types.Nil
func (n NilValue) Type() types.ID {
	return types.Nil
} // func (n NilValue) Type() types.ID

// String returns a string representation of the Lisp value.
func (n NilValue) String() string {
	return "NIL"
} // func (n NilValue) String() string

// Bool returns the "truthiness" of a Lisp value.
func (n NilValue) Bool() bool {
	return false
} // func (n NilValue) Bool() bool

// Eq compares the receiver with the argument for identity.
func (n NilValue) Eq(other LispValue) bool {
	if other == nil {
		return true
	} else if other.Type() == types.List {
		var l = other.(*List)

		if l.Length == 0 || l.Car == nil {
			return true
		}
	} else if other.Type() == types.Nil {
		return true
	}

	return false
} // func (n NilValue) Eq(other LispValue) bool

// NIL is the canonical nil value. In theory, we could get away with just
// having a single value.
const NIL NilValue = 1

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

// Bool returns the "truthiness" of a Lisp value.
func (i IntValue) Bool() bool {
	return true
} // func (i IntValue) Bool() bool

// Eq compares the receiver with the argument for identity.
func (i IntValue) Eq(other LispValue) bool {
	if other == nil {
		return false
	} else if other.Type() != types.Number {
		return false
	}

	return i == other.(IntValue)
} // func (i IntValue) Eq(other LispValue) bool

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

// Bool returns the "truthiness" of a Lisp value.
func (s StringValue) Bool() bool {
	return true
} // func (s StringValue) Bool() bool

// Eq compares the receiver with the argument for identity.
func (s StringValue) Eq(other LispValue) bool {
	if other == nil {
		return false
	} else if other.Type() != types.String {
		return false
	}

	return s == other.(StringValue)
} // func (s StringValue) Eq(other LispValue) bool

// I am not sure if should represent symbols as plain strings.
// But for now I cannot think of a good reason not to.
// If I were to intern symbols so I only need to compare
// hash codes or something, it might make sense. For now,
// I have no clue, yet, how well or badly my Lisp interpreter
// is going to perform, so I will use a plain string.

// Symbol represents a Lisp symbol.
type Symbol string

// Type returns the value's type ID, in this case types.Symbol
func (s Symbol) Type() types.ID {
	return types.Symbol
} // func (s Symbol) Type() types.ID

// String returns a string representation of the Lisp value.
func (s Symbol) String() string {
	return string(s)
} // func (s Symbol) String() string

// Bool returns the "truthiness" of a Lisp value.
func (s Symbol) Bool() bool {
	return s != "NIL"
} // func (s Symbol) Bool() bool

// Eq compares the receiver with the argument for identity.
func (s Symbol) Eq(other LispValue) bool {
	if other == nil {
		return s == "NIL"
	}

	switch v := other.(type) {
	case NilValue:
		return s == "NIL"
	case *List:
		return (s == "NIL") && (v.Length == 0 || v.Car == nil)
	case Symbol:
		return s == v
	default:
		return false
	}
} // func (s Symbol) Eq(other LispValue) bool

// IsKeyword returns true if the symbol gets special treatment by the
// interpreter.
// Usually, this is either because some primitives need to be implemented
// outside Lisp, or because of efficiency considerations.
func (s Symbol) IsKeyword() bool {
	return s[0] == ':'
} // func (s Symbol) IsKeyword() bool

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
} // func (s *ConsCell) String() string

// Bool returns the "truthiness" of a Lisp value.
func (s *ConsCell) Bool() bool {
	return true
} // func (s *ConsCell) Bool() bool

// IsList returns true if the receiver is a proper list.
// Semantically, this should work the same as listp in Common Lisp.
func (s *ConsCell) IsList() bool {
	var cell = s

	for cell != nil {
		// I wonder what is more efficient - calling the Type method
		// or using a type switch.
		if cell.Cdr == NIL || cell.Cdr == nil {
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

// ActualLength returns the length of a cons'ed structure.
func (s *ConsCell) ActualLength() int {
	if s == nil {
		return 0
	} else if s.Car == nil && s.Cdr == nil {
		return 0
	}

	var cnt = 0
	var cell = s

	for cell != nil {
		cnt++
		if cell.Cdr != NIL {
			switch v := cell.Cdr.(type) {
			case *ConsCell:
				cell = v
			default:
				cell = nil
			}
		} else {
			cell = nil
		}
	}

	return cnt
} // func (s *ConsCell) ActualLength() int

// Eq compares the receiver with the argument for identity.
func (s *ConsCell) Eq(other LispValue) bool {
	// Do I compare for equality (i.e. equivalent values) or identity?
	// My first thought was to have Equal be the equivalent to Lisp's eq
	// operator, which does compare for identity, except for integers and
	// floats.
	// ... I'll have this method be Eq, once I have implemented it for all types,
	// I will rename it to Eq for clarity.
	if other == nil {
		return s == nil
	} else if c, ok := other.(*ConsCell); ok && s == c {
		return true
	}

	return false
} // func (s *ConsCell) Eq(other LispValue) bool

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
		return "NIL"
	} else if l.Car == nil {
		return "NIL"
	}

	var elements = make([]string, l.Length)
	var idx = 0
	var cell = l.Car

	for ; cell != nil; idx++ {
		if cell.Car != nil {
			elements[idx] = cell.Car.String()
		} else {
			elements[idx] = "NIL"
		}

		if !(cell.Cdr == nil || cell.Cdr == NIL) {
			cell = cell.Cdr.(*ConsCell)
		} else {
			cell = nil
		}
	}

	return "(" + strings.Join(elements, " ") + ")"
} // func (l *List) String() string

// Bool returns the "truthiness" of a Lisp value.
func (l *List) Bool() bool {
	return l.Car != nil
} // func (l *List) Bool() bool

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
		return NIL
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

// IsLambda returns true if the given list is a lambda list.
// This method recognizes any method as a lambda list whose
// first element is the symbol lambda, and whose second element is a list.
func (l *List) IsLambda() bool {
	if l.Car == nil || l.Car.Car == nil {
		return false
	}

	return l.Car.Car.Type() == types.Symbol &&
		l.Car.Car.(Symbol) == Symbol("LAMBDA") &&
		l.Car.Cdr.(*ConsCell).Car.Type() == types.List
} // func (l *List) IsLambda() bool

// Eq compares the receiver with the argument for identity.
func (l *List) Eq(other LispValue) bool {
	if other == nil {
		return false
	} else if other.Type() != types.List {
		return false
	}

	return l.Car == other.(*List).Car
} // func (l *List) Eq(other LispValue) bool

// Eq return true if the receiver and the argument are the same, i.e. if both
// lists' Car member points to the same ConsCell.
// func (l *List) Eq(other *List) bool {
// 	return l.Car == other.Car
// } // func (l *List) Eq(other *List) bool

// ActualLength counts the number of elements in the list and returns the
// result. The main purpose is for debugging/testing, but it is also used
// in the parser.
func (l *List) ActualLength() int {
	if l == nil {
		return 0
	} else if l.Car == nil {
		return 0
	}

	var cnt = 0
	var cell = l.Car

	for cell != nil {
		cnt++
		if !(cell.Cdr == nil || cell.Cdr == NIL) {
			cell = cell.Cdr.(*ConsCell)
		} else {
			cell = nil
		}
	}

	l.Length = cnt

	return cnt
} // func (l *List) ActualLength() int

// Nth returns the nth element of a List.
func (l *List) Nth(n int) (LispValue, error) {
	if l == nil {
		return NIL, errors.New("Receiver is NIL")
	}
	if n < 0 || n >= l.Length {
		return NIL, fmt.Errorf("Index is out of range: %d (valid: [0 - %d])",
			n,
			l.Length-1)
	}

	var elt = l.Car

	for idx := 0; idx < n; idx++ {
		elt = elt.Cdr.(*ConsCell)
	}

	return elt.Car, nil
} // func (l *List) Nth(n int) (LispValue, error)

// Function represents a Lisp function.
// Technically, one could implement functions purely in terms of lists,
// but for efficiency reasons - and functions are used a *lot* in Lisp,
// obviously - they get their own type.
type Function struct {
	Env  *Environment
	Args []Symbol
	Body []LispValue
}

// Type returns the type ID of the value, in this case types.Function
func (f *Function) Type() types.ID {
	return types.Function
} // func (f *Function) Type() types.ID

// String returns a string representation of the Lisp value.
func (f *Function) String() string {
	var out bytes.Buffer
	var argCnt = len(f.Args)

	out.WriteString("(lambda (")
	for i := 0; i < argCnt; i++ {
		out.WriteString(f.Args[i].String())
		if i < argCnt-1 {
			out.WriteString(" ")
		}
	}
	out.WriteString(") ")

	for _, exp := range f.Body {
		out.WriteString(exp.String())
		out.WriteString("\n")
	}

	out.WriteString(")")
	return out.String()
} // func (f *Function) String() string

// Bool returns the "truthiness" of a Lisp value.
func (f *Function) Bool() bool {
	return true
} // func (f *Function) Bool() bool

// Eq compares the receiver with the argument for identity.
func (f *Function) Eq(other LispValue) bool {
	if other == nil {
		return false
	} else if v, ok := other.(*Function); ok && v == f {
		return true
	}

	return false
} // func (f *Function) Eq(other LispValue) bool
