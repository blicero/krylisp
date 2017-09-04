// /home/krylon/go/src/krylisp/token/token.go
// -*- mode: go; coding: utf-8; -*-
// Created on 04. 09. 2017 by Benjamin Walkenhorst
// (c) 2017 Benjamin Walkenhorst
// Time-stamp: <2017-09-04 20:04:12 krylon>
//
// Montag, 04. 09. 2017, 20:03
// Ich bin mir noch nicht sicher, wie weit sich diese Datei gegenüber dem
// Monkey-Interpreter unterscheiden muss. Ich übernehme erstmal die Elemente,
// die mir sinnvoll erscheinen und schaue dann im Weiteren mal.

package token

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
}

const (
	ILLEGAL   = "ILLEGAL"
	EOF       = "EOF"
	IDENT     = "IDENT"
	SYMBOL    = "SYMBOL"
	LAMBDA    = "LAMBDA"
	NIL       = "NIL"
	PLUS      = "PLUS"
	MINUS     = "-"
	BANG      = "!"
	ASTERISK  = "*"
	SLASH     = "/"
	PERCENT   = "%"
	LT        = "<"
	GT        = ">"
	SEMICOLON = ";"
	LPAREN    = "("
	RPAREN    = ")"
	LBRACE    = "{"
	RBRACE    = "}"
	LBRACKET  = "["
	RBRACKET  = "]"
	LET       = "LET"
	TRUE      = "TRUE"
	FALSE     = "FALSE"
	IF        = "IF"
)

var keywords = map[string]TokenType{
	"lambda": LAMBDA,
	"if":     IF,
	"t":      TRUE,
	"nil":    NIL,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	} else {
		return IDENT
	}
} // func LookupIdent(ident string) TokenType
