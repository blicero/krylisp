// /home/krylon/go/src/krylisp/interpreter/interpreter.go
// -*- mode: go; coding: utf-8; -*-
// Created on 08. 09. 2017 by Benjamin Walkenhorst
// (c) 2017 Benjamin Walkenhorst
// Time-stamp: <2017-11-10 17:36:06 krylon>
//
// Donnerstag, 19. 10. 2017, 19:17
// Mmmh, adding floating point numbers makes all the arithmetic code a lot more
// complex. I am going to consider type promotion and such.
// On the other hand, if I do that well, adding additional numeric types is going
// to be relatively straightforward.
//
// Freitag, 20. 10. 2017, 17:16
// Adding support for multiple numeric types is not going to be easy. I basically
// need to rewrite all the arithmetic functions.
//
// Dienstag, 31. 10. 2017, 01:59
// Oh my, when I added numeric types, I forgot the comparison operators!
//
// Freitag, 03. 11. 2017, 19:23
// So, I added regex as a distinct type. I am not sure, yet, what I want the
// regex-API to look like from Lisp.
//
// Okay, I have hashtables and arrays. Now, maybe I want to add some looping
// constructs?
//
// Dienstag, 07. 11. 2017, 21:21
// This file has gotten ... quite large. Maybe I should break it up into
// several smaller files?
// I'll keep that in mind, but for now I am sticking with the big file.
// I am not even sure if Go allows spreading the methods for a type across
// several files.

// Package interpreter implements the actual interpreter.
// The first time 'round, the interpreter is simply going to walk the parse tree
// recursively and evaluate each node.
//
// Later on, I might make a try at performing some basic optimizations or
// generating byte code?
package interpreter

import (
	"errors"
	"fmt"
	"io"
	"krylib"
	"krylisp/compare"
	"krylisp/types"
	"krylisp/value"
	"os"
	"regexp"

	"github.com/davecgh/go-spew/spew"
)

// specialSymbols refer to values or syntactic constructs that are defined in the
// Interpreter itself, not in Lisp.
var specialSymbols = map[string]bool{
	"T":              true,
	"NIL":            true,
	"+":              true,
	"-":              true,
	"*":              true,
	"/":              true,
	"<":              true,
	">":              true,
	">=":             true,
	"<=":             true,
	"EQ":             true,
	"FN":             true,
	"DEFUN":          true,
	"IF":             true,
	"LET":            true,
	"DO":             true,
	"PRINT":          true,
	"CONS":           true,
	"CAR":            true,
	"CDR":            true,
	"SET!":           true,
	"DEFINE":         true,
	"GOTO":           true,
	"QUOTE":          true,
	"NOT":            true,
	"AND":            true,
	"OR":             true,
	"APPLY":          true,
	"LAMBDA":         true,
	"NIL?":           true,
	"LIST":           true,
	"AREF":           true,
	"APUSH":          true,
	"MAKE-ARRAY":     true,
	"MAKE-HASH":      true,
	"HASHREF":        true,
	"HASH-SET":       true,
	"HASH-DELETE":    true,
	"HAS-KEY":        true,
	"DEFMACRO":       true,
	"REGEXP-COMPILE": true,
	"REGEXP-MATCH":   true,
	"LENGTH":         true,
	"CONCAT":         true,
	"GETENV":         true,
	"SETENV":         true,
	//	"FOR-EACH":       true,
}

// IsSpecial returns true if the given symbols has special significance
// to the Lisp Interpreter.
func IsSpecial(s fmt.Stringer) bool {
	_, ok := specialSymbols[s.String()]
	return ok
} // func IsSpecial(s value.Symbol) bool

// IsNumber returns true if the type of the given value is numeric.
func IsNumber(v value.LispValue) bool {
	_, ok := v.(value.Number)
	return ok
} // func IsNumber(v value.LispValue) bool

// Interpreter is my first shot at a tree-walking interpreter for my
// toy Lisp dialect.
type Interpreter struct {
	debug         bool
	gensymCounter int
	env           *value.Environment
	fnEnv         *value.Environment
	stdout        io.Writer
	stderr        io.Writer
	stdin         io.Reader
}

// New returns a fresh, initialized Interpreter instance with an
// empty Environment. It passes the debug flag to the Interpreter.
func New(debug bool) *Interpreter {
	var inter = &Interpreter{
		debug:         debug,
		gensymCounter: 1,
		env:           value.NewEnvironment(nil),
		fnEnv:         value.NewEnvironment(nil),
		stdin:         os.Stdin,
		stdout:        os.Stdout,
		stderr:        os.Stderr,
	}

	return inter
} // func New(debug bool) *Interpreter

// Eval evaluates a Lisp value and returns the result.
// nolint: gocyclo
func (inter *Interpreter) Eval(lval value.LispValue) (value.LispValue, error) {
	if lval == nil {
		return value.NIL, nil
	} /*else if inter.debug {
		spew.Printf("EVAL %#v\n",
			lval)
	}*/

	switch v := lval.(type) {
	case value.IntValue,
		value.FloatValue,
		*value.BigInt,
		value.StringValue,
		value.NilValue:
		return v, nil
	case value.Symbol:
		return inter.evalSymbol(v)
	case *value.List:
		if v.Car.Car.Type() == types.Symbol {
			if IsSpecial(v.Car.Car.(value.Symbol)) {
				return inter.evalSpecialForm(v)
			}
		}

		return inter.evalFuncall(v)
	case value.Program:
		var res value.LispValue
		var err error
		for _, clause := range v {
			if res, err = inter.Eval(clause); err != nil {
				return value.NIL, err
			}
		}

		return res, nil
	case value.Array:
		return v, nil
	case value.Hashtable:
		return v, nil
	case *value.Regexp:
		return v, nil
	default:
		return nil, &TypeError{
			expected: "Atom or List",
			actual:   v.Type().String(),
		}
	}
} // func (inter *Interpreter) Eval(v value.LispValue) (value.LispValue, error)

// nolint: gocyclo
func (inter *Interpreter) evalSpecialForm(l *value.List) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}
	var sym = l.Car.Car.(value.Symbol)

	// This is going to be tedious...
	// Maybe I should use a lookup table. As the number of special forms
	// grows - and it WILL grow -, a switch might not be the most
	// efficient solution.
	//
	// Dienstag, 07. 11. 2017, 18:56
	// I wonder what number branches is the tipping point...
	switch sym {
	case "IF":
		return inter.evalIf(l)
	case "+":
		return inter.evalPlus(l)
	case "-":
		return inter.evalMinus(l)
	case "*":
		return inter.evalMultiply(l)
	case "/":
		return inter.evalDivide(l)
	case "<":
		return inter.evalLessThan(l)
	case ">":
		return inter.evalGreaterThan(l)
	case "<=":
		return inter.evalLessEqual(l)
	case ">=":
		return inter.evalGreaterEqual(l)
	case "DEFUN":
		return inter.evalDefun(l)
	case "LAMBDA":
		return inter.evalLambda(l)
	case "LET":
		return inter.evalLet(l)
	case "EQ":
		return inter.evalEq(l)
	case "QUOTE":
		var retval value.LispValue
		if l.Car.Cdr.Type() == types.ConsCell {
			retval = l.Car.Cdr.(*value.ConsCell).Car
		} else if l.Car.Cdr.Type() == types.List {
			retval = l.Car.Cdr.(*value.List).Car
		}
		return retval, nil
	case "NOT":
		return inter.evalNot(l)
	case "AND":
		return inter.evalAnd(l)
	case "OR":
		return inter.evalOr(l)
	case "DEFINE":
		return inter.evalDefine(l)
	case "SET!":
		return inter.evalSet(l)
	case "PRINT":
		return inter.evalPrint(l)
	case "APPLY":
		return inter.evalApply(l)
	case "CONS":
		return inter.evalCons(l)
	case "CAR":
		return inter.evalCar(l)
	case "CDR":
		return inter.evalCdr(l)
	case "FN":
		return inter.evalFn(l)
	case "NIL?":
		return inter.evalIsNil(l)
	case "LIST":
		return inter.evalList(l)
	case "AREF":
		return inter.evalAref(l)
	case "MAKE-ARRAY":
		return inter.evalMakeArray(l)
	case "APUSH":
		return inter.evalApush(l)
	case "HAS-KEY":
		return inter.evalHasKey(l)
	case "MAKE-HASH":
		return inter.evalMakeHash(l)
	case "HASHREF":
		return inter.evalHashref(l)
	case "HASH-SET":
		return inter.evalHashSet(l)
	case "HASH-DELETE":
		return inter.evalHashDelete(l)
	case "REGEXP-COMPILE":
		return inter.evalRegexpCompile(l)
	case "REGEXP-MATCH":
		return inter.evalRegexpMatch(l)
	case "DO":
		return inter.evalDoLoop(l)
	case "LENGTH":
		return inter.evalLength(l)
	case "CONCAT":
		return inter.evalConcat(l)
	case "GETENV":
		return inter.evalGetenv(l)
	case "SETENV":
		return inter.evalSetenv(l)
	default:
		return value.NIL, fmt.Errorf("Special form %s is not implemented, yet",
			sym)

	}
} // func (inter *Interpreter) evalSpecialForm(l *value.List) (value.LispValue, error)

func (inter *Interpreter) evalSymbol(s value.Symbol) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}
	if s.IsKeyword() {
		// Keyword symbols evaluate to themselves:
		return s, nil
	} else if s == "NIL" || s == "T" {
		return s, nil
	} else if inter.debug {
		fmt.Printf("EVAL Symbol %s\n",
			s)
		inter.env.Dump(inter.stdout)
	}

	var v value.LispValue
	var found bool

	if v, found = inter.env.Get(string(s)); found {
		return v, nil
	}

	return nil, &NoBindingError{sym: s}
} // func (inter *Interpreter) evalSymbol(s value.Symbol) (value.LispValue, error)

