package postgres

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"regexp"
	"testing"

	"bitbucket.org/andyfusniakteam/ecom-api-go/model"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

var ecomPgPassword = flag.String("pgpass", "postgres", "Set the postgres password")

func setup(t *testing.T) (*PgModel, func()) {
	dsn := fmt.Sprintf("host=localhost port=5432 user=postgres password=%s dbname=ecom_dev sslmode=disable connect_timeout=10", *ecomPgPassword)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
		return nil, func() {}
	}
	err = db.Ping()
	if err != nil {
		t.Fatalf("failed to verify db connection: %v", err)
		return nil, func() {}
	}
	model := NewPgModel(db)
	return model, func() {
		if err := db.Close(); err != nil {
			t.Errorf("db.Close(): %s", err)
		}
	}
}

func isValidUUID(uuid string) bool {
	r := regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$")
	return r.MatchString(uuid)
}

func TestGetCustomers(t *testing.T) {
	m, teardown := setup(t)
	defer teardown()

	ctx := context.Background()

	prs, err := m.GetCustomers(ctx, "firstname", "ASC", 2, "1bb7faa4-0435-4cf5-978d-3ee76327c32a")
	if err != nil {
		t.Fatalf("m.GetCustomers() :%v", err)
	}
	fmt.Println(prs.RContext)
	if err != nil {
		t.Errorf("model.GetCatalogNestedSet(ctx): %v", err)
	}

	for i, v := range prs.Rset.([]*model.Customer) {
		fmt.Println(i, v)
	}
}

func TestGetCatalogNestedSet(t *testing.T) {
	model, teardown := setup(t)
	defer teardown()

	ctx := context.Background()
	nodes, err := model.GetCatalogNestedSet(ctx)
	if err != nil {
		t.Errorf("model.GetCatalogNestedSet(ctx): %v", err)
	}

	assert.Equal(t, len(nodes), 14, "should be 14 nodes in the set")

	expected := map[int]struct {
		segment string
		path    string
		lft     int
		rgt     int
		depth   int
	}{
		0:  {segment: "a", path: "/a", lft: 1, rgt: 28, depth: 0},
		1:  {segment: "b", path: "/a/b", lft: 2, rgt: 5, depth: 1},
		2:  {segment: "e", path: "/a/b/e", lft: 3, rgt: 4, depth: 2},
		3:  {segment: "c", path: "/a/c", lft: 6, rgt: 19, depth: 1},
		4:  {segment: "f", path: "/a/c/f", lft: 7, rgt: 16, depth: 2},
		5:  {segment: "i", path: "/a/c/f/i", lft: 8, rgt: 9, depth: 3},
		6:  {segment: "j", path: "/a/c/f/j", lft: 10, rgt: 15, depth: 3},
		7:  {segment: "m", path: "/a/c/f/j/m", lft: 11, rgt: 12, depth: 4},
		8:  {segment: "n", path: "/a/c/f/j/n", lft: 13, rgt: 14, depth: 4},
		9:  {segment: "g", path: "/a/c/g", lft: 17, rgt: 18, depth: 2},
		10: {segment: "d", path: "/a/d", lft: 20, rgt: 27, depth: 1},
		11: {segment: "h", path: "/a/d/h", lft: 21, rgt: 26, depth: 2},
		12: {segment: "k", path: "/a/d/h/k", lft: 22, rgt: 23, depth: 3},
		13: {segment: "l", path: "/a/d/h/l", lft: 24, rgt: 25, depth: 3},
	}

	for i, n := range nodes {
		assert.Equal(t, expected[i].segment, n.Segment, fmt.Sprintf("Node %d name should be %q; got %q", i, expected[i].segment, n.Segment))
		assert.Equal(t, expected[i].path, n.Path, fmt.Sprintf("Node %d path should be %q; got %q", i, expected[i].path, n.Path))
		assert.Equal(t, expected[i].lft, n.Lft, fmt.Sprintf("Node %d lft should be %d; got %d", i, expected[i].lft, n.Lft))
		assert.Equal(t, expected[i].rgt, n.Rgt, fmt.Sprintf("Node %d rgt should be %d; got %d", i, expected[i].rgt, n.Rgt))
		assert.Equal(t, expected[i].depth, n.Depth, fmt.Sprintf("Node %d depth should be %d; got %d", i, expected[i].depth, n.Depth))
	}
}

func TestCart(t *testing.T) {
	model, teardown := setup(t)
	defer teardown()

	ctx := context.Background()
	uuid, err := model.CreateCart(ctx)
	if err != nil {
		t.Errorf("model.CreateCart(ctx): %v", err)
	}

	if !isValidUUID(*uuid) {
		t.Errorf("got invalid uuid: %s", *uuid)
	}

	t.Run("AddItemToCart", func(t *testing.T) {
		_, err := model.AddItemToCart(ctx, *uuid, "default", "WATER", 1)
		if err != nil {
			t.Errorf("AddItemToCart(...): %v", err)
		}
	})
}
