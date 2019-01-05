package sqi

import (
	"errors"
	"strconv"
)

// parse() converts tokens into an AST.
func parse(tokens []token_t) (AstNode, error) {
	tree, err := make_tree(tokens)
	if err != nil {
		return nil, err
	}
	if tree == nil || len(tree.children) != 1 {
		return nil, errors.New("sqi: parse created empty tree")
	}
	return tree.children[0].asAst()
}

// make_tree() creates the tree structure. It is solely concerned about
// the structure -- i.e. it cares that a token is a binary, but does not
// care what type of binary it is.
func make_tree(tokens []token_t) (*tree_node, error) {
	root := &tree_node{}
	cur := root
	for _, t := range tokens {
		tn := &tree_node{t: t}
		if t.isBinary() {
			if cur.parent == nil {
				return nil, errors.New("sqi: parse has binary with no parent")
			}
			err := cur.parent.replaceChild(cur, tn)
			if err != nil {
				return nil, err
			}
			cur = tn
		} else {
			cur.addChild(tn)
			cur = tn
		}
	}
	return root, nil
}

// ----------------------------------------
// TREE-NODE

// tree_node is used to assemble the tokens into a tree.
type tree_node struct {
	parent   *tree_node
	children []*tree_node
	t        token_t
}

func (n *tree_node) addChild(child *tree_node) {
	child.parent = n
	n.children = append(n.children, child)
}

func (n *tree_node) replaceChild(oldchild, newchild *tree_node) error {
	for i, c := range n.children {
		if c == oldchild {
			newchild.addChild(oldchild)
			newchild.parent = n
			n.children[i] = newchild
			return nil
		}
	}
	return errors.New("sqi: parse missing child")
}

func (n *tree_node) asAst() (AstNode, error) {
	switch n.t.tok {
	case eql_token:
		lhs, rhs, err := n.makeBinary()
		if err != nil {
			return nil, err
		}
		return &binaryOpNode{Op: eql_token, Lhs: lhs, Rhs: rhs}, nil
	case path_token:
		lhs, rhs, err := n.makeBinary()
		if err != nil {
			return nil, err
		}
		return &pathNode{Lhs: lhs, Rhs: rhs}, nil
	case string_token:
		if len(n.children) != 0 {
			return nil, errors.New("sqi: parse field has wrong number of children: " + strconv.Itoa(len(n.children)))
		}
		// There are rules on strings -- based on context I can be either a field or string node
		ctx := n.getCtx()
		if ctx == string_tree_ctx {
			return &stringNode{Value: n.t.text}, nil
		}
		return &fieldNode{Field: n.t.text}, nil
	}
	return nil, errors.New("sqi: parse on unknown token: " + strconv.Itoa(int(n.t.tok)))
}

func (n *tree_node) makeBinary() (AstNode, AstNode, error) {
	if len(n.children) != 2 {
		return nil, nil, errors.New("sqi: parse path has wrong number of children: " + strconv.Itoa(len(n.children)))
	}
	lhs, err := n.children[0].asAst()
	if err != nil {
		return nil, nil, err
	}
	rhs, err := n.children[1].asAst()
	if err != nil {
		return nil, nil, err
	}
	return lhs, rhs, nil
}

// getCtx() answers the context for this token, based on its position in the syntax tree.
func (n *tree_node) getCtx() tree_ctx {
	// Not sure how this will evolve, but currently only strings that are rhs of binaries have meaning.
	if n.t.tok != string_token || n.parent == nil || len(n.parent.children) != 2 {
		return empty_tree_ctx
	}
	if !n.parent.isToken(string_capable_rhs...) {
		return empty_tree_ctx
	}
	if n.parent.children[1] == n {
		return string_tree_ctx
	}
	return empty_tree_ctx
}

func (n *tree_node) isToken(tokens ...Token) bool {
	for _, t := range tokens {
		if n.t.tok == t {
			return true
		}
	}
	return false
}

// ----------------------------------------
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
