// /home/krylon/go/src/krylisp/interpreter/interpreter.go
// -*- mode: go; coding: utf-8; -*-
// Created on 08. 09. 2017 by Benjamin Walkenhorst
// (c) 2017 Benjamin Walkenhorst
// Time-stamp: <2024-05-23 18:31:02 krylon>
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
//
// Dienstag, 14. 11. 2017, 18:17
// While trying to add I/O, I discovered that I would probably like to have
// keyword arguments that are passed in a hashtable-like structure.
// The downside is that I will have do major work on the function call
// mechanism, the upside is that keyword arguments are a desirable for
// purposes other than I/O.
//
// Donnerstag, 16. 11. 2017, 09:42
// So, I added keyword arguments, and then I realized that it was not a good
// idea to have all those "functions" implemented in Go be special forms
// from the interpreters point of view.
// So I created a new data type, GoFunction, that encapsulates a function
// written in, well, Go. The next step is to check all the eval* functions
// here if they need to be special forms and to convert them to GoFunctions
// otherwise, AND adjust the function call mechanism to handle GoFunctions,
// AND adjust all the tests.
// This is going to be a lot of work.
//
// Freitag, 17. 11. 2017, 12:12
// Okay, I think I haved most of the way towards ... let's call them native
// functions.
// Now I need to hook up the interpreter to these, so it detects when a function
// call refers to a native function and chooses the appropriate path.
// ...
// Maybe, just maybe, if I am really lucky, all I need to do is to "implant" those
// functions into the interpreter's function environment (we're a Lisp-2, remember?)
// From there, I am almost sure it should Just Work(tm).
//
// Freitag, 17. 11. 2017, 13:04
// With my single test case, the humble "+", it worked perfectly.
// So now I am stuck with the rather tedious task inspect all eval* methods,
// and converting those that do not need to be treated as special form.
//
// Samstag, 25. 11. 2017, 16:36
// The next big step is macros, and right now I have no clue whatsoever how to
// implement them.
//
// Montag, 11. 12. 2017, 18:42
// I have decided to have Environments store data, functions, and macros.
// So the Interpreter no longer needs two separate environments.

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
	"os"
	"regexp"
	"strings"

	"github.com/blicero/krylib"
	"github.com/blicero/krylisp/compare"
	"github.com/blicero/krylisp/filemode"
	"github.com/blicero/krylisp/lexer"
	"github.com/blicero/krylisp/parser"
	"github.com/blicero/krylisp/types"
	"github.com/blicero/krylisp/value"

	"github.com/davecgh/go-spew/spew"
)

const one = value.IntValue(1)

// specialSymbols refer to values or syntactic constructs that are defined in the
// Interpreter itself, not in Lisp.
var specialSymbols = map[string]bool{
	"T":        true,
	"NIL":      true,
	"FN":       true,
	"DEFUN":    true,
	"DEFMACRO": true,
	"IF":       true,
	"LET":      true,
	"DO":       true,
	"WHILE":    true,
	"PRINT":    true,
	"CONS":     true,
	"CAR":      true,
	"CDR":      true,
	"SET!":     true,
	"DEFINE":   true,
	"GOTO":     true,
	"QUOTE":    true,
	"AND":      true,
	"OR":       true,
	"APPLY":    true,
	"LAMBDA":   true,
	"CONCAT":   true,
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
	// fnEnv         *value.Environment
	stdout io.Writer
	stderr io.Writer
	stdin  io.Reader
}

// New returns a fresh, initialized Interpreter instance with an
// empty Environment. It passes the debug flag to the Interpreter.
func New(debug bool) *Interpreter {
	var inter = &Interpreter{
		debug:         debug,
		gensymCounter: 1,
		env:           value.NewEnvironment(nil),
		// fnEnv:         value.NewEnvironment(nil),
		stdin:  os.Stdin,
		stdout: os.Stdout,
		stderr: os.Stderr,
	}

	inter._initNativeFunctions()

	return inter
} // func New(debug bool) *Interpreter

