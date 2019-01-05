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
	case string_token:
		if len(n.children) != 0 {
			return nil, errors.New("sqi: parse field has wrong number of children: " + strconv.Itoa(len(n.children)))
		}
		return &fieldNode{Field: n.t.text}, nil
	case path_token:
		lhs, rhs, err := n.makeBinary()
		if err != nil {
			return nil, err
		}
		return &pathNode{Lhs: lhs, Rhs: rhs}, nil
		/*
			case eql_token:
				lhs, rhs, err := n.makeBinary()
				if err != nil {
					return nil, err
				}
				return &binaryOpNode{Op: `==`, Lhs: lhs, Rhs: rhs}, nil
		*/
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
