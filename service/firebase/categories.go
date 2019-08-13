package firebase

import (
	"context"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"bitbucket.org/andyfusniakteam/ecom-api-go/model/postgres"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// ErrAssocsAlreadyExist is returned by UpdateCatalog when any associations
// already exist.
var ErrAssocsAlreadyExist = errors.New("service: associations already exist")

// CategoryList is a container for a list of category objects
type CategoryList struct {
	Object string      `json:"object"`
	Data   []*Category `json:"data"`
}

// A Category represents an individual category in the catalog hierarchy.
type Category struct {
	Object   string `json:"object"`
	ID       string `json:"id"`
	Segment  string `json:"segment"`
	path     string
	Name     string `json:"name"`
	lft      int
	rgt      int
	depth    int
	parent   *Category
	Nodes    *CategoryList    `json:"categories"`
	Products *ProductSlimList `json:"products"`
}

// AddChild attaches a Category to its parent Category.
func (n *Category) AddChild(c *Category) {
	c.parent = n
	n.Nodes.Data = append(n.Nodes.Data, c)
}

// IsRoot returns true for the root node only.
func (n *Category) IsRoot() bool {
	return n.parent == nil
}

// IsLeaf return true if the node is a leaf node.
func (n *Category) IsLeaf() bool {
	return len(n.Nodes.Data) == 0
}

// NestedSet uses preorder traversal of the tree to return a
// slice of NestedSetNodes.
func (n *Category) NestedSet(ns *[]*postgres.NestedSetNode) {
	n.preorderTraversalNS(ns)
}

func (n *Category) preorderTraversalNS(ns *[]*postgres.NestedSetNode) {
	nsn := &postgres.NestedSetNode{
		Segment: n.Segment,
		Path:    n.path,
		Name:    n.Name,
		Lft:     n.lft,
		Rgt:     n.rgt,
		Depth:   n.depth,
	}
	*ns = append(*ns, nsn)
	for _, i := range n.Nodes.Data {
		i.preorderTraversalNS(ns)
	}
}

// PreorderTraversalPrint provides a depth first search printout of each node
// in the hierarchy.
func (n *Category) PreorderTraversalPrint(w io.Writer) {
	tw := new(tabwriter.Writer).Init(w, 0, 8, 2, ' ', 0)
	n.preorderTraversalWrite(tw)
	tw.Flush()
}

func (n *Category) preorderTraversalWrite(w io.Writer) {
	fmt.Fprintf(w, "segment: %s\t path: %q\tname: %q\tlft: %d\t rgt: %d\t depth %d\n", n.Segment, n.path, n.Name, n.lft, n.rgt, n.depth)

	for _, i := range n.Nodes.Data {
		i.preorderTraversalWrite(w)
	}
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

// NewCategory creates a new Category.
func NewCategory(segment, name string) *Category {
	return &Category{
		Segment: segment,
		path:    "",
		Name:    name,
		lft:     -1,
		rgt:     -1,
		depth:   -1,
		parent:  nil,
		Nodes: &CategoryList{
			Object: "list",
			Data:   make([]*Category, 0),
		},
	}
}

// GenerateNestedSet performs a pre-order tree traversal wiring the tree.
func (n *Category) GenerateNestedSet(lft, depth int, path string) int {
	rgt := lft + 1
	for _, i := range n.Nodes.Data {
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

// UpdateCatalog takes a root tree Node and converts it to a nested set
// representation before calling the model to persist the replacement
// catalog.
func (s *Service) UpdateCatalog(ctx context.Context, root *Category) error {
	hasAssocs, err := s.HasCategoryProductAssocs(ctx)
	if err != nil {
		return errors.Wrapf(err, "HasCategoryProductAssocs(ctx) error")
	}
	if hasAssocs {
		return ErrAssocsAlreadyExist
	}

	root.GenerateNestedSet(1, 0, "")
	ns := make([]*postgres.NestedSetNode, 0, 128)
	root.NestedSet(&ns)
	if err := s.model.BatchCreateNestedSet(ctx, ns); err != nil {
		return errors.Wrap(err, "service: replace catalog")
	}
	return nil
}

func (n *Category) addChild(c *Category) {
	n.Nodes.Data = append(n.Nodes.Data, c)
}

func (n *Category) hasChildren() bool {
	return len(n.Nodes.Data) > 0
}

func (n *Category) findNode(segment string) *Category {
	if !n.hasChildren() {
		return nil
	}
	for _, node := range n.Nodes.Data {
		if node.Segment == segment {
			return node
		}
	}
	return nil
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
func BuildTree(nestedset []*postgres.NestedSetNode, cmap map[string][]*ProductSlim) *Category {
	context := &Category{
		Object:  "category",
		ID:      nestedset[0].UUID,
		Segment: nestedset[0].Segment,
		path:    nestedset[0].Path,
		Name:    nestedset[0].Name,
		parent:  nil,
		Nodes: &CategoryList{
			Object: "list",
			Data:   make([]*Category, 0),
		},
		Products: &ProductSlimList{
			Object: "list",
			Data:   make([]*ProductSlim, 0),
		},
		lft: nestedset[0].Lft,
		rgt: nestedset[0].Rgt,
	}
	for i := 1; i < len(nestedset); i++ {
		cur := nestedset[i]
		products, ok := cmap[cur.Path]
		if !ok {
			products = nil
		}
		n := &Category{
			Object:  "category",
			ID:      cur.UUID,
			Segment: cur.Segment,
			path:    cur.Path,
			Name:    cur.Name,
			parent:  context,
			Nodes: &CategoryList{
				Object: "list",
				Data:   make([]*Category, 0),
			},
			Products: &ProductSlimList{
				Object: "list",
				Data:   products,
			},
			lft: cur.Lft,
			rgt: cur.Rgt,
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
	log.WithContext(ctx).Debug("Service: GetCatalog started")
	ns, err := s.model.GetCatalogNestedSet(ctx)
	if err != nil {
		return nil, err
	}
	if len(ns) == 0 {
		log.WithContext(ctx).Debug("s.model.GetCatalogNestedSet(ctx) returned an empty list")
		return nil, nil
	}
	cpas, err := s.model.GetCategoryProductAssocsFull(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "service: GetCategoryProductAssocsFull(ctx) failed")
	}

	// convert slice into map
	cmap := make(map[string][]*ProductSlim)
	for _, cpf := range cpas {
		cmap[cpf.CategoryPath] = append(cmap[cpf.CategoryPath], &ProductSlim{
			Object:   "product_slim",
			ID:       cpf.ProductUUID,
			SKU:      cpf.SKU,
			EAN:      cpf.EAN,
			Path:     cpf.ProductPath,
			Name:     cpf.Name,
			Created:  cpf.ProductCreated,
			Modified: cpf.ProductModified,
		})
	}
	tree := BuildTree(ns, cmap)
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
