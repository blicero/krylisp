// /home/krylon/go/src/github.com/blicero/krylisp/logdomain/logdomain.go
// -*- mode: go; coding: utf-8; -*-
// Created on 21. 12. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2025-02-23 20:42:56 krylon>

package logdomain

//go:generate stringer -type=ID

type ID uint8

const (
	Parser ID = iota
	Interpreter
)

func AllDomains() []ID {
	return []ID{
		Parser,
		Interpreter,
	}
} // func AllDomains() []ID