// nolint: gocyclo
func (inter *Interpreter) evalLambda(lst *value.List) (*value.Function, error) {
	if inter.debug {
		krylib.Trace()
	}

	if lst == nil || lst.Car == nil || lst.Car.Car == nil {
		return nil, errors.New("Argument is not a lambda list")
	} else if lst.Car.Car.Type() != types.Symbol || lst.Car.Car.(value.Symbol) != "LAMBDA" {
		return nil, errors.New("Argument is not a lambda list")
	} else if lst.Car.Cdr.(*value.ConsCell).Car.Type() != types.List {
		//return nil, errors.New("Second element in List should be a list (of arguments)")
		return nil, fmt.Errorf("Second element in lambda list should be a list (of arguments), not a %s",
			lst.Car.Cdr.(*value.ConsCell).Car.Type())
	}

	var (
		args = lst.Car.Cdr.(*value.ConsCell).Car.(*value.List)
		idx  = 0
		fn   = &value.Function{
			Env:  inter.env,
			Args: make([]value.Symbol, args.Length),
		}
	)

	for symlist := args.Car; symlist != nil; symlist = symlist.Cdr.(*value.ConsCell) {
		if symlist.Car.Type() != types.Symbol {
			return nil, &TypeError{
				expected: types.Symbol.String(),
				actual:   symlist.Car.Type().String(),
			}
		}

		var car = symlist.Car.(value.Symbol)

		fn.Args[idx] = car
		idx++

		if symlist.Cdr == nil {
			break
		}
	}

	var body = lst.Car.Cdr.(*value.ConsCell).Cdr.(*value.ConsCell)
	var len = body.ActualLength()
	fn.Body = make([]value.LispValue, len)

	idx = 0

	for ; body != nil; body = body.Cdr.(*value.ConsCell) {
		fn.Body[idx] = body.Car
		idx++
		if body.Cdr == nil {
			break
		}
	}

	return fn, nil
} // func (inter *Interpreter) evalLambda(lst *value.List) (*value.Function, error)

// nolint: gocyclo
func (inter *Interpreter) evalFuncall(inv *value.List) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}

	var fn *value.Function
	var err error

	if inv == nil || inv.Car == nil {
		return nil, nil
	} else if inter.debug {
		// fmt.Printf("DBG FUNCALL %s\n",
		// 	spew.Sdump(inv))
	}

	// Sonntag, 10. 09. 2017, 18:55
	// I need to handle lists, too. Scheme allows a lambda list at the first
	// position. I would like to support that, too.
	// One day...
	switch f := inv.Car.Car.(type) {
	case *value.Function:
		fn = f
	case value.Symbol:
		if IsSpecial(f) {
			return inter.evalSpecialForm(inv)
		} else if v, ok := inter.fnEnv.Get(string(f)); ok {
			fn = v.(*value.Function)
		} else {
			if inter.debug {
				fmt.Printf("No such function: %s\n",
					spew.Sdump(f))
			}
			return nil, MissingFunctionError(f)
		}
	case *value.List:
		if inter.debug {
			fmt.Printf("Evaluate function call: %s\n",
				//f.String())
				spew.Sdump(f))
		}

		if f.IsLambda() {
			if fn, err = inter.evalLambda(f); err != nil {
				return nil, err
			}
		} else {
			return nil, &TypeError{
				expected: "Lambda List",
				actual:   "Not a Lambda List",
			}
		}
	default:
		return nil, &TypeError{
			expected: "Symbol or function literal",
			actual:   fmt.Sprintf("%T", f),
		}
	}

	// So, if we arrive here, we have a function object, next we should
	// check out the arguments.
	//
	// Montag, 02. 10. 2017, 23:42
	// If a function is called without any parameters, the argument list
	// might be nil!

	var argList *value.ConsCell
	var ok bool
	var argCnt, idx int

	if argList, ok = inv.Car.Cdr.(*value.ConsCell); !ok {
		if value.IsNil(inv.Car.Cdr) {
			argList = new(value.ConsCell)
		} else {
			return value.NIL, SyntaxErrorf("Malformed list in function call: CDR is neither NIL nor a cons cell: %T",
				inv.Car.Cdr)
		}
	}

	if argCnt = argList.ActualLength(); argCnt != len(fn.Args) {
		return nil, fmt.Errorf("Wrong number of arguments for funcall: Expected %d, got %d %s",
			len(fn.Args),
			argCnt,
			argList.String())
	}

	// Dienstag, 03. 10. 2017, 00:12
	// This is not right - I need to use the *function's* environment, not
	// that of the current call-stack!
	//var env = value.NewEnvironment(inter.env)
	var env = value.NewEnvironment(fn.Env)

	// Is there a more elegant way to skip this? The loop over the
	// arguments causes a panic if the argument list is empty.
	if argCnt == 0 {
		goto EVALUATE
	}

	for ; argList != nil; argList = argList.Cdr.(*value.ConsCell) {
		var sym = fn.Args[idx]
		var val value.LispValue

		if val, err = inter.Eval(argList.Car); err != nil {
			return nil, err
		}

		env.Data[string(sym)] = val
		idx++
		if argList.Cdr == nil {
			break
		}
	}

	// Once we have environment put together, it's just walking over the
	// body and evaluating each element in turn, returning the value of the
	// last element.
EVALUATE:
	if inter.debug {
		env.Dump(os.Stdout)
	}
	var oldEnv = inter.env
	defer func() { inter.env = oldEnv }()
	inter.env = env
	//defer func() { inter.env = inter.env.Parent }()
	var res value.LispValue

	for _, exp := range fn.Body {
		if res, err = inter.Eval(exp); err != nil {
			return nil, err
		}
	}

	return res, nil
} // func (inter *Interpreter) evalFuncall(fun value.Function) (value.LispValue, errror)

func (inter *Interpreter) evalIf(l *value.List) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}

	if l.Length < 3 || l.Length > 4 {
		return nil, SyntaxError(fmt.Sprintf("Invalid number of elements for IF-clause: %d (expected 3 or 4)", l.Length))
	}

	// Freitag, 15. 09. 2017, 21:28
	// I need to *EVALUATE* the condition first!
	var condVal = l.Car.Cdr.(*value.ConsCell).Car //.Bool()
	var cond value.LispValue
	var err error

	if cond, err = inter.Eval(condVal); err != nil {
		return value.NIL, err
	} else if cond.Bool() {
		return inter.Eval(l.Car.Cdr.(*value.ConsCell).Cdr.(*value.ConsCell).Car)
	} else if l.Length == 4 {
		return inter.Eval(l.Car.Cdr.(*value.ConsCell).Cdr.(*value.ConsCell).Cdr.(*value.ConsCell).Car)
	}

	return value.NIL, nil
} // func (inter *Interpreter) evalIf(l *value.List) (value.LispValue, error)

/////////////////////////////////////////////////////////////////////////////
// Arithmetic ///////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////

// I think it would be preferrable to have arithmetic use a matrix to determine what
// operand gets promoted to what type.

func (inter *Interpreter) evalPlus(l *value.List) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
		fmt.Println(l.String())
		spew.Dump(l)
	}

	var cnt value.Number = value.IntValue(0)

	for v := l.Car.Cdr; v != nil; v = v.(*value.ConsCell).Cdr {
		var val value.LispValue
		var err error

		if v.(*value.ConsCell).Car == nil {
			return nil, &TypeError{expected: "Number", actual: "nil"}
		} else if val, err = inter.Eval(v.(*value.ConsCell).Car); err != nil {
			return nil, err
		} else if !value.IsNumber(val) {
			return nil, &TypeError{
				expected: "Number",
				actual:   val.Type().String(),
			}
		} else if cnt, err = evalAddition(cnt, val.(value.Number)); err != nil {
			return nil, err
		}
		/*else if val.Type() != types.Integer {
			return nil, &TypeError{
				expected: "Number",
				actual:   val.Type().String(),
			}
		} */

		//cnt += val.(value.IntValue)

	}

	return cnt, nil
} // func (inter *Interpreter) evalPlus(l *value.List) (value.LispValue, error)

func (inter *Interpreter) evalMinus(l *value.List) (value.LispValue, error) {
	var cnt value.Number
	var val value.LispValue
	var err error

	// Mittwoch, 25. 10. 2017, 15:01
	// XXX Ich glaube, folgendes Programm funktioniert im Moment nicht:
	// (define x 10)
	// (print (- x))

	if inter.debug {
		krylib.Trace()
	}

	if l.Length < 2 {
		return value.NIL, SyntaxError("Too few arguments for -")
	} else if l.Length == 2 {
		// FIXME Evaluate the argument!!!
		if val, err = inter.Eval(l.Car.Cdr.(*value.ConsCell).Car); err != nil {
			return nil, err
		} else if !value.IsNumber(val) {
			return nil, &TypeError{
				expected: "Number",
				actual:   l.Car.Cdr.(*value.ConsCell).Car.Type().String(),
			}
		}

		//return -(l.Car.Cdr.(*value.ConsCell).Car.(value.IntValue)), nil
		return evalNegate(val.(value.Number))
	}

	// I need to eval all arguments!!!
	//cnt = l.Car.Cdr.(*value.ConsCell).Car.(value.IntValue)
	if val, err = inter.Eval(l.Car.Cdr.(*value.ConsCell).Car); err != nil {
		return value.NIL, err
	} else if !value.IsNumber(val) {
		return value.NIL, &TypeError{
			expected: "Number",
			actual:   val.Type().String(),
		}
	}

	cnt = val.(value.Number)

	for v := l.Car.Cdr.(*value.ConsCell).Cdr; v != nil; v = v.(*value.ConsCell).Cdr {
		if val, err = inter.Eval(v.(*value.ConsCell).Car); err != nil {
			return value.NIL, err
		} else if !value.IsNumber(val) {
			return value.NIL, &TypeError{
				expected: "Number",
				actual:   val.Type().String(),
			}
		} else if cnt, err = evalSubtraction(cnt, val.(value.Number)); err != nil {
			return nil, err
		}
	}

	return cnt, nil
} // func (inter *Interpreter) evalMinus(l *value.List) (value.LispValue, error)

func (inter *Interpreter) evalMultiply(l *value.List) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}

	if l.Length == 1 {
		return value.IntValue(1), nil
	} else if l.Length == 2 {
		return inter.Eval(l.Car.Cdr.(*value.ConsCell).Car)
	} else if inter.debug {
		spew.Printf("evalMultiply %#v\n",
			l)
	}

	var (
		err    error
		resRaw value.LispValue
		res    value.Number
	)

	if resRaw, err = inter.Eval(l.Car.Cdr.(*value.ConsCell).Car); err != nil {
		return value.NIL, err
		//} else if resRaw.Type() == types.Integer {
	} else if !value.IsNumber(resRaw) {
		return value.NIL, &TypeError{
			expected: "Number",
			actual:   resRaw.Type().String(),
		}
	}

	res = resRaw.(value.Number)

	if inter.debug {
		spew.Printf("MULTIPLY the following numbers: %#v\n",
			l.Car.Cdr.(*value.ConsCell).Cdr.(*value.ConsCell))
	}

	for v := l.Car.Cdr.(*value.ConsCell).Cdr.(*value.ConsCell); v != nil; v = v.Cdr.(*value.ConsCell) {
		// Ich muss hier v.Car evaluieren!
		var cval value.LispValue
		if cval, err = inter.Eval(v.Car); err != nil {
			return value.NIL, err
		} else if !value.IsNumber(cval) {
			return value.NIL, &TypeError{
				expected: "Number",
				actual:   cval.Type().String(),
			}
		} else if res, err = evalMultiplication(res, cval.(value.Number)); err != nil {
			return nil, err
		} else if v.Cdr == nil {
			break
		}
	}

	return res, nil
} // func (inter *Interpreter) evalMultiply(l *value.List) (value.LispValue, error)

