// /home/krylon/go/src/krylisp/permission/permission.go
// -*- mode: go; coding: utf-8; -*-
// Created on 11. 11. 2017 by Benjamin Walkenhorst
// (c) 2017 Benjamin Walkenhorst
// Time-stamp: <2017-12-11 18:52:42 krylon>

// Package filemode implements symbolic constants for file access modes.
package filemode

//go:generate stringer -type=FileMode

// FileMode is a "bitfield", i.e. a set of flags that indicate what permissions
// a given I/O handle has on the underlying file.
type FileMode uint16

// I hope these are self-explanatory.
const (
	Read FileMode = 1 << iota
	Write
	Append
	Sync
)
