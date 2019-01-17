package sqi

import (
	"errors"
)

// --------------------------------------------------------------------------------------
// EXPR

// Expr is an interface for anything that can evaluate input.
type Expr interface {
	Eval(interface{}, *Opt) (interface{}, error)
}

// MakeExpr() converts an expression string into an evaluatable object.
func MakeExpr(term string) (Expr, error) {
	tokens, err := scan(term)
	if err != nil {
		return nil, err
	}
	tree, err := parse(tokens)
	if err != nil {
		return nil, err
	}
	if tree == nil {
		return nil, newParseError("no result")
	}
	tree, err = contextualize(tree)
	if err != nil {
		return nil, err
	}
	ast, err := tree.asAst()
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

func (e *expr_t) Eval(input interface{}, opt *Opt) (interface{}, error) {
	if e.ast == nil {
		return nil, errors.New("Missing AST")
	}
	return e.ast.Eval(input, opt)
}
