package firebase

import (
	"context"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"

	"bitbucket.org/andyfusniakteam/ecom-api-go/model/postgres"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// ErrCategoryNotFound error
var ErrCategoryNotFound = errors.New("service: category not found")

// ErrCategoryNotLeaf error
var ErrCategoryNotLeaf = errors.New("service: category not a leaf")

// ErrAssocsAlreadyExist is returned by UpdateCatalog when any associations
// already exist.
var ErrAssocsAlreadyExist = errors.New("service: associations already exist")

// ErrLeafCategoryNotFound error
var ErrLeafCategoryNotFound = errors.New("service: category not found")

// ErrCategoriesEmpty error
var ErrCategoriesEmpty = errors.New("service: categories empty")

// CategoryList is a container for a list of category objects
type CategoryList struct {
	Object string          `json:"object"`
	Data   []*CategoryNode `json:"data"`
}

// CategoryRequest represents the request body for an update categories operation.
type CategoryRequest struct {
	Segment string `json:"segment"`
	Name    string `json:"name"`
	path    string
	lft     int
	rgt     int
	depth   int
	Nodes   []*CategoryRequest `json:"categories"`
}

// A CategoryNode represents an individual category in the catalog hierarchy.
type CategoryNode struct {
	Object   string `json:"object,omitempty"`
	ID       string `json:"id,omitempty"`
	Segment  string `json:"segment"`
	path     string
	Name     string `json:"name"`
	lft      int
	rgt      int
	depth    int
	parent   *CategoryNode
	Nodes    *CategoryList `json:"categories"`
	Products *ProductList  `json:"products,omitempty"`
}

// Category represents a single entry from the nested set
type Category struct {
	ID       string `json:"id"`
	Segment  string `json:"segment"`
	Path     string `json:"path"`
	Name     string `json:"name"`
	Lft      int    `json:"lft"`
	Rgt      int    `json:"rgt"`
	Depth    int    `json:"depth"`
	Created  time.Time
	Modified time.Time
}

// AddChild attaches a Category to its parent Category.
func (n *CategoryNode) AddChild(c *CategoryNode) {
	c.parent = n
	n.Nodes.Data = append(n.Nodes.Data, c)
}

// IsRoot returns true for the root node only.
func (n *CategoryNode) IsRoot() bool {
	return n.parent == nil
}

// IsLeaf return true if the node is a leaf node.
func (n *CategoryNode) IsLeaf() bool {
	return len(n.Nodes.Data) == 0
}

// NestedSet uses preorder traversal of the tree to return a
// slice of CategoryRow.
func (n *CategoryRequest) NestedSet(ns *[]*postgres.CategoryRow) {
	n.preorderTraversalNS(ns)
}

func (n *CategoryRequest) preorderTraversalNS(ns *[]*postgres.CategoryRow) {
	nsn := &postgres.CategoryRow{
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

// PreorderTraversalPrint provides a depth first search printout of each node
// in the hierarchy.
func (n *CategoryNode) PreorderTraversalPrint(w io.Writer) {
	tw := new(tabwriter.Writer).Init(w, 0, 8, 2, ' ', 0)
	n.preorderTraversalWrite(tw)
	tw.Flush()
}

func (n *CategoryNode) preorderTraversalWrite(w io.Writer) {
	fmt.Fprintf(w, "segment: %s\t path: %q\tname: %q\tlft: %d\t rgt: %d\t depth %d\n", n.Segment, n.path, n.Name, n.lft, n.rgt, n.depth)

	for _, i := range n.Nodes.Data {
		i.preorderTraversalWrite(w)
	}
}

func moveContext(context *CategoryNode) *CategoryNode {
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
func NewCategory(segment, name string) *CategoryNode {
	return &CategoryNode{
		Segment: segment,
		path:    "",
		Name:    name,
		lft:     -1,
		rgt:     -1,
		depth:   -1,
		parent:  nil,
		Nodes: &CategoryList{
			Object: "list",
			Data:   make([]*CategoryNode, 0),
		},
	}
}

// GenerateNestedSet performs a pre-order tree traversal wiring the tree.
func (n *CategoryRequest) GenerateNestedSet(lft, depth int, path string) int {
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

// UpdateCatalog takes a root tree Node and converts it to a nested set
// representation before calling the model to persist the replacement
// catalog.
func (s *Service) UpdateCatalog(ctx context.Context, root *CategoryRequest) error {
	hasAssocs, err := s.HasProductCategoryRelations(ctx)
	if err != nil {
		return errors.Wrapf(err, "HasProductCategoryRelations(ctx) error")
	}
	if hasAssocs {
		return ErrAssocsAlreadyExist
	}

	root.GenerateNestedSet(1, 0, "")
	ns := make([]*postgres.CategoryRow, 0, 128)
	root.NestedSet(&ns)
	if err := s.model.BatchCreateNestedSet(ctx, ns); err != nil {
		return errors.Wrap(err, "service: replace catalog")
	}
	return nil
}

func (n *CategoryNode) addChild(c *CategoryNode) {
	n.Nodes.Data = append(n.Nodes.Data, c)
}

func (n *CategoryNode) hasChildren() bool {
	return len(n.Nodes.Data) > 0
}

func (n *CategoryNode) findNode(segment string) *CategoryNode {
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
func (n *CategoryNode) FindNodeByPath(path string) *CategoryNode {
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
func BuildTree(nestedset []*postgres.CategoryRow, cmap map[string][]*Product) *CategoryNode {
	context := &CategoryNode{
		Object:  "category",
		ID:      nestedset[0].UUID,
		Segment: nestedset[0].Segment,
		path:    nestedset[0].Path,
		Name:    nestedset[0].Name,
		parent:  nil,
		Nodes: &CategoryList{
			Object: "list",
			Data:   make([]*CategoryNode, 0),
		},
		lft: nestedset[0].Lft,
		rgt: nestedset[0].Rgt,
	}

	// If the tree is just a single node there could be products
	// attached to it. Otherwise, products can't be attached
	// to non-leafs.
	if context.lft == context.rgt-1 {
		context.Products = &ProductList{
			Object: "list",
			Data:   make([]*Product, 0),
		}
	} else {
		context.Products = nil
	}

	for i := 1; i < len(nestedset); i++ {
		cur := nestedset[i]
		products, ok := cmap[cur.Path]
		if !ok {
			products = nil
		}
		n := &CategoryNode{
			Object:  "category",
			ID:      cur.UUID,
			Segment: cur.Segment,
			path:    cur.Path,
			Name:    cur.Name,
			parent:  context,
			Nodes: &CategoryList{
				Object: "list",
				Data:   make([]*CategoryNode, 0),
			},
			lft: cur.Lft,
			rgt: cur.Rgt,
		}

		// if the current node is a leaf node we either need
		// to attach a container with products
		//
		// "object": "list"
		// "data": [
		//   {}, {}, {}, ...
		// ]
		//
		// or attach a container with an empty list
		//
		// "object": "list"
		// "data": []
		//
		// If it's not a leaf node we set Products slice to nil
		// and the json omitempty will not render it to the client
		if cur.Lft == cur.Rgt-1 {
			if products != nil {
				n.Products = &ProductList{
					Object: "list",
					Data:   products,
				}
			} else {
				n.Products = &ProductList{
					Object: "list",
					Data:   make([]*Product, 0, 0),
				}
			}
		} else {
			n.Products = nil
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

// GetCategoriesTree returns a tree of all categories as a hierarchy of nodes.
func (s *Service) GetCategoriesTree(ctx context.Context) (*CategoryNode, error) {
	log.WithContext(ctx).Debug("service: GetCatalog started")
	ns, err := s.model.GetCategories(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "s.model.GetCategories(ctx) failed")
	}
	if len(ns) == 0 {
		log.WithContext(ctx).Debug("service: s.model.GetCategories(ctx) returned an empty list")
		return nil, ErrCategoriesEmpty
	}
	cpas, err := s.model.GetProductCategoryRelationsFull(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "service: GetProductCategoryAssocsFull(ctx) failed")
	}

	// convert slice into map
	cmap := make(map[string][]*Product)
	for _, cpf := range cpas {
		cmap[cpf.CategoryPath] = append(cmap[cpf.CategoryPath], &Product{
			Object:   "product",
			ID:       cpf.ProductUUID,
			SKU:      cpf.ProductSKU,
			Path:     cpf.ProductPath,
			Name:     cpf.ProductName,
			Created:  cpf.ProductCreated,
			Modified: cpf.ProductModified,
		})
	}
	tree := BuildTree(ns, cmap)
	return tree, nil
}

// DeleteCategories deletes all categories effectively purging the
// entire tree.
func (s *Service) DeleteCategories(ctx context.Context) error {
	err := s.model.DeleteCategories(ctx)
	if err != nil {
		return errors.Wrap(err, "service: delete categories failed")
	}
	return nil
}

// Category raw nested set

// GetCategories returns a list of categories.
func (s *Service) GetCategories(ctx context.Context) ([]*Category, error) {
	cats, err := s.model.GetCategories(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "service: s.model.GetCategories(ctx) failed")
	}

	categories := make([]*Category, 0, len(cats))
	for _, c := range cats {
		category := Category{
			ID:       c.UUID,
			Segment:  c.Segment,
			Path:     c.Path,
			Name:     c.Name,
			Lft:      c.Lft,
			Rgt:      c.Rgt,
			Depth:    c.Depth,
			Created:  c.Created,
			Modified: c.Modified,
		}
		categories = append(categories, &category)
	}
	return categories, nil
}