func (inter *Interpreter) evalDivide(l *value.List) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}
	// Montag, 11. 09. 2017, 20:12
	// In Common Lisp and Scheme, passing a single argument x returns
	// 1/x, as a rational number. We do not support rational numbers, yet.
	if l.Length < 3 {
		return nil, SyntaxError("Too few arguments for division (need at least 2)")
	}

	var val value.LispValue
	var err error

	if val, err = inter.Eval(l.Car.Cdr.(*value.ConsCell).Car); err != nil {
		return value.NIL, err
	} else if !value.IsNumber(val) {
		return value.NIL, &TypeError{
			expected: "Number",
			actual:   val.Type().String(),
		}
	}

	var res = val.(value.Number)

	for c := l.Car.Cdr.(*value.ConsCell).Cdr; c != nil; c = c.(*value.ConsCell).Cdr {
		v := c.(*value.ConsCell)
		if val, err = inter.Eval(v.Car); err != nil {
			return value.NIL, err
		} else if value.IsNumber(val) {
			var n = val.(value.Number)
			if !n.IsZero() {
				if res, err = evalDivision(res, n); err != nil {
					return value.NIL, err
				}
			} else {
				return nil, &ValueError{val: n}
			}

			if v.Cdr == nil {
				v = nil
			}
		} else {
			return nil, &TypeError{
				expected: "Number",
				actual:   v.Car.Type().String(),
			}
		}
	}

	return res, nil
} // func (inter *Interpreter) evalDivide(l *value.List) (value.LispValue, error)

/////////////////////////////////////////////////////////////////////////////
// Fundamental Lisp stuff ///////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////

func (inter *Interpreter) evalDefun(l *value.List) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}
	// (defun square (x) (* x x))
	// Nah, that is not sufficient - in Common Lisp, a function can also
	// have a documentation string.
	// So I need to check if the third element of the list is a string.
	var fn *value.Function
	var err error
	var ok bool
	var name value.Symbol
	var val value.LispValue
	var docstring value.StringValue
	var lambdaList = &value.List{
		Length: l.Length - 1,
		Car: &value.ConsCell{
			Car: value.Symbol("LAMBDA"),
			Cdr: l.Car.Cdr.(*value.ConsCell).Cdr.(*value.ConsCell),
		},
	}

	// Position 1 should be the symbol that is going to be the function's name.
	// Position 2 should be the (possibly empty) argument list
	// Position 3, IF it is a string, is used as documentation for the function.
	//
	// The rest is the body of the function; when the function is called, the
	// body evaluated sequentially.
	if val, _ = l.Nth(1); err != nil {
		return value.NIL, err
	} else if name, ok = val.(value.Symbol); !ok {
		return value.NIL, fmt.Errorf("First argument to defun must be a symbol, not a %T (%s)",
			val.Type().String(),
			val.String())
	} else if val, _ = l.Nth(2); val.Type() == types.String {
		docstring = val.(value.StringValue)
		lambdaList.Car.Cdr = lambdaList.Car.Cdr.(*value.ConsCell).Cdr
	}

	if fn, err = inter.evalLambda(lambdaList); err != nil {
		return value.NIL, err
	}

	fn.DocString = docstring

	inter.fnEnv.Set(name.String(), fn)

	return name, nil
} // func (inter *Interpreter) evalDefun(l *value.List) (value.LispValue, error)

func (inter *Interpreter) evalEq(l *value.List) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}
	if l.Length != 3 {
		return value.NIL, SyntaxError(fmt.Sprintf("Invalid number of arguments to EQ: %d (expected 2)",
			l.Length-1))
	}

	// Samstag, 16. 09. 2017, 13:37
	// Damn it, I need to evaluate these arguments, too.
	var raw1, raw2, v1, v2 value.LispValue
	var err error

	raw1, _ = l.Nth(1)
	raw2, _ = l.Nth(2)

	if v1, err = inter.Eval(raw1); err != nil {
		return value.NIL, err
	} else if v2, err = inter.Eval(raw2); err != nil {
		return value.NIL, err
	} else if v1.Type() != v2.Type() {
		if value.IsNumber(v1) && value.IsNumber(v2) {
			var res compare.Result
			if res, err = evalPolymorphCmp(v1.(value.Number), v2.(value.Number)); err != nil {
				return value.NIL, err
			}

			if res == compare.Equal {
				return value.T, nil
			}
		}

		return value.NIL, nil
	}

	if v1.Eq(v2) {
		return value.T, nil
	}

	return value.NIL, nil
} // func (inter *Interpreter) evalEq(l *value.List) (value.LispValue, error)

// nolint: gocyclo
func (inter *Interpreter) evalLessThan(l *value.List) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}

	if l.Length < 2 {
		return value.NIL, SyntaxError("Too few arguments for <")
	} else if l.Length == 2 {
		return value.T, nil
	}

	var v1, v2, raw1, raw2 value.LispValue
	var err error

	raw1 = l.Car.Cdr.(*value.ConsCell).Car
	raw2 = l.Car.Cdr.(*value.ConsCell).Cdr.(*value.ConsCell).Car

	if v1, err = inter.Eval(raw1); err != nil {
		return value.NIL, err
	} else if v2, err = inter.Eval(raw2); err != nil {
		return value.NIL, err
	} else if !value.IsNumber(v1) {
		return value.NIL, &TypeError{
			expected: "Number",
			actual:   v1.Type().String(),
		}
	} else if !value.IsNumber(v2) {
		return value.NIL, &TypeError{
			expected: "Number",
			actual:   v2.Type().String(),
		}
	} else if l.Length == 2 {
		return value.T, nil
	}

	var cmpResult compare.Result

	if cmpResult, err = evalPolymorphCmp(v1.(value.Number), v2.(value.Number)); err != nil {
		return value.NIL, err
	} else if cmpResult != compare.LessThan {
		return value.NIL, err
	} else if l.Length == 3 {
		return value.T, nil
	}

	for c := l.Car.Cdr.(*value.ConsCell).Cdr.(*value.ConsCell).Cdr.(*value.ConsCell); !value.IsNil(c); c = c.Cdr.(*value.ConsCell) {
		var res compare.Result
		v1 = v2
		raw2 = c.Car

		if v2, err = inter.Eval(raw2); err != nil {
			return value.NIL, err
		} else if !value.IsNumber(v2) {
			return value.NIL, &TypeError{
				expected: "Number",
				actual:   v2.Type().String(),
			}
		} else if res, err = evalPolymorphCmp(v1.(value.Number), v2.(value.Number)); err != nil {
			return value.NIL, err
		} else if res != compare.LessThan {
			return value.NIL, err
		} else if value.IsNil(c.Cdr) {
			break
		}
	}

	return value.T, nil
} // func (inter *Interpreter) evalLessThan(l *value.List) (value.LispValue, error)

// nolint: gocyclo
func (inter *Interpreter) evalGreaterThan(l *value.List) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}

	if l.Length < 2 {
		return value.NIL, SyntaxError("Too few arguments for <")
	} else if l.Length == 2 {
		return value.T, nil
	}

	var v1, v2, raw1, raw2 value.LispValue
	var err error
	var res compare.Result

	raw1 = l.Car.Cdr.(*value.ConsCell).Car
	raw2 = l.Car.Cdr.(*value.ConsCell).Cdr.(*value.ConsCell).Car

	if v1, err = inter.Eval(raw1); err != nil {
		return value.NIL, err
	} else if v2, err = inter.Eval(raw2); err != nil {
		return value.NIL, err
	} else if !value.IsNumber(v1) {
		return value.NIL, &TypeError{
			expected: "Number",
			actual:   v1.Type().String(),
		}
	} else if !value.IsNumber(v2) {
		return value.NIL, &TypeError{
			expected: "Number",
			actual:   v2.Type().String(),
		}
	} else if res, err = evalPolymorphCmp(v1.(value.Number), v2.(value.Number)); err != nil {
		return value.NIL, err
	} else if l.Length == 3 {
		if res == compare.GreaterThan {
			return value.T, nil
		}

		return value.NIL, nil
	}

	for c := l.Car.Cdr.(*value.ConsCell).Cdr.(*value.ConsCell).Cdr.(*value.ConsCell); c != nil; c = c.Cdr.(*value.ConsCell) {
		v1 = v2
		raw2 = c.Car

		if v2, err = inter.Eval(raw2); err != nil {
			return value.NIL, err
		} else if !value.IsNumber(v2) {
			return value.NIL, &TypeError{
				expected: "Number",
				actual:   v2.Type().String(),
			}
		} else if res, err = evalPolymorphCmp(v1.(value.Number), v2.(value.Number)); err != nil {
			return value.NIL, err
		} else if res != compare.GreaterThan {
			return value.NIL, nil
		} else if c.Cdr == nil {
			break
		}
	}

	return value.T, nil
} // func (inter *Interpreter) evalGreaterThan(l *value.List) (value.LispValue, error)

