// /home/krylon/go/src/krylisp/interpreter/arithmetic.go
// -*- mode: go; coding: utf-8; -*-
// Created on 20. 10. 2017 by Benjamin Walkenhorst
// (c) 2017 Benjamin Walkenhorst
// Time-stamp: <2017-10-25 21:05:39 krylon>

package interpreter

import (
	"fmt"
	"krylisp/types"
	"krylisp/value"
)

// Zero is the number 0, as you probably have guessed.
const (
	Zero = value.IntValue(0)
)

type constellation [2]types.ID

type promotionrule struct {
	input  constellation
	left   bool
	output types.ID
}

var arithRules = map[constellation]promotionrule{
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
	constellation{types.Float, types.Float}: promotionrule{
		input:  constellation{types.Float, types.Float},
		output: types.Float,
	},
	constellation{types.Float, types.Integer}: promotionrule{
		input:  constellation{types.Integer, types.Integer},
		output: types.Float,
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
	default:
		return nil, fmt.Errorf("I do not know how to divide a %T by a %T",
			l, r)
	}

	return result, nil
} // func evalDivision(l, r value.Number) (value.Number, error)
