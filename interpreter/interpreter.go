// /home/krylon/go/src/krylisp/interpreter/interpreter.go
// -*- mode: go; coding: utf-8; -*-
// Created on 08. 09. 2017 by Benjamin Walkenhorst
// (c) 2017 Benjamin Walkenhorst
// Time-stamp: <2017-09-16 14:16:32 krylon>

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
	"krylib"
	"krylisp/types"
	"krylisp/value"
)

// specialSymbols refer to values or syntactic constructs that are defined in the
// Interpreter itself, not in Lisp.
var specialSymbols = map[string]bool{
	"T":      true,
	"NIL":    true,
	"+":      true,
	"-":      true,
	"*":      true,
	"/":      true,
	"<":      true,
	">":      true,
	">=":     true,
	"<=":     true,
	"EQ":     true,
	"DEFUN":  true,
	"IF":     true,
	"LET":    true,
	"DO":     true,
	"PRINT":  true,
	"CONS":   true,
	"CAR":    true,
	"CDR":    true,
	"SET!":   true,
	"DEFINE": true,
	"GOTO":   true,
}

// IsSpecial returns true if the given symbols has special significance
// to the Lisp Interpreter.
func IsSpecial(s fmt.Stringer) bool {
	_, ok := specialSymbols[s.String()]
	return ok
} // func IsSpecial(s value.Symbol) bool

// Interpreter is my first shot at a tree-walking interpreter for my
// toy Lisp dialect.
type Interpreter struct {
	debug         bool
	gensymCounter int
	env           *value.Environment
	fnEnv         *value.Environment
}

// Eval evaluates a Lisp value and returns the result.
func (inter *Interpreter) Eval(lval value.LispValue) (value.LispValue, error) {
	switch v := lval.(type) {
	case value.IntValue:
		return v, nil
	case value.StringValue:
		return v, nil
	case value.Symbol:
		return inter.evalSymbol(v)
	case value.NilValue:
		return v, nil
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
	default:
		return nil, &TypeError{
			expected: "Atom or List",
			actual:   v.Type().String(),
		}
	}
} // func (inter *Interpreter) Eval(v value.LispValue) (value.LispValue, error)

func (inter *Interpreter) evalSpecialForm(l *value.List) (value.LispValue, error) {
	//return nil, krylib.NotImplemented
	var sym = l.Car.Car.(value.Symbol)

	// This is going to be tedious...
	// Maybe I should use a lookup table. As the number of special forms
	// grows - and it WILL grow -, a switch might not be the most
	// efficient solution.
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
	// case ">":
	// 	return inter.evalGreaterThan(l)
	// case "<=":
	// 	return inter.evalLessEqual(l)
	// case ">=":
	// 	return inter.evalGreaterEqual(l)
	case "DEFUN":
		return inter.evalDefun(l)
	case "LAMBDA":
		return inter.evalLambda(l)
	case "EQ":
		return inter.evalEq(l)
	}

	return nil, krylib.NotImplemented
} // func (inter *Interpreter) evalSpecialForm(l *value.List) (value.LispValue, error)

func (inter *Interpreter) evalSymbol(s value.Symbol) (value.LispValue, error) {
	if s.IsKeyword() {
		// Keyword symbols evaluate to themselves:
		return s, nil
	} else if s == "NIL" || s == "T" {
		return s, nil
	}

	var v value.LispValue
	var found bool

	if v, found = inter.env.Get(string(s)); found {
		return v, nil
	}

	return nil, &NoBindingError{sym: s}
} // func (inter *Interpreter) evalSymbol(s value.Symbol) (value.LispValue, error)