// nolint: gocyclo
func (inter *Interpreter) evalLessEqual(l *value.List) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}

	if l.Length < 2 {
		return value.NIL, SyntaxError("Too few arguments for <")
	} else if l.Length == 2 {
		return value.T, nil
	}

	var v1, v2, raw1, raw2 value.LispValue
	var err error
	var res compare.Result

	raw1 = l.Car.Cdr.(*value.ConsCell).Car
	raw2 = l.Car.Cdr.(*value.ConsCell).Cdr.(*value.ConsCell).Car

	if v1, err = inter.Eval(raw1); err != nil {
		return value.NIL, err
	} else if v2, err = inter.Eval(raw2); err != nil {
		return value.NIL, err
	} else if !value.IsNumber(v1) {
		return value.NIL, &TypeError{
			expected: "Number",
			actual:   v1.Type().String(),
		}
	} else if value.IsNumber(v2) {
		return value.NIL, &TypeError{
			expected: "Number",
			actual:   v2.Type().String(),
		}
	} else if res, err := evalPolymorphCmp(v1.(value.Number), v2.(value.Number)); err != nil {
		return value.NIL, err
	} else if l.Length == 3 {
		if res == compare.GreaterThan {
			return value.NIL, nil
		}

		return value.T, nil
	}

	for c := l.Car.Cdr.(*value.ConsCell).Cdr.(*value.ConsCell).Cdr.(*value.ConsCell); c != nil; c = c.Cdr.(*value.ConsCell) {
		v1 = v2
		raw2 = c.Car

		if v2, err = inter.Eval(raw2); err != nil {
			return value.NIL, err
		} else if !value.IsNumber(v2) {
			return value.NIL, &TypeError{
				expected: "Number",
				actual:   v2.Type().String(),
			}
		} else if res, err = evalPolymorphCmp(v1.(value.Number), v2.(value.Number)); err != nil {
			return value.NIL, err
		} else if res == compare.GreaterThan {
			return value.NIL, nil
		} else if c.Cdr == nil {
			break
		}
	}

	return value.T, nil
} // func (inter *Interpreter) evalLessEqual(l *value.List) (value.LispValue, error)

// nolint: gocyclo
func (inter *Interpreter) evalGreaterEqual(l *value.List) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}

	if l.Length < 2 {
		return value.NIL, SyntaxError("Too few arguments for <")
	} else if l.Length == 2 {
		return value.T, nil
	}

	var v1, v2, raw1, raw2 value.LispValue
	var err error
	var res compare.Result

	raw1 = l.Car.Cdr.(*value.ConsCell).Car
	raw2 = l.Car.Cdr.(*value.ConsCell).Cdr.(*value.ConsCell).Car

	if v1, err = inter.Eval(raw1); err != nil {
		return value.NIL, err
	} else if v2, err = inter.Eval(raw2); err != nil {
		return value.NIL, err
	} else if !value.IsNumber(v1) {
		return value.NIL, &TypeError{
			expected: "Number",
			actual:   v1.Type().String(),
		}
	} else if !value.IsNumber(v2) {
		return value.NIL, &TypeError{
			expected: "Number",
			actual:   v2.Type().String(),
		}
	} else if res, err = evalPolymorphCmp(v1.(value.Number), v2.(value.Number)); err != nil {
		return value.NIL, err
	} else if l.Length == 3 {
		if res == compare.LessThan {
			return value.NIL, nil
		}

		return value.T, nil
	}

	for c := l.Car.Cdr.(*value.ConsCell).Cdr.(*value.ConsCell).Cdr.(*value.ConsCell); c != nil; c = c.Cdr.(*value.ConsCell) {
		v1 = v2
		raw2 = c.Car

		if v2, err = inter.Eval(raw2); err != nil {
			return value.NIL, err
		} else if !value.IsNumber(v2) {
			return value.NIL, &TypeError{
				expected: "Number",
				actual:   v2.Type().String(),
			}
		} else if res, err = evalPolymorphCmp(v1.(value.Number), v2.(value.Number)); err != nil {
			return value.NIL, err
		} else if res == compare.LessThan {
			return value.NIL, nil
		} else if c.Cdr == nil {
			break
		}
	}

	return value.T, nil
} // func (inter *Interpreter) evalGreaterEqual(l *value.List) (value.LispValue, error)

func (inter *Interpreter) evalCons(l *value.List) (v value.LispValue, e error) {
	if inter.debug {
		krylib.Trace()
	}

	// Strictly speaking, I should check if the second value is nil.
	// cons'ing some value to nil gives a list.
	// Also, consing to a list should return another list.
	if l.Length != 3 {
		return value.NIL, SyntaxError("CONS takes exactly TWO (2) arguments")
	}

	var raw1, raw2, val1, val2 value.LispValue
	var err error

	raw1, _ = l.Nth(1)
	raw2, _ = l.Nth(2)

	if val1, err = inter.Eval(raw1); err != nil {
		return value.NIL, err
	} else if val2, err = inter.Eval(raw2); err != nil {
		return value.NIL, err
	}

	if value.IsNil(val2) {
		return &value.List{
			Car: &value.ConsCell{
				Car: val1,
				Cdr: nil,
			},
			Length: 1,
		}, nil
	} else if val2.Type() == types.List {
		var v2l = val2.(*value.List)
		return &value.List{
			Car: &value.ConsCell{
				Car: val1,
				Cdr: v2l.Car,
			},
			Length: v2l.Length + 1,
		}, nil
	}

	var cell = &value.ConsCell{
		Car: val1,
		Cdr: val2,
	}

	return cell, nil
} // func (inter *Interpreter) evalCons(l *value.List) (value.LispValue, error)

// nolint: gocyclo
func (inter *Interpreter) evalLet(l *value.List) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}

	if l.Length < 2 {
		return value.NIL, SyntaxError("Too few parameters for LET")
	}

	if inter.debug {
		spew.Dump(l)
	}

	var bindings value.LispValue
	var err error

	if bindings, err = l.Nth(1); err != nil {
		return value.NIL, err
	} else if bindings.Type() != types.List {
		return value.NIL, &TypeError{
			expected: "List",
			actual:   bindings.Type().String(),
		}
	}

	var env = value.NewEnvironment(inter.env)

	for v := bindings.(*value.List).Car; v != nil; v = v.Cdr.(*value.ConsCell) {
		var symbol = v.Car.(*value.List).Car.Car
		var rawValue = v.Car.(*value.List).Car.Cdr.(*value.ConsCell).Car
		var val value.LispValue

		if inter.debug {
			fmt.Printf("LET-form: evaluate binding %s\n",
				spew.Sdump(v.Car))
		}

		if symbol.Type() != types.Symbol {
			return value.NIL, &TypeError{
				expected: "Symbol",
				actual:   symbol.Type().String(),
			}
		} else if val, err = inter.Eval(rawValue); err != nil {
			return value.NIL, err
		}

		env.Ins(string(symbol.(value.Symbol)), val)
		if v.Cdr == nil {
			break
		}
	}

	if inter.debug {
		env.Dump(os.Stdout)
	}

	var val value.LispValue
	var expr *value.ConsCell

	inter.env = env
	defer func() { inter.env = inter.env.Parent }()

	expr = l.Car.Cdr.(*value.ConsCell).Cdr.(*value.ConsCell)

	for expr != nil {
		if val, err = inter.Eval(expr.Car); err != nil {
			return value.NIL, err
		} else if expr.Cdr != nil {
			expr = expr.Cdr.(*value.ConsCell)
		} else {
			expr = nil
		}
	}

	return val, nil
} // func (inter *Interpreter) evalLet(l *value.List) (value.LispValue, error)

func (inter *Interpreter) evalNot(l *value.List) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}

	var val value.LispValue
	var err error

	if l.Length != 2 {
		return value.NIL, SyntaxError("Wrong number of arguments for NOT")
	} else if val, err = inter.Eval(l.Car.Cdr.(*value.ConsCell).Car); err != nil {
		return nil, err
	} else if value.IsNil(val) {
		return value.T, nil
	} else if val.Bool() {
		return value.NIL, nil
	} else {
		return value.T, nil
	}
} // func (inter *Interpreter) evalNot(l *value.List) (value.LispValue, error)

func (inter *Interpreter) evalAnd(l *value.List) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}

	var err error
	var val value.LispValue

	// (and) evaluates to true, both in Common Lisp and in Scheme
	if l.Car == nil || l.Length == 1 {
		return value.T, nil
	}

	for cell := l.Car.Cdr.(*value.ConsCell); cell != nil; cell = cell.Next() {
		if value.IsNil(cell.Car) {
			return value.NIL, nil
		}

		if val, err = inter.Eval(cell.Car); err != nil {
			return nil, err
		} else if !val.Bool() {
			return value.NIL, nil
		}
	}

	return val, nil
} // func (inter *Interpreter) evalAnd(l *value.List) (value.LispValue, error)

func (inter *Interpreter) evalOr(l *value.List) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}

	var err error
	var val value.LispValue

	// (or) evaluates to false, both in Common Lisp and in Scheme
	if l.Car == nil || l.Length == 1 {
		return value.NIL, nil
	}

	for cell := l.Car.Cdr.(*value.ConsCell); cell != nil; cell = cell.Next() {
		if val, err = inter.Eval(cell.Car); err != nil {
			return nil, err
		} else if val.Bool() {
			return val, nil
		}
	}

	return value.NIL, nil
} // func (inter *Interpreter) evalOr(l *value.List) (value.LispValue, error)

func (inter *Interpreter) evalDefine(l *value.List) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}

	// (define x 3)
	var err error
	var val value.LispValue

	if l == nil || l.Car == nil || l.Length != 3 {
		return value.NIL, SyntaxErrorf("DEFINE expects a list of exactly three arguments (%d were given: %s)",
			l.Length, l.String())
	} else if l.Car.Cdr.(*value.ConsCell).Car.Type() != types.Symbol {
		return value.NIL, SyntaxErrorf("Second argument to DEFINE must be a symbol, not a %s",
			l.Car.Cdr.(*value.ConsCell).Car.Type())
	} else if val, err = inter.Eval(l.Car.Cdr.(*value.ConsCell).Cdr.(*value.ConsCell).Car); err != nil {
		return value.NIL, err
	}

	inter.env.Ins(
		l.Car.Cdr.(*value.ConsCell).Car.String(),
		val)

	return val, nil
} // func (inter *Interpreter) evalDefine(l *value.List) (value.LispValue, error)

func (inter *Interpreter) evalSet(l *value.List) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}
	// (set! x 3)
	var err error
	var val value.LispValue

	if l == nil || l.Car == nil || l.Length != 3 {
		return value.NIL, SyntaxError("DEFINE expects a list of exactly three arguments")
	} else if l.Car.Cdr.(*value.ConsCell).Car.Type() != types.Symbol {
		return value.NIL, SyntaxErrorf("Second argument to DEFINE must be a symbol, not a %s",
			l.Car.Cdr.(*value.ConsCell).Car.Type())
	} else if val, err = inter.Eval(l.Car.Cdr.(*value.ConsCell).Cdr.(*value.ConsCell).Car); err != nil {
		return value.NIL, err
	}

	inter.env.Set(
		l.Car.Cdr.(*value.ConsCell).Car.String(),
		val)

	return val, nil
} // func (inter *Interpreter) evalSet(l *value.List) (value.LispValue, error)

