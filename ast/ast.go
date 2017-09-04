// /home/krylon/go/src/krylisp/ast/ast.go
// -*- mode: go; coding: utf-8; -*-
// Created on 04. 09. 2017 by Benjamin Walkenhorst
// (c) 2017 Benjamin Walkenhorst
// Time-stamp: <2017-09-04 19:54:52 krylon>

package ast

type Node interface {
	TokenLiteral() string
}

// In Lisp brauche ich ja eigentlich keine reinen Statements, die keine
// Expressions sind, oder?
type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

type Program struct {
	Expressions []Expression
}

func (p *Program) TokenLiteral() string {
	if len(p.Expressions) > 0 {
		return p.Expressions[0].TokenLiteral()
	} else {
		return ""
	}
} // func (p *Program) TokenLiteral() string

type Atom interface {
	Node
	atomNode()
}
