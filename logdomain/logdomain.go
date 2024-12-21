// /home/krylon/go/src/github.com/blicero/krylisp/logdomain/logdomain.go
// -*- mode: go; coding: utf-8; -*-
// Created on 21. 12. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2024-12-21 21:15:47 krylon>

package logdomain

//go:generate stringer -type=ID

type ID uint8

const (
	Parser ID = iota
)