func (inter *Interpreter) evalPrint(l *value.List) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}

	if l == nil || l.Car == nil {
		return value.NIL, nil
	} else if l.Length < 2 {
		return value.NIL, nil
	}

	for cell := l.Car.Cdr.(*value.ConsCell); cell != nil; cell = cell.Next() {
		var val value.LispValue
		var err error

		if val, err = inter.Eval(cell.Car); err != nil {
			return value.NIL, err
		} else if val == nil {
			val = value.NIL
		}

		inter.stdout.Write([]byte(val.String() + "\n"))
	}

	return value.NIL, nil
} // func (inter *Interpreter) evalPrint(l *value.List) (value.LispValue, error)

func (inter *Interpreter) evalApply(l *value.List) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
		inter.stdout.Write([]byte(l.String()))
	}

	// (apply #somefn '(arg1 arg2 arg3)) <=> (somefn arg1 arg2 arg3)
	var (
		err                      error
		fnspec, fn, val, arglist value.LispValue
	)

	if l == nil || l.Car == nil || l.Length < 3 {
		return value.NIL, SyntaxError("APPLY must be called with at least two arguments (function and one or more arguments)")
	} else if fnspec, err = l.Nth(1); err != nil {
		return value.NIL, err
	} else if fn, err = inter.Eval(fnspec); err != nil {
		return value.NIL, err
	} else if fn.Type() != types.Function {
		return value.NIL, &TypeError{
			expected: "Function",
			actual:   fn.Type().String(),
		}
	} else if val, err = l.Nth(2); err != nil {
		return value.NIL, err
	} else if arglist, err = inter.Eval(val); err != nil {
		return value.NIL, err
	} else if arglist.Type() != types.List {
		return value.NIL, &TypeError{
			expected: "List",
			actual:   arglist.Type().String(),
		}
	}

	var funcall = &value.List{
		Car: &value.ConsCell{
			Car: fn,
			Cdr: arglist.(*value.List).Car,
		},
		Length: arglist.(*value.List).Length + 1,
	}

	if inter.debug {
		spew.Printf("APPLY: %#v\n",
			funcall)
	}

	return inter.evalFuncall(funcall)
} // func (inter *Interpreter) evalApply(l *value.List) (value.LispValue, error)

func (inter *Interpreter) evalCar(l *value.List) (v value.LispValue, e error) {
	if inter.debug {
		krylib.Trace()
		inter.stdout.Write([]byte(l.String()))
	}

	defer func() {
		spew.Printf("CAR returns %#v\n", v)
	}()

	if value.IsNil(l.Car.Cdr.(*value.ConsCell).Car) {
		inter.stdout.Write([]byte("CAR of NIL is NIL"))
		return value.NIL, nil
	}

	// Dienstag, 10. 10. 2017, 20:27
	// Should I evaluate the argument to car?

	var (
		val value.LispValue
		err error
	)

	if val, err = inter.Eval(l.Car.Cdr.(*value.ConsCell).Car); err != nil {
		return value.NIL, err
	}

	switch v := val.(type) {
	case *value.List:
		return v.Car.Car, nil
	case *value.ConsCell:
		return v.Car, nil
	case value.NilValue:
		return value.NIL, nil
	default:
		return value.NIL, &TypeError{
			expected: "ConsCell or List",
			actual:   val.Type().String(),
		}
	}
} // func (inter *Interpreter) evalCar(l *value.List) (value.LispValue, error)

func (inter *Interpreter) evalCdr(l *value.List) (v value.LispValue, e error) {
	if inter.debug {
		krylib.Trace()
		defer func() {
			var msg string
			if v != nil {
				msg = v.String()
			} else {
				msg = "NIL"
			}

			fmt.Printf("Return from CDR %s: %s\n",
				l.String(),
				msg)
		}()
	}

	if l == nil || l.Car == nil || l.Length != 2 {
		return value.NIL, SyntaxErrorf("Malformed call to CDR: %s",
			spew.Sdump(l))
	}

	var err error
	var input, val value.LispValue

	if input, err = l.Nth(1); err != nil {
		return value.NIL, err
	} else if val, err = inter.Eval(input); err != nil {
		return value.NIL, err
	} //else if val.Type() == types.List {

	switch val.Type() {
	case types.List:
		var vl = val.(*value.List)
		if vl.Length < 2 {
			return value.NIL, nil
		}

		return &value.List{
			Car:    vl.Car.Cdr.(*value.ConsCell),
			Length: vl.Length - 1,
		}, nil
	case types.ConsCell:
		return val.(*value.ConsCell).Cdr, nil
	case types.Nil:
		return value.NIL, nil
	default:
		return nil, &TypeError{
			expected: "List or ConsCell",
			actual:   val.Type().String(),
		}
	}
} // func (inter *Interpreter) evalCdr(l *value.List) (value.LispValue, error)

func (inter *Interpreter) evalFn(l *value.List) (value.LispValue, error) {
	// evalFn does not evaluate a function CALL, but the function OBJECT itself.
	// This may be a lambda list, or a function object.
	if inter.debug {
		krylib.Trace()
	}

	var (
		cadr, fn value.LispValue
		sym      value.Symbol
		err      error
	)

	if l == nil || l.Length != 2 {
		return value.NIL, SyntaxError("FN requires exactly one argument")
	} else if cadr, err = l.Nth(1); err != nil {
		return value.NIL, err
	}

	switch cadr.Type() {
	case types.Symbol:
		sym = cadr.(value.Symbol)
		var found bool
		if fn, found = inter.fnEnv.Get(string(sym)); !found {
			return value.NIL, MissingFunctionError(sym)
		}
	case types.Function:
		fn = cadr.(*value.Function)
	default:
		fmt.Printf("FN: Invalid argument %s\n", spew.Sdump(cadr))
		return value.NIL, &TypeError{
			expected: "Symbol",
			actual:   cadr.Type().String(),
		}
	}

	return fn, nil
} // func (inter *Interpreter) evalFn(l *value.List) (value.LispValue, error)

func (inter *Interpreter) evalIsNil(l *value.List) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}

	if l.Length != 2 {
		return value.NIL, SyntaxError("NIL? expects exactly one argument")
	} else if inter.debug {
		spew.Printf("(NIL? %#v)\n",
			l.Car.Cdr)
	}

	// Donnerstag, 19. 10. 2017, 17:26
	// Abobo - I need to ____ING EVALUATE the ____ING argument!!!

	var val = l.Car.Cdr.(*value.ConsCell).Car
	var res value.LispValue
	var err error

	if res, err = inter.Eval(val); err != nil {
		return value.NIL, err
	} else if value.IsNil(res) {
		if inter.debug {
			spew.Printf("evalIsNil: Value %#v is NIL indeed!\n",
				val)
		}
		return value.T, nil
	} else if inter.debug {
		spew.Printf("evalIsNil: Value %#v is NOT NIL!\n",
			val)
	}

	return value.NIL, nil
} // func (inter *Interpreter) evalIsNil(l *value.List) (value.LispValue, error)

func (inter *Interpreter) evalList(l *value.List) (v value.LispValue, e error) {
	if inter.debug {
		krylib.Trace()
		defer func() {
			spew.Printf("LIST returns %#v\n", v)
		}()
	}

	if l.Length == 1 {
		return value.NIL, nil
	}

	var res = &value.List{
		Length: l.Length - 1,
		Car:    new(value.ConsCell),
	}

	var tail = res.Car

	if inter.debug {
		spew.Printf("LIST form: %#v\n", l)
	}

	for cell := l.Car.Cdr.(*value.ConsCell); !value.IsNil(cell); cell = cell.Cdr.(*value.ConsCell) {
		var val value.LispValue
		var err error

		if val, err = inter.Eval(cell.Car); err != nil {
			return value.NIL, err
		} else if inter.debug {
			spew.Printf("LIST form: %#v => %#v\n",
				cell.Car,
				val)
		}

		tail.Car = val
		if cell.Cdr == nil {
			break
		} else {
			tail.Cdr = new(value.ConsCell)
			tail = tail.Cdr.(*value.ConsCell)
			tail.Car = value.NIL
		}
	}

	return res, nil
} // func (inter *Interpreter) evalList(l *value.List) (value.LispValue, error)

/////////////////////////////////////////////////////////////////////////////
// Arrays ///////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////

func (inter *Interpreter) evalAref(l *value.List) (v value.LispValue, e error) {
	if inter.debug {
		krylib.Trace()
		defer func() {
			spew.Printf("AREF returns %#v\n", v)
		}()
	}

	if l == nil || l.Length != 3 {
		return value.NIL, SyntaxError("Usage: (aref <array> <index>)")
	}

	var (
		arrRaw, idxRaw, tmp value.LispValue
		index               int
		arr                 value.Array
		err                 error
	)

	arrRaw, _ = l.Nth(1)
	idxRaw, _ = l.Nth(2)

	if tmp, err = inter.Eval(arrRaw); err != nil {
		return value.NIL, err
	} else if tmp.Type() != types.Array {
		if inter.debug {
			spew.Printf("First argument (the array) to AREF is not an array: %#v\n",
				tmp)
		}
		return value.NIL, &TypeError{
			expected: "Array",
			actual:   tmp.Type().String(),
		}
	}

	arr = tmp.(value.Array)

	if tmp, err = inter.Eval(idxRaw); err != nil {
		return value.NIL, err
	} else if tmp.Type() != types.Integer {
		return value.NIL, &TypeError{
			expected: "Integer",
			actual:   tmp.Type().String(),
		}
	}

	index = int(tmp.(value.IntValue))

	if index < 0 || index >= len(arr) {
		return value.NIL, &ValueError{val: value.IntValue(index)}
	}

	return arr[index], nil
} // func evalAref(l *value.List) (value.LispValue, error)

func (inter *Interpreter) evalApush(l *value.List) (v value.LispValue, e error) {
	if inter.debug {
		krylib.Trace()
		defer func() {
			spew.Printf("APUSH returns %#v\n", v)
		}()
	}

	if l == nil || l.Length < 3 {
		return value.NIL, SyntaxError("Usage: (apush <array> <val1>... )")
	}

	var tmp value.LispValue
	var arr value.Array
	var err error
	var ok bool

	if tmp, err = l.Nth(1); err != nil {
		return value.NIL, err
	} else if arr, ok = tmp.(value.Array); !ok {
		return value.NIL, &TypeError{
			expected: "Array",
			actual:   tmp.Type().String(),
		}
	} else if arr == nil {
		return value.NIL, &TypeError{
			expected: "Array",
			actual:   "nil",
		}
	} else if arr.Type() != types.Array {
		return value.NIL, &TypeError{
			expected: "Array",
			actual:   arr.Type().String(),
		}
	}

	for cell := l.Car.Cdr.(*value.ConsCell).Cdr.(*value.ConsCell); cell != nil; cell = cell.Cdr.(*value.ConsCell) {
		arr = append(arr, cell.Car)
		if cell.Cdr == nil {
			break
		}
	}

	return arr, nil
} // func (inter *Interpreter) evalApush(l *value.List) (v value.LispValue, e error)