// I sorta kinda suspect, that I could do this part with code generation.
// It would probably be rather difficult and tedious, but in the long run,
// it could make dealing with native functions A LOT easier.
func (inter *Interpreter) _initNativeFunctions() {
	if inter.debug {
		krylib.Trace()
	}

	var nativeFunctions = map[string]*value.GoFunction{
		"+": &value.GoFunction{
			Fn:   inter.evalPlus,
			Name: "+",
		},
		"-": &value.GoFunction{
			Fn:   inter.evalMinus,
			Name: "-",
		},
		"*": &value.GoFunction{
			Fn:   inter.evalMultiply,
			Name: "*",
		},
		"/": &value.GoFunction{
			Fn:   inter.evalDivide,
			Name: "/",
		},
		"EQ": &value.GoFunction{
			Fn:   inter.evalEq,
			Name: "EQ",
		},
		"<": &value.GoFunction{
			Fn:   inter.evalLessThan,
			Name: "<",
		},
		">": &value.GoFunction{
			Fn:   inter.evalGreaterThan,
			Name: ">",
		},
		"<=": &value.GoFunction{
			Fn:   inter.evalLessEqual,
			Name: "<=",
		},
		">=": &value.GoFunction{
			Fn:   inter.evalGreaterEqual,
			Name: ">=",
		},
		"NOT": &value.GoFunction{
			Fn:   inter.evalNot,
			Name: "NOT",
		},
		"NIL?": &value.GoFunction{
			Fn:   inter.evalIsNil,
			Name: "NIL?",
		},
		"LIST": &value.GoFunction{
			Fn:   inter.evalList,
			Name: "LIST",
		},
		"AREF": &value.GoFunction{
			Fn:   inter.evalAref,
			Name: "AREF",
		},
		"APUSH": &value.GoFunction{
			Fn:   inter.evalApush,
			Name: "APUSH",
		},
		"MAKE-ARRAY": &value.GoFunction{
			Fn:   inter.evalMakeArray,
			Name: "MAKE-ARRAY",
		},
		"MAKE-HASH": &value.GoFunction{
			Fn:   inter.evalMakeHash,
			Name: "MAKE-HASH",
		},
		"HASHREF": &value.GoFunction{
			Fn:   inter.evalHashref,
			Name: "HASHREF",
		},
		"HAS-KEY": &value.GoFunction{
			Fn:   inter.evalHasKey,
			Name: "HAS-KEY",
		},
		"HASH-SET": &value.GoFunction{
			Fn:   inter.evalHashSet,
			Name: "HASH-SET",
		},
		"HASH-DELETE": &value.GoFunction{
			Fn:   inter.evalHashDelete,
			Name: "HASH-DELETE",
		},
		"REGEXP-COMPILE": &value.GoFunction{
			Fn:   inter.evalRegexpCompile,
			Name: "REGEXP-COMPILE",
		},
		"REGEXP-MATCH": &value.GoFunction{
			Fn:   inter.evalRegexpMatch,
			Name: "REGEXP-MATCH",
		},
		"LENGTH": &value.GoFunction{
			Fn:   inter.evalLength,
			Name: "LENGTH",
		},
		"GETENV": &value.GoFunction{
			Fn:   inter.evalGetEnv,
			Name: "GETENV",
		},
		"SETENV": &value.GoFunction{
			Fn:   inter.evalSetEnv,
			Name: "SETENV",
		},
		"FOPEN": &value.GoFunction{
			Fn:   inter.evalFopen,
			Name: "FOPEN",
		},
		"FCLOSE": &value.GoFunction{
			Fn:   inter.evalFclose,
			Name: "FCLOSE",
		},
		"FREAD-LINE": &value.GoFunction{
			Fn:   inter.evalFreadLine,
			Name: "FREADLINE",
		},
		"FWRITE": &value.GoFunction{
			Fn:   inter.evalFwrite,
			Name: "FWRITE",
		},
		"FEOF": &value.GoFunction{
			Fn:   inter.evalFeof,
			Name: "FEOF",
		},
		"FCLEAREOF": &value.GoFunction{
			Fn:   inter.evalClearEOF,
			Name: "FCLEAREOF",
		},
		"READ-FROM-STRING": &value.GoFunction{
			Fn:   inter.evalReadString,
			Name: "READ-FROM-STRING",
		},
	}

	for sym, fn := range nativeFunctions {
		inter.env.InsFn(sym, fn)
	}

	if inter.debug {
		inter.env.Dump(inter.stdout)
	}
} // func (inter *Interpreter) _initNativeFunctions()

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
		return nil, &value.TypeError{
			Expected: "Atom or List",
			Actual:   v.Type().String(),
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
	case "DEFUN":
		return inter.evalDefun(l)
	case "DEFMACRO":
		return inter.evalDefmacro(l)
	case "LAMBDA":
		return inter.evalLambda(l)
	case "LET":
		return inter.evalLet(l)
	case "QUOTE":
		var retval value.LispValue
		if l.Car.Cdr.Type() == types.ConsCell {
			retval = l.Car.Cdr.(*value.ConsCell).Car
		} else if l.Car.Cdr.Type() == types.List {
			retval = l.Car.Cdr.(*value.List).Car
		}
		return retval, nil
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
	case "DO":
		return inter.evalDoLoop(l)
	case "WHILE":
		return inter.evalWhile(l)
	case "CONCAT":
		return inter.evalConcat(l)
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
func (inter *Interpreter) evalLambda(lst *value.List) (*value.LispFunction, error) {
	if inter.debug {
		krylib.Trace()
	}

	if lst == nil || lst.Car == nil || lst.Car.Car == nil {
		return nil, errors.New("Argument is not a lambda list")
	} else if lst.Car.Car.Type() != types.Symbol || lst.Car.Car.(value.Symbol) != "LAMBDA" {
		return nil, errors.New("Argument is not a lambda list")
	} else if lst.Car.Cdr.(*value.ConsCell).Car.Type() != types.List {
		//return nil, errors.New("Second element in List should be a list (of arguments)")
		return nil, fmt.Errorf("Second element in lambda list should be a list (of arguments), not a %s: %s",
			lst.Car.Cdr.(*value.ConsCell).Car.Type(),
			lst.String())
	}

	// Mittwoch, 15. 11. 2017, 17:56
	// XXX When I use keyword arguments, I need to to consider them in a
	//     special way. The &key symbol does not count towards the length of
	//     the arg list.
	//     I have to keep track of how many elements I add to the arg list
	//     and then make a slice of it at the end to cut off the "empty" rest.

	var (
		keywords bool
		args     = lst.Car.Cdr.(*value.ConsCell).Car.(*value.List)
		idx      = 0
		fn       = &value.LispFunction{
			Env:         inter.env,
			Args:        make([]value.Symbol, args.Length),
			Keywordargs: make(map[value.Symbol]value.LispValue),
		}
	)

	for symlist := args.Car; symlist != nil; symlist = symlist.Cdr.(*value.ConsCell) {
		if !keywords {
			if symlist.Car.Type() != types.Symbol {
				return nil, &value.TypeError{
					Expected: types.Symbol.String(),
					Actual:   symlist.Car.Type().String(),
				}
			}

			var car = symlist.Car.(value.Symbol)

			if car == "&KEY" {
				keywords = true
				continue
			}

			fn.Args[idx] = car
			idx++

		} else {
			if symlist.Car.Type() != types.List {
				return nil, &value.TypeError{
					Expected: "List (keyword argument)",
					Actual:   symlist.Car.Type().String(),
				}
			}

			var (
				keyword      = symlist.Car.(*value.List)
				key          value.Symbol
				defaultValue value.LispValue
				ok           bool
			)

			if keyword.Length != 2 {
				return nil, SyntaxError("Keyword arguments must be declared as pairs")
			} else if key, ok = keyword.Car.Car.(value.Symbol); !ok {
				return nil, &value.TypeError{
					Expected: "Symbol (keyword name)",
					Actual:   keyword.Car.Car.Type().String(),
				}
			}

			defaultValue = keyword.Car.Cdr.(*value.ConsCell).Car

			fn.Keywordargs[key] = defaultValue
		}

		if symlist.Cdr == nil {
			break
		}
	}

	fn.Args = fn.Args[:idx]

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

	if inter.debug {
		_, _ = spew.Printf("Evaluated LAMBDA list %s -> %#v\n",
			lst,
			fn)
	}

	return fn, nil
} // func (inter *Interpreter) evalLambda(lst *value.List) (*value.LispFunction, error)

// nolint: gocyclo
func (inter *Interpreter) evalFuncall(inv *value.List) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}

	var fn *value.LispFunction
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
	case *value.LispFunction:
		fn = f
	case value.Symbol:
		if inter.debug {
			fmt.Printf("DBG FUNCALL => Resolving symbol %s\n",
				f)
		}

		if IsSpecial(f) {
			return inter.evalSpecialForm(inv)
		} else if v := inter.env.GetFn(string(f)); v != nil {
			// Hier muss ich den Code anpassen, dass GoFunction auch
			// berücksichtigt wird.
			// Oder ich mache "Function" zu einem Interface, dass
			// dann von GoFunction und LispFunction implementiert
			// wird. Von der Performance wäre das wahrscheinlich
			// nicht so prall, aber wenn ich drüber nachdenke,
			// stecke ich in dem Schlamassel ohnehin schon
			// knietief, wenn nicht mehr.
			if gfn, ok := v.(*value.GoFunction); ok {
				return inter.evalGoFunction(gfn, inv)
				// return value.NIL, krylib.NotImplemented
			} else if fn, ok = v.(*value.LispFunction); !ok {
				return value.NIL, fmt.Errorf("Function lookup returned a %s",
					v.Type().String())
			}
		} else {
			if inter.debug {
				fmt.Printf("No such function: %s\n",
					f)
			}
			return nil, MissingFunctionError(f)
		}
	case *value.List:
		if inter.debug {
			_, _ = spew.Printf("Evaluate function call: %#v\n",
				f)
			// fmt.Printf("Evaluate function call: %s\n",
			// 	//f.String())
			// 	spew.Sdump(f))
		}

		if f.IsLambda() {
			if fn, err = inter.evalLambda(f); err != nil {
				return nil, err
			}
		} else {
			return nil, &value.TypeError{
				Expected: "Lambda List",
				Actual:   "Not a Lambda List",
			}
		}
	default:
		return nil, &value.TypeError{
			Expected: "Symbol or function literal",
			Actual:   fmt.Sprintf("%T", f),
		}
	}

	// So, if we arrive here, we have a function object, next we should
	// check out the arguments.
	//
	// Montag, 02. 10. 2017, 23:42
	// If a function is called without any parameters, the argument list
	// might be nil!

	if inter.debug {
		_, _ = spew.Printf("Evaluating call to this function: %#v\n",
			fn)
	}

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

	// Mittwoch, 15. 11. 2017, 06:36
	// FIXME I will have to adjust this check to account for keyword arguments
	//       and maybe later on optional arguments, too.
	//       I need to check if argCnt is between the number of mandatory
	//       and the maximum number of optional arguments.
	//
	//       Mittwoch, 15. 11. 2017, 17:52
	//       More precisely, when a function is called with keyword
	//       arguments, there are two elements in the list per
	//       keyword. I have to account for that.
	//       Then again, I thought I was doing exactly that already
	//       by... wait a second.
	argCnt = argList.ActualLength()
	// I need to add twice the number of keyword arguments, because the
	// keyword arguments in the function call come in pairs.
	var maxCnt = len(fn.Args) + len(fn.Keywordargs)*2
	//if argCnt = argList.ActualLength(); argCnt != len(fn.Args) {
	if argCnt < len(fn.Args) || argCnt > maxCnt {
		return nil, fmt.Errorf("Wrong number of arguments for funcall: Expected %d, got %d %s\n%s",
			len(fn.Args),
			argCnt,
			argList.String(),
			fn.ArglistString())
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
		var val value.LispValue

		if argList.Car.Type() == types.Symbol && argList.Car.(value.Symbol).IsKeyword() {
			var name = argList.Car.(value.Symbol)[1:]
			if argList.Cdr == nil {
				var msg = fmt.Sprintf("Keyword argument %s without matching value",
					name)
				fmt.Println(msg)
				return value.NIL, SyntaxError(msg)
			}

			argList = argList.Cdr.(*value.ConsCell)

			if val, err = inter.Eval(argList.Car); err != nil {
				return value.NIL, err
			}

			env.Data[string(name)] = val
		} else {
			var sym = fn.Args[idx]

			if val, err = inter.Eval(argList.Car); err != nil {
				return value.NIL, err
			}

			env.Data[string(sym)] = val
			idx++
		}

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

// nolint: gocyclo
func (inter *Interpreter) evalGoFunction(fn *value.GoFunction, lst *value.List) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}

	if fn == nil {
		panic("evalGoFunction was called with a nil function pointer")
	} else if lst == nil {
		panic("evalGoFunction was called with an empty invocation list")
	}

	var args = &value.Arguments{
		Positional: make([]value.LispValue, 0, 5),
		Keyword:    make(map[value.Symbol]value.LispValue),
	}

	if lst.Car.Cdr == nil {
		goto FUNCALL
	}

	for cell := lst.Car.Cdr.(*value.ConsCell); cell != nil; cell = cell.Cdr.(*value.ConsCell) {
		var (
			val value.LispValue
			err error
		)

		if cell.Car.Type() == types.Symbol && cell.Car.(value.Symbol).IsKeyword() {
			var name = cell.Car.(value.Symbol)[1:]
			if cell.Cdr == nil {
				var msg = fmt.Sprintf("GoFunc: Keyword argument %s without matching value",
					name)
				fmt.Println(msg)
				return value.NIL, SyntaxError(msg)
			}

			cell = cell.Cdr.(*value.ConsCell)

			if val, err = inter.Eval(cell.Car); err != nil {
				var msg = fmt.Sprintf("GoFunc: Error in call to GoFunc %s while evaluating arguments: %s",
					fn.Name,
					err.Error())

				if inter.debug {
					fmt.Println(msg)
				}

				return value.NIL, errors.New(msg)
			}

			args.Keyword[name] = val
			goto ENDCHECK
		} else if val, err = inter.Eval(cell.Car); err != nil {
			var msg = fmt.Sprintf("GoFunc: Error evaluating argument %s: %s",
				cell.Car,
				err.Error())
			fmt.Println(msg)
			return value.NIL, errors.New(msg)
		}

		args.Positional = append(args.Positional, val)

	ENDCHECK:
		if cell.Cdr == nil {
			break
		}
	}

