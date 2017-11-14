// /home/krylon/go/src/krylisp/value.go
// -*- mode: go; coding: utf-8; -*-
// Created on 06. 09. 2017 by Benjamin Walkenhorst
// (c) 2017 Benjamin Walkenhorst
// Time-stamp: <2017-11-13 21:27:16 krylon>
//
// Donnerstag, 07. 09. 2017, 17:33
// Aus ... Gründen, werden im Paket types nur die symbolischen Konstanten
// für die verschiedenen Typen definiert, die mein Lisp-Interpreter später
// verstehen soll.
// Hier werden die eigentlichen Datentypen definiert, die mein Interpreter
// dann verwendet um Lisp-Daten darzustellen.
//
// Freitag, 20. 10. 2017, 18:03
// Die Logik für Arithmetik-Gedöns in den Methoden der numerischen Typen
// unterzubringen ist keine so gute Idee, fällt mir auf, weil ich dann
// zum Hinzufügen eine Typen *alle* bisher existenten Typen anpacken muss
// und das eine Menge Fleißarbeit wird.
// Aber ohne generische Typen und Operatoren-Überladung bleibt mir nicht
// viel übrig, oder?
//
// Donnerstag, 09. 11. 2017, 22:08
// I do not remember why, but I decided to finally add an Equal-method to
// the LispValue interface. Now that I have that, I think I could make
// Eq stricter. Right now, Eq checks for structural equality on most
// types.

package value

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"krylib"
	"krylisp/permission"
	"krylisp/types"
	"math"
	"math/big"
	"nosy/common"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

///////////////////////////////////////////////////////////////////////
// General stuff //////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////

const (
	// T is for truth.
	// In a boolean context, anything but NIL is considered "true", but
	// if a function wants to make a point of it, it return T to indicate something
	// to be true.
	// For compatibility with Common Lisp, T gets special treatment and evaluates
	// to itself.
	T         = Symbol("T")
	nilString = "NIL"
	ioBufSize = 1 << 18
)

// LispValue is the "abstract base class", so to speak, for Lisp data.
// All values usable in kryLisp internally implement this interface.
type LispValue interface {
	Type() types.ID
	String() string
	Bool() bool
	Eq(other LispValue) bool
	Equal(other LispValue) bool
	Convert(id types.ID) (LispValue, error)
}

// Number is an interface - kind of an abstract base class, if you will -
// for numeric types in Lisp.
type Number interface {
	LispValue
	Num()
	IsZero() bool
}

// IOHandle is a Handle used for general I/O, regardless of whether it is file
// I/O or network I/O.
type IOHandle struct {
	LispValue
	io.ReadWriter
	io.Closer
}

///////////////////////////////////////////////////////////////////////
// Nil ////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////

// A NilValue represents nil, the strange list-symbol duality.
type NilValue int

// Type returns the type ID of the Lisp value, in this case types.Nil
func (n NilValue) Type() types.ID {
	return types.Nil
} // func (n NilValue) Type() types.ID

// String returns a string representation of the Lisp value.
func (n NilValue) String() string {
	return nilString
} // func (n NilValue) String() string

// Bool returns the "truthiness" of a Lisp value.
func (n NilValue) Bool() bool {
	return false
} // func (n NilValue) Bool() bool

// Eq compares the receiver with the argument for identity.
func (n NilValue) Eq(other LispValue) bool {
	// if other == nil {
	// 	return true
	// } else if other.Type() == types.List {
	// 	var l = other.(*List)

	// 	if l.Length == 0 || l.Car == nil {
	// 		return true
	// 	}
	// } else if other.Type() == types.Nil {
	// 	return true
	// }

	// return false
	return IsNil(other)
} // func (n NilValue) Eq(other LispValue) bool

// Equal compares two Lisp values for equality.
func (n NilValue) Equal(other LispValue) bool {
	return IsNil(other)
} // func (n NilValue) Equal(other LispValue) bool

// Convert attempts to convert the receiver to a LispValue of the given type.
// Converting a value to its own type always returns the receiver.
// Converting a value to types.String may invoke the type's String method to
// perform the conversion.
func (n NilValue) Convert(id types.ID) (LispValue, error) {
	switch id {
	case types.Nil:
		return NIL, nil
	case types.String:
		return StringValue(n.String()), nil
	case types.List:
		return NIL, nil
	case types.Symbol:
		return Symbol("NIL"), nil
	default:
		return nil, &TypeConversionError{types.Nil, id}
	}
} // func (n NilValue) Convert(id types.ID) (LispValue, error)

// NIL is the canonical nil value. In theory, we could get away with just
// having a single value.
const NIL NilValue = 1

///////////////////////////////////////////////////////////////////////
// IntValue ///////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////

// IntValue is an integer.
// At this point, a signed, 64-bit integer is the only numeric type supported.
// Eventually, I want to add support for bignums and floating point numbers,
// possibly more (rational numbers would be nice).
type IntValue int64

// Type returns the type ID of the Lisp value, in this case
// types.Number
func (i IntValue) Type() types.ID {
	return types.Integer
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
	} else if other.Type() != types.Integer {
		return false
	}

	return i == other.(IntValue)
} // func (i IntValue) Eq(other LispValue) bool