// Samstag, 04. 11. 2017, 16:30
// I really have no idea if I actually need make-array, and if so, how I want
// to use it.
// But I think, it should take a single list as a parameter and convert that to
// an Array.
// There is, of course, the special case of no parameter: "(make-array)"
// In this case, it would make sense to return a new Array of length zero.
// Can I do that in Go?
func (inter *Interpreter) evalMakeArray(l *value.List) (v value.LispValue, e error) {
	if inter.debug {
		krylib.Trace()
		defer func() {
			spew.Printf("APUSH returns %#v\n", v)
		}()
	}

	if l == nil {
		return value.NIL, SyntaxError("Usage: (make-array '(a b c d))")
	} else if l.Length == 1 {
		return make(value.Array, 0, 10), nil
	}

	var err error
	var tmp1 = l.Car.Cdr.(*value.ConsCell).Car
	var tmp2 value.LispValue
	var ok bool
	var lst *value.List

	if tmp2, err = inter.Eval(tmp1); err != nil {
		return value.NIL, err
	} else if lst, ok = tmp2.(*value.List); !ok {
		return value.NIL, &TypeError{
			expected: "List",
			actual:   tmp2.Type().String(),
		}
	}

	if inter.debug {
		fmt.Printf("MAKE-ARRAY: Evaluate argument list: %s\n",
			lst.String())
	}

	var arr = make(value.Array, lst.Length)
	var idx int

	for cell := lst.Car; cell != nil; cell = cell.Cdr.(*value.ConsCell) {
		var val value.LispValue

		if inter.debug {
			spew.Printf("Argument #%d to MAKE-ARRAY: %#v",
				idx,
				cell.Car)
		}

		// if val, err = inter.Eval(cell.Car); err != nil {
		// 	return value.NIL, err
		// }
		val = cell.Car

		arr[idx] = val
		idx++
		if cell.Cdr == nil {
			break
		}
	}

	return arr, nil
} // func (inter *Interpreter) evalMakeArray(l *value.List) (v value.LispValue, e error)

/////////////////////////////////////////////////////////////////////////////
// Hash tables //////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////

// Since we have literal syntax for hash tables, this function does not really need
// any arguments, now, does it?
func (inter *Interpreter) evalMakeHash(l *value.List) (v value.LispValue, e error) {
	if inter.debug {
		krylib.Trace()
	}
	return make(value.Hashtable), nil
} // func (inter *Interpreter) evalMakeHash(l *value.List) (v value.LispValue, e error)

func (inter *Interpreter) evalHashref(l *value.List) (v value.LispValue, e error) {
	if inter.debug {
		krylib.Trace()
	}

	if l == nil || l.Length != 3 {
		return value.NIL, SyntaxError("HASHREF takes exactly *two* arguments")
	}

	var tmp, key value.LispValue
	var err error

	if tmp, err = l.Nth(1); err != nil {
		return value.NIL, err
	} else if tmp, err = inter.Eval(tmp); err != nil {
		return value.NIL, err
	} else if tmp.Type() != types.Hashtable {
		return value.NIL, &TypeError{
			expected: "Hashtable",
			actual:   tmp.Type().String(),
		}
	}

	var tbl = tmp.(value.Hashtable)

	if tmp, err = l.Nth(2); err != nil {
		return value.NIL, err
	} else if key, err = inter.Eval(tmp); err != nil {
		return value.NIL, err
	}

	var val value.LispValue
	var ok bool

	// Oh, wait. I had not really thought of the interface before.
	// In Common Lisp, hashref returns *two* values, the second value
	// is a flag indicating if the key was found in the hash table,
	// so users can distinguish between a given key not being present
	// in a hash table and a given key being present and having the value
	// NIL.
	//
	// While I like the idea, I have a hunch that it would be really
	// painful, exhausting and frustrating to implement. So I am
	// going to skip on that one and act like Lua.
	// I can add another special to explicitly check for the presence of
	// a given key to compensate.
	if val, ok = tbl[key]; !ok {
		return value.NIL, nil
	}

	return val, nil
} // func (inter *Interpreter) evalHashref(l *value.List) (v value.LispValue, e error)

func (inter *Interpreter) evalHasKey(l *value.List) (v value.LispValue, e error) {
	if inter.debug {
		krylib.Trace()
	}

	if l == nil || l.Length != 3 {
		return value.NIL, SyntaxError("HASHREF takes exactly *two* arguments")
	}

	var tmp, key value.LispValue
	var err error

	if tmp, err = l.Nth(1); err != nil {
		return value.NIL, err
	} else if tmp, err = inter.Eval(tmp); err != nil {
		return value.NIL, err
	} else if tmp.Type() != types.Hashtable {
		return value.NIL, &TypeError{
			expected: "Hashtable",
			actual:   tmp.Type().String(),
		}
	}

	var tbl = tmp.(value.Hashtable)

	if tmp, err = l.Nth(2); err != nil {
		return value.NIL, err
	} else if key, err = inter.Eval(tmp); err != nil {
		return value.NIL, err
	}

	var ok bool

	// Oh, wait. I had not really thought of the interface before.
	// In Common Lisp, hashref returns *two* values, the second value
	// is a flag indicating if the key was found in the hash table,
	// so users can distinguish between a given key not being present
	// in a hash table and a given key being present and having the value
	// NIL.
	//
	// While I like the idea, I have a hunch that it would be really
	// painful, exhausting and frustrating to implement. So I am
	// going to skip on that one and act like Lua.
	// I can add another special to explicitly check for the presence of
	// a given key to compensate.
	if _, ok = tbl[key]; !ok {
		return value.NIL, nil
	}

	return value.T, nil
} // func (inter *Interpreter) evalHasKey(l *value.List) (v value.LispValue, e error)

func (inter *Interpreter) evalHashSet(l *value.List) (v value.LispValue, e error) {
	if inter.debug {
		krylib.Trace()
	}

	// (hash-set tbl key val)
	if l == nil || l.Length != 4 {
		return value.NIL, SyntaxError("HASH-SET takes exactly *three* arguments")
	}

	var (
		tmp1, tmp2, key, val value.LispValue
		tbl                  value.Hashtable
		ok                   bool
		err                  error
	)

	if tmp1, err = l.Nth(1); err != nil {
		return value.NIL, err
	} else if tmp2, err = inter.Eval(tmp1); err != nil {
		return value.NIL, err
	} else if tmp2.Type() != types.Hashtable {
		return value.NIL, &TypeError{
			expected: "Hashtable",
			actual:   tmp2.Type().String(),
		}
	} else if tbl, ok = tmp2.(value.Hashtable); !ok {
		// CANTHAPPEN
		return value.NIL, &TypeError{
			expected: "Hashtable",
			actual:   tmp2.Type().String(),
		}
	} else if tmp1, err = l.Nth(2); err != nil {
		return value.NIL, err
	} else if key, err = inter.Eval(tmp1); err != nil {
		return value.NIL, err
	} else if tmp1, err = l.Nth(3); err != nil {
		return value.NIL, err
	} else if val, err = inter.Eval(tmp1); err != nil {
		return value.NIL, err
	}

	tbl[key] = val
	return val, nil
} // func (inter *Interpreter) evalHashSet(l *value.List) (v value.LispValue, e error)

func (inter *Interpreter) evalHashDelete(l *value.List) (v value.LispValue, e error) {
	if inter.debug {
		krylib.Trace()
	}

	// (hash-delete tbl key)
	if l == nil || l.Length != 3 {
		return value.NIL, SyntaxError("HASH-DELETE takes exactly *two* arguments")
	}

	var (
		tmp1, tmp2, key value.LispValue
		tbl             value.Hashtable
		ok              bool
		err             error
	)

	if tmp1, err = l.Nth(1); err != nil {
		return value.NIL, err
	} else if tmp2, err = inter.Eval(tmp1); err != nil {
		return value.NIL, err
	} else if tmp2.Type() != types.Hashtable {
		return value.NIL, &TypeError{
			expected: "Hashtable",
			actual:   tmp2.Type().String(),
		}
	} else if tbl, ok = tmp2.(value.Hashtable); !ok {
		// CANTHAPPEN
		return value.NIL, &TypeError{
			expected: "Hashtable",
			actual:   tmp2.Type().String(),
		}
	} else if tmp1, err = l.Nth(2); err != nil {
		return value.NIL, err
	} else if key, err = inter.Eval(tmp1); err != nil {
		return value.NIL, err
	}

	_, ok = tbl[key]
	delete(tbl, key)

	if ok {
		return value.T, nil
	}

	return value.NIL, nil
} // func (inter *Interpreter) evalHashDelete(l *value.List) (v value.LispValue, e error)

/////////////////////////////////////////////////////////////////////////////
// Regular expressions //////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////

func (inter *Interpreter) evalRegexpCompile(l *value.List) (v value.LispValue, e error) {
	if inter.debug {
		krylib.Trace()
	}

	if l == nil || l.Length != 2 {
		return value.NIL, SyntaxError("REGEXP-COMPILE takes exactly one argument")
	}

	var err error
	var tmp, arg value.LispValue

	tmp = l.Car.Cdr.(*value.ConsCell).Car

	if arg, err = inter.Eval(tmp); err != nil {
		return value.NIL, err
	} else if arg.Type() != types.String {
		return value.NIL, &TypeError{
			expected: "String",
			actual:   arg.Type().String(),
		}
	}

	var re *regexp.Regexp

	if re, err = regexp.Compile(string(arg.(value.StringValue))); err != nil {
		return value.NIL, &ValueError{
			val: arg,
			msg: err.Error(),
		}
	}

	var lispRe = &value.Regexp{Pat: re}

	return lispRe, nil
} // func (inter *Interpreter) evalRegexpCompile(l *value.List) (v value.LispValue, e error)

