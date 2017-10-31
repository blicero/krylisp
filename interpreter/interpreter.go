// /home/krylon/go/src/krylisp/interpreter/interpreter.go
// -*- mode: go; coding: utf-8; -*-
// Created on 08. 09. 2017 by Benjamin Walkenhorst
// (c) 2017 Benjamin Walkenhorst
// Time-stamp: <2017-10-31 22:23:48 krylon>
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

	"github.com/davecgh/go-spew/spew"
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
	"FN":     true,
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
	"QUOTE":  true,
	"NOT":    true,
	"AND":    true,
	"OR":     true,
	"APPLY":  true,
	"LAMBDA": true,
	"NIL?":   true,
	"LIST":   true,
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
		fmt.Printf("DBG FUNCALL %s\n",
			spew.Sdump(inv))
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

	if cmpResult, err = evalPolymorphLT(v1.(value.Number), v2.(value.Number)); err != nil {
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
		} else if res, err = evalPolymorphLT(v1.(value.Number), v2.(value.Number)); err != nil {
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

	raw1 = l.Car.Cdr.(*value.ConsCell).Car
	raw2 = l.Car.Cdr.(*value.ConsCell).Cdr.(*value.ConsCell).Car

	if v1, err = inter.Eval(raw1); err != nil {
		return value.NIL, err
	} else if v2, err = inter.Eval(raw2); err != nil {
		return value.NIL, err
	} else if v1.Type() != types.Integer {
		return value.NIL, &TypeError{
			expected: "Number",
			actual:   v1.Type().String(),
		}
	} else if v2.Type() != types.Integer {
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
		} else if v2.Type() != types.Integer {
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

	raw1 = l.Car.Cdr.(*value.ConsCell).Car
	raw2 = l.Car.Cdr.(*value.ConsCell).Cdr.(*value.ConsCell).Car

	if v1, err = inter.Eval(raw1); err != nil {
		return value.NIL, err
	} else if v2, err = inter.Eval(raw2); err != nil {
		return value.NIL, err
	} else if v1.Type() != types.Integer {
		return value.NIL, &TypeError{
			expected: "Number",
			actual:   v1.Type().String(),
		}
	} else if v2.Type() != types.Integer {
		return value.NIL, &TypeError{
			expected: "Number",
			actual:   v2.Type().String(),
		}
	} else if v1.(value.IntValue) > v2.(value.IntValue) {
		return value.NIL, nil
	} else if l.Length == 3 {
		return value.T, nil
	}

	for c := l.Car.Cdr.(*value.ConsCell).Cdr.(*value.ConsCell).Cdr.(*value.ConsCell); c != nil; c = c.Cdr.(*value.ConsCell) {
		v1 = v2
		raw2 = c.Car

		if v2, err = inter.Eval(raw2); err != nil {
			return value.NIL, err
		} else if v2.Type() != types.Integer {
			return value.NIL, &TypeError{
				expected: "Number",
				actual:   v2.Type().String(),
			}
		} else if v1.(value.IntValue) > v2.(value.IntValue) {
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

	raw1 = l.Car.Cdr.(*value.ConsCell).Car
	raw2 = l.Car.Cdr.(*value.ConsCell).Cdr.(*value.ConsCell).Car

	if v1, err = inter.Eval(raw1); err != nil {
		return value.NIL, err
	} else if v2, err = inter.Eval(raw2); err != nil {
		return value.NIL, err
	} else if v1.Type() != types.Integer {
		return value.NIL, &TypeError{
			expected: "Number",
			actual:   v1.Type().String(),
		}
	} else if v2.Type() != types.Integer {
		return value.NIL, &TypeError{
			expected: "Number",
			actual:   v2.Type().String(),
		}
	} else if v1.(value.IntValue) < v2.(value.IntValue) {
		return value.NIL, nil
	} else if l.Length == 3 {
		return value.T, nil
	}

	for c := l.Car.Cdr.(*value.ConsCell).Cdr.(*value.ConsCell).Cdr.(*value.ConsCell); c != nil; c = c.Cdr.(*value.ConsCell) {
		v1 = v2
		raw2 = c.Car

		if v2, err = inter.Eval(raw2); err != nil {
			return value.NIL, err
		} else if v2.Type() != types.Integer {
			return value.NIL, &TypeError{
				expected: "Number",
				actual:   v2.Type().String(),
			}
		} else if v1.(value.IntValue) < v2.(value.IntValue) {
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