// Equal compares two Lisp values for equality.
func (i IntValue) Equal(other LispValue) bool {
	switch o := other.(type) {
	case NilValue:
		return false
	case IntValue:
		return i == o
	case FloatValue:
		if IsFloatInteger(o) {
			return FloatValue(float64(i)) == o
		}

		return false
	case *BigInt:
		var tmp = big.NewInt(int64(i))
		return tmp.Cmp(o.Value) == 0
	default:
		return false
	}
} // func (i IntValue) Equal(other LispValue) bool

// Num identifies the receiver as kind of Number.
func (i IntValue) Num() {
} // func (i IntValue) Num()

// IsZero returns true if the receiver's value is equal to zero.
func (i IntValue) IsZero() bool {
	return i == 0
} // func (i IntValue) IsZero() bool

// Convert attempts to convert the receiver to a LispValue of the given type.
// Converting a value to its own type always returns the receiver.
// Converting a value to types.String may invoke the type's String method to
// perform the conversion.
func (i IntValue) Convert(id types.ID) (LispValue, error) {
	switch id {
	case types.Integer:
		return i, nil
	case types.Float:
		return FloatValue(float64(i)), nil
	case types.BigInt:
		return &BigInt{Value: big.NewInt(int64(i))}, nil
	case types.String:
		//return StringValue(strconv.Itoa(int(i))), nil
		return StringValue(i.String()), nil
	default:
		return nil, &TypeConversionError{types.Integer, id}
	}
} // func (i IntValue) Convert(id types.ID) (LispValue, error)

///////////////////////////////////////////////////////////////////////
// FloatValue /////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////

// FloatValue is a 64-bit floating point number.
type FloatValue float64

// Type returns the type ID of the Lisp value, in this case
// types.Float
func (f FloatValue) Type() types.ID {
	return types.Float
} // func (i FloatValue) Type() types.ID

// String returns a string representation of the Lisp value.
func (f FloatValue) String() string {
	//return fmt.Sprintf("%f", f)
	return strconv.FormatFloat(float64(f), 'f', -1, 64)
} // func (f FloatValue) String() string

// Bool returns the "truthiness" of a Lisp value.
func (f FloatValue) Bool() bool {
	return true
} // func (f FloatValue) Bool() bool

// Num identifies the receiver as kind of Number.
func (f FloatValue) Num() {
} // func (f FloatValue) Num()

// IsZero returns true if the receiver's value is equal to zero.
func (f FloatValue) IsZero() bool {
	return f == 0
} // func (f FloatValue) IsZero() bool

// Eq return true if the receiver and the argument are the same, i.e. if both
// are of the same type and have the same value.
// (Floating Point equality is determined by value! All the usual caveats for
// floating point equality apply, i.e. two different computations that on paper
// yield the same result may give *very* slightly different results with
// floating point numbers.)
func (f FloatValue) Eq(other LispValue) bool {
	if other == nil {
		return false
	} else if other.Type() != types.Float {
		return false
	}

	return f == other.(FloatValue)
} // func (f FloatValue) Eq(other LispValue) bool

// Equal compares two Lisp values for equality.
func (f FloatValue) Equal(other LispValue) bool {
	switch other.Type() {
	case types.Integer:
		if IsFloatInteger(f) {
			return FloatValue(float64(other.(IntValue))) == f
		}

		return false
	case types.Float:
		return f == other.(FloatValue)
	case types.BigInt:
		if IsFloatInteger(f) {
			var tmp = big.NewInt(int64(f))
			return tmp.Cmp(other.(*BigInt).Value) == 0
		}

		return false
	default:
		return false
	}
} // func (f FloatValue) Equal(other LispValue) bool

// Convert attempts to convert the receiver to a LispValue of the given type.
// Converting a value to its own type always returns the receiver.
// Converting a value to types.String may invoke the type's String method to
// perform the conversion.
func (f FloatValue) Convert(id types.ID) (LispValue, error) {
	switch id {
	case types.Integer:
		return IntValue(f), nil
	case types.Float:
		return f, nil
	case types.BigInt:
		//return &BigInt{num: big.NewFloat(float64(f)).Int()}, nil
		var bf = big.NewFloat(float64(f))
		var bi *big.Int

		bi, _ = bf.Int(nil)
		return &BigInt{Value: bi}, nil
	case types.String:
		return StringValue(f.String()), nil
	default:
		return NIL, &TypeConversionError{types.Float, id}
	}
} // func (f FloatValue) Convert(types.ID) (LispValue, error)

///////////////////////////////////////////////////////////////////////
// BigInt /////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////

// Donnerstag, 26. 10. 2017, 10:50
// I think I have to wrap bignums in a struct or array because of the method
// namespaces. While I am at it, could I add anything else that is worthwhile?
// Cache the string representation or something like that?

// BigInt is an arbitrary-precision integer value, using Go's big.Int
// under the hood.
type BigInt struct {
	Value *big.Int
	str   string
}

// BigIntFromString attempts to parse a string into a BigInt value.
func BigIntFromString(s string) (*BigInt, error) {
	var n = new(big.Int)
	var ok bool

	if _, ok = n.SetString(s, 10); !ok {
		return nil, fmt.Errorf("Could not parse string to BigInt: %s",
			s)
	}

	return &BigInt{Value: n}, nil
} // func BigIntFromString(s string) (*BigInt, error)

// BigZero is the BigInt value of zero.
var BigZero = &BigInt{
	Value: big.NewInt(0),
	str:   "0",
}

