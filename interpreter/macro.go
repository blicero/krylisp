// /home/krylon/go/src/krylisp/interpreter/macro.go
// -*- mode: go; coding: utf-8; -*-
// Created on 08. 12. 2017 by Benjamin Walkenhorst
// (c) 2017 Benjamin Walkenhorst
// Time-stamp: <2017-12-12 18:26:55 krylon>
//
// This file implements macro expansion.
// My intention is to at least minimize the performance hit by performing
// macro expansion before the code actually evaluated.
// We'll see how that works out.

package interpreter

import (
	"krylib"
	"krylisp/value"
)

func (inter *Interpreter) expandMacro(lst *value.List) (*value.List, error) {
	return nil, krylib.NotImplemented
} // func (inter *Interpreter) expandMacro(lst *value.List) (*value.List, error)
