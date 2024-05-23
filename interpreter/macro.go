// /home/krylon/go/src/krylisp/interpreter/macro.go
// -*- mode: go; coding: utf-8; -*-
// Created on 08. 12. 2017 by Benjamin Walkenhorst
// (c) 2017 Benjamin Walkenhorst
// Time-stamp: <2024-05-23 14:00:45 krylon>
//
// This file implements macro expansion.
// My intention is to at least minimize the performance hit by performing
// macro expansion before the code actually evaluated.
// We'll see how that works out.

package interpreter

// func (inter *Interpreter) expandMacro(lst *value.List) (*value.List, error) {
// 	return nil, krylib.ErrNotImplemented
// } // func (inter *Interpreter) expandMacro(lst *value.List) (*value.List, error)