// IntMax is the largest value that can be represented as an int64/IntValue
var IntMax = &BigInt{
	Value: big.NewInt(math.MaxInt64),
	str:   strconv.FormatInt(math.MaxInt64, 10),
}

// IntMin is the smallest (most negative) value that can be represented as an
// int64/IntValue
var IntMin = &BigInt{
	Value: big.NewInt(math.MinInt64),
	str:   strconv.FormatInt(math.MinInt64, 10),
}

// Type returns the type ID of the Lisp value, in this case
// types.BigInt
func (b *BigInt) Type() types.ID {
	return types.BigInt
} // func (b *BigInt) Type() types.ID

// String returns a string representation of the Lisp value.
func (b *BigInt) String() string {
	if b.str != "" {
		return b.str
	}

	b.str = b.Value.String()
	return b.str
} // func (b *BigInt) String() string

// Bool returns the "truthiness" of a Lisp value.
func (b *BigInt) Bool() bool {
	return true
} // func (b *BigInt) Bool() bool

// Eq return true if the receiver and the argument are the same.
// If the argument is a Number, the two are compared for value equality.
func (b *BigInt) Eq(other LispValue) bool {
	if other == nil || other.Type() != types.BigInt {
		return false
	}

	return b == other.(*BigInt)

	// if other == nil {
	// 	return false
	// } else if !IsNumber(other) {
	// 	return false
	// }

	// switch ot := other.(type) {
	// case *BigInt:
	// 	return b.Value.Cmp(ot.Value) == 0
	// case IntValue:
	// 	var ob = big.NewInt(int64(ot))
	// 	return b.Value.Cmp(ob) == 0
	// case FloatValue:
	// 	var of = new(big.Float)
	// 	of.SetFloat64(float64(ot))
	// 	if of.IsInt() {
	// 		var oi, _ = of.Int(nil)
	// 		return b.Value.Cmp(oi) == 0
	// 	}

	// 	return false
	// default:

	// 	return false
	// }
} // func (b *BigInt) Eq(other LispValue) bool

// Equal compares two Lisp values for equality.
func (b *BigInt) Equal(other LispValue) bool {
	switch other.Type() {
	case types.Integer:
		var tmp = big.NewInt(int64(other.(IntValue)))
		return b.Value.Cmp(tmp) == 0
	case types.Float:
		if IsFloatInteger(other.(FloatValue)) {
			var tmp1 = big.NewFloat(float64(other.(FloatValue)))
			tmp2, _ := tmp1.Int(nil)

			return b.Value.Cmp(tmp2) == 0
		}

		return false
	case types.BigInt:
		return b.Value.Cmp(other.(*BigInt).Value) == 0
	default:
		return false
	}
} // func (b *BigInt) Equal(other LispValue) bool

// Num identifies the receiver as kind of Number.
func (b *BigInt) Num() {
}

// IsInt64 returns true if the value of the receiver is within the range
// that can be represented by int64/IntValue
func (b *BigInt) IsInt64() bool {
	return b.Value.Cmp(IntMin.Value) >= 0 && b.Value.Cmp(IntMax.Value) <= 0
} // func (b *BigInt) IsInt64() bool

// IsZero returns true if the receiver's value is equal to zero.
func (b *BigInt) IsZero() bool {
	return b.Value.Cmp(BigZero.Value) == 0
} // func (b *BigInt) IsZero() bool

// Convert attempts to convert the receiver to a LispValue of the given type.
func (b *BigInt) Convert(id types.ID) (LispValue, error) {
	switch id {
	case types.Integer:
		if b.IsInt64() {
			return IntValue(b.Value.Int64()), nil
		}

		return NIL, fmt.Errorf("BigInt Value is out range for conversion to int64: %s",
			b.Value.String())
	case types.BigInt:
		return b, nil
	case types.Float:
		var tmpFloat = new(big.Float)
		var f float64

		tmpFloat.SetInt(b.Value)
		f, _ = tmpFloat.Float64()
		return FloatValue(f), nil
	case types.String:
		return StringValue(b.Value.String()), nil
	default:
		return NIL, &TypeConversionError{
			source:      types.BigInt,
			destination: id,
		}
	}
}

// Clone creates and returns a new BigInt with the same value as the receiver.
func (b *BigInt) Clone() *BigInt {
	var n = &BigInt{Value: new(big.Int)}
	n.Value.Set(b.Value)
	return n
} // func (b *BigInt) Clone() *BigInt

///////////////////////////////////////////////////////////////////////
// StringValue ////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////

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

	//return s == other.(StringValue)
	return 0 == strings.Compare(string(s), string(other.(StringValue)))
} // func (s StringValue) Eq(other LispValue) bool

// Equal compares two Lisp values for equality.
func (s StringValue) Equal(other LispValue) bool {
	switch other.Type() {
	case types.String:
		return s == other.(StringValue)
	default:
		return false
	}
} // func (s StringValue) Equal(other LispValue) bool

// Convert attempts to convert the receiver to a LispValue of the given type.
// Converting a value to its own type always returns the receiver.
// Converting a value to types.String may invoke the type's String method to
// perform the conversion.
func (s StringValue) Convert(id types.ID) (LispValue, error) {
	var err error
	var i int64
	var f float64
	switch id {
	case types.Integer:
		if i, err = strconv.ParseInt(string(s), 10, 64); err != nil {
			return NIL, err
		}

		return IntValue(i), nil
	case types.Float:
		if f, err = strconv.ParseFloat(string(s), 64); err != nil {
			return NIL, err
		}

		return FloatValue(f), nil
	case types.String:
		return s, nil
	case types.Symbol:
		return Symbol(strings.ToUpper(string(s))), nil
	case types.KeySym:
		return Symbol(":" + strings.ToUpper(string(s))), nil
	default:
		return NIL, &TypeConversionError{types.String, id}
	}
} // func (s StringValue) Convert(id types.ID) (LispValue, error)

