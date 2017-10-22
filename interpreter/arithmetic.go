// /home/krylon/go/src/krylisp/interpreter/arithmetic.go
// -*- mode: go; coding: utf-8; -*-
// Created on 20. 10. 2017 by Benjamin Walkenhorst
// (c) 2017 Benjamin Walkenhorst
// Time-stamp: <2017-10-22 16:27:21 krylon>

package interpreter

import (
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
	constellation{types.Integer, types.Integer}: promotionrule{
		input:  constellation{types.Integer, types.Integer},
		output: types.Integer,
	},
}

func evalAddition(l, r value.Number) (value.Number, error) {
	var (
		result promotionrule
		ok     bool
		input  = constellation{l.Type(), r.Type()}
		err    error
	)

	if result, ok = arithRules[input]; !ok {
		return nil, &TypePromotionError{
			inputLeft:  l.Type(),
			inputRight: r.Type(),
		}
	} else if result.left {
		if result.output != l.Type() {
			if l, err = l.Convert(result.output); err != nil {
				return nil, err
			}
		}
	} else if result.output != r.Type() {
		if r, err = r.Convert(result.output); err != nil {
			return nil, err
		}
	}

	return nil, nil
} // func evalAddition(l, r value.Number) (value.Number, error)