func (inter *Interpreter) evalRegexpMatch(l *value.List) (v value.LispValue, e error) {
	if inter.debug {
		krylib.Trace()
	}

	if l == nil || l.Length != 3 {
		return value.NIL, SyntaxError("REGEXP-MATCH takes exactly two arguments")
	}

	var raw1, raw2, val1, val2 value.LispValue
	var err error

	raw1 = l.Car.Cdr.(*value.ConsCell).Car
	raw2 = l.Car.Cdr.(*value.ConsCell).Cdr.(*value.ConsCell).Car

	if val1, err = inter.Eval(raw1); err != nil {
		return value.NIL, err
	} else if val1.Type() != types.Regexp {
		// We could make a special case here strings and compile them
		// on the fly.
		// But we don't have to. The interpreter is going to be slow
		// enough as it is, and adding to that the cost of compiling
		// regular expressions on very match does not make it better.
		return value.NIL, &TypeError{
			expected: "Regexp",
			actual:   val1.Type().String(),
		}
	} else if val2, err = inter.Eval(raw2); err != nil {
		return value.NIL, err
	} else if val2.Type() != types.String {
		return value.NIL, &TypeError{
			expected: "String",
			actual:   val2.Type().String(),
		}
	}

	// This is going to be slightly more complex than compiling regexps,
	// because Go has such a rich API for that, and I would like to keep
	// the Lisp API as simple as possible.
	// So what do I need?
	// Check if the string matches at all.
	// If it does, all the matches
	// Plus, if there are groupings in the regexp, we want those, too.
	//
	// That leaves us with ... FindAllStringSubmatch

	var (
		pat     = val1.(*value.Regexp)
		str     = val2.(value.StringValue)
		matches [][]string
	)

	// We want all matches, so the count argument to FindAllStringSubmatch
	// is -1.
	if matches = pat.Pat.FindAllStringSubmatch(string(str), -1); matches == nil {
		return value.NIL, nil
	}

	// If we arrive here, the regexp did match the string, and now we have
	// to create a Lisp data structure analogous to the return value
	// of FindAllStringSubmatch.

	var result = make(value.Array, len(matches))

	for i, match := range matches {
		var groups = make(value.Array, len(match))
		for j, sub := range match {
			groups[j] = value.StringValue(sub)
		}
		result[i] = groups
	}

	return result, nil
} // func (inter *Interpreter) evalRegexpMatch(l *value.List) (v value.LispValue, e error)

type loopVariable struct {
	sym  value.Symbol
	step value.LispValue
}

func (inter *Interpreter) evalDoLoop(l *value.List) (v value.LispValue, e error) {
	if inter.debug {
		krylib.Trace()
	}

	// A DO loop requires at least three arguments:
	// - Variables (may be an empty list)
	// - Exit condition (may be nil)
	// - Return value (may be nil)
	// Anything beyond these three is the loop body, but it is legal to
	// have loop with an empty body.
	if l == nil || l.Length < 4 {
		return value.NIL, SyntaxError("DO requires at least three arguments")
	}

	// The intimidating thing about DO is that is takes a rather large
	// number of arguments..

	var (
		err                           error
		varList, endTest, returnValue value.LispValue
	)

	if varList, err = l.Nth(1); err != nil {
		return value.NIL, err
	} else if endTest, err = l.Nth(2); err != nil {
		return value.NIL, err
	} else if returnValue, err = l.Nth(3); err != nil {
		return value.NIL, err
	}

	var (
		body *value.ConsCell
		//stepForms map[value.Symbol]value.LispValue
		stepForms []loopVariable
		loopEnv   *value.Environment //= value.NewEnvironment(inter.env)
		tmp       value.LispValue
	)

	if l.Length > 4 {
		//       DO                        vars                  exit condition        return value
		body = l.Car.Cdr.(*value.ConsCell).Cdr.(*value.ConsCell).Cdr.(*value.ConsCell).Cdr.(*value.ConsCell).Cdr.(*value.ConsCell)
	}

	if value.IsNil(varList) {
		goto LOOP
	} else if varList.Type() != types.List {
		fmt.Printf("Variable list in a DO-loop must be a *list*, not a %T\n",
			varList.Type().String())
		return value.NIL, &TypeError{
			expected: "List",
			actual:   varList.Type().String(),
		}
	} else if loopEnv, stepForms, err = inter.evalDoVariables(varList.(*value.List)); err != nil {
		return value.NIL, err
	}

	inter.env = loopEnv
	defer func() { inter.env = inter.env.Parent }()

	// Now we have the loop environment-frame with the freshly initialized
	// loop variables, it is time to execute the loop
LOOP:

	// Check for abort
	if tmp, err = inter.Eval(endTest); err != nil {
		var msg = fmt.Sprintf("Error evaluating end-test for DO-loop %s: %s",
			endTest.String(),
			err.Error())
		fmt.Println(msg)
		return value.NIL, errors.New(msg)
	} else if tmp.Bool() {
		// In the DO loop from Common Lisp, the test does not check,
		// like C's for loop, if the loop should *continue* to run,
		// but if it should be *ended*.
		// So if our test gives a true value,
		// we are done.
		goto END
	}

	// Run the loop body
	for cell := body; cell != nil; cell = cell.Cdr.(*value.ConsCell) {
		if tmp, err = inter.Eval(cell.Car); err != nil {
			var msg = fmt.Sprintf("Error evaluating loop body %s: %s",
				cell.Car.String(),
				err.Error())
			fmt.Println(msg)
			return value.NIL, errors.New(msg)
		}
	}

	// Evaluate the step forms for all loop variables
	for _, loopVar := range stepForms {
		var oldValue value.LispValue

		if inter.debug {
			oldValue, _ = loopEnv.Get(loopVar.sym.String())
		}

		if tmp, err = inter.Eval(loopVar.step); err != nil {
			var msg = fmt.Sprintf("Error updating loop variable %s with form %s: %s",
				loopVar.sym,
				loopVar.step,
				err.Error())
			fmt.Println(msg)
			return value.NIL, errors.New(msg)
		}

		loopEnv.Ins(loopVar.sym.String(), tmp)

		if inter.debug {
			fmt.Printf("DO: Update loop variable %s from %s to %s\n",
				loopVar.sym,
				oldValue,
				tmp)
		}
	}

	// Play it again, Sam...
	goto LOOP

END:
	if tmp, err = inter.Eval(returnValue); err != nil {
		var msg = fmt.Sprintf("Error evaluating return form %s: %s",
			returnValue.String(),
			err.Error())
		fmt.Println(msg)
		return value.NIL, errors.New(msg)
	}

	return tmp, nil
} // func (inter *Interpreter) evalDoLoop(l *value.List) (v value.LispValue, e error)

func (inter *Interpreter) evalDoVariables(varList *value.List) (*value.Environment, []loopVariable, error) {
	var (
		env       = value.NewEnvironment(inter.env)
		stepForms = make([]loopVariable, 0, varList.Length)
		err       error
	)

	for cell := varList.Car; !value.IsNil(cell); cell = cell.Cdr.(*value.ConsCell) {
		// Dienstag, 07. 11. 2017, 13:59
		// Bindings have the following form:
		// (<identifier> <init-form> <step-form>)
		// So I have to take this apart, first.
		var (
			loopvDecl               *value.List
			ok                      bool
			identifier              value.Symbol
			initForm, stepForm, tmp value.LispValue
		)

		if loopvDecl, ok = cell.Car.(*value.List); !ok {
			var msg = fmt.Sprintf("Loop variable declaration must be a List, not a %s",
				cell.Car.Type().String())
			fmt.Println("Error in DO-form: " + msg)
			return nil, nil, errors.New(msg)
		} else if loopvDecl.Length != 3 {
			var msg = fmt.Sprintf("Declaration of DO-loop variable must contain exactly three elements, found %d: %s",
				loopvDecl.Length,
				loopvDecl.String())
			fmt.Println(msg)
			return nil, nil, errors.New(msg)
		} else if identifier, ok = loopvDecl.Car.Car.(value.Symbol); !ok {
			return nil, nil, SyntaxErrorf("First element of declaration of loop variable must be a Symbol, not a %s",
				loopvDecl.Car.Car.Type().String())
		} else if initForm, err = loopvDecl.Nth(1); err != nil {
			fmt.Printf("DO-loop: Error getting init-form from loop variable declaration %s: %s",
				loopvDecl.String(),
				err.Error())
			return nil, nil, err
		} else if stepForm, err = loopvDecl.Nth(2); err != nil {
			fmt.Printf("DO-loop: Error getting step-form from loop variable declaration %s: %s",
				loopvDecl.String(),
				err.Error())
			return nil, nil, err
		} else if tmp, err = inter.Eval(initForm); err != nil {
			var msg = fmt.Sprintf("Error initializing loop variable %s with %s: %s",
				identifier.String(),
				loopvDecl.String(),
				err.Error())
			fmt.Println(msg)
			return nil, nil, errors.New(msg)
		}

		env.Ins(identifier.String(), tmp)
		//stepForms[identifier] = stepForm
		stepForms = append(stepForms, loopVariable{identifier, stepForm})

		if cell.Cdr == nil {
			break
		}
	}

	return env, stepForms, nil
} // func (inter *Interpreter) evalDoVariables(varList *value.List) (*value.Environment, map[value.Symbol]value.LispValue, error)

func (inter *Interpreter) evalLength(l *value.List) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}

	if l == nil || l.Length != 2 {
		return value.NIL, SyntaxError("LENGTH requires exactly one argument")
	}

	var arg = l.Car.Cdr.(*value.ConsCell).Car

	if value.IsNil(arg) {
		return value.IntValue(0), nil
	}

	var tmp value.LispValue
	var err error

	if tmp, err = inter.Eval(arg); err != nil {
		return value.NIL, err
	}

	switch v := tmp.(type) {
	case value.StringValue:
		return value.IntValue(len(string(v))), nil
	case *value.ConsCell:
		return value.IntValue(v.ActualLength()), nil
	case *value.List:
		return value.IntValue(v.Length), nil
	case value.Array:
		return value.IntValue(len(v)), nil
	case value.Hashtable:
		return value.IntValue(len(v)), nil
	default:
		return value.NIL, &TypeError{
			expected: "String, List, Array, or Hashtable",
			actual:   tmp.Type().String(),
		}
	}
} // func (inter *Interpreter) evalLength(l *value.List) (value.LispValue, error)

