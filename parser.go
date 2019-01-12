package sqi

import (
	"errors"
	"strconv"
	"strings"
)

// parse() converts tokens into an AST.
func parse(tokens []*token_t) (AstNode, error) {
	tree, err := make_tree(tokens)
	if err != nil {
		return nil, err
	}
	if tree == nil || len(tree.Children) != 1 {
		return nil, errors.New("sqi: parse created empty tree")
	}
	// fmt.Println("tree:", toJson(tree))
	return tree.Children[0].asAst()
}

// make_tree() creates the tree structure. It is solely concerned about
// the structure -- i.e. it cares that a token is a binary, but does not
// care what type of binary it is.
func make_tree(tokens []*token_t) (*token_t, error) {
	root := &token_t{}
	cur := root
	var paren []*token_t
	for _, t := range tokens {
		if t.isBinary() {
			if cur.Parent == nil {
				return nil, errors.New("sqi: parse has binary with no parent")
			}
			err := cur.Parent.replaceChild(cur, t)
			if err != nil {
				return nil, err
			}
			cur = t
			// Special insertion rules: Conditionals can only exist inside condition nodes
			if t.needsCondition() {
				err := cur.setCondition()
				if err != nil {
					return nil, err
				}
			}
		} else if t.isOpenParen() {
			cur.addChild(t)
			paren = append(paren, t)
			cur = t
		} else if t.isCloseParen() {
			if len(paren) < 1 {
				return nil, errors.New("sqi: mismatched parentheses")
			}
			cur = paren[len(paren)-1]
			paren = paren[:len(paren)-1]
		} else {
			cur.addChild(t)
			cur = t
		}
	}
	return root, nil
}

/*
func make_tree(tokens []token_t) (*tree_node, error) {
	root := &tree_node{}
	cur := root
	var paren []*tree_node
	for _, t := range tokens {
		tn := &tree_node{T: t}
		if t.isBinary() {
			if cur.Parent == nil {
				return nil, errors.New("sqi: parse has binary with no parent")
			}
			err := cur.Parent.replaceChild(cur, tn)
			if err != nil {
				return nil, err
			}
			cur = tn
			// Special insertion rules: Conditionals can only exist inside condition nodes
			if t.needsCondition() {
				err := cur.setCondition()
				if err != nil {
					return nil, err
				}
			}
		} else if t.isOpenParen() {
			cur.addChild(tn)
			paren = append(paren, tn)
			cur = tn
		} else if t.isCloseParen() {
			if len(paren) < 1 {
				return nil, errors.New("sqi: mismatched parentheses")
			}
			cur = paren[len(paren)-1]
			paren = paren[:len(paren)-1]
		} else {
			cur.addChild(tn)
			cur = tn
		}
	}
	return root, nil
}
*/

// ------------------------------------------------------------
// PARSE-NODE
// Additional behaviour on tokens so they can be assembled into a tree.

func (n *token_t) addChild(child *token_t) {
	child.Parent = n
	n.Children = append(n.Children, child)
}

func (n *token_t) replaceChild(oldchild, newchild *token_t) error {
	for i, c := range n.Children {
		if c == oldchild {
			newchild.addChild(oldchild)
			newchild.Parent = n
			n.Children[i] = newchild
			return nil
		}
	}
	return errors.New("sqi: parse missing child")
}

// setCondition() finds the proper node to insert a condtion node. This is used
// by boolean conditions: Every subgraph that needs to evaluate to true/false
// must be wrapped in a condition. Currently that means any comparison booleans,
// and the conditionals that can contain them.
func (n *token_t) setCondition() error {
	if n.Parent == nil || !n.Parent.canHaveCondition() {
		return n.setInsert(condition_token)
	}
	return n.Parent.setCondition()
}

// setInsert() sets the insert value for this node. A node can only have a single
// insert type set -- any change will result in an error.
func (n *token_t) setInsert(t Token) error {
	if n.Insert == illegal_token || n.Insert == t {
		n.Insert = t
		return nil
	}
	return newMismatchError("tree insert " + strconv.Itoa(int(n.Insert)) + " and " + strconv.Itoa(int(t)))
}

