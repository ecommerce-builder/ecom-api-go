package nestedset

import (
	"fmt"
	"io"
	"text/tabwriter"
	"time"
)

type NestedSetNode struct {
	ID       int       `json:"id"`
	Parent   *int      `json:"parent"`
	Segment  string    `json:"segment"`
	Path     string    `json:"path"`
	Name     string    `json:"name"`
	Lft      int       `json:"lft"`
	Rgt      int       `json:"rgt"`
	Depth    int       `json:"depth"`
	Created  time.Time `json:"created"`
	Modified time.Time `json:"modified"`
}

type Node struct {
	segment string
	path    string
	name    string
	lft     int
	rgt     int
	depth   int
	parent  *Node
	nodes   []*Node
}

func NewNode(segment, name string) *Node {
	return &Node{
		segment: segment,
		path:    "",
		name:    name,
		lft:     -1,
		rgt:     -1,
		depth:   -1,
		parent:  nil,
		nodes:   make([]*Node, 0),
	}
}

func GetNestedSet() []*NestedSetNode {
	return nil
}

// BuildTree builds a Tree hierarchy from a Nested Set.
func BuildTree(nestedset []*NestedSetNode) *Node {
	context := &Node{
		parent:  nil,
		segment: nestedset[0].Segment,
		path:    nestedset[0].Path,
		name:    nestedset[0].Name,
		nodes:   make([]*Node, 0),
		lft:     nestedset[0].Lft,
		rgt:     nestedset[0].Rgt,
	}

	for i := 1; i < len(nestedset); i++ {
		cur := nestedset[i]
		n := &Node{
			parent:  context,
			segment: cur.Segment,
			path:    cur.Path,
			name:    cur.Name,
			nodes:   make([]*Node, 0),
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

// Walk traverses a tree depth-first,
func (n *Node) GenerateNestedSet(lft, depth int, path string) int {
	rgt := lft + 1
	for _, i := range n.nodes {
		if path == "" {
			rgt = i.GenerateNestedSet(rgt, depth+1, n.segment)
		} else {
			rgt = i.GenerateNestedSet(rgt, depth+1, path+"/"+n.segment)
		}
	}
	n.lft = lft
	n.rgt = rgt
	n.depth = depth

	if path == "" {
		n.path = n.segment
	} else {
		n.path = path + "/" + n.segment
	}
	return rgt + 1
}

func (n *Node) PreorderTraversalPrint(w io.Writer) {
	tw := new(tabwriter.Writer).Init(w, 0, 8, 2, ' ', 0)
	n.preorderTraversalWrite(tw)
	tw.Flush()
}

func (n *Node) preorderTraversalWrite(w io.Writer) {
	fmt.Fprintf(w, "segment: %s\t path: %q\tname: %q\tlft: %d\t rgt: %d\t depth %d\n", n.segment, n.path, n.name, n.lft, n.rgt, n.depth)

	for _, i := range n.nodes {
		i.preorderTraversalWrite(w)
	}
}

func (n *Node) AddChild(c *Node) {
	c.parent = n
	n.nodes = append(n.nodes, c)
}

func (n *Node) IsRoot() bool {
	return n.parent == nil
}

func (n *Node) IsLeaf() bool {
	return len(n.nodes) == 0
}