///////////////////////////////////////////////////////////////////////
// Symbol /////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////

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
	return s != nilString
} // func (s Symbol) Bool() bool

// Eq compares the receiver with the argument for identity.
func (s Symbol) Eq(other LispValue) bool {
	if other == nil {
		return s == nilString
	}

	switch v := other.(type) {
	case NilValue:
		return s == nilString
	case *List:
		return (s == nilString) && (v.Length == 0 || v.Car == nil)
	case Symbol:
		return s == v
	default:
		return false
	}
} // func (s Symbol) Eq(other LispValue) bool

// Equal compares two Lisp values for equality.
func (s Symbol) Equal(other LispValue) bool {
	switch other.Type() {
	case types.Nil:
		return s == nilString
	case types.List:
		return (s == nilString) && (other.(*List).Length == 0 || other.(*List).Car == nil)
	case types.Symbol:
		return s == other.(Symbol)
	default:
		return false
	}
} // func (s Symbol) Equal(other LispValue) bool

// Convert attempts to convert the receiver to a LispValue of the given type.
// Converting a value to its own type always returns the receiver.
// Converting a value to types.String may invoke the type's String method to
// perform the conversion.
func (s Symbol) Convert(id types.ID) (LispValue, error) {
	switch id {
	case types.String:
		return StringValue(s), nil
	case types.Symbol:
		return s, nil
	default:
		return NIL, &TypeConversionError{types.Symbol, id}
	}
} // func (s Symbol) Convert(id types.ID) (LispValue, error)

// IsKeyword returns true if the symbol gets special treatment by the
// interpreter.
// Usually, this is either because some primitives need to be implemented
// outside Lisp, or because of efficiency considerations.
func (s Symbol) IsKeyword() bool {
	return s[0] == ':'
} // func (s Symbol) IsKeyword() bool

///////////////////////////////////////////////////////////////////////
// ConsCell ///////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////

// NB, ConsCell is not really meant to be exposed to the Lisp-programmer.

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

	if s.Car == nil || s.Car == NIL {
		s1 = nilString
	} else {
		s1 = s.Car.String()
	}

	if s.Cdr == nil || s.Cdr == NIL {
		s2 = nilString
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
			var ok bool
			// fmt.Println(spew.Sdump(cell.Cdr))
			// fmt.Printf("Type of CDR is %s\n", cell.Cdr.Type().String())
			if cell, ok = cell.Cdr.(*ConsCell); !ok {
				cell = nil
			}
		}
	}

	// Hier sollten wir nur ankommen, wenn s selbst schon nil ist.
	// Ich weiß gar nicht, ob das in Go erlaubt ist. In C++ ist das im
	// Prinzip erlaubt, meine ich, aber ... das ist auch einer der Gründe,
	// aus denen ich um C++ nach Kräften einen Bogen mache.
	return true
} // func (s *ConsCell) IsList() bool

// Next returns the next ConsCell in a chain, or nil if there is none.
func (s *ConsCell) Next() *ConsCell {
	if s.Cdr == nil {
		return nil
	} else if s.Cdr.Type() != types.ConsCell {
		return nil
	} else {
		return s.Cdr.(*ConsCell)
	}
} // func (s *ConsCell) Next() *ConsCell

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

// Equal compares two Lisp values for equality.
func (s *ConsCell) Equal(other LispValue) bool {
	if other == nil {
		return s == nil
	} else if c, ok := other.(*ConsCell); ok {
		var cell1, cell2 *ConsCell

		cell1 = s
		cell2 = c

		for cell1 != nil && cell2 != nil {
			if !cell1.Car.Equal(cell2.Car) {
				return false
			} else if IsNil(cell1.Cdr) {
				return IsNil(cell2.Cdr)
			} else if IsNil(cell2.Cdr) {
				return false
			}

			cell1 = cell1.Cdr.(*ConsCell)
			cell2 = cell2.Cdr.(*ConsCell)
		}

		return true
	}

	return false
} // func (s *ConsCell) Equal(other LispValue) bool

// Convert attempts to convert the receiver to a LispValue of the given type.
// Converting a value to its own type always returns the receiver.
// Converting a value to types.String may invoke the type's String method to
// perform the conversion.
func (s *ConsCell) Convert(id types.ID) (LispValue, error) {
	switch id {
	case types.String:
		return StringValue(s.String()), nil
	case types.List:
		return &List{Car: s, Length: s.ActualLength()}, nil
	case types.ConsCell:
		return s, nil
	default:
		return NIL, &TypeConversionError{types.ConsCell, id}
	}
} // func (s *ConsCell) Convert(id types.ID) (LispValue, error)

///////////////////////////////////////////////////////////////////////
// List ///////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////

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

// ListNil returns a new, empty List
func ListNil() *List {
	return &List{
		Car:    nil,
		Length: 0,
	}
}

// Type returns the type ID of the receiver.
// (In this case, types.List
func (l *List) Type() types.ID {
	return types.List
} // func (l *List) Type() types.ID

