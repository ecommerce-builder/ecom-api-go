package nestedset

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNestedSet(t *testing.T) {
	nodes := map[string]*Node{
		"a": NewNode("a", "Category A"),
		"b": NewNode("b", "Category B"),
		"c": NewNode("c", "Category C"),
		"d": NewNode("d", "Category D"),
		"e": NewNode("e", "Category E"),
		"f": NewNode("f", "Category F"),
		"g": NewNode("g", "Category G"),
		"h": NewNode("h", "Category H"),
		"i": NewNode("i", "Category I"),
		"j": NewNode("j", "Category J"),
		"k": NewNode("k", "Category K"),
		"l": NewNode("l", "Category L"),
		"m": NewNode("m", "Category M"),
		"n": NewNode("n", "Category N"),
	}

	// Build a tree
	nodes["a"].AddChild(nodes["b"])
	nodes["a"].AddChild(nodes["c"])
	nodes["a"].AddChild(nodes["d"])
	nodes["b"].AddChild(nodes["e"])
	nodes["c"].AddChild(nodes["f"])
	nodes["c"].AddChild(nodes["g"])
	nodes["d"].AddChild(nodes["h"])
	nodes["f"].AddChild(nodes["i"])
	nodes["f"].AddChild(nodes["j"])
	nodes["h"].AddChild(nodes["k"])
	nodes["h"].AddChild(nodes["l"])
	nodes["j"].AddChild(nodes["m"])
	nodes["j"].AddChild(nodes["n"])

	nodes["a"].GenerateNestedSet(1, 0, "")

	buf := new(bytes.Buffer)
	nodes["a"].PreorderTraversalPrint(buf)

	expected := map[string]struct {
		segment string
		path    string
		name    string
		lft     int
		rgt     int
		depth   int
	}{
		"a: {segment: "a", path: "/a", name: "Category A", lft: 1, rgt: 28, depth: 0},
		"b": {segment: "b", path: "/a/b", name: "Category B", lft: 2, rgt: 5, depth: 1},
		"e": {segment: "e", path: "/a/b/e", name: "Category E", lft: 3, rgt: 4, depth: 2},
		"c": {segment: "c", path: "/a/c", name: "Category C", lft: 6, rgt: 19, depth: 1},
		"f": {segment: "f", path: "/a/c/f", name: "Category F", lft: 7, rgt: 16, depth: 2},
		"i": {segment: "i", path: "/a/c/f/i", name: "Category I", lft: 8, rgt: 9, depth: 3},
		"j": {segment: "j", path: "/a/c/f/j", name: "Category J", lft: 10, rgt: 15, depth: 3},
		"m": {segment: "m", path: "/a/c/f/j/m", name: "Category M", lft: 11, rgt: 12, depth: 4},
		"n": {segment: "n", path: "/a/c/f/j/n", name: "Category N", lft: 13, rgt: 14, depth: 4},
		"g": {segment: "g", path: "/a/c/g", name: "Category G", lft: 17, rgt: 18, depth: 2},
		"d": {segment: "d", path: "/a/d", name:"Category D", lft: 20, rgt: 27, depth: 1},
		"h": {segment: "h", path: "/a/d/h", name:"Category H",  lft: 21, rgt: 26, depth: 2},
		"k": {segment: "k", path: "/a/d/h/k", name:"Category K", lft: 22, rgt: 23, depth: 3},
		"l": {segment: "l", path: "/a/d/h/l", name:"Category L", lft: 24, rgt: 25, depth: 3},
	}

	for k, n := range nodes {
		assert.Equal(t, expected[k].segment, n.segment, fmt.Sprintf("Node %q segment should be %q; got %q", k, expected[k].segment, n.segment))
		assert.Equal(t, expected[k].path, n.path, fmt.Sprintf("Node %q path should be %q; got %q", k, expected[k].path, n.path))
		assert.Equal(t, expected[k].name, n.name, fmt.Sprintf("Node %q name should be %q; got %q", k, expected[k].name, n.name))
		assert.Equal(t, expected[k].lft, n.lft, fmt.Sprintf("Node %q lft should be %d; got %d", k, expected[k].lft, n.lft))
		assert.Equal(t, expected[k].rgt, n.rgt, fmt.Sprintf("Node %q rgt should be %d; got %d", k, expected[k].rgt, n.rgt))
		assert.Equal(t, expected[k].depth, n.depth, fmt.Sprintf("Node %q depth should be %d; got %d", k, expected[k].depth, n.depth))
	}
	t.Logf("\n%s\n", buf.String())
}

func TestBuildTree(t *testing.T) {
	nodes := []*NestedSetNode{
		{Segment: "a", Path: "/a", Name: "Category A", Lft: 1, Rgt: 28, Depth: 0},
		{Segment: "b", Path: "/a/b", Name: "Category B", Lft: 2, Rgt: 5, Depth: 1},
		{Segment: "e", Path: "/a/b/e", Name: "Category E", Lft: 3, Rgt: 4, Depth: 2},
		{Segment: "c", Path: "/a/c", Name: "Category C", Lft: 6, Rgt: 19, Depth: 1},
		{Segment: "f", Path: "/a/c/f", Name: "Category F", Lft: 7, Rgt: 16, Depth: 2},
		{Segment: "i", Path: "/a/c/f/i", Name: "Category I", Lft: 8, Rgt: 9, Depth: 3},
		{Segment: "j", Path: "/a/c/f/j", Name: "Category J", Lft: 10, Rgt: 15, Depth: 3},
		{Segment: "m", Path: "/a/c/f/j/m", Name: "Category M", Lft: 11, Rgt: 12, Depth: 4},
		{Segment: "n", Path: "/a/c/f/j/n", Name: "Category N", Lft: 13, Rgt: 14, Depth: 4},
		{Segment: "g", Path: "/a/c/g", Name: "Category G", Lft: 17, Rgt: 18, Depth: 2},
		{Segment: "d", Path: "/a/d", Name: "Category D", Lft: 20, Rgt: 27, Depth: 1},
		{Segment: "h", Path: "/a/d/h", Name: "Category H", Lft: 21, Rgt: 26, Depth: 2},
		{Segment: "k", Path: "/a/d/h/k", Name: "Category K", Lft: 22, Rgt: 23, Depth: 3},
		{Segment: "l", Path: "/a/d/h/l", Name: "Category L", Lft: 24, Rgt: 25, Depth: 3},
	}

	root := BuildTree(nodes)

	buf := new(bytes.Buffer)
	root.PreorderTraversalPrint(buf)
	t.Logf("\n%s\n", buf.String())
}
