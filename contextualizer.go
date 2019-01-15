package sqi

import (
	"fmt"
)

// contextualize() insert rules specific to evaluating our rules.
func contextualize(tree *node_t) (*node_t, error) {
	if tree == nil {
		return tree, nil
	}
	return tree.contextualize()
}

// ------------------------------------------------------------
// NODE_T (contextualizing)

func (n *node_t) contextualize() (*node_t, error) {
	n2, err := n.performContextualize()
	if err != nil {
		return nil, err
	}
	for i, c := range n.Children {
		c2, err := c.contextualize()
		if err != nil {
			return nil, err
		}
		n.Children[i] = c2
	}
	return n2, nil
}

func (n *node_t) performContextualize() (*node_t, error) {
	// Strings on the right of certain tokens will be strings,
	// but all other strings are field selectors.
	if n.isFieldContext() {
		n.Token = token_map[field_token]
	}
	return n, nil
}

// isFieldContext() answers true if this string should be a field.
func (n *node_t) isFieldContext() bool {
	if n.Token.Symbol != string_token {
		return false
	}
	// Strings are the right-hand size of a string-capabale binary.
	if n.Parent != nil && len(n.Parent.Children) == 2 && n.Parent.isToken(string_capable_rhs...) && n.Parent.Children[1] == n {
		return false
	}
	return true
}

// ------------------------------------------------------------
// CONST and VAR

var (
	string_capable_rhs = []symbol{assign_token, eql_token, neq_token}
)

// ------------------------------------------------------------
// BOILERPLATE

func contextualizerFakeFmt() {
	fmt.Println()
}