FUNCALL:

	return fn.Fn(args)
} // func (inter *Interpreter) evalGoFunction(fn *value.GoFunction, lst *value.List) (LispValue, error)

/////////////////////////////////////////////////////////////////////////////
// Macros ///////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////

// Mittwoch, 13. 12. 2017, 20:12
// I feel strangely insecure when thinking about how to implement macros.
// After thinking about it for a bit, I decided that it is probably best
// to start by implementing macro creation.
// And I just realize it has been *years* since I have done anything in
// Lisp, so my memory is rather rusty.
// So I guess, I'll have to read up on macros before I start implementing them.
//
// Another thing I will have to keep in mind is that I am effectively designing
// a programming language. So I can have macros work any way I want.
// In particular, I might want to look at Racket before blindly copying
// Common Lisp.
//
// Mittwoch, 13. 12. 2017, 21:33
// Okay, I just took a quick look at Racket's documentation.
// Consider me confused.
// So, for the moment I am going to try and replicate Common Lisp's macro
// system. But I solemnly promise to look into Racket's approach to macros,
// and if I feel I learn something significant from it, I will make an attempt
// to bring that to kryLisp.
// Perhaps, as they say, another day.

func (inter *Interpreter) evalDefmacro(lst *value.List) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}

	if lst == nil || lst.Car == nil || lst.Car.Car == nil {
		return nil, errors.New("Argument is not a lambda list")
	} else if !lst.Car.Car.Equal(value.Symbol("DEFMACRO")) {
		return nil, SyntaxErrorf("The first element of a macro definition must be the Symbol DEFMACRO, not %s",
			lst.Car.Car)
	} else if lst.Car.Cdr.(*value.ConsCell).Car.Type() != types.Symbol {
		return nil, SyntaxErrorf("The first argument to DEFMACRO must be a symbol, not a %s",
			lst.Car.Cdr.(*value.ConsCell).Car.Type())
	}

	var (
		args, body *value.List
		err        error
		val        value.LispValue
		m          *value.Macro
	)

	if val, err = lst.Nth(2); err != nil {
		if inter.debug {
			fmt.Fprintf(inter.stderr, "Error getting third element (argument list) from macro definition: %s\n",
				err.Error())
		}
		return nil, err
	} else if val.Type() != types.List {
		return nil, &value.TypeError{
			Expected: "(Argument) List",
			Actual:   val.Type().String(),
		}
	}

	args = val.(*value.List)
	body = &value.List{
		Car:    lst.Car.Cdr.(*value.ConsCell).Cdr.(*value.ConsCell),
		Length: lst.Length - 2,
	}

	m = &value.Macro{
		Name: lst.Car.Cdr.(*value.ConsCell).Car.String(),
		Args: make([]value.LispValue, args.Length),
		Body: make([]value.LispValue, body.Length),
	}

	var idx int

	// Process arguments
	for cell := args.Car; cell != nil; cell = cell.Cdr.(*value.ConsCell) {
		m.Args[idx] = cell.Car
		idx++
		if cell.Cdr == nil {
			break
		}
	}

	idx = 0
	// Process body
	for cell := body.Car; cell != nil; cell = cell.Cdr.(*value.ConsCell) {
		m.Body[idx] = cell.Car
		idx++
		if cell.Cdr == nil {
			break
		}
	}

	return m, krylib.ErrNotImplemented
} // func (inter *Interpreter) evalDefmacro(l *value.List) (value.LispValue, error)

// If

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
//
// Donnerstag, 16. 11. 2017, 17:18
// I am about to embark on a big experiment, turning many special forms
// into the newly created GoFunctions, so they will behave more like functions
// from the Lisp point of view. I think the arithmnetic functions are
// predestined to be the starting point for this.
//
// Freitag, 08. 12. 2017, 19:51
// Sooner or later, I probably should clean up the arithmetic part a bit.

func (inter *Interpreter) evalPlus(arg *value.Arguments) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}

	if len(arg.Positional) < 1 {
		return value.IntValue(0), nil
	} else if len(arg.Positional) == 1 {
		if !value.IsNumber(arg.Positional[0]) {
			return value.NIL, &value.TypeError{
				Expected: "Number",
				Actual:   arg.Positional[0].Type().String(),
			}
		}

		return arg.Positional[0], nil
	}

	var acc value.Number = value.IntValue(0)

	for _, item := range arg.Positional {
		var (
			num, tmp value.Number
			ok       bool
			err      error
		)

		if num, ok = item.(value.Number); !ok {
			return value.NIL, &value.TypeError{
				Expected: "Number",
				Actual:   item.Type().String(),
			}
		} else if tmp, err = evalAddition(acc, num); err != nil {
			return value.NIL, err
		}

		acc = tmp
	}

	return acc, nil
} // func (inter *Interpreter) evalPlus(arg *value.Arguments) (value.LispValue, error)

func (inter *Interpreter) evalMinus(arg *value.Arguments) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}

	if len(arg.Positional) < 1 {
		return value.IntValue(0), nil
	} else if len(arg.Positional) == 1 {
		if value.IsNumber(arg.Positional[0]) {
			return evalNegate(arg.Positional[0].(value.Number))
		}

		return value.NIL, &value.TypeError{
			Expected: "Number",
			Actual:   arg.Positional[0].Type().String(),
		}
	}

	var (
		acc value.Number
		ok  bool
	)

	if acc, ok = arg.Positional[0].(value.Number); !ok {
		return value.NIL, &value.TypeError{
			Expected: "Number",
			Actual:   arg.Positional[0].Type().String(),
		}
	}

	for _, n := range arg.Positional[1:] {
		if !value.IsNumber(n) {
			return value.NIL, &value.TypeError{
				Expected: "Number",
				Actual:   n.Type().String(),
			}
		}

		var (
			num, tmp value.Number
			err      error
		)

		num = n.(value.Number)

		if tmp, err = evalSubtraction(acc, num); err != nil {
			return value.NIL, err
		}

		acc = tmp

	}

	return acc, nil
} // func (inter *Interpreter) evalMinus(arg *value.Arguments) (value.LispValue, error)

