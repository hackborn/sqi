package sqi

import (
	"fmt"
)

// contextualize() insert rules specific to evaluating our rules.
func contextualize(tree *node_t) (*node_t, error) {
	if tree == nil {
		return tree, nil
	}
	args := &contextualizeArgs{}
	return tree.contextualize(args)
}

// ------------------------------------------------------------
// NODE-T (contextualizing)

func (n *node_t) contextualize(args *contextualizeArgs) (*node_t, error) {
	// Strings on the right of certain tokens will be strings,
	// but all other strings are field selectors.
	if n.isFieldContext() {
		n.Token = token_map[field_token]
	}

	// Begin/continue tracking if I need a conditional inserted.
	condparent := false
	if args.condctx == nil &&  n.canHaveConditionalContext() {
		args.condctx = &conditionalContext{needed: true}
		condparent = true
	} else if args.condctx != nil && n.cannotHaveConditionalContext() {
		args.condctx.needed = false
	}

	err := n.contextualizeChildren(args)
	if err != nil {
		return nil, err
	}

	ans := n
	if args.condctx != nil && condparent {
		if args.condctx.needed {
			ans = newNode(condition_token, "")
			ans.addChild(n)
		}
		args.condctx = nil
	}
	return ans, nil
}

func (n *node_t) contextualizeChildren(args *contextualizeArgs) error {
	for i, c := range n.Children {
		c2, err := c.contextualize(args)
		if err != nil {
			return err
		}
		n.Children[i] = c2
	}
	return nil
}

// isFieldContext() answers true if this string should be a field.
func (n *node_t) isFieldContext() bool {
	if n.Token.Symbol != string_token {
		return false
	}
	// Strings are the right-hand size of a string-capabale binary.
	if n.Parent != nil && len(n.Parent.Children) == 2 && n.Parent.Token.any(string_capable_rhs...) && n.Parent.Children[1] == n {
		return false
	}
	return true
}

func (n *node_t) canHaveConditionalContext() bool {
	return n.Token.inside(start_comparison, end_comparison) || n.Token.inside(start_conditional, end_conditional)
}

func (n *node_t) cannotHaveConditionalContext() bool {
	return n.Token.any(condition_disable_fields...)
}

// ------------------------------------------------------------
// CONTEXTUALIZE-ARGS

type contextualizeArgs struct {
	condctx *conditionalContext
}

// ------------------------------------------------------------
// CONDITIONAL-CONTEXT

type conditionalContext struct {
	needed bool
}

// ------------------------------------------------------------
// CONST and VAR

var (
	string_capable_rhs = []symbol{assign_token, eql_token, neq_token}
	condition_disable_fields = []symbol{assign_token, path_token}
)

// ------------------------------------------------------------
// BOILERPLATE

func contextualizerFakeFmt() {
	fmt.Println()
}