func (inter *Interpreter) evalLambda(lst *value.List) (*value.Function, error) {
	if lst == nil || lst.Car == nil || lst.Car.Car == nil {
		return nil, errors.New("Argument is not a lambda list")
	} else if lst.Car.Car.Type() != types.Symbol || lst.Car.Car.(value.Symbol) != "LAMBDA" {
		return nil, errors.New("Argument is not a lambda list")
	} else if lst.Car.Cdr.(*value.ConsCell).Car.Type() != types.List {
		return nil, errors.New("Second element in List should be a list (of arguments)")
	}

	var args = lst.Car.Cdr.(*value.ConsCell).Car.(*value.List)
	var idx = 0

	var fn = &value.Function{
		Env:  inter.env,
		Args: make([]value.Symbol, args.Length),
	}

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
	var fn *value.Function
	var err error

	if inv == nil || inv.Car == nil {
		return nil, nil
	}

	// Sonntag, 10. 09. 2017, 18:55
	// I need to handle lists, too. Scheme allows a lambda list at the first
	// position. I would like to support that, too.
	// One day...
	switch f := inv.Car.Car.(type) {
	case *value.Function:
		fn = f
	case value.Symbol:
		if v, ok := inter.fnEnv.Get(string(f)); ok {
			fn = v.(*value.Function)
		} else {
			return nil, MissingFunctionError(f)
		}
	case *value.List:
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

	var argList *value.ConsCell = inv.Car.Cdr.(*value.ConsCell)
	var argCnt, idx int

	if argCnt = argList.ActualLength(); argCnt != len(fn.Args) {
		return nil, fmt.Errorf("Wrong number of arguments: Expected %d, got %d",
			len(fn.Args),
			argCnt)
	}

	var env = value.NewEnvironment(inter.env)

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

	inter.env = env
	defer func() { inter.env = inter.env.Parent }()
	var res value.LispValue

	for _, exp := range fn.Body {
		if res, err = inter.Eval(exp); err != nil {
			return nil, err
		}
	}

	return res, nil
} // func (inter *Interpreter) evalFuncall(fun value.Function) (value.LispValue, errror)

func (inter *Interpreter) evalIf(l *value.List) (value.LispValue, error) {
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

func (inter *Interpreter) evalPlus(l *value.List) (value.LispValue, error) {
	var cnt value.IntValue

	for v := l.Car.Cdr; v != nil; v = v.(*value.ConsCell).Cdr {
		if v.(*value.ConsCell).Car == nil {
			return nil, &TypeError{expected: "Number", actual: "nil"}
		} else if v.(*value.ConsCell).Car.Type() != types.Number {
			return nil, &TypeError{expected: "Number", actual: "nil"}
		}

		cnt += v.(*value.ConsCell).Car.(value.IntValue)
	}

	return cnt, nil
} // func (inter *Interpreter) evalPlus(l *value.List) (value.LispValue, error)

func (inter *Interpreter) evalMinus(l *value.List) (value.LispValue, error) {
	var cnt value.IntValue
	var val value.LispValue
	var err error

	if l.Length < 2 {
		return value.NIL, SyntaxError("Too few arguments for -")
	} else if l.Length == 2 {
		if l.Car.Cdr.(*value.ConsCell).Car.Type() != types.Number {
			return nil, &TypeError{
				expected: "Number",
				actual:   l.Car.Cdr.(*value.ConsCell).Car.Type().String(),
			}
		}

		return -(l.Car.Cdr.(*value.ConsCell).Car.(value.IntValue)), nil
	}

	// I need to eval all arguments!!!
	//cnt = l.Car.Cdr.(*value.ConsCell).Car.(value.IntValue)
	if val, err = inter.Eval(l.Car.Cdr.(*value.ConsCell).Car); err != nil {
		return value.NIL, err
	} else if val.Type() != types.Number {
		return value.NIL, &TypeError{
			expected: "Number",
			actual:   val.Type().String(),
		}
	}

	cnt = val.(value.IntValue)

	for v := l.Car.Cdr.(*value.ConsCell).Cdr; v != nil; v = v.(*value.ConsCell).Cdr {
		if val, err = inter.Eval(v.(*value.ConsCell).Car); err != nil {
			return value.NIL, err
		} else if val.Type() != types.Number {
			return value.NIL, &TypeError{
				expected: "Number",
				actual:   val.Type().String(),
			}
		}
		cnt -= val.(value.IntValue)
	}

	return cnt, nil
} // func (inter *Interpreter) evalMinus(l *value.List) (value.LispValue, error)

func (inter *Interpreter) evalMultiply(l *value.List) (value.LispValue, error) {
	if l.Length == 1 {
		return value.IntValue(1), nil
	} else if l.Length == 2 {
		return l.Car.Cdr.(*value.ConsCell).Car, nil
	}

	// if inter.debug {
	// 	spew.Dump(l)
	// }

	var err error
	var resRaw value.LispValue
	var res value.IntValue

	if resRaw, err = inter.Eval(l.Car.Cdr.(*value.ConsCell).Car); err != nil {
		return value.NIL, err
	} else if resRaw.Type() == types.Number {
		res = resRaw.(value.IntValue)
	} else {
		return value.NIL, &TypeError{
			expected: "Number",
			actual:   resRaw.Type().String(),
		}
	}

	for v := l.Car.Cdr.(*value.ConsCell).Cdr.(*value.ConsCell); v != nil; v = v.Cdr.(*value.ConsCell) {
		// Ich muss hier v.Car evaluieren!
		var cval value.LispValue
		if cval, err = inter.Eval(v.Car); err != nil {
			return value.NIL, err
		} else if cval.Type() == types.Number {
			res *= cval.(value.IntValue)
			if v.Cdr == nil {
				break
			}
		} else {
			return value.NIL, &TypeError{
				expected: "Number",
				actual:   cval.Type().String(),
			}
		}
	}

	return res, nil
} // func (inter *Interpreter) evalMultiply(l *value.List) (value.LispValue, error)

func (inter *Interpreter) evalDivide(l *value.List) (value.LispValue, error) {
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
	} else if val.Type() != types.Number {
		return value.NIL, &TypeError{
			expected: "Number",
			actual:   val.Type().String(),
		}
	}

	var res = val.(value.IntValue)

	for c := l.Car.Cdr.(*value.ConsCell).Cdr; c != nil; c = c.(*value.ConsCell).Cdr {
		v := c.(*value.ConsCell)
		if val, err = inter.Eval(v.Car); err != nil {
			return value.NIL, err
		} else if val.Type() == types.Number {
			var n = val.(value.IntValue)
			if n != 0 {
				res /= n
			} else {
				return nil, &ValueError{n}
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

func (inter *Interpreter) evalDefun(l *value.List) (value.LispValue, error) {
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
		Car: &value.ConsCell{
			Car: value.Symbol("LAMBDA"),
			Cdr: l.Car.Cdr.(*value.ConsCell).Cdr.(*value.ConsCell),
		},
		Length: l.Length - 1,
	}

	if val, err = l.Nth(1); err != nil {
		return value.NIL, err
	} else if name, ok = val.(value.Symbol); !ok {
		return value.NIL, fmt.Errorf("First argument to defun must be a symbol, not a %T (%s)",
			val.Type().String(),
			val.String())
	} else if val, err = l.Nth(2); val.Type() == types.String {
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
	}

	if v1.Eq(v2) {
		// if inter.debug {
		// 	fmt.Printf("(EQ %s %s) => T\n",
		// 		v1.String(),
		// 		v2.String())
		// }
		return value.T, nil
	} /*else if inter.debug {
		fmt.Printf("(EQ %s %s) => NIL\n",
			v1.String(),
			v2.String())
	}*/

	return value.NIL, nil
} // func (inter *Interpreter) evalEq(l *value.List) (value.LispValue, error)

func (inter *Interpreter) evalLessThan(l *value.List) (value.LispValue, error) {
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
	} else if v1.Type() != types.Number {
		return value.NIL, &TypeError{
			expected: "Number",
			actual:   v1.Type().String(),
		}
	} else if v2.Type() != types.Number {
		return value.NIL, &TypeError{
			expected: "Number",
			actual:   v2.Type().String(),
		}
	} else if v1.(value.IntValue) >= v2.(value.IntValue) {
		return value.NIL, nil
	} else if l.Length == 3 {
		return value.T, nil
	}

	for c := l.Car.Cdr.(*value.ConsCell).Cdr.(*value.ConsCell).Cdr.(*value.ConsCell); c != nil; c = c.Cdr.(*value.ConsCell) {
		v1 = v2
		raw2 = c.Car

		if v2, err = inter.Eval(raw2); err != nil {
			return value.NIL, err
		} else if v2.Type() != types.Number {
			return value.NIL, &TypeError{
				expected: "Number",
				actual:   v2.Type().String(),
			}
		} else if v1.(value.IntValue) >= v2.(value.IntValue) {
			return value.NIL, nil
		} else if c.Cdr == nil {
			break
		}
	}

	return value.T, nil
} // func (inter *Interpreter) evalLessThan(l *value.List) (value.LispValue, error)

func (inter *Interpreter) evalGreaterThan(l *value.List) (value.LispValue, error) {
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
	} else if v1.Type() != types.Number {
		return value.NIL, &TypeError{
			expected: "Number",
			actual:   v1.Type().String(),
		}
	} else if v2.Type() != types.Number {
		return value.NIL, &TypeError{
			expected: "Number",
			actual:   v2.Type().String(),
		}
	} else if v1.(value.IntValue) <= v2.(value.IntValue) {
		return value.NIL, nil
	} else if l.Length == 3 {
		return value.T, nil
	}

	for c := l.Car.Cdr.(*value.ConsCell).Cdr.(*value.ConsCell).Cdr.(*value.ConsCell); c != nil; c = c.Cdr.(*value.ConsCell) {
		v1 = v2
		raw2 = c.Car

		if v2, err = inter.Eval(raw2); err != nil {
			return value.NIL, err
		} else if v2.Type() != types.Number {
			return value.NIL, &TypeError{
				expected: "Number",
				actual:   v2.Type().String(),
			}
		} else if v1.(value.IntValue) <= v2.(value.IntValue) {
			return value.NIL, nil
		} else if c.Cdr == nil {
			break
		}
	}

	return value.T, nil
} // func (inter *Interpreter) evalGreaterThan(l *value.List) (value.LispValue, error)