// String returns a string representation of the Lisp value.
func (l *List) String() string {
	if l == nil {
		return nilString
	} else if l.Car == nil {
		return nilString
	}

	// spew.Dump(l)

	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("Error converting list to string: %s\n",
				err)
		}
	}()

	var elements = make([]string, l.Length)
	var idx = 0
	//var cell = l.Car

	for cell := l.Car; cell != nil; idx++ {
		if cell.Car != nil {
			elements[idx] = cell.Car.String()
		} else {
			elements[idx] = nilString
		}

		if !(cell.Cdr == nil || cell.Cdr == NIL) {
			var ok bool
			if cell, ok = cell.Cdr.(*ConsCell); !ok {
				// elements[idx+1] = cell.Cdr.String()
				elements = elements[:idx]
				break
			}
		} else {
			cell = nil
		}
	}

	//return "(" + strings.Join(elements, " ") + ")"
	return fmt.Sprintf("(%s)", strings.Join(elements, " "))
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

// Equal compares two Lisp values for equality.
func (l *List) Equal(other LispValue) bool {
	if other == nil || other.Type() != types.List {
		return false
	}

	var o = other.(*List)

	var cell1, cell2 *ConsCell

	cell1 = l.Car
	cell2 = o.Car

	for cell1 != nil && cell2 != nil {
		if !cell1.Car.Equal(cell2.Car) {
			return false
		} else if IsNil(cell1.Cdr) {
			return IsNil(cell2.Cdr)
		} else if IsNil(cell2.Cdr) {
			return false
		}

		cell1 = cell1.Cdr.(*ConsCell)
		cell2 = cell2.Cdr.(*ConsCell)
	}

	return true
} // func (l *List) Equal(other LispValue) bool

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
		if elt.Cdr.Type() == types.ConsCell {
			elt = elt.Cdr.(*ConsCell)
		} else {
			return NIL, errors.New("Not a proper list")
		}
	}

	return elt.Car, nil
} // func (l *List) Nth(n int) (LispValue, error)

// Convert attempts to convert the receiver to a LispValue of the given type.
// Converting a value to its own type always returns the receiver.
// Converting a value to types.String may invoke the type's String method to
// perform the conversion.
func (l *List) Convert(id types.ID) (LispValue, error) {
	// I *could* support converting a list of Strings
	// to a string by joining them.
	// Should I?

	switch id {
	case types.String:
		return StringValue(l.String()), nil
	case types.List:
		return l, nil
	case types.Array:
		arr := make(Array, l.Length)
		var idx = 0
		for cell := l.Car; cell != nil; cell = cell.Cdr.(*ConsCell) {
			arr[idx] = cell.Car
			idx++
			if cell.Cdr == nil {
				break
			}
		}

		return arr, nil
	case types.ConsCell:
		return l.Car, nil
	default:
		return NIL, &TypeConversionError{types.List, id}
	}
} // func (l *List) Convert(id types.ID) (LispValue, error)

///////////////////////////////////////////////////////////////////////
// Function ///////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////