func (inter *Interpreter) evalMultiply(arg *value.Arguments) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}

	if len(arg.Positional) == 0 {
		return value.IntValue(1), nil
	} else if len(arg.Positional) == 1 {
		return arg.Positional[0], nil
	}

	var (
		acc, tmp value.Number
		err      error
	)

	acc = one

	for _, n := range arg.Positional {
		if !value.IsNumber(n) {
			return value.NIL, &value.TypeError{
				Expected: "Number",
				Actual:   n.Type().String(),
			}
		} else if tmp, err = evalMultiplication(acc, n.(value.Number)); err != nil {
			return value.NIL, err
		}

		acc = tmp
	}

	return acc, nil
} // func (inter *Interpreter) evalMultiply(arg *value.Arguments) (value.LispValue, error)

func (inter *Interpreter) evalDivide(arg *value.Arguments) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()

		_, _ = spew.Printf("DIV - args: %#v\n",
			arg.Positional)
	}

	// This makes me think how nice it would be to have rational numbers.
	// Ah, one day...
	if len(arg.Positional) == 0 {
		return value.NIL, SyntaxError("Calling divide (/) without arguments is not legal")
	} else if len(arg.Positional) == 1 {
		return evalDivision(value.FloatValue(1.0), arg.Positional[0].(value.Number))
	}

	var (
		acc, tmp, num value.Number
		err           error
		ok            bool
	)

	if acc, ok = arg.Positional[0].(value.Number); !ok {
		return value.NIL, fmt.Errorf("All arguments to DIVIDE must be numbers; %s is a %T",
			arg.Positional[0],
			arg.Positional[0])
	}

	for _, n := range arg.Positional[1:] {
		if num, ok = n.(value.Number); !ok {
			return value.NIL, &value.TypeError{
				Expected: "Number",
				Actual:   n.Type().String(),
			}
		} else if tmp, err = evalDivision(acc, num); err != nil {
			return value.NIL, err
		}

		acc = tmp
	}

	return acc, nil
} // func (inter *Interpreter) evalDivide(arg *value.Arguments) (value.LispValue, error)

/////////////////////////////////////////////////////////////////////////////
// Fundamental Lisp stuff ///////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////

func (inter *Interpreter) evalDefun(l *value.List) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}

	// Mittwoch, 13. 12. 2017, 19:36
	// Now that I have nested function environments, I realize that may not
	// be all that smart.
	// In Common Lisp, DEFUN installs the function into the current package,
	// regardless of scope.

	// (defun square (x) (* x x))
	// Nah, that is not sufficient - in Common Lisp, a function can also
	// have a documentation string.
	// So I need to check if the third element of the list is a string.
	var fn *value.LispFunction
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
	//
	// Dienstag, 14. 11. 2017, 21:40
	// I thought I had found a bug here and set the index of the doc string to 3,
	// which is correct for the full DEFUN-form, but at this point... wait.
	// It SHOULD be 3, but for some reason that causes various tests to fail.
	// That is weird.
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

	//inter.env.SetFn(name.String(), fn)
	var env = inter.env

	for env.Parent != nil {
		env = env.Parent
	}

	env.SetFn(name.String(), fn)

	return name, nil
} // func (inter *Interpreter) evalDefun(l *value.List) (value.LispValue, error)

func (inter *Interpreter) evalEq(args *value.Arguments) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}

	var argCnt int

	if argCnt = len(args.Positional); argCnt < 2 {
		return value.NIL, SyntaxError("calling < for LESS THAN two two arguments does not make sense")
	}

	var acc = args.Positional[0]

	for _, x := range args.Positional[1:] {
		if !acc.Eq(x) {
			return value.NIL, nil
		}
	}

	return value.T, nil
} // func (inter *Interpreter) evalEq(args *value.Arguments) (value.LispValue, error)

func (inter *Interpreter) evalLessThan(arg *value.Arguments) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}

	if len(arg.Positional) < 2 {
		return value.NIL, SyntaxError("< needs at least two arguments")
	}

	var (
		r  value.Number
		ok bool
	)

	if r, ok = arg.Positional[0].(value.Number); !ok {
		return value.NIL, &value.TypeError{
			Expected: "Number",
			Actual:   arg.Positional[0].Type().String(),
		}
	}

	for _, v := range arg.Positional[1:] {
		var res compare.Result
		var err error
		var n value.Number

		if n, ok = v.(value.Number); !ok {
			return value.NIL, &value.TypeError{
				Expected: "Number",
				Actual:   v.Type().String(),
			}
		} else if res, err = evalPolymorphCmp(r, n); err != nil {
			return value.NIL, err
		} else if res != compare.LessThan {
			return value.NIL, nil
		}

		r = n
	}

	return value.T, nil
} // func (inter *Interpreter) evalLessThan(arg *value.Arguments) (value.LispValue, error)

func (inter *Interpreter) evalGreaterThan(arg *value.Arguments) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}

	if len(arg.Positional) < 2 {
		return value.NIL, SyntaxError("< need at least two arguments")
	}

	var (
		r  value.Number
		ok bool
	)

	if r, ok = arg.Positional[0].(value.Number); !ok {
		return value.NIL, &value.TypeError{
			Expected: "Number",
			Actual:   arg.Positional[0].Type().String(),
		}
	}

	for _, v := range arg.Positional[1:] {
		var (
			res compare.Result
			n   value.Number
			err error
		)

		if n, ok = v.(value.Number); !ok {
			return value.NIL, &value.TypeError{
				Expected: "Number",
				Actual:   arg.Positional[0].Type().String(),
			}
		} else if res, err = evalPolymorphCmp(r, n); err != nil {
			return value.NIL, err
		} else if res != compare.GreaterThan {
			return value.NIL, nil
		}

		r = n
	}

	return value.T, nil
} // func (inter *Interpreter) evalGreaterThan(arg *value.Arguments) (value.LispValue, error)

func (inter *Interpreter) evalLessEqual(arg *value.Arguments) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}

	if len(arg.Positional) < 2 {
		return value.NIL, SyntaxError("<= needs at least two arguments.")
	}

	var (
		r  value.Number
		ok bool
	)

	if r, ok = arg.Positional[0].(value.Number); !ok {
		return value.NIL, &value.TypeError{
			Expected: "Number",
			Actual:   arg.Positional[0].Type().String(),
		}
	}

	for _, v := range arg.Positional[1:] {
		var (
			res compare.Result
			n   value.Number
			err error
		)

		if n, ok = v.(value.Number); !ok {
			return value.NIL, &value.TypeError{
				Expected: "Number",
				Actual:   arg.Positional[0].Type().String(),
			}
		} else if res, err = evalPolymorphCmp(r, n); err != nil {
			return value.NIL, err
		} else if res == compare.GreaterThan {
			return value.NIL, nil
		}

		r = n
	}

	return value.T, nil
} // func (inter *Interpreter) evalLessEqual(arg *value.Arguments) (value.LispValue, error)

func (inter *Interpreter) evalGreaterEqual(arg *value.Arguments) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}

	if len(arg.Positional) < 2 {
		return value.NIL, SyntaxError("<= needs at least two arguments.")
	}

	var (
		r  value.Number
		ok bool
	)

	if r, ok = arg.Positional[0].(value.Number); !ok {
		return value.NIL, &value.TypeError{
			Expected: "Number",
			Actual:   arg.Positional[0].Type().String(),
		}
	}

	for _, v := range arg.Positional[1:] {
		var (
			res compare.Result
			n   value.Number
			err error
		)

		if n, ok = v.(value.Number); !ok {
			return value.NIL, &value.TypeError{
				Expected: "Number",
				Actual:   arg.Positional[0].Type().String(),
			}
		} else if res, err = evalPolymorphCmp(r, n); err != nil {
			return value.NIL, err
		} else if res == compare.LessThan {
			return value.NIL, nil
		}

		r = n
	}

	return value.T, nil
} // func (inter *Interpreter) evalGreaterEqual(arg *value.Arguments) (value.LispValue, error)

