package sqi

// --------------------------------------------------------------------------------------
// EXPR

// Expr is an interface for anything that can evaluate input.
type Expr interface {
	Eval(interface{}, *Opt) (interface{}, error)
}

// MakeExpr converts an expression string into an evaluatable object.
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
	return &exprT{ast: ast}, nil
}

// --------------------------------------------------------------------------------------
// EXPR-T

type exprT struct {
	ast AstNode
}

func (e *exprT) Eval(input interface{}, opt *Opt) (interface{}, error) {
	if e.ast == nil {
		return nil, newEvalError("missing AST")
	}
	return e.ast.Eval(input, opt)
}