func (inter *Interpreter) evalConcat(l *value.List) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}

	if value.IsNil(l) {
		// We need the two-step form to shut up golint
		var msg = "CANTHAPPEN: Call to CONCAT is nil?!?!"
		return value.NIL, errors.New(msg)
	} else if l.Length == 1 {
		return value.NIL, nil
	} else if l.Length == 2 {
		// FIXME I really think I should add the same kind of type checking here
		//       I will (hopefully) use later on!
		return l.Car.Cdr.(*value.ConsCell).Car, nil
	}

	var (
		acc  value.LispValue //  = l.Car.Cdr.(*value.ConsCell).Car
		err  error
		rest = l.Car.Cdr.(*value.ConsCell).Cdr.(*value.ConsCell)
	)

	if acc, err = inter.Eval(l.Car.Cdr.(*value.ConsCell).Car); err != nil {
		return value.NIL, err
	}

	for cell := rest; cell != nil; cell = cell.Cdr.(*value.ConsCell) {
		// var val  = cell.Car
		var tmp value.LispValue
		var val value.LispValue

		if val, err = inter.Eval(cell.Car); err != nil {
			return value.NIL, err
		}

		switch acc.Type() {
		case types.Nil:
			acc = cell.Car
			continue
		case types.String:
			tmp, err = inter.evalConcatString(acc.(value.StringValue), val)
		case types.ConsCell:
			// This should not really happen in real life?
			// ... This should only ever happen if acc is not
			// a proper list, but a cons pair.
			// In that case we cannot really append anything
			// to it, right?
			var msg = "Cannot append stuff to a ConsCell"
			fmt.Println(msg)
			return value.NIL, errors.New(msg)
		case types.List:
			tmp, err = inter.evalConcatList(acc.(*value.List), val)
		case types.Array:
			tmp, err = inter.evalConcatArray(acc.(value.Array), val)
		case types.Hashtable:
			tmp, err = inter.evalConcatHashtable(acc.(value.Hashtable), val)
		default:
			return value.NIL, fmt.Errorf("Cannot append anything to a %s",
				acc.Type())
		}

		if err != nil {
			var e = fmt.Errorf("Error appending a %s to a %s: %s",
				val.Type(),
				acc.Type(),
				err.Error())
			fmt.Println(e.Error())
			return value.NIL, e
		}

		acc = tmp
		if cell.Cdr == nil {
			break
		}
	}

	return acc, nil
} // func (inter *Interpreter) evalConcat(l *value.List) (value.LispValue, error)

func mkConcatList(v1, v2 value.LispValue) *value.List {
	return &value.List{
		Car: &value.ConsCell{
			Car: value.Symbol("CONCAT"),
			Cdr: &value.ConsCell{
				Car: v1,
				Cdr: &value.ConsCell{
					Car: v2,
				},
			},
		},
		Length: 2,
	}
} // func mkConcatList(v1, v2 value.LispValue) *value.List

func (inter *Interpreter) evalConcatString(acc value.StringValue, other value.LispValue) (value.StringValue, error) {
	var ok bool
	var err error

	// We are skipping Hashtable deliberately, because in a Hashtable the
	// ordering of entries is kind of arbitrary/random, and there is no
	// good way to decide upon a certain order.
	// So (concat <string> <hash-table>) is "undefined, if you will.
	switch other.Type() {
	case types.Nil:
		return acc, nil
	case types.Integer, types.Float, types.BigInt:
		return acc + value.StringValue(other.String()), nil
	case types.String:
		return acc + value.StringValue(string(string(other.(value.StringValue)))), nil
	case types.Symbol:
		var sym = other.(value.Symbol)
		var retval value.StringValue
		if sym.IsKeyword() {
			retval = acc + value.StringValue(string(value.StringValue(sym))[1:])
		} else {
			retval = value.StringValue(string(sym))
		}
		return retval, nil
	case types.List:
		var iter = other.(*value.List).Car
		var tmp = acc
		for cell := iter; !value.IsNil(cell); cell = cell.Cdr.(*value.ConsCell) {
			if tmp, err = inter.evalConcatString(tmp, cell.Car); err != nil {
				return "", err
			}
		}

		return tmp, nil
	case types.Array:
		var arr = other.(value.Array)
		var tmp = acc
		for _, elt := range arr {
			var strVal value.LispValue
			if strVal, err = inter.evalConcat(mkConcatList(tmp, elt)); err != nil {
				return "", err
			}
			if tmp, ok = strVal.(value.StringValue); !ok {
				return "", err
			}
		}

		return tmp, nil
	default:
		var msg = fmt.Sprintf("ERROR Don't know how to append %s to String",
			other.Type())
		fmt.Println(msg)
		return "", errors.New(msg)
	}
} // func (inter *Interpreter) evalConcatString(acc value.StringValue, other value.LispValue) (value.StringValue, error)

func (inter *Interpreter) evalConcatList(acc *value.List, other value.LispValue) (*value.List, error) {
	var last *value.ConsCell

	for cell := acc.Car; cell != nil; cell = cell.Cdr.(*value.ConsCell) {
		if cell.Cdr == nil {
			last = cell
			break
		}
	}

	if last == nil {
		var msg = spew.Sprintf("Did not find end-of-list in %v",
			acc)
		fmt.Println(msg)
		return value.ListNil(), errors.New(msg)
	}

	switch other.Type() {
	case types.Nil:
		return acc, nil
	case types.Integer, types.Float, types.BigInt, types.String:
		last.Cdr = &value.ConsCell{
			Car: other,
		}
		acc.Length++
		return acc, nil
	case types.ConsCell:
		// Mittwoch, 08. 11. 2017, 00:42
		// It would not be *that* hard to deal with this case, but
		// it really should not happen under real-world conditions.
		// Not ever.
		var msg = "CANTHAPPEN: Cannot append raw ConsCell to List"
		fmt.Println(msg)
		return value.ListNil(), errors.New(msg)
	case types.List:
		last.Cdr = other.(*value.List).Car
		acc.Length += other.(*value.List).Length
		return acc, nil
	case types.Array:
		// Donnerstag, 09. 11. 2017, 19:41
		// Not sure if this matters or how much, but I could make this
		// slightly less inefficient, if I built up the second list from
		// the back and then set the first list's tail to the head of the
		// second one.
		// Since our second argument is an Array, that should not be
		// a problem, right?
		for _, v := range other.(value.Array) {
			// if val, err = inter.Eval(v); err != nil {
			// 	return value.ListNil(), err
			// }

			var newLast = &value.ConsCell{
				Car: v,
			}

			last.Cdr = newLast
			//last = last.Cdr.(*value.ConsCell)
			last = newLast
			acc.Length++
		}

		return acc, nil
	default:
		return value.ListNil(), &TypeError{
			expected: "Number, or List, or Array",
			actual:   other.Type().String(),
		}
	}
} // func (inter *Interpreter) evalConcatList(acc *value.List, other value.LispValue) (*value.List, error)

func (inter *Interpreter) evalConcatArray(acc value.Array, other value.LispValue) (value.Array, error) {
	switch other.Type() {
	case types.Integer, types.Float, types.BigInt, types.String, types.Regexp, types.Symbol, types.KeySym:
		return append(acc, other), nil
	case types.ConsCell:
		var msg = "CANTHAPPEN: Raw ConsCell value should NEVER appear in real life"
		fmt.Println(msg)
		return value.EmptyArray(), nil
	case types.List:
		var idx = len(acc)
		var newArray = make([]value.LispValue, idx+other.(*value.List).Length)
		copy(newArray, acc)

		for cell := other.(*value.List).Car; !value.IsNil(cell); cell = cell.Cdr.(*value.ConsCell) {
			// Donnerstag, 09. 11. 2017, 19:35
			// XXX I do not need to evaluate each member of the list
			// if tmp, err = inter.Eval(cell.Car); err != nil {
			// 	var msg = fmt.Sprintf("Error evaluating List member %s: %s",
			// 		cell.Car,
			// 		err.Error())
			// 	fmt.Println(msg)
			// 	return nil, nil
			// }

			newArray[idx] = cell.Car
			idx++
			if cell.Cdr == nil {
				break
			}
		}

		return newArray, nil
	default:
		fmt.Printf("Don't know how to append a %s to an Array\n",
			other.Type().String())
		return value.EmptyArray(), nil
	}
} // func (inter *Interpreter) evalConcatArray(acc value.Array) (value.Array, error)

func (inter *Interpreter) evalConcatHashtable(acc value.Hashtable, other value.LispValue) (value.Hashtable, error) {
	// Do I really need this?
	return acc, nil
} // func (inter *Interpreter) evalConcatHashtable(acc value.Hashtable, other value.LispValue) (value.Hashtable, error)

func (inter *Interpreter) evalGetenv(l *value.List) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}

	if l == nil || l.Length != 2 {
		return value.NIL, SyntaxError("GETENV takes exactly one argument")
	}

	var err error
	var val, raw value.LispValue

	raw = l.Car.Cdr.(*value.ConsCell).Car

	if val, err = inter.Eval(raw); err != nil {
		return value.NIL, err
	} else if val.Type() != types.String {
		return value.NIL, &TypeError{
			expected: "String",
			actual:   val.Type().String(),
		}
	}

	var rawEnv = os.Getenv(string(val.(value.StringValue)))

	return value.StringValue(rawEnv), nil
} // func (inter *Interpreter) evalGetenv(l *value.List) (value.LispValue, error)

func (inter *Interpreter) evalSetenv(l *value.List) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}

	if l == nil || l.Length != 3 {
		return value.NIL, SyntaxError("SETENV takes exactly two arguments")
	}

	var env, newValue, rawEnv, rawVal value.LispValue
	var err error

	rawEnv, _ = l.Nth(1)
	rawVal, _ = l.Nth(2)

	if env, err = inter.Eval(rawEnv); err != nil {
		return value.NIL, err
	} else if newValue, err = inter.Eval(rawVal); err != nil {
		return value.NIL, err
	} else if env.Type() != types.String {
		return value.NIL, &TypeError{
			expected: "String",
			actual:   env.Type().String(),
		}
	} else if newValue.Type() != types.String {
		return value.NIL, &TypeError{
			expected: "String",
			actual:   newValue.Type().String(),
		}
	}

	var (
		key = string(env.(value.StringValue))
		val = string(newValue.(value.StringValue))
	)

	if err = os.Setenv(key, val); err != nil {
		fmt.Printf("Error setting environment variable %s to value %s: %s",
			key,
			val,
			err.Error())
		return value.NIL, err
	}

	return newValue, nil
} // func (inter *Interpreter) evalSetenv(l *value.List) (value.LispValue, error)