// Muss CONS eine special form sein? Ich sehe erstmal keinen Grund dafür - die
// Auswertung der Argumente functioniert ja wie bei regulären Funktionen auch.
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
		return value.NIL, &value.TypeError{
			Expected: "List",
			Actual:   bindings.Type().String(),
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
			return value.NIL, &value.TypeError{
				Expected: "Symbol",
				Actual:   symbol.Type().String(),
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

func (inter *Interpreter) evalNot(arg *value.Arguments) (v value.LispValue, e error) {
	if inter.debug {
		krylib.Trace()
		defer func() {
			fmt.Printf("(not %s) => %s\n",
				arg.Positional[0],
				v)
		}()
	}

	if len(arg.Positional) != 1 {
		return value.NIL, SyntaxError("NOT needs exactly one argument")
	}

	if value.IsNil(arg.Positional[0]) {
		return value.T, nil
	}

	return value.NIL, nil
} // func (inter *Interpreter) evalNot(arg *value.Arguments) (value.LispValue, error)

// Dienstag, 21. 11. 2017, 14:33
// For the time being, AND remains a special form, because I want to support
// short circuit evaluation, which requires a special form, obviously.
// (I mean, one could hack something together with macros later on,
// but I have a hunch that in this case the difference in performance
// will be so large, that I probably would not want to replace the
// current arrangement. If I did measurements on that. Which I intend
// to, eventually, but not right now.)
//
// The same goes for OR and AND.
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

// Dienstag, 21. 11. 2017, 14:37
// Does SET! need to be a special form?
// Good question, actually. It needs to be implemented in Go, that's for sure.
// But the rules for argument evaluation are the same as for regular functions.
// ...
// No, actually, they are not. Not right now, at least. The destination
// is treated specially.
// I could follow Common Lisp and implement SET! as a native function and then
// add SETQ! as a macro or special form. But what advantage would that have?
// Until I can think of a good answer, SET! stays a special form.
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

// PRINT is a another interesting case - the way it is implemented right now,
// it stops evaluating arguments as soon as it encounters an error, and I
// can easily imagine how one might put that behavior to interesting use.
// Okay, for now it stays a special form, until I can think of a compelling
// reason to change that.
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

		inter.stdout.Write([]byte(val.String() + "\n")) // nolint: errcheck
	}

	return value.NIL, nil
} // func (inter *Interpreter) evalPrint(l *value.List) (value.LispValue, error)

// nolint: gocyclo
func (inter *Interpreter) evalApply(l *value.List) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
		inter.stdout.Write([]byte(l.String())) // nolint: errcheck
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
		return value.NIL, &value.TypeError{
			Expected: "Function",
			Actual:   fn.Type().String(),
		}
	} else if val, err = l.Nth(2); err != nil {
		return value.NIL, err
	} else if arglist, err = inter.Eval(val); err != nil {
		return value.NIL, err
	} else if arglist.Type() != types.List {
		return value.NIL, &value.TypeError{
			Expected: "List",
			Actual:   arglist.Type().String(),
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
		spew.Printf("APPLY: %#v\n", // nolint: errcheck
			funcall)
	}

	return inter.evalFuncall(funcall)
} // func (inter *Interpreter) evalApply(l *value.List) (value.LispValue, error)

func (inter *Interpreter) evalCar(l *value.List) (v value.LispValue, e error) {
	if inter.debug {
		krylib.Trace()
		inter.stdout.Write([]byte(l.String())) // nolint: errcheck
	}

	defer func() {
		spew.Printf("CAR returns %#v\n", v) // nolint: errcheck
	}()

	if value.IsNil(l.Car.Cdr.(*value.ConsCell).Car) {
		inter.stdout.Write([]byte("CAR of NIL is NIL")) // nolint: errcheck
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
		return value.NIL, &value.TypeError{
			Expected: "ConsCell or List",
			Actual:   val.Type().String(),
		}
	}
} // func (inter *Interpreter) evalCar(l *value.List) (value.LispValue, error)

// nolint: gocyclo
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
		return nil, &value.TypeError{
			Expected: "List or ConsCell",
			Actual:   val.Type().String(),
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
		if fn = inter.env.GetFn(string(sym)); fn == nil {
			return value.NIL, MissingFunctionError(sym)
		}
	case types.Function:
		fn = cadr.(value.Function)
	default:
		fmt.Printf("FN: Invalid argument %s\n", spew.Sdump(cadr))
		return value.NIL, &value.TypeError{
			Expected: "Symbol",
			Actual:   cadr.Type().String(),
		}
	}

	return fn, nil
} // func (inter *Interpreter) evalFn(l *value.List) (value.LispValue, error)

func (inter *Interpreter) evalIsNil(arg *value.Arguments) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}

	if len(arg.Positional) != 1 {
		return value.NIL, SyntaxError("NIL? requires exactly one argument")
	}

	if value.IsNil(arg.Positional[0]) {
		return value.T, nil
	}

	return value.NIL, nil
} // func (inter *Interpreter) evalIsNil(arg *value.Arguments) (value.LispValue, error)

func (inter *Interpreter) evalList(arg *value.Arguments) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}

	if len(arg.Positional) == 0 {
		return value.NIL, nil
	}

	var acc *value.ConsCell

	for i := len(arg.Positional) - 1; i >= 0; i-- {
		var cell = &value.ConsCell{
			Car: arg.Positional[i],
			Cdr: acc,
		}

		acc = cell
	}

	return &value.List{Car: acc, Length: len(arg.Positional)}, nil
} // func (inter *Interpreter) evalList(arg *value.Arguments) (value.LispValue, error)

/////////////////////////////////////////////////////////////////////////////
// Arrays ///////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////

func (inter *Interpreter) evalAref(arg *value.Arguments) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}

	if len(arg.Positional) != 2 {
		return value.NIL, SyntaxError("AREF requires exactly two arguments")
	} else if arg.Positional[0].Type() != types.Array {
		return value.NIL, &value.TypeError{
			Expected: "Array",
			Actual:   arg.Positional[0].Type().String(),
		}
	}

	var (
		arr   = arg.Positional[0].(value.Array)
		index int
		err   error
	)

	switch v := arg.Positional[1].(type) {
	case value.IntValue:
		index = int(v)
	case *value.BigInt:
		var tmp value.LispValue
		if tmp, err = v.Convert(types.Integer); err != nil {
			return value.NIL, &ValueError{
				val: v,
				msg: fmt.Sprintf("Index %s is too large",
					v.String()),
			}
		}

		index = int(tmp.(value.IntValue))
	default:
		return value.NIL, &value.TypeError{
			Expected: "Number",
			Actual:   v.Type().String(),
		}
	}

	if index < 0 || index >= len(arr) {
		return value.NIL, &ValueError{
			val: value.IntValue(index),
			msg: fmt.Sprintf("Index is out of range [0-%d)",
				len(arr)),
		}
	}

	return arr[index], nil
} // func (inter *Interpreter) evalAref(arg *value.Arguments) (value.LispValue, error)

func (inter *Interpreter) evalApush(arg *value.Arguments) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}

	if len(arg.Positional) < 2 {
		return value.NIL, SyntaxError("Usage: (apush <array> <val1>... )")
	}

	var (
		arr value.Array
		ok  bool
	)

	if arr, ok = arg.Positional[0].(value.Array); !ok {
		return value.NIL, &value.TypeError{
			Expected: "Array",
			Actual:   arg.Positional[0].Type().String(),
		}
	}

	arr = append(arr, arg.Positional[1:]...)

	return arr, nil
} // func (inter *Interpreter) evalApush(arg *value.Arguments) (value.LispValue, error)

func (inter *Interpreter) evalMakeArray(arg *value.Arguments) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}

	if len(arg.Positional) > 1 {
		return value.NIL, SyntaxError("MAKE-ARRAY requires exactly one argument!")
	} else if len(arg.Positional) == 0 {
		return make(value.Array, 0, 10), nil
	} else if arg.Positional[0].Type() != types.List {
		return value.NIL, &value.TypeError{
			Expected: "List",
			Actual:   arg.Positional[0].Type().String(),
		}
	}

	var (
		lst = arg.Positional[0].(*value.List)
		arr = make(value.Array, lst.Length)
		idx = 0
	)

	for cell := lst.Car; cell != nil; cell = cell.Cdr.(*value.ConsCell) {
		arr[idx] = cell.Car
		idx++
		if cell.Cdr == nil {
			break
		}
	}

	if inter.debug && idx != lst.Length {
		fmt.Printf("ERROR Making an array from a list of %d elements results in an array of %d elements\n",
			lst.Length,
			idx)
	}

	return arr, nil
} // func (inter *Interpreter) evalMakeArray(arg *value.Arguments) (value.LispValue, error)

/////////////////////////////////////////////////////////////////////////////
// Hash tables //////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////

