package sqi

import (
	"fmt"
)

// contextualize() insert nodes specific to evaluating our rules.
func contextualize(tree *nodeT) (*nodeT, error) {
	if tree == nil {
		return tree, nil
	}
	args := &contextualizeArgs{}
	return tree.contextualize(args)
}

// ------------------------------------------------------------
// NODE-T (contextualizing)

func (n *nodeT) contextualize(args *contextualizeArgs) (*nodeT, error) {
	// Begin/continue tracking if I need a select inserted.
	selectparent := false
	if args.selectctx == nil && n.canHaveSelect() {
		args.selectctx = &selectContext{needed: true, notneeded: false}
		selectparent = true
	} else if args.selectctx != nil && n.doesNotNeedSelect() {
		args.selectctx.notneeded = true
	}

	err := n.contextualizeChildren(args)
	if err != nil {
		return nil, err
	}

	ans := n
	if args.selectctx != nil && selectparent {
		if args.selectctx.needed && args.selectctx.notneeded {
			return nil, newParseError("conflicting select conditions")
		}
		if args.selectctx.needed {
			ans = newNode(selectToken, "")
			ans.addChild(n)
		}
		args.selectctx = nil
	}
	return ans, nil
}

func (n *nodeT) contextualizeChildren(args *contextualizeArgs) error {
	for i, c := range n.Children {
		c2, err := c.contextualize(args)
		if err != nil {
			return err
		}
		n.Children[i] = c2
	}
	return nil
}

func (n *nodeT) canHaveSelect() bool {
	if n.Parent == nil || n.Parent.Token.Symbol != pathToken {
		return false
	}
	return n.Token.any(selectNeededFields...)
}

func (n *nodeT) doesNotNeedSelect() bool {
	return n.Token.any(selectNotNeededFields...)
}

// ------------------------------------------------------------
// CONTEXTUALIZE-ARGS

type contextualizeArgs struct {
	selectctx *selectContext
}

// ------------------------------------------------------------
// SELECT-CONTEXT

type selectContext struct {
	needed    bool
	notneeded bool
}

// ------------------------------------------------------------
// CONST and VAR

var (
	selectNeededFields    = []symbol{eqlToken, neqToken}
	selectNotNeededFields = []symbol{assignToken}
)

// ------------------------------------------------------------
// BOILERPLATE

func contextualizerFakeFmt() {
	fmt.Println()
}
