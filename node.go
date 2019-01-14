package sqi

import (
	"strconv"
	"strings"
)

// ------------------------------------------------------------
// NODE_T

func newToken(s symbol, text string) *node_t {
	token, ok := token_map[s]
	if !ok {
		token = token_map[illegal_token]
	}
	return &node_t{Token: token, Text: text}
}

// node_t serves two purposes: it is a token generated by the
// lexer, and it is a node in the tree constructed by the parser.
type node_t struct {
	// Lexing
	Token *token_t
	Text  string

	// Parsing
	Parent   *node_t `json:"-"`
	Left     *node_t
	Right    *node_t
	Children []*node_t

	// Contextualizing
	Insert symbol // A command to replace this node with a unary that contains it.
}

// reclassify() converts this token into one of the defined
// keywords, if appropriate. Ideally this is done directly
// in the scanning stage, but I'm not sure how to get the
// scanner to do that.
func (n *node_t) reclassify() *node_t {
	if n.Token.Symbol != string_token {
		return n
	}
	if found, ok := keyword_map[n.Text]; ok {
		return newToken(found.Symbol, n.Text)
	}
	return n
}

// ------------------------------------------------------------
// PARSE-NODE
// Additional behaviour on tokens so they can be assembled into a tree.

// setCondition() finds the proper node to insert a condtion node. This is used
// by boolean conditions: Every subgraph that needs to evaluate to true/false
// must be wrapped in a condition. Currently that means any comparison booleans,
// and the conditionals that can contain them.
func (n *node_t) setCondition() error {
	if n.Parent == nil || !n.Parent.canHaveCondition() {
		return n.setInsert(condition_token)
	}
	return n.Parent.setCondition()
}

// setInsert() sets the insert value for this node. A node can only have a single
// insert type set -- any change will result in an error.
func (n *node_t) setInsert(t symbol) error {
	if n.Insert == illegal_token || n.Insert == t {
		n.Insert = t
		return nil
	}
	return newMismatchError("tree insert " + strconv.Itoa(int(n.Insert)) + " and " + strconv.Itoa(int(t)))
}

// asAst() converts this node into an AST node, including special rules like the insert.
func (n *node_t) asAst() (AstNode, error) {
	node, err := n.nodeAsAst()
	if err != nil {
		return nil, err
	}
	if n.Insert == condition_token {
		node = &conditionNode{n.Insert, node}
	}
	return node, nil
}

// nodeAsAst() returns the AST node for this tree node.
func (n *node_t) nodeAsAst() (AstNode, error) {
	//	fmt.Println("ast", n.Text)
	switch n.Token.Symbol {
	case eql_token, neq_token, and_token, or_token:
		lhs, rhs, err := n.makeBinary()
		if err != nil {
			return nil, err
		}
		return &binaryNode{Op: n.Token.Symbol, Lhs: lhs, Rhs: rhs}, nil
	case float_token:
		if len(n.Children) != 0 {
			return nil, newParseError("float has wrong number of children: " + strconv.Itoa(len(n.Children)))
		}
		f64, err := strconv.ParseFloat(n.Text, 64)
		if err != nil {
			return nil, err
		}
		return &constantNode{Value: f64}, nil
	case int_token:
		if len(n.Children) != 0 {
			return nil, newParseError("int has wrong number of children: " + strconv.Itoa(len(n.Children)))
		}
		i, err := strconv.ParseInt(n.Text, 10, 32)
		if err != nil {
			return nil, err
		}
		return &constantNode{Value: int(i)}, nil
	case open_token:
		child, err := n.makeUnary()
		if err != nil {
			return nil, err
		}
		return &unaryNode{Op: open_token, Child: child}, nil
	case path_token:
		lhs, rhs, err := n.makeBinary()
		if err != nil {
			return nil, err
		}
		return &pathNode{Lhs: lhs, Rhs: rhs}, nil
	case string_token:
		if len(n.Children) != 0 {
			return nil, newParseError("string has wrong number of children: " + strconv.Itoa(len(n.Children)))
		}
		// Unwrap quoted text, which has served its purpose of allowing special characters.
		text := strings.Trim(n.Text, `"`)

		// There are rules on strings -- based on context I can be either a field or string node
		return &constantNode{Value: text}, nil
		/*
			ctx := n.getCtx()
			if ctx == string_tree_ctx {
				return &constantNode{Value: text}, nil
			}
			return &fieldNode{Field: text}, nil
		*/
	}
	return nil, newParseError("on unknown token: " + strconv.Itoa(int(n.Token.Symbol)))
}

/*
// asAst() converts this node into an AST node, including special rules like the insert.
func (n *node_t) asAst_orig() (AstNode, error) {
	node, err := n.nodeAsAst_orig()
	if err != nil {
		return nil, err
	}
	if n.Insert == condition_token {
		node = &conditionNode{n.Insert, node}
	}
	return node, nil
}
*/

func (n *node_t) makeBinary() (AstNode, AstNode, error) {
	if len(n.Children) != 2 {
		return nil, nil, newParseError("binary has wrong number of children: " + strconv.Itoa(len(n.Children)))
	}
	lhs, err := n.Children[0].asAst()
	if err != nil {
		return nil, nil, err
	}
	rhs, err := n.Children[1].asAst()
	if err != nil {
		return nil, nil, err
	}
	return lhs, rhs, nil
}

func (n *node_t) makeUnary() (AstNode, error) {
	if len(n.Children) != 1 {
		return nil, newParseError("unary has wrong number of children: " + strconv.Itoa(len(n.Children)))
	}
	return n.Children[0].asAst()
}

// getCtx() answers the context for this token, based on its position in the syntax tree.
func (n *node_t) getCtx() tree_ctx {
	// Not sure how this will evolve, but currently only strings that are rhs of binaries have meaning.
	if n.Token.Symbol != string_token || n.Parent == nil || n.Parent.Left == nil || n.Parent.Right == nil {
		return empty_tree_ctx
	}
	if !n.Parent.isToken(string_capable_rhs...) {
		return empty_tree_ctx
	}
	if n.Parent.Right == n {
		return string_tree_ctx
	}
	return empty_tree_ctx
}

func (n *node_t) getCtx_orig() tree_ctx {
	// Not sure how this will evolve, but currently only strings that are rhs of binaries have meaning.
	if n.Token.Symbol != string_token || n.Parent == nil || len(n.Parent.Children) != 2 {
		return empty_tree_ctx
	}
	if !n.Parent.isToken(string_capable_rhs...) {
		return empty_tree_ctx
	}
	if n.Parent.Children[1] == n {
		return string_tree_ctx
	}
	return empty_tree_ctx
}

func (n *node_t) isToken(tokens ...symbol) bool {
	for _, t := range tokens {
		if n.Token.Symbol == t {
			return true
		}
	}
	return false
}

func (n *node_t) needsCondition() bool {
	return n.Token.Symbol == eql_token || n.Token.Symbol == neq_token
}

func (n *node_t) canHaveCondition() bool {
	return n.needsCondition() || (n.Token.Symbol > start_conditional && n.Token.Symbol < end_conditional)
}
