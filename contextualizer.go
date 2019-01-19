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
	// Begin/continue tracking if I need a conditional inserted.
	condparent := false
	if args.condctx == nil && n.canHaveConditionalContext() {
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
	condition_disable_fields = []symbol{assign_token, path_token}
)

// ------------------------------------------------------------
// BOILERPLATE

func contextualizerFakeFmt() {
	fmt.Println()
}
