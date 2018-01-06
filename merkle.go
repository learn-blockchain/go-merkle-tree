package merkle

// note: this code is not well though out or organized and was built for the purpose of excercise...
// furthermore, this tree does not follow the standard conventions detailed here: https://bitcoin.stackexchange.com/questions/30329/what-shape-of-merkle-tree-does-the-bitcoin-client-build

import (
	"crypto/sha256"
	"errors"
)

// Datum is a bytes array that represents an individual block to be hashed
type Datum []byte

// Data is the collection of Datum blocks
type Data []Datum

func (d Data) split() (Data, Data) {
	len := len(d)
	if len%2 != 0 {
		var newD Datum
		d = append(d, newD)
		len++
	}
	mid := len / 2

	return d[:mid], d[mid:]
}

func (d Data) buildLeadfs() Leafs {
	var leafs Leafs
	for idx := range d {
		leaf := Leaf{
			DatumLoc: &d[idx],
		}
		leaf.setHash()

		leafs = append(leafs, leaf)
	}

	return leafs
}

// Leaf is the edge of the Merkle Tree. i.e. the first hash after the data
type Leaf struct {
	DatumLoc  *Datum
	ParentLoc *Node
	Hash      []byte
}

func (l *Leaf) setHash() {
	hasher := sha256.New()
	hasher.Write(*l.DatumLoc)
	l.Hash = hasher.Sum(nil)
}

// Leafs is a slice of Leaf structs
type Leafs []Leaf

func (l Leafs) split() (Leafs, Leafs) {
	len := len(l)
	if len%2 != 0 {
		var newL Leaf
		l = append(l, newL)
		len++
	}
	mid := len / 2

	return l[:mid], l[mid:]
}

// Node is a non-leaf hash block of the Merkle Tree
type Node struct {
	LeftChildLoc  interface{}
	RightChildLoc interface{}
	ParentLoc     interface{}
	Hash          []byte
}

func (n *Node) setHash() error {
	var leftHash []byte
	var rightHash []byte
	var combinedBytes []byte

	left, ok := (n.LeftChildLoc).(*Node)
	if !ok {
		left2, ok2 := (n.LeftChildLoc).(*Leaf)
		if !ok2 {
			return errors.New("could not convert left node child to node or leaf pointer")
		}
		leftHash = left2.Hash
	} else {
		leftHash = left.Hash
	}

	right, ok := (n.RightChildLoc).(*Node)
	if !ok {
		right2, ok2 := (n.RightChildLoc).(*Leaf)
		if !ok2 {
			return errors.New("could not convert right node child to node or leaf pointer")
		}
		rightHash = right2.Hash
	} else {
		rightHash = right.Hash
	}

	combinedBytes = append(leftHash, rightHash...)

	hasher := sha256.New()
	hasher.Write(combinedBytes)
	n.Hash = hasher.Sum(nil)
	return nil
}

func buildTree(children interface{}) (*Node, error) {
	c, ok := children.(Leafs)
	if !ok {
		c2, ok2 := children.(Nodes)
		if !ok2 {
			return nil, errors.New("could not cast children as leafs or nodes")
		}

		switch len(c2) {
		case 0:
			return &Node{}, nil

		case 1:
			c2 = append(c2, c2...)
			fallthrough

		case 2:
			topNode := &Node{
				LeftChildLoc:  &c2[0],
				RightChildLoc: &c2[1],
			}
			topNode.setHash()

			c2[0].ParentLoc = topNode
			c2[1].ParentLoc = topNode

			return topNode, nil

		default:
			leftChildren, rightChildren := c2.split()

			leftNode, err := buildTree(leftChildren)
			if err != nil {
				return nil, err
			}
			rightNode, err := buildTree(rightChildren)
			if err != nil {
				return nil, err
			}

			topNode := &Node{
				LeftChildLoc:  leftNode,
				RightChildLoc: rightNode,
			}
			err = topNode.setHash()
			if err != nil {
				return nil, err
			}

			leftNode.ParentLoc = topNode
			rightNode.ParentLoc = topNode

			return topNode, nil
		}
	}

	switch len(c) {
	case 0:
		return &Node{}, nil

	case 1:
		c = append(c, c...)
		fallthrough

	case 2:
		topNode := &Node{
			LeftChildLoc:  &c[0],
			RightChildLoc: &c[1],
		}
		err := topNode.setHash()
		if err != nil {
			return nil, err
		}

		c[0].ParentLoc = topNode
		c[1].ParentLoc = topNode

		return topNode, nil

	default:
		leftChildren, rightChildren := c.split()
		leftNode, err := buildTree(leftChildren)
		if err != nil {
			return nil, err
		}
		rightNode, err := buildTree(rightChildren)
		if err != nil {
			return nil, err
		}

		topNode := &Node{
			LeftChildLoc:  leftNode,
			RightChildLoc: rightNode,
		}
		err = topNode.setHash()
		if err != nil {
			return nil, err
		}

		leftNode.ParentLoc = topNode
		rightNode.ParentLoc = topNode

		return topNode, nil
	}
}

// Nodes is a slice of Node structs
type Nodes []Node

func (n Nodes) split() (Nodes, Nodes) {
	len := len(n)
	if len%2 != 0 {
		var newN Node
		n = append(n, newN)
		len++
	}
	mid := len / 2

	return n[:mid], n[mid:]
}

// New builds a Merkle Tree from data and returns the Top Hash i.e. the top most Node of the tree.
func New(d Data) (*Node, error) {
	// 1. build the leafs
	leafs := d.buildLeadfs()
	if len(leafs) == 2 {
		topNode := Node{
			LeftChildLoc:  &leafs[0],
			RightChildLoc: &leafs[1],
		}
		err := topNode.setHash()

		return &topNode, err
	}
	leftLeafs, rightLeafs := leafs.split()

	// 2. build the left tree
	leftTop, err := buildTree(leftLeafs)
	if err != nil {
		return nil, err
	}

	// 3. build the right tree
	rightTop, err := buildTree(rightLeafs)
	if err != nil {
		return nil, err
	}

	// 4. combine the trees
	top := &Node{
		LeftChildLoc:  leftTop,
		RightChildLoc: rightTop,
	}
	err = top.setHash()
	if err != nil {
		return nil, err
	}

	leftTop.ParentLoc = &top
	rightTop.ParentLoc = &top

	return top, nil
}
