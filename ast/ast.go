// /home/krylon/go/src/krylisp/ast/ast.go
// -*- mode: go; coding: utf-8; -*-
// Created on 04. 09. 2017 by Benjamin Walkenhorst
// (c) 2017 Benjamin Walkenhorst
// Time-stamp: <2017-09-06 10:26:27 krylon>

package ast

import (
	"bytes"
	"go/token"
	"strings"
)

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

func (p *Program) String() string {
	var out bytes.Buffer

	for _, s := range p.Expressions {
		out.WriteString(s.String())
	}

	return out.String()
} // func (p *Program) String() string

type Atom interface {
	Node
	atomNode()
}

type ExpressionStatement struct {
	Token      token.Token
	Expression Expression
}

func (es *ExpressionStatement) statementNode() {}
func (es *ExpressionStatement) TokenLiteral() string {
	return es.Token.Literal
} // func (es *ExpressionStatement) TokenLiteral() string

type Symbol struct {
	Token token.Token
	Value string
}

func (sym *Symbol) expressionNode() {}
func (sym *Symbol) atomNode()       {}
func (sym *Symbol) TokenLiteral() string {
	return sym.Token.Literal
} // func (sym *Symbol) TokenLiteral() string

func (sym *Symbol) String() string {
	return sym.Value
} // func (sym *Symbol) String() string

// Montag, 04. 09. 2017, 20:18
// Mittel- bis langfristig möchte ich natürlich gern bignum-Support
// einbauen, was ja auch nicht allzu schwer werden dürfte, weil Go
// das ja seinerseits auch schon mitbringt.
// Aber fangen wir erstmal einfach an.
type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode() {}
func (il *IntegerLiteral) atomNode()       {}
func (il *IntegerLiteral) TokenLiteral() string {
	return il.Token.Literal
} // func (il *IntegerLiteral) TokenLiteral() string

type IfExpression struct {
	Token       token.Token
	Condition   Expression
	Consequence Expression
	Alternative Expression
}

func (ie *IfExpression) expressionNode()      {}
func (ie *IfExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IfExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(if ")
	out.WriteString(ie.Condition.String())
	out.WriteString(ie.Consequence.String())
	if ie.Alternative != nil {
		out.WriteString(ie.Alternative.String())
	}
	out.WriteString(")")

	return out.String()
} // func (ie *IfExpression) String() string

type BlockExpression struct {
	Token       token.Token
	Expressions []Expression
}

func (be *BlockExpression) statementNode()       {}
func (be *BlockExpression) expressionNode()      {}
func (be *BlockExpression) TokenLiteral() string { return be.Token.Literal }
func (be *BlockExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(begin ")
	for _, s := range be.Expressions {
		out.WriteString("\n")
		out.WriteString(s.String())
	}
	out.WriteString(")")

	return out.String()
} // func (be *BlockExpression) String() string

type FunctionLiteral struct {
	Token      token.Token
	Parameters []Symbol
	Body       *BlockExpression
}

func (fl *FunctionLiteral) expressionNode()      {}
func (fl *FunctionLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FunctionLiteral) String() string {
	var out bytes.Buffer
	var params = make([]string, len(fl.Parameters))

	for idx, arg := range fl.Parameters {
		params[idx] = arg.String()
	}

	//out.WriteString(fl.TokenLiteral())
	out.WriteString("(lambda ")
	out.WriteString("(")
	out.WriteString(strings.Join(params, " "))
	out.WriteString(")")
	out.WriteString(fl.Body.String())
	out.WriteString(")")

	return out.String()
} // func (fl *FunctionLiteral) String() string

type CallExpression struct {
	Token     token.Token // The '('
	Function  Expression
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) String() string {
	var out bytes.Buffer
	var args = make([]string, len(ce.Arguments))

	for i, a := range ce.Arguments {
		args[i] = ce.Arguments[i].String()
	}

	out.WriteString("(")
	out.WriteString(ce.Function.String())
	out.WriteString(" ")
	out.Writestring(strings.Join(args), " ")
	out.WriteString(")")

	return out.String()
} // func (ce *CallExpression) String() string

type StringLiteral struct {
	Token token.Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) atomNode()            {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return sl.Token.Literal }

// Mittwoch, 06. 09. 2017, 10:25
// Im monkey/ast package kommt hier ein ArrayLiteral, weiß gar nicht sicher, ob
// ich das brauche oder will.
// Mittelfristig wäre es nett, aber für den Anfang will ich die grundlegenden
// Sachen klar machen: Atome, Listen, Functionen.
