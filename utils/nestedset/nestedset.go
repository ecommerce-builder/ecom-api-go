package nestedset

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"
)

// A NestedSetNode represents a single node in the nested set.
type NestedSetNode struct {
	ID       int
	Segment  string
	Path     string
	Name     string
	Lft      int
	Rgt      int
	Depth    int
	Created  time.Time
	Modified time.Time
}

// A Node represents a hierarchical tree structure.
type Node struct {
	Segment string `json:"segment"`
	path    string
	Name    string `json:"name"`
	lft     int
	rgt     int
	depth   int
	parent  *Node
	Nodes   []*Node `json:"categories"`
}

// NewNode creates a new Node.
func NewNode(segment, name string) *Node {
	return &Node{
		Segment: segment,
		path:    "",
		Name:    name,
		lft:     -1,
		rgt:     -1,
		depth:   -1,
		parent:  nil,
		Nodes:   make([]*Node, 0),
	}
}

// BuildTree builds a Tree hierarchy from a Nested Set.
func BuildTree(nestedset []*NestedSetNode) *Node {
	context := &Node{
		parent:  nil,
		Segment: nestedset[0].Segment,
		path:    nestedset[0].Path,
		Name:    nestedset[0].Name,
		Nodes:   make([]*Node, 0),
		lft:     nestedset[0].Lft,
		rgt:     nestedset[0].Rgt,
	}
	for i := 1; i < len(nestedset); i++ {
		cur := nestedset[i]
		n := &Node{
			parent:  context,
			Segment: cur.Segment,
			path:    cur.Path,
			Name:    cur.Name,
			Nodes:   make([]*Node, 0),
			lft:     cur.Lft,
			rgt:     cur.Rgt,
		}
		context.AddChild(n)

		// Is Leaf node and the context needs moving.
		if cur.Lft == cur.Rgt-1 {
			if cur.Rgt == context.rgt-1 {
				context = moveContext(context)
			}
		} else {
			context = n
		}
	}
	return context
}

// FindNodeByPath traverses the tree looking for a Node with a matching path.
func (n *Node) FindNodeByPath(path string) *Node {
	// example without leading forwardslash 'a/c/f/j/n'
	segments := strings.Split(path, "/")
	if len(segments) == 0 {
		return nil
	}
	if segments[0] != n.Segment {
		return nil
	}
	context := n
	for i := 1; i < len(segments); i++ {
		context = context.findNode(segments[i])
		if context == nil {
			return nil
		}
	}
	return context
}

func (n *Node) hasChildren() bool {
	return len(n.Nodes) > 0
}

// findCategory Looks through the child nodes for a matching segment.
// Runs in O(n) time. Example segment 'shoes', 'widgets' etc.
func (n *Node) findNode(segment string) *Node {
	if !n.hasChildren() {
		return nil
	}
	for _, node := range n.Nodes {
		if node.Segment == segment {
			return node
		}
	}
	return nil
}

func moveContext(context *Node) *Node {
	if context.parent == nil {
		return context
	}
	prev := context
	context = context.parent

	for prev.rgt == context.rgt-1 && context.parent != nil {
		prev = context
		context = context.parent
	}
	return context
}

// GenerateNestedSet performs a pre-order tree traversal wiring the tree.
func (n *Node) GenerateNestedSet(lft, depth int, path string) int {
	rgt := lft + 1
	for _, i := range n.Nodes {
		if path == "" {
			rgt = i.GenerateNestedSet(rgt, depth+1, n.Segment)
		} else {
			rgt = i.GenerateNestedSet(rgt, depth+1, path+"/"+n.Segment)
		}
	}
	n.lft = lft
	n.rgt = rgt
	n.depth = depth

	if path == "" {
		n.path = n.Segment
	} else {
		n.path = path + "/" + n.Segment
	}
	return rgt + 1
}

// PreorderTraversalPrint provides a depth first search printout of each node
// in the hierarchy.
func (n *Node) PreorderTraversalPrint(w io.Writer) {
	tw := new(tabwriter.Writer).Init(w, 0, 8, 2, ' ', 0)
	n.preorderTraversalWrite(tw)
	tw.Flush()
}

func (n *Node) preorderTraversalWrite(w io.Writer) {
	fmt.Fprintf(w, "segment: %s\t path: %q\tname: %q\tlft: %d\t rgt: %d\t depth %d\n", n.Segment, n.path, n.Name, n.lft, n.rgt, n.depth)

	for _, i := range n.Nodes {
		i.preorderTraversalWrite(w)
	}
}

// NestedSet uses preorder traversal of the tree to return a
// slice of NestedSetNodes.
func (n *Node) NestedSet(ns *[]*NestedSetNode) {
	n.preorderTraversalNS(ns)
}

func (n *Node) preorderTraversalNS(ns *[]*NestedSetNode) {
	nsn := &NestedSetNode{
		Segment: n.Segment,
		Path:    n.path,
		Name:    n.Name,
		Lft:     n.lft,
		Rgt:     n.rgt,
		Depth:   n.depth,
	}
	*ns = append(*ns, nsn)
	for _, i := range n.Nodes {
		i.preorderTraversalNS(ns)
	}
}

// AddChild attaches a node to its parent node.
func (n *Node) AddChild(c *Node) {
	c.parent = n
	n.Nodes = append(n.Nodes, c)
}

// IsRoot returns true for the root node only.
func (n *Node) IsRoot() bool {
	return n.parent == nil
}

// IsLeaf return true if the node is a leaf node.
func (n *Node) IsLeaf() bool {
	return len(n.Nodes) == 0
}
