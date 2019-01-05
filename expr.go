package sqi

import (
	"errors"
)

// --------------------------------------------------------------------------------------
// EXPR

// Expr is an interface for anything that can evaluate input.
type Expr interface {
	Eval(interface{}) (interface{}, error)
}

// MakeExpr() converts an expression string into an evaluatable object.
func MakeExpr(term string) (Expr, error) {
	tokens, err := scan(term)
	if err != nil {
		return nil, err
	}
	ast, err := parse(tokens)
	if err != nil {
		return nil, err
	}
	return &expr_t{ast: ast}, nil
}

// --------------------------------------------------------------------------------------
// EXPR_T

type expr_t struct {
	ast AstNode
}

func (e *expr_t) Eval(input interface{}) (interface{}, error) {
	if e.ast == nil {
		return nil, errors.New("Missing AST")
	}
	return e.ast.Run(input)
}
