// /home/krylon/go/src/krylisp/interpreter/arithmetic.go
// -*- mode: go; coding: utf-8; -*-
// Created on 20. 10. 2017 by Benjamin Walkenhorst
// (c) 2017 Benjamin Walkenhorst
// Time-stamp: <2017-10-31 21:32:27 krylon>
//
// Donnerstag, 26. 10. 2017, 17:00
// I would like to have seamless transitions between Fixnum and Bignum,
// i.e. if the result of an operation fits in an int64, it returns an int64,
// and a BigInt if not.
// But that means I would have to check for over/underflow in all arithmetic
// operations. That seems a bit harsh.
// I *could* try to mitigate the performance hit by stealing a trick from the
// good old Lisp Machine perform the check in parallel. ???
// Or would that just make things more convoluted?

package interpreter

import (
	"fmt"
	"krylisp/compare"
	"krylisp/types"
	"krylisp/value"
	"math/big"
)

// Zero is the number 0, as you probably have guessed.
const (
	Zero = value.IntValue(0)
)

/*
   Some non-optimized but clean code from Rob Pike himself:
*/
const mostNegative = -(mostPositive + 1)
const mostPositive = 1<<63 - 1

func signedMulOverflows(a, b int64) bool {
	if a == 0 || b == 0 || a == 1 || b == 1 {
		return false
	}
	if a == mostNegative || b == mostNegative {
		return true
	}
	c := a * b
	return c/b != a
}

type constellation [2]types.ID

type promotionrule struct {
	input  constellation
	left   bool
	output types.ID
}

var arithRules = map[constellation]promotionrule{
	// Integer
	constellation{types.Integer, types.Integer}: promotionrule{
		input:  constellation{types.Integer, types.Integer},
		left:   false,
		output: types.Integer,
	},
	constellation{types.Integer, types.Float}: promotionrule{
		input:  constellation{types.Integer, types.Float},
		left:   true,
		output: types.Float,
	},
	constellation{types.Integer, types.BigInt}: promotionrule{
		input:  constellation{types.Integer, types.BigInt},
		left:   true,
		output: types.BigInt,
	},

	// Float
	constellation{types.Float, types.Float}: promotionrule{
		input:  constellation{types.Float, types.Float},
		output: types.Float,
	},
	constellation{types.Float, types.Integer}: promotionrule{
		input:  constellation{types.Integer, types.Integer},
		output: types.Float,
	},
	constellation{types.Float, types.BigInt}: promotionrule{
		input:  constellation{types.Float, types.BigInt},
		output: types.Float,
	},

	// BigInt
	constellation{types.BigInt, types.Integer}: promotionrule{
		input:  constellation{types.BigInt, types.Integer},
		output: types.BigInt,
	},
	// Since double precision floats can have values that are
	// pretty *!&%$ huge, I pretend this is not ever going to
	// cause problems.
	constellation{types.BigInt, types.Float}: promotionrule{
		input:  constellation{types.BigInt, types.Float},
		output: types.Float,
		left:   true,
	},
	constellation{types.BigInt, types.BigInt}: promotionrule{
		input:  constellation{types.BigInt, types.BigInt},
		output: types.BigInt,
	},
}

func promoteTypes(l, r value.Number) (value.Number, value.Number, error) {
	var (
		promoResult promotionrule
		ok          bool
		input       = constellation{l.Type(), r.Type()}
		err         error
		tmp         value.LispValue
	)

	if promoResult, ok = arithRules[input]; !ok {
		return nil, nil, &TypePromotionError{
			inputLeft:  l.Type(),
			inputRight: r.Type(),
		}
	} else if promoResult.left {
		if promoResult.output != l.Type() {
			if tmp, err = l.Convert(promoResult.output); err != nil {
				return nil, nil, err
			}

			l = tmp.(value.Number)
		}
	} else if promoResult.output != r.Type() {
		if tmp, err = r.Convert(promoResult.output); err != nil {
			return nil, nil, err
		}

		r = tmp.(value.Number)
	}

	return l, r, nil
} // func promoteTypes(l, r value.Number) (value.Number, value.Number)

// I will have to check for overflow, too, if I want to promote integers
// to bignums automatically.

func evalNegate(x value.Number) (value.Number, error) {
	switch n := x.(type) {
	case value.IntValue:
		return -n, nil
	case value.FloatValue:
		return -n, nil
	case *value.BigInt:
		var tmp = &value.BigInt{Value: new(big.Int)}
		tmp.Value.Neg(n.Value)
		return tmp, nil
	default:
		return nil, fmt.Errorf("I do not know how to negate a %T value", x)
	}
}

