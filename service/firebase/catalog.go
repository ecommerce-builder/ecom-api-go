package firebase

import (
	"context"
	"strings"

	"bitbucket.org/andyfusniakteam/ecom-api-go/model/postgres"
	"github.com/pkg/errors"
)

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

// NestedSet uses preorder traversal of the tree to return a
// slice of NestedSetNodes.
func (n *Node) NestedSet(ns *[]*postgres.NestedSetNode) {
	n.preorderTraversalNS(ns)
}

func (n *Node) preorderTraversalNS(ns *[]*postgres.NestedSetNode) {
	nsn := &postgres.NestedSetNode{
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

// UpdateCatalog takes a root tree Node and converts it to a nested set
// representation before calling the model to persist the replacement
// catalog.
func (s *Service) UpdateCatalog(ctx context.Context, root *Node) error {
	root.GenerateNestedSet(1, 0, "")
	ns := make([]*postgres.NestedSetNode, 0, 128)
	root.NestedSet(&ns)
	if err := s.model.BatchCreateNestedSet(ctx, ns); err != nil {
		return errors.Wrap(err, "service: replace catalog")
	}
	return nil
}

// A Category represents an individual category in the catalog hierarchy.
type Category struct {
	Segment  string `json:"segment"`
	path     string
	Name     string `json:"name"`
	lft      int
	rgt      int
	parent   *Category
	Nodes    []*Category `json:"categories"`
	Products []*struct {
		SKU string `json:"sku"`
	} `json:"products"`
}

func (n *Category) addChild(c *Category) {
	n.Nodes = append(n.Nodes, c)
}

func (n *Category) hasChildren() bool {
	return len(n.Nodes) > 0
}

func (n *Category) findNode(segment string) *Category {
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

// IsLeaf return true if the node is a leaf node.
func (n *Category) IsLeaf() bool {
	return len(n.Nodes) == 0
}

func moveContext(context *Category) *Category {
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

// FindNodeByPath traverses the tree looking for a Node with a matching path.
func (n *Category) FindNodeByPath(path string) *Category {
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

// BuildTree builds a Tree hierarchy from a Nested Set.
func buildCategoryTree(nestedset []*postgres.NestedSetNode, cmap map[string][]string) *Category {
	context := &Category{
		Segment: nestedset[0].Segment,
		path:    nestedset[0].Path,
		Name:    nestedset[0].Name,
		parent:  nil,
		Nodes:   make([]*Category, 0),
		Products: make([]*struct {
			SKU string `json:"sku"`
		}, 0),
		lft: nestedset[0].Lft,
		rgt: nestedset[0].Rgt,
	}
	for i := 1; i < len(nestedset); i++ {
		cur := nestedset[i]
		skus, ok := cmap[cur.Path]
		if !ok {
			skus = nil
		}
		products := make([]*struct {
			SKU string `json:"sku"`
		}, 0)
		for _, s := range skus {
			products = append(products, &struct {
				SKU string `json:"sku"`
			}{
				SKU: s,
			})
		}
		n := &Category{
			Segment:  cur.Segment,
			path:     cur.Path,
			Name:     cur.Name,
			parent:   context,
			Nodes:    make([]*Category, 0),
			Products: products,
			lft:      cur.Lft,
			rgt:      cur.Rgt,
		}
		context.addChild(n)

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

// HasCatalog returns true if the catalog exists.
func (s *Service) HasCatalog(ctx context.Context) (bool, error) {
	has, err := s.model.HasCatalog(ctx)
	if err != nil {
		return false, errors.Wrap(err, "service: has catalog")
	}
	return has, nil
}

// GetCatalog returns the catalog as a hierarchy of nodes.
func (s *Service) GetCatalog(ctx context.Context) (*Category, error) {
	ns, err := s.model.GetCatalogNestedSet(ctx)
	if err != nil {
		return nil, err
	}
	if len(ns) == 0 {
		return nil, nil
	}
	cpas, err := s.model.GetCatalogProductAssocs(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "service: get catalog product assocs")
	}
	// convert slice into map
	cmap := make(map[string][]string)
	for _, cp := range cpas {
		_, ok := cmap[cp.Path]
		if ok {
			cmap[cp.Path] = append(cmap[cp.Path], cp.SKU)
		} else {
			cmap[cp.Path] = []string{cp.SKU}
		}
	}
	tree := buildCategoryTree(ns, cmap)
	return tree, nil
}

// DeleteCatalog purges the entire catalog hierarchy.
func (s *Service) DeleteCatalog(ctx context.Context) error {
	err := s.model.DeleteCatalogNestedSet(ctx)
	if err != nil {
		return errors.Wrap(err, "delete catalog nested set")
	}
	return nil
}