// Function represents a Lisp function.
// Technically, one could implement functions purely in terms of lists,
// but for efficiency reasons - and functions are used a *lot* in Lisp,
// obviously - they get their own type.
type Function struct {
	Env       *Environment
	Args      []Symbol
	Body      []LispValue
	DocString StringValue
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

// Equal compares two Lisp values for equality.
func (f *Function) Equal(other LispValue) bool {
	return f.Eq(other)
} // func (f *Function) Equal(other LispValue) bool

// Convert attempts to convert the receiver to a LispValue of the given type.
// Converting a value to its own type always returns the receiver.
// Converting a value to types.String may invoke the type's String method to
// perform the conversion.
func (f *Function) Convert(id types.ID) (LispValue, error) {
	if id == types.Function {
		return f, nil
	} else if id == types.String {
		return StringValue(f.String()), nil
	}

	return NIL, &TypeConversionError{types.Function, id}
} // func (f *Function) Convert(id types.ID) (LispValue, error)

// Program represents a sequence of lisp expressions.
type Program []LispValue

// Type returns the type ID of the value, in this case types.Program
func (p Program) Type() types.ID {
	return types.Program
} // func (p Program) Type() types.ID

// String returns a string representation of the Lisp value.
func (p Program) String() string {
	var out bytes.Buffer

	out.WriteString("(begin")

	for _, clause := range p {
		out.WriteString(fmt.Sprintf("\n\t%s", clause))
	}

	out.WriteString(")")

	return out.String()
} // func (p Program) String() string

// Bool returns the "truthiness" of a Lisp value.
func (p Program) Bool() bool {
	return len(p) != 0
} // func (p Program) Bool() bool

// Eq compares the receiver with the argument for identity.
func (p Program) Eq(other LispValue) bool {
	if other == nil || other == NIL {
		return false
	} else if other.Type() != types.Program {
		return false
	}

	var op = other.(Program)

	if len(p) != len(op) {
		return false
	}

	for i := 0; i < len(p); i++ {
		if !p[i].Eq(op[i]) {
			return false
		}
	}

	return true
} // func (p Program) Eq(other LispValue) bool

// Equal compares two Lisp values for equality.
func (p Program) Equal(other LispValue) bool {
	return p.Eq(other)
} // func (p Program) Equal(other LispValue) bool

// Convert attempts to convert the receiver to a LispValue of the given type.
// Converting a value to its own type always returns the receiver.
// Converting a value to types.String may invoke the type's String method to
// perform the conversion.
func (p Program) Convert(id types.ID) (LispValue, error) {
	if id == types.Program {
		return p, nil
	} else if id == types.String {
		return StringValue(p.String()), nil
	}

	return NIL, &TypeConversionError{types.Program, id}
} // func (p Program) Convert(id types.ID) (LispValue, error)

///////////////////////////////////////////////////////////////////////
// Regexp /////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////

// Regexp is a regular expression. Duh!
// Since Go kindly provides a regex engine, we can map this directly to
// a Go *regexp.Regexp.
type Regexp struct {
	Pat *regexp.Regexp
}

// Type returns the type ID of the value, in this case types.Regexp
func (re *Regexp) Type() types.ID {
	return types.Regexp
} // func (re *Regexp) Type() types.ID

func (re *Regexp) String() string {
	return re.Pat.String()
} // func (re *Regexp) String() string

// Bool returns the "truthiness" of a Lisp value.
func (re *Regexp) Bool() bool {
	return true
} // func (re *Regexp) Bool() bool

// Eq compares the receiver with the argument for identity.
func (re *Regexp) Eq(other LispValue) bool {
	if other == nil || other.Type() != types.Regexp {
		return false
	}

	return re == other.(*Regexp)
} // func (re *Regexp) Eq(other value.LispValue) bool

// Equal compares two Lisp values for equality.
func (re *Regexp) Equal(other LispValue) bool {
	if nil == other || other.Type() != types.Regexp {
		return false
	}

	return re.String() == other.(*Regexp).String()
} // func (re *Regexp) Equal(other LispValue) bool

// Convert attempts to convert the receiver to a LispValue of the given type.
// Converting a value to its own type always returns the receiver.
// Converting a value to types.String may invoke the type's String method to
// perform the conversion.
func (re *Regexp) Convert(id types.ID) (LispValue, error) {
	if id == types.Regexp {
		return re, nil
	} else if id == types.String {
		return StringValue(re.Pat.String()), nil
	}

	return NIL, &TypeConversionError{types.Regexp, id}
} // func (re *Regexp) Convert(id types.ID) (LispValue, error)

///////////////////////////////////////////////////////////////////////
// Array //////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////

// Array is a one-dimensional array of Lisp values.
type Array []LispValue

// EmptyArray returns a new, empty Array
func EmptyArray() Array {
	return make(Array, 0)
}

// Type returns the type ID of the value, in this case types.Regexp
func (arr Array) Type() types.ID {
	return types.Array
} // func (arr Array) Type() types.ID

// String returns a string representation of the Lisp value.
func (arr Array) String() string {
	var out bytes.Buffer

	out.WriteString("[")
	var svals = make([]string, len(arr))
	for idx, val := range arr {
		svals[idx] = val.String()
	}
	out.WriteString(strings.Join(svals, " "))
	out.WriteString("]")

	return out.String()
} // func (arr Array) String() string

// Bool returns the "truthiness" of a Lisp value.
func (arr Array) Bool() bool {
	return true
} // func (arr Array) Bool() bool

// Eq compares the receiver with the argument for identity.
func (arr Array) Eq(other LispValue) bool {
	if other == nil || other.Type() != types.Array {
		return false
	}

	var oarr = other.(Array)

	for idx, val := range arr {
		if !val.Eq(oarr[idx]) {
			return false
		}
	}

	return true
} // func Eq(other LispValue) bool

// Equal compares two Lisp values for equality.
func (arr Array) Equal(other LispValue) bool {
	if other == nil || other.Type() != types.Array {
		return false
	}

	var oarr = other.(Array)

	for idx, val := range arr {
		if !val.Equal(oarr[idx]) {
			return false
		}
	}

	return true
} // func (arr Array) Equal(other LispValue) bool

// Convert attempts to convert the receiver to a LispValue of the given type.
// Converting a value to its own type always returns the receiver.
// Converting a value to types.String may invoke the type's String method to
// perform the conversion.
func (arr Array) Convert(id types.ID) (LispValue, error) {
	switch id {
	case types.String:
		return StringValue(arr.String()), nil
	case types.List:
		var res *ConsCell

		for i := len(arr) - 1; i >= 0; i-- {
			res = &ConsCell{
				Car: arr[i],
				Cdr: res,
			}
		}

		return &List{Car: res, Length: len(arr)}, nil
	case types.Array:
		return arr, nil
	default:
		return NIL, &TypeConversionError{types.Array, id}
	}
} // func (arr Array) Convert(id types.ID) (LispValue, error)

///////////////////////////////////////////////////////////////////////
// Hashtable //////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////

// Hashtable is ... a hash table. (Under the hood, it's a Go map)
type Hashtable map[LispValue]LispValue

// Type returns the type ID of the value, in this case types.Regexp
func (ht Hashtable) Type() types.ID {
	return types.Hashtable
} // func (ht Hashtable) Type() types.ID

// String returns a string representation of the Lisp value.
func (ht Hashtable) String() string {
	var pairs = make([]string, len(ht))
	var idx = 0

	for k, v := range ht {
		s := fmt.Sprintf("%s : %s",
			k, v)
		pairs[idx] = s
		idx++
	}

	return "{" + strings.Join(pairs, ", ") + "}"
} // func (ht Hashtable) String() string

// Bool returns the "truthiness" of a Lisp value.
func (ht Hashtable) Bool() bool {
	return true
} // func (ht Hashtable) Bool() bool

// Eq compares the receiver with the argument for identity.
func (ht Hashtable) Eq(other LispValue) bool {
	if other == nil || other.Type() != types.Hashtable {
		return false
	}

	var ot = other.(Hashtable)

	if len(ot) != len(ht) {
		return false
	}

	// To be really sure, I need to check if the other table
	// has keys that we do not have. ...
	// I promise to implement that sometimes.
	for key, val := range ht {
		var (
			oval LispValue
			ok   bool
		)

		if oval, ok = ot[key]; !ok {
			return false
		} else if !val.Eq(oval) {
			return false
		}
	}

	return true
} // func (ht Hashtable) Eq() bool

// Equal compares two Lisp values for equality.
func (ht Hashtable) Equal(other LispValue) bool {
	if other == nil || other.Type() != types.Hashtable {
		return false
	}

	var ot = other.(Hashtable)

	if len(ot) != len(ht) {
		return false
	}

	// To be really sure, I need to check if the other table
	// has keys that we do not have. ...
	// I promise to implement that sometimes.
	for key, val := range ht {
		var (
			oval LispValue
			ok   bool
		)

		if oval, ok = ot[key]; !ok {
			return false
		} else if !val.Equal(oval) {
			return false
		}
	}

	return true
} // func (ht Hashtable) Eq(other LispValue) bool

// Convert attempts to convert the receiver to a LispValue of the given type.
// Converting a value to its own type always returns the receiver.
// Converting a value to types.String may invoke the type's String method to
// perform the conversion.
func (ht Hashtable) Convert(id types.ID) (LispValue, error) {
	switch id {
	case types.Hashtable:
		return ht, nil
	default:
		return NIL, &TypeConversionError{
			types.Hashtable,
			id,
		}
	}
} // func (ht Hashtable) Convert(id types.ID) (LispValue, error)

///////////////////////////////////////////////////////////////////////
// FileHandle /////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////

// Montag, 13. 11. 2017, 19:06
// I just realized that it is slightly more complicated: I assume that file I/O
// means reading and writing text, line by line, most of the time.
// To get that, I need to wrap the *os.File in a bufio.Reader, or alternatively
// implement my own way of reading text files line by line.
// ... Mmmh, that is not going to be easy.
// The buffered reader has the disadvantage that it makes writing more
// complicated.
// ...
// Aaaah, there is bufio.ReadWriter, which is a struct of a Reader and
// a Writer...

// FileHandle represents a file on disk or on some other mass storage medium. Name one.
type FileHandle struct {
	path        string
	raw         *os.File
	r           *bufio.Reader
	w           *bufio.Writer
	permissions permission.FilePermission
	bufRead     bool
	bufWrite    bool
}

// OpenFile opens a file at the given location with the given access rights.
func OpenFile(path string, perm int, access permission.FilePermission) (*FileHandle, error) {
	// First, we map the access flags to the file flags in the OS package:
	var flags = os.O_CREATE
	const rw = permission.Read | permission.Write

	if access&rw == rw {
		flags |= os.O_RDWR
	} else if access&permission.Read != 0 {
		flags |= os.O_RDONLY
	} else if access&permission.Write != 0 {
		flags |= os.O_WRONLY
	} else {
		var msg = "Opening a file with neither read nor write access does not make sense"
		fmt.Println(msg)
		return nil, errors.New(msg)
	}

	if (access & permission.Append) != 0 {
		flags |= os.O_APPEND
	}

	if (access & permission.Sync) != 0 {
		flags |= os.O_SYNC
	}

	var fh = &FileHandle{
		path:        path,
		permissions: access,
	}
	var err error

	//if fh.raw, err = os.OpenFile(path, perm, os.FileMode(flags|os.O_CREATE)); err != nil {
	if fh.raw, err = os.OpenFile(path, flags, os.FileMode(perm)); err != nil {
		fmt.Printf("Error opening %s: %s",
			path,
			err.Error())
		return nil, err
	}

	runtime.SetFinalizer(fh, func(arg *FileHandle) {
		if arg.raw != nil {
			arg.raw.Close()
			arg.raw = nil
		}
	})

	if fh.isRead() {
		fh.r = bufio.NewReader(fh.raw)
		fh.bufRead = true
	}

	if fh.isWrite() {
		if !fh.isSync() {
			fh.w = bufio.NewWriter(fh.raw)
			fh.bufWrite = true
		}
	}

	return fh, nil
} // func OpenFile(path string, access permission.FilePermission) (*FileHandle, error)

// Close closes the file. Using a filehandle after it has been closed will
// result in an error.
func (fh *FileHandle) Close() error {
	err := fh.raw.Close()
	//fh.handle = nil
	return err
} // func (fh *FileHandle) Close() error

// Type returns the type ID of the value, in this case types.FileHandle
func (fh *FileHandle) Type() types.ID {
	return types.FileHandle
} // func (fh *FileHandle) Type() types.ID

func (fh *FileHandle) String() string {
	if common.Debug {
		krylib.Trace()
	}

	return fmt.Sprintf("FileHandle<@path=%s, @read=%t, @write=%t>",
		fh.path,
		fh.isRead(),
		fh.isWrite(),
	)
} // func (fh *FileHandle) String() string

// Bool returns the "truthiness" of a Lisp value.
func (fh *FileHandle) Bool() bool {
	if common.Debug {
		krylib.Trace()
	}

	if IsNil(fh) {
		return false
	}

	return !IsNil(fh)
} // func (fh *FileHandle) Bool() bool

// Eq compares the receiver with the argument for identity.
func (fh *FileHandle) Eq(other LispValue) bool {
	if common.Debug {
		krylib.Trace()
	}

	if other == nil {
		return false
	} else if other.Type() != types.FileHandle {
		return false
	}

	var cmp = other.(*FileHandle)

	return fh == cmp
} // func (fh *FileHandle) Eq(other LispValue) bool

// Equal compares two Lisp values for equality.
func (fh *FileHandle) Equal(other LispValue) bool {
	if common.Debug {
		krylib.Trace()
	}

	if other.Type() != types.FileHandle {
		return false
	}

	if other == nil {
		return false
	}

	var of = other.(*FileHandle)

	if fh.path != of.path {
		return false
	} else if fh.isRead() != of.isRead() || fh.isWrite() != of.isWrite() {
		return false
	}

	return true
} // func (fh *FileHandle) Equal(other LispValue) bool

// Convert attempts to convert the receiver to a LispValue of the given type.
func (fh *FileHandle) Convert(id types.ID) (LispValue, error) {
	if common.Debug {
		krylib.Trace()
	}

	if id == types.FileHandle {
		return fh, nil
	}

	return NIL, &TypeConversionError{
		source:      types.FileHandle,
		destination: id,
	}
} // func (fh *FileHandle) Convert(id types.ID) (LispValue, error)

func (fh *FileHandle) isRead() bool {
	return (fh.permissions & permission.Read) != 0
} // func (fh *FileHandle) isRead() bool

func (fh *FileHandle) isWrite() bool {
	return (fh.permissions & permission.Write) != 0
} // func (fh *FileHandle) isWrite() bool

func (fh *FileHandle) isAppend() bool {
	return (fh.permissions & permission.Append) != 0
} // func (fh *FileHandle) isAppend() bool

func (fh *FileHandle) isSync() bool {
	return (fh.permissions & permission.Sync) != 0
} // func (fh *FileHandle) isSync() bool

func (fh *FileHandle) Read(b []byte) (n int, e error) {
	if common.Debug {
		krylib.Trace()
	}

	var (
		buf = make([]byte, 262144)
		err error
		// numSize = len(buf)
	)

	if _, err = fh.raw.Read(buf); err != nil {
		fmt.Printf("Error reading from I/O stream @(%s) -- %s",
			fh.path,
			err.Error())
		return 0, err
	}

	return 0, krylib.NotImplemented
} // func (fh *FileHandle) Read([]byte) (n int, e error)

func (fh *FileHandle) readBuffered(b []byte) (n int, e error) {
	if common.Debug {
		krylib.Trace()
	}

	var (
		buf = make([]byte, 262144)
		err error
		// numSize = len(buf)
	)

	if _, err = fh.r.Read(buf); err != nil {
		fmt.Printf("Error reading from I/O stream @(%s) -- %s",
			fh.path,
			err.Error())
		return 0, err
	}

	return 0, krylib.NotImplemented
} // func (fh *FileHandle) readBuffered([]byte) (n int, e error)

// ReadLine attempts to read a line of text from the file handle.
func (fh *FileHandle) ReadLine() (string, error) {
	if common.Debug {
		krylib.Trace()
	}

	return fh.r.ReadString('\n')
} // func (fh *FileHandle) ReadLine() (string, error)

func (fh *FileHandle) Write(b []byte) (n int, e error) {
	if common.Debug {
		krylib.Trace()
	}

	var (
		bytesWritten int
		err          error
	)

	if bytesWritten, err = fh.raw.Write(b); err != nil {
		fmt.Printf("Error writing to I/O stream @(%s) -- %s",
			fh.path,
			err.Error())
		return 0, err
	}

	return bytesWritten, nil
} // func (fh *FileHandle) Write(b []byte) (n int, e error)

// Seek the file handle to the given position, analog to File.Seek from the
// os package.
func (fh *FileHandle) Seek(offset int64, whence int) (int64, error) {
	if common.Debug {
		krylib.Trace()
	}

	var (
		newOffset int64
		err       error
	)

	if fh.isAppend() {
		return 0, fmt.Errorf("FH %s is append-only", fh)
	}

	if newOffset, err = fh.raw.Seek(offset, whence); err != nil {
		fmt.Printf("Error seeking %s to (%d)%d: %s\n",
			fh,
			whence,
			offset,
			err.Error())
		return 0, err
	}

	if fh.bufRead {
		fh.r.Reset(fh.raw)
	}

	if fh.bufWrite {
		fh.w.Reset(fh.raw)
	}

	return newOffset, nil
} // func (fh *FileHandle) Seek(offset int64, orig int) (int64, error)

// Tell returns the current absolute offset within the file.
func (fh *FileHandle) Tell() (int64, error) {
	return fh.raw.Seek(0, 1)
} // func (fh *FileHandle) Tell() (int64, error)

// Sync forces all unwritten data to be written to disk.
func (fh *FileHandle) Sync() error {
	var err error

	if err = fh.raw.Sync(); err != nil {
		var msg = fmt.Sprintf("Error sync()ing %s: %s",
			fh.path,
			err.Error())
		fmt.Println(msg)
		return err
	}

	return nil
} // func (fh *FileHandle) Sync() error