// In Common Lisp, Hash tables are a bit fancier, and I may get around to copying
// that, too. But for now, I would like to keep things simple.
func (inter *Interpreter) evalMakeHash(arg *value.Arguments) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}

	var (
		sizeval value.LispValue
		size    int
		ok      bool
	)

	if sizeval, ok = arg.Keyword["SIZE"]; !ok {
		size = 16
	} else if sizeval.Type() == types.Integer {
		size = int(sizeval.(value.IntValue))
	} else if sizeval.Type() == types.BigInt {
		var biggie = sizeval.(*value.BigInt)

		if biggie.IsInt64() {
			var (
				v   value.LispValue
				err error
			)

			if v, err = biggie.Convert(types.Integer); err != nil {
				return value.NIL, err
			}

			size = int(v.(value.IntValue))
		}
	} else {
		return value.NIL, &value.TypeError{
			Expected: "Integer or BigInt",
			Actual:   sizeval.Type().String(),
		}
	}

	return make(value.Hashtable, size), nil
} // func (inter *Interpreter) evalMakeHash(arg *value.Arguments) (value.LispValue, error)

// Since we have literal syntax for hash tables, this function does not really need
// any arguments, now, does it?
// func (inter *Interpreter) evalMakeHash(l *value.List) (v value.LispValue, e error) {
// 	if inter.debug {
// 		krylib.Trace()
// 	}
// 	return make(value.Hashtable), nil
// } // func (inter *Interpreter) evalMakeHash(l *value.List) (v value.LispValue, e error)

func (inter *Interpreter) evalHashref(arg *value.Arguments) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}

	if len(arg.Positional) != 2 {
		return value.NIL, SyntaxError("HASHREF takes exactly two arguments")
	} else if arg.Positional[0].Type() != types.Hashtable {
		return value.NIL, &value.TypeError{
			Expected: "Hashtable",
			Actual:   arg.Positional[0].Type().String(),
		}
	}

	var (
		ht       value.Hashtable
		key, res value.LispValue
		ok       bool
	)

	ht = arg.Positional[0].(value.Hashtable)
	key = arg.Positional[1]

	if res, ok = ht[key]; !ok {
		return value.NIL, nil
	}

	return res, nil
} // func (inter *Interpreter) evalHashref(arg *value.Arguments) (value.LispValue, error)

func (inter *Interpreter) evalHasKey(arg *value.Arguments) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}

	if len(arg.Positional) != 2 {
		return value.NIL, SyntaxError("HASKEY takes exactly two arguments")
	}

	var (
		tbl value.Hashtable
		key value.LispValue
		ok  bool
	)

	if tbl, ok = arg.Positional[0].(value.Hashtable); !ok {
		return value.NIL, &value.TypeError{
			Expected: "Hashtable",
			Actual:   arg.Positional[0].Type().String(),
		}
	}

	key = arg.Positional[1]

	if _, ok = tbl[key]; !ok {
		return value.NIL, nil
	}

	return value.T, nil
} // func (inter *Interpreter) evalHashref(arg *value.Arguments) (value.LispValue, error)

func (inter *Interpreter) evalHashSet(arg *value.Arguments) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}

	if len(arg.Positional) != 3 {
		return value.NIL, SyntaxError("HASH-SET takes exactly three arguments (hash table, key, value)")
	}

	var (
		tbl      value.Hashtable
		key, val value.LispValue
		ok       bool
	)

	if tbl, ok = arg.Positional[0].(value.Hashtable); !ok {
		return value.NIL, &value.TypeError{
			Expected: "Hashtable",
			Actual:   arg.Positional[0].Type().String(),
		}
	}

	key = arg.Positional[1]
	val = arg.Positional[2]

	tbl[key] = val

	return val, nil
} // func (inter *Interpreter) evalHashSet(arg *value.Arguments) (value.LispValue, error)

// func (inter *Interpreter) evalHashSet(l *value.List) (v value.LispValue, e error) {
// 	if inter.debug {
// 		krylib.Trace()
// 	}

// 	// (hash-set tbl key val)
// 	if l == nil || l.Length != 4 {
// 		return value.NIL, SyntaxError("HASH-SET takes exactly *three* arguments")
// 	}

// 	var (
// 		tmp1, tmp2, key, val value.LispValue
// 		tbl                  value.Hashtable
// 		ok                   bool
// 		err                  error
// 	)

// 	if tmp1, err = l.Nth(1); err != nil {
// 		return value.NIL, err
// 	} else if tmp2, err = inter.Eval(tmp1); err != nil {
// 		return value.NIL, err
// 	} else if tmp2.Type() != types.Hashtable {
// 		return value.NIL, &value.TypeError{
// 			Expected: "Hashtable",
// 			Actual:   tmp2.Type().String(),
// 		}
// 	} else if tbl, ok = tmp2.(value.Hashtable); !ok {
// 		// CANTHAPPEN
// 		return value.NIL, &value.TypeError{
// 			Expected: "Hashtable",
// 			Actual:   tmp2.Type().String(),
// 		}
// 	} else if tmp1, err = l.Nth(2); err != nil {
// 		return value.NIL, err
// 	} else if key, err = inter.Eval(tmp1); err != nil {
// 		return value.NIL, err
// 	} else if tmp1, err = l.Nth(3); err != nil {
// 		return value.NIL, err
// 	} else if val, err = inter.Eval(tmp1); err != nil {
// 		return value.NIL, err
// 	}

// 	tbl[key] = val
// 	return val, nil
// } // func (inter *Interpreter) evalHashSet(l *value.List) (v value.LispValue, e error)

func (inter *Interpreter) evalHashDelete(arg *value.Arguments) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}

	var (
		ok  bool
		tbl value.Hashtable
		key value.LispValue
	)

	if len(arg.Positional) != 2 {
		return value.NIL, SyntaxError("HASH-DELETE requires exactly two arguments")
	} else if tbl, ok = arg.Positional[0].(value.Hashtable); !ok {
		return value.NIL, &value.TypeError{
			Expected: "Hashtable",
			Actual:   arg.Positional[0].Type().String(),
		}
	}

	key = arg.Positional[1]

	_, ok = tbl[key]
	delete(tbl, key)

	if ok {
		return value.T, nil
	}

	return value.NIL, nil
} // func (inter *Interpreter) evalHashDelete(arg *value.Arguments) (value.LispValue, error)

// func (inter *Interpreter) evalHashDelete(l *value.List) (v value.LispValue, e error) {
// 	if inter.debug {
// 		krylib.Trace()
// 	}

// 	// (hash-delete tbl key)
// 	if l == nil || l.Length != 3 {
// 		return value.NIL, SyntaxError("HASH-DELETE takes exactly *two* arguments")
// 	}

// 	var (
// 		tmp1, tmp2, key value.LispValue
// 		tbl             value.Hashtable
// 		ok              bool
// 		err             error
// 	)

// 	if tmp1, err = l.Nth(1); err != nil {
// 		return value.NIL, err
// 	} else if tmp2, err = inter.Eval(tmp1); err != nil {
// 		return value.NIL, err
// 	} else if tmp2.Type() != types.Hashtable {
// 		return value.NIL, &value.TypeError{
// 			Expected: "Hashtable",
// 			Actual:   tmp2.Type().String(),
// 		}
// 	} else if tbl, ok = tmp2.(value.Hashtable); !ok {
// 		// CANTHAPPEN
// 		return value.NIL, &value.TypeError{
// 			Expected: "Hashtable",
// 			Actual:   tmp2.Type().String(),
// 		}
// 	} else if tmp1, err = l.Nth(2); err != nil {
// 		return value.NIL, err
// 	} else if key, err = inter.Eval(tmp1); err != nil {
// 		return value.NIL, err
// 	}

// 	_, ok = tbl[key]
// 	delete(tbl, key)

// 	if ok {
// 		return value.T, nil
// 	}

// 	return value.NIL, nil
// } // func (inter *Interpreter) evalHashDelete(l *value.List) (v value.LispValue, e error)

/////////////////////////////////////////////////////////////////////////////
// Regular expressions //////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////

