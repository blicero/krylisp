// /home/krylon/go/src/krylisp/interpreter/arithmetic.go
// -*- mode: go; coding: utf-8; -*-
// Created on 20. 10. 2017 by Benjamin Walkenhorst
// (c) 2017 Benjamin Walkenhorst
// Time-stamp: <2017-10-25 14:27:13 krylon>

package interpreter

import (
	"fmt"
	"krylisp/types"
	"krylisp/value"
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

func evalAddition(l, r value.Number) (value.Number, error) {
	var (
		promoResult promotionrule
		ok          bool
		input       = constellation{l.Type(), r.Type()}
		err         error
		tmp         value.LispValue
	)

	if promoResult, ok = arithRules[input]; !ok {
		return nil, &TypePromotionError{
			inputLeft:  l.Type(),
			inputRight: r.Type(),
		}
	} else if promoResult.left {
		if promoResult.output != l.Type() {
			if tmp, err = l.Convert(promoResult.output); err != nil {
				return nil, err
			}

			l = tmp.(value.Number)
		}
	} else if promoResult.output != r.Type() {
		if tmp, err = r.Convert(promoResult.output); err != nil {
			return nil, err
		}

		r = tmp.(value.Number)
	}

	// For simplicity's sake, I will assume that after type promotion,
	// both arguments are of the same type.
	var resultValue value.Number

	switch lv := l.(type) {
	case value.IntValue:
		resultValue = lv + r.(value.IntValue)
	case value.FloatValue:
		resultValue = lv + r.(value.FloatValue)
	default:
		return nil, fmt.Errorf("Don't know how to handle numeric type %T", l)
	}

	return resultValue, nil
} // func evalAddition(l, r value.Number) (value.Number, error)