func evalAddition(l, r value.Number) (value.Number, error) {
	var lop, rop, resultValue value.Number
	var err error

	// For simplicity's sake, I will assume that after type promotion,
	// both arguments are of the same type.
	//var resultValue value.Number

	if lop, rop, err = promoteTypes(l, r); err != nil {
		return nil, err
	}

	switch lv := lop.(type) {
	case value.IntValue:
		resultValue = lv + rop.(value.IntValue)
	case value.FloatValue:
		resultValue = lv + rop.(value.FloatValue)
	case *value.BigInt:
		var tmp = &value.BigInt{Value: new(big.Int)}
		tmp.Value.Add(lv.Value, rop.(*value.BigInt).Value)
		resultValue = tmp
	default:
		return nil, fmt.Errorf("Don't know how to handle numeric type %T", lop)
	}

	return resultValue, nil
} // func evalAddition(l, r value.Number) (value.Number, error)

func evalSubtraction(l, r value.Number) (value.Number, error) {
	var lop, rop, resultValue value.Number
	var err error

	// For simplicity's sake, I will assume that after type promotion,
	// both arguments are of the same type.
	//var resultValue value.Number

	if lop, rop, err = promoteTypes(l, r); err != nil {
		return nil, err
	}

	switch lv := lop.(type) {
	case value.IntValue:
		resultValue = lv - rop.(value.IntValue)
	case value.FloatValue:
		resultValue = lv - rop.(value.FloatValue)
	case *value.BigInt:
		var tmp = &value.BigInt{Value: new(big.Int)}
		tmp.Value.Sub(lv.Value, rop.(*value.BigInt).Value)
		resultValue = tmp
	default:
		return nil, fmt.Errorf("Don't know how to handle numeric type %T", lop)
	}

	return resultValue, nil
} // func evalSubtraction(l, r value.Number) (value.Number, error)

func evalMultiplication(l, r value.Number) (value.Number, error) {
	var (
		lop, rop, result value.Number
		err              error
	)

	if lop, rop, err = promoteTypes(l, r); err != nil {
		return Zero, err
	}

	switch lv := lop.(type) {
	case value.IntValue:
		result = lv * rop.(value.IntValue)
	case value.FloatValue:
		result = lv * rop.(value.FloatValue)
	case *value.BigInt:
		var tmp = &value.BigInt{Value: new(big.Int)}
		tmp.Value.Mul(lv.Value, rop.(*value.BigInt).Value)
		result = tmp
	default:
		return nil, fmt.Errorf("I do not know how to multiply a %T and a %T",
			lop, rop)
	}

	return result, nil
} // func evalMultiplication(l, r value.Number) (value.Number, error)

func evalDivision(l, r value.Number) (value.Number, error) {
	var (
		lop, rop, result value.Number
		err              error
	)

	if lop, rop, err = promoteTypes(l, r); err != nil {
		return Zero, err
	} else if rop.IsZero() {
		return Zero, &ValueError{rop}
	}

	switch lv := lop.(type) {
	case value.IntValue:
		result = lv / rop.(value.IntValue)
	case value.FloatValue:
		result = lv / rop.(value.FloatValue)
	case *value.BigInt:
		var tmp = &value.BigInt{Value: new(big.Int)}
		tmp.Value.Div(lv.Value, rop.(*value.BigInt).Value)
		result = tmp
	default:
		return nil, fmt.Errorf("I do not know how to divide a %T by a %T",
			l, r)
	}

	return result, nil
} // func evalDivision(l, r value.Number) (value.Number, error)

func evalPolymorphLT(l, r value.Number) (compare.Result, error) {
	var (
		lop, rop value.Number
		err      error
	)

	if lop, rop, err = promoteTypes(l, r); err != nil {
		return compare.Undefined, err
	}

	switch lv := lop.(type) {
	case value.IntValue:
		return cmpInt(lv, rop.(value.IntValue)), nil
	case value.FloatValue:
		return cmpFloat(lv, rop.(value.FloatValue)), nil
	case *value.BigInt:
		return cmpBigInt(lv, rop.(*value.BigInt)), nil
	default:
		return compare.Undefined, fmt.Errorf("Less-Than comparison is not implemented for type %T",
			lop)
	}
} // func evalPolymorphLT(l, r value.Number) (value.Number, error)

func cmpInt(l, r value.IntValue) compare.Result {
	if l < r {
		return compare.LessThan
	} else if l == r {
		return compare.Equal
	} else if l > r {
		return compare.GreaterThan
	}

	panic("CANTHAPPEN: two Integer values cannot be sorted")
} // func cmpInt(l, r value.IntValue) int

func cmpFloat(l, r value.FloatValue) compare.Result {
	if l < r {
		return compare.LessThan
	} else if l == r {
		return compare.Equal
	} else if l > r {
		return compare.GreaterThan
	}

	panic("CANTHAPPEN: two FloatValues cannot be sorted")
} // func cmpFloat(l, r, value.FloatValue) value.LispValue

func cmpBigInt(l, r *value.BigInt) compare.Result {
	v := value.IntValue(l.Value.Cmp(r.Value))
	switch v {
	case -1:
		return compare.LessThan
	case 0:
		return compare.Equal
	case 1:
		return compare.GreaterThan
	default:
		panic(fmt.Errorf("CANTHAPPEN - Undefined comparison for BigInt values: %v", v))
	}

} // func cmpBigInt(l, r value.BigInt) value.LispValue