func (inter *Interpreter) evalRegexpCompile(arg *value.Arguments) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}

	if len(arg.Positional) != 1 {
		return value.NIL, SyntaxError("REGEXP-COMPILE takes exactly one argument")
	}

	var (
		rawPat value.StringValue
		pat    *value.Regexp
		err    error
		ok     bool
	)

	if rawPat, ok = arg.Positional[0].(value.StringValue); !ok {
		return value.NIL, &value.TypeError{
			Expected: "String",
			Actual:   arg.Positional[0].Type().String(),
		}
	}

	pat = new(value.Regexp)

	if pat.Pat, err = regexp.Compile(string(rawPat)); err != nil {
		return value.NIL, &ValueError{
			val: rawPat,
			msg: err.Error(),
		}
	}

	return pat, nil
} // func (inter *Interpreter) evalRegexpCompile(arg *value.Arguments) (value.LispValue, error)

func (inter *Interpreter) evalRegexpMatch(arg *value.Arguments) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}

	if len(arg.Positional) != 2 {
		return value.NIL, SyntaxError("REGEXP-MATCH takes exactly two arguments")
	}

	var (
		re      *value.Regexp
		sample  value.StringValue
		ok      bool
		matches [][]string
	)

	if re, ok = arg.Positional[0].(*value.Regexp); !ok {
		return value.NIL, &value.TypeError{
			Expected: "Regexp",
			Actual:   arg.Positional[0].Type().String(),
		}
	} else if sample, ok = arg.Positional[1].(value.StringValue); !ok {
		return value.NIL, &value.TypeError{
			Expected: "String",
			Actual:   arg.Positional[1].Type().String(),
		}
	}

	if matches = re.Pat.FindAllStringSubmatch(string(sample), -1); matches == nil {
		return value.NIL, nil
	}

	var result = make(value.Array, len(matches))

	for i, match := range matches {
		var groups = make(value.Array, len(match))
		for j, sub := range match {
			groups[j] = value.StringValue(sub)
		}

		result[i] = groups
	}

	return result, nil
} // func (inter *Interpreter) evalRegexpMatch(arg *value.Arguments) (value.LispValue, error)

type loopVariable struct {
	sym  value.Symbol
	step value.LispValue
}

// nolint: gocyclo
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
		return value.NIL, &value.TypeError{
			Expected: "List",
			Actual:   varList.Type().String(),
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
		if _, err = inter.Eval(cell.Car); err != nil {
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

func (inter *Interpreter) evalWhile(l *value.List) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}

	// Freitag, 24. 11. 2017, 22:22
	// I am not sure how useful that is, but we allow for a while loop with an empty body.
	// If the condition has side-effects this makes sense. If not, the condition never
	// changes, so the loop either does not execute at all, or it runs endlessly.

	if l.Length < 3 {
		return value.NIL, SyntaxError("(WHILE <CONDITION> ...)")
	}

	var (
		condition, check, res value.LispValue
		body                  *value.ConsCell
		err                   error
	)

	condition = l.Car.Cdr.(*value.ConsCell).Car
	//body = l.Car.Cdr.(*value.ConsCell).Cdr.(*value.ConsCell)
	if l.Length >= 3 {
		body = l.Car.Cdr.(*value.ConsCell).Cdr.(*value.ConsCell)
	}

	for check, err = inter.Eval(condition); err == nil && !value.IsNil(check); check, err = inter.Eval(condition) {
		for cell := body; cell != nil; cell = cell.Cdr.(*value.ConsCell) {
			if res, err = inter.Eval(cell.Car); err != nil {
				return value.NIL, err
			} else if cell.Cdr == nil {
				break
			}
		}
	}

	if err != nil {
		return value.NIL, err
	}

	return res, nil
} // func (inter *Interpreter) evalWhile(l *value.List) (value.LispValue, error)

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

func (inter *Interpreter) evalLength(arg *value.Arguments) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}

	if len(arg.Positional) != 1 {
		return value.NIL, SyntaxError("LENGTH requires exactly one argument")
	}

	var length int

	switch val := arg.Positional[0].(type) {
	case value.StringValue:
		length = len(string(val))
	case *value.ConsCell:
		length = val.ActualLength()
	case *value.List:
		length = val.Length
	case value.Array:
		length = len(val)
	case value.Hashtable:
		length = len(val)
	default:
		return value.NIL, &value.TypeError{
			Expected: "String, List, Array, or Hashtable",
			Actual:   arg.Positional[0].Type().String(),
		}
	}

	return value.IntValue(length), nil
} // func (inter *Interpreter) evalLength(arg *value.Arguments) (value.LispValue, error)

// Dienstag, 21. 11. 2017, 19:39
// Da die ganze Batterie an Funktionen, die CONCAT implementiert, recht
// umfangreich ist, lasse ich das erstmal so stehen.

// nolint: gocyclo
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

// nolint: gocyclo
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
		return acc + value.StringValue(string(other.(value.StringValue))), nil
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

// nolint: gocyclo
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
		return value.ListNil(), &value.TypeError{
			Expected: "Number, or List, or Array",
			Actual:   other.Type().String(),
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

func (inter *Interpreter) evalGetEnv(arg *value.Arguments) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}

	if len(arg.Positional) != 1 {
		return value.NIL, SyntaxError("GETENV takes exactly one argument")
	}

	var (
		key value.StringValue
		env string
		ok  bool
	)

	if key, ok = arg.Positional[0].(value.StringValue); !ok {
		return value.NIL, &value.TypeError{
			Expected: "String",
			Actual:   arg.Positional[0].Type().String(),
		}
	}

	env = os.Getenv(string(key))

	return value.StringValue(env), nil
} // func (inter *Interpreter) evalGetEnv(arg *value.Arguments) (value.LispValue, error)

func (inter *Interpreter) evalSetEnv(arg *value.Arguments) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}

	var (
		key, val value.StringValue
		ok       bool
		err      error
	)

	if len(arg.Positional) != 2 {
		return value.NIL, SyntaxError("SETENV takes exactly two arguments")
	} else if key, ok = arg.Positional[0].(value.StringValue); !ok {
		return value.NIL, &value.TypeError{
			Expected: "String",
			Actual:   arg.Positional[1].Type().String(),
		}
	} else if val, ok = arg.Positional[1].(value.StringValue); !ok {
		return value.NIL, &value.TypeError{
			Expected: "String",
			Actual:   arg.Positional[1].Type().String(),
		}
	}

	if err = os.Setenv(string(key), string(val)); err != nil {
		fmt.Printf("Error setting environment variable %s to value %s: %s",
			key,
			val,
			err.Error())
		return value.NIL, err
	}

	return val, nil
} // func (inter *Interpreter) evalSetEnv(arg *value.Arguments) (value.LispValue, error)

// Montag, 13. 11. 2017, 22:25
// How do I pass permissions and flags from Lisp?
// ... How does Common Lisp do this?
// Ah, like this: (open <path> :direction :input)
// That sounds kind of nice, I guess...

// nolint: gocyclo
func (inter *Interpreter) evalFopen(arg *value.Arguments) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}

	// This is not the most satisfactory solution ever, but I think 0644 is
	// the right permission set 90% of the time, and it's just the default,
	// anyway.
	// const defaultAccessRights = 0644

	var (
		path                        value.StringValue
		rawmode, append, sync, perm value.LispValue
		lmode                       value.Symbol
		mode                        filemode.FileMode
		accessRights                int
		ok                          bool
	)

	if len(arg.Positional) != 1 {
		return value.NIL, SyntaxError("FOPEN takes exactly one argument")
	} else if path, ok = arg.Positional[0].(value.StringValue); !ok {
		return value.NIL, &value.TypeError{
			Expected: "String",
			Actual:   arg.Positional[0].Type().String(),
		}
	} else if rawmode, ok = arg.Keyword["DIRECTION"]; !ok {
		// Dienstag, 21. 11. 2017, 20:40
		// Can I define a reasonable default value, or should I bail?
		return value.NIL, SyntaxError("FOPEN needs the :DIRECTION keyword")
	} else if lmode, ok = rawmode.(value.Symbol); !ok {
		return value.NIL, &ValueError{
			val: rawmode,
			msg: "Argument :DIRECTION to FOPEN must be a keyword symbol",
		}
	}

	switch lmode {
	case ":READ":
		mode = filemode.Read
	case ":WRITE":
		mode = filemode.Write
	case ":BOTH":
		mode = filemode.Read | filemode.Write
	default:
		return value.NIL, &ValueError{
			val: lmode,
			msg: ":DIRECTTION must be :READ, :WRITE or :BOTH",
		}
	}

	if append, ok = arg.Keyword["APPEND"]; ok {
		if !value.IsNil(append) {
			mode |= filemode.Append
		}
	}

	if sync, ok = arg.Keyword["SYNC"]; ok {
		if !value.IsNil(sync) {
			mode |= filemode.Sync
		}
	}

	// Maybe I should also allow users to pass in strings like the ones
	// one can pass to chmod.
	// I think that is a neat idea, but I will save it for later.
	if perm, ok = arg.Keyword["PERMISSION"]; !ok {
		// accessRights = defaultAccessRights
	} else if value.IsNil(perm) {
		return value.NIL, &ValueError{
			val: perm,
			msg: "File permissions must not be nil",
		}
	} else if perm.Type() != types.Integer {
		return value.NIL, &value.TypeError{
			Expected: "Integer",
			Actual:   perm.Type().String(),
		}
	}

	accessRights = int(perm.(value.IntValue))

	var (
		fh  *value.FileHandle
		err error
	)

	if fh, err = value.OpenFile(string(path), accessRights, mode); err != nil {
		return value.NIL, err
	}

	return fh, nil
} // func (inter *Interpreter) evalFopen(arg *value.Arguments) (value.LispValue, error)