// asAst() converts this node into an AST node, including special rules like the insert.
func (n *token_t) asAst() (AstNode, error) {
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
func (n *token_t) nodeAsAst() (AstNode, error) {
	switch n.Tok {
	case eql_token, neq_token, and_token, or_token:
		lhs, rhs, err := n.makeBinary()
		if err != nil {
			return nil, err
		}
		return &binaryNode{Op: n.Tok, Lhs: lhs, Rhs: rhs}, nil
	case float_token:
		if len(n.Children) != 0 {
			return nil, errors.New("sqi: parse float has wrong number of children: " + strconv.Itoa(len(n.Children)))
		}
		f64, err := strconv.ParseFloat(n.Text, 64)
		if err != nil {
			return nil, err
		}
		return &constantNode{Value: f64}, nil
	case int_token:
		if len(n.Children) != 0 {
			return nil, errors.New("sqi: parse int has wrong number of children: " + strconv.Itoa(len(n.Children)))
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
			return nil, errors.New("sqi: parse string has wrong number of children: " + strconv.Itoa(len(n.Children)))
		}
		// Unwrap quoted text, which has served its purpose of allowing special characters.
		text := strings.Trim(n.Text, `"`)
		// There are rules on strings -- based on context I can be either a field or string node
		ctx := n.getCtx()
		if ctx == string_tree_ctx {
			return &constantNode{Value: text}, nil
		}
		return &fieldNode{Field: text}, nil
	}
	return nil, errors.New("sqi: parse on unknown token: " + strconv.Itoa(int(n.Tok)))
}

func (n *token_t) makeBinary() (AstNode, AstNode, error) {
	if len(n.Children) != 2 {
		return nil, nil, errors.New("sqi: parse binary has wrong number of children: " + strconv.Itoa(len(n.Children)))
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

func (n *token_t) makeUnary() (AstNode, error) {
	if len(n.Children) != 1 {
		return nil, errors.New("sqi: parse unary has wrong number of children: " + strconv.Itoa(len(n.Children)))
	}
	return n.Children[0].asAst()
}

// getCtx() answers the context for this token, based on its position in the syntax tree.
func (n *token_t) getCtx() tree_ctx {
	// Not sure how this will evolve, but currently only strings that are rhs of binaries have meaning.
	if n.Tok != string_token || n.Parent == nil || len(n.Parent.Children) != 2 {
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

func (n *token_t) isToken(tokens ...Token) bool {
	for _, t := range tokens {
		if n.Tok == t {
			return true
		}
	}
	return false
}

// ------------------------------------------------------------
// TREE-NODE
/*
// tree_node is used to assemble the tokens into a tree.
type tree_node struct {
	Parent   *tree_node `json:"-"`
	Children []*tree_node
	T        token_t
	Insert   Token // A command to replace this node with a unary that contains it.
}

func (n *tree_node) addChild(child *tree_node) {
	child.Parent = n
	n.Children = append(n.Children, child)
}

func (n *tree_node) replaceChild(oldchild, newchild *tree_node) error {
	for i, c := range n.Children {
		if c == oldchild {
			newchild.addChild(oldchild)
			newchild.Parent = n
			n.Children[i] = newchild
			return nil
		}
	}
	return errors.New("sqi: parse missing child")
}

// setCondition() finds the proper node to insert a condtion node. This is used
// by boolean conditions: Every subgraph that needs to evaluate to true/false
// must be wrapped in a condition. Currently that means any comparison booleans,
// and the conditionals that can contain them.
func (n *tree_node) setCondition() error {
	if n.Parent == nil || !n.Parent.T.canHaveCondition() {
		return n.setInsert(condition_token)
	}
	return n.Parent.setCondition()
}

// setInsert() sets the insert value for this node. A node can only have a single
// insert type set -- any change will result in an error.
func (n *tree_node) setInsert(t Token) error {
	if n.Insert == illegal_token || n.Insert == t {
		n.Insert = t
		return nil
	}
	return newMismatchError("tree insert " + strconv.Itoa(int(n.Insert)) + " and " + strconv.Itoa(int(t)))
}

// asAst() converts this node into an AST node, including special rules like the insert.
func (n *tree_node) asAst() (AstNode, error) {
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
func (n *tree_node) nodeAsAst() (AstNode, error) {
	switch n.T.Tok {
	case eql_token, neq_token, and_token, or_token:
		lhs, rhs, err := n.makeBinary()
		if err != nil {
			return nil, err
		}
		return &binaryNode{Op: n.T.Tok, Lhs: lhs, Rhs: rhs}, nil
	case float_token:
		if len(n.Children) != 0 {
			return nil, errors.New("sqi: parse float has wrong number of children: " + strconv.Itoa(len(n.Children)))
		}
		f64, err := strconv.ParseFloat(n.T.Text, 64)
		if err != nil {
			return nil, err
		}
		return &constantNode{Value: f64}, nil
	case int_token:
		if len(n.Children) != 0 {
			return nil, errors.New("sqi: parse int has wrong number of children: " + strconv.Itoa(len(n.Children)))
		}
		i, err := strconv.ParseInt(n.T.Text, 10, 32)
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
			return nil, errors.New("sqi: parse string has wrong number of children: " + strconv.Itoa(len(n.Children)))
		}
		// Unwrap quoted text, which has served its purpose of allowing special characters.
		text := strings.Trim(n.T.Text, `"`)
		// There are rules on strings -- based on context I can be either a field or string node
		ctx := n.getCtx()
		if ctx == string_tree_ctx {
			return &constantNode{Value: text}, nil
		}
		return &fieldNode{Field: text}, nil
	}
	return nil, errors.New("sqi: parse on unknown token: " + strconv.Itoa(int(n.T.Tok)))
}

func (n *tree_node) makeBinary() (AstNode, AstNode, error) {
	if len(n.Children) != 2 {
		return nil, nil, errors.New("sqi: parse binary has wrong number of children: " + strconv.Itoa(len(n.Children)))
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

func (n *tree_node) makeUnary() (AstNode, error) {
	if len(n.Children) != 1 {
		return nil, errors.New("sqi: parse unary has wrong number of children: " + strconv.Itoa(len(n.Children)))
	}
	return n.Children[0].asAst()
}

// getCtx() answers the context for this token, based on its position in the syntax tree.
func (n *tree_node) getCtx() tree_ctx {
	// Not sure how this will evolve, but currently only strings that are rhs of binaries have meaning.
	if n.T.Tok != string_token || n.Parent == nil || len(n.Parent.Children) != 2 {
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

func (n *tree_node) isToken(tokens ...Token) bool {
	for _, t := range tokens {
		if n.T.Tok == t {
			return true
		}
	}
	return false
}
*/

// ------------------------------------------------------------
// CONST and VAR

type tree_ctx int

const (
	// Special tokens
	empty_tree_ctx tree_ctx = iota

	field_tree_ctx
	string_tree_ctx
)

var (
	string_capable_rhs = []Token{assign_token, eql_token, neq_token}
)