func (inter *Interpreter) evalFclose(arg *value.Arguments) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}

	if len(arg.Positional) != 1 {
		return value.NIL, SyntaxError("FCLOSE needs exactly one argument")
	} else if arg.Positional[0].Type() != types.FileHandle {
		return value.NIL, &value.TypeError{
			Expected: "FileHandle",
			Actual:   arg.Positional[0].Type().String(),
		}
	}

	if err := arg.Positional[0].(*value.FileHandle).Close(); err != nil {
		return value.NIL, err
	}

	return value.NIL, nil
} // func (inter *Interpreter) evalFclose(arg *value.Arguments) (value.LispValue, error)

func (inter *Interpreter) evalFreadLine(arg *value.Arguments) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}

	// Mittwoch, 22. 11. 2017, 17:37
	// I haven't really thought about how I want to call this function from Lisp.
	// And I haven't done much I/O in Common Lisp or Scheme, either, so I have
	// no clue how to proceed.
	// I think it would be better to write a small test script first, so I
	// can figure out what kind of API I would be comfortable with.
	// And keep in mind, of course, that it would super nice to eventually
	// have a unified interface for file and network I/O.
	// Mmmmh. It appears I have already started writing a small test script,
	// and it ... mmmh.
	// I might choose an approach not entirely unlike Perl, with one very
	// simple API for line-by-line text data (because that is probably
	// what we will deal with 95% of the time), and one more complex
	// API that also allows to do stuff like binary I/O and is (hopefully)
	// more effcicient than the line-by-line approach.
	//
	// Mittwoch, 22. 11. 2017, 18:36
	// Okay, for the moment I will keep the interface relatively primitive.
	// But if I wanted somebody else to ever use this language, I would definitely
	// have to include a good I/O API.
	// My hope is that by building something really primitive, I get working
	// code faster, then I can write Lisp code that does I/O faster, and
	// then I have some experience to shape a new, improved I/O interface.

	var (
		line string
		err  error
		fh   *value.FileHandle
		ok   bool
	)

	if len(arg.Positional) < 1 {
		return value.NIL, SyntaxError("FREADLINE must be called with at least one argument")
	} else if fh, ok = arg.Positional[0].(*value.FileHandle); !ok {
		return value.NIL, &value.TypeError{
			Expected: "FileHandle",
			Actual:   arg.Positional[0].Type().String(),
		}
	} else if line, err = fh.ReadLine(); err != nil {
		if err != io.EOF {
			return value.NIL, err
		}
	} /* else if inter.debug {
		fmt.Printf("Read one line from %s: %s\n",
			fh,
			line)
	} */

	return value.StringValue(strings.TrimRight(line, "\r\n")), nil
} // func (inter *Interpreter) evalFreadLine(arg *value.Arguments) (value.LispValue, error)

func (inter *Interpreter) evalFwrite(arg *value.Arguments) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}

	if len(arg.Positional) < 2 {
		return value.NIL, SyntaxError("FWRITE requires at least two arguments (one I/O handle, one or more things to write)")
	} else if arg.Positional[0].Type() != types.FileHandle {
		return value.NIL, &value.TypeError{
			Expected: "FileHandle",
			Actual:   arg.Positional[0].Type().String(),
		}
	}

	var fh = arg.Positional[0].(*value.FileHandle)

	for idx, val := range arg.Positional[1:] {
		var (
			out            = val.String()
			outbuf         []byte
			err            error
			written, total int
		)

		outbuf = []byte(out)

	WRITE:
		if written, err = fh.Write(outbuf[total:]); err != nil {
			if !fh.IsEOF() {

				var msg = fmt.Sprintf("Error writing argument %d to %s: %s",
					idx+1,
					fh.Path,
					err.Error())
				fmt.Println(msg)
				return value.NIL, errors.New(msg)
			}
		} else if written == 0 {
			// Does this ever happen?
			var msg = "CANTHAPPEN - FWRITE did not return an error, but wrote 0 bytes!"
			fmt.Println(msg)
			if total != len(outbuf) {
				return value.NIL, errors.New(msg)
			}
		} else if total += written; total < len(outbuf) {
			goto WRITE
		}
	}

	return value.NIL, nil
} // func (inter *Interpreter) evalFwrite(arg *value.Arguments) (value.LispValue, error)

func (inter *Interpreter) evalFeof(arg *value.Arguments) (v value.LispValue, e error) {
	if inter.debug {
		krylib.Trace()
	}

	var (
		fh *value.FileHandle
		ok bool
	)

	if inter.debug {
		defer func() {
			fmt.Printf("FEOF %s --> %s\n",
				fh,
				v)
		}()
	}

	if len(arg.Positional) != 1 {
		return value.NIL, SyntaxError("FEOF takes exactly one argument")
	} else if fh, ok = arg.Positional[0].(*value.FileHandle); !ok {
		return value.NIL, &value.TypeError{
			Expected: "FileHandle",
			Actual:   arg.Positional[0].Type().String(),
		}
	} else if fh.IsEOF() {
		return value.T, nil
	}

	return value.NIL, nil
} // func (inter *Interpreter) evalFeof(arg *value.Arguments) (value.LispValue, error)

func (inter *Interpreter) evalClearEOF(arg *value.Arguments) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}

	var (
		fh *value.FileHandle
		ok bool
	)

	if len(arg.Positional) != 1 {
		return value.NIL, SyntaxError("CLEAREOF takes exactly one argument")
	} else if fh, ok = arg.Positional[0].(*value.FileHandle); !ok {
		return value.NIL, &value.TypeError{
			Expected: "FileHandle",
			Actual:   arg.Positional[0].Type().String(),
		}
	}

	fh.ClearEOF()

	return fh, nil
} // func (inter *Interpreter) evalClearEOF(arg *value.Arguments) (value.LispValue, error)

func (inter *Interpreter) evalReadString(arg *value.Arguments) (value.LispValue, error) {
	if inter.debug {
		krylib.Trace()
	}

	var (
		line   value.StringValue
		err    error
		ok     bool
		parsed interface{}
		prog   value.Program
		p      = parser.NewParser()
	)

	if len(arg.Positional) != 1 {
		return value.NIL, SyntaxError("READ-FROM-STRING expects exactly one argument")
	} else if line, ok = arg.Positional[0].(value.StringValue); !ok {
		return value.NIL, &value.TypeError{
			Expected: "String",
			Actual:   arg.Positional[0].Type().String(),
		}
	} else if line == "" {
		return value.NIL, nil
	} else if parsed, err = p.Parse(lexer.NewLexer([]byte(line))); err != nil {
		fmt.Printf("Error parsing input %s: %s\n",
			line,
			err.Error())
		return value.NIL, err
	} else if prog, ok = parsed.([]value.LispValue); !ok {
		return value.NIL, fmt.Errorf("Parser returned unexpected data type: %T",
			parsed)
	}

	return prog[0], nil
} // func (inter *Interpreter) evalReadString(arg *value.Arguments) (value.LispValue, error)
