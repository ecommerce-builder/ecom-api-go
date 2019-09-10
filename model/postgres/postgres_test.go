package postgres

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"regexp"
	"testing"

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

func TestUpdateProduct(t *testing.T) {
	model, teardown := setup(t)
	defer teardown()

	ctx := context.Background()
	pu := &ProductCreateUpdate{
		Path: "updated-url",
		Name: "Updated Name",
	}
	pr, err := model.UpdateProduct(ctx, "DESK-SKU", pu)
	if err != nil {
		t.Fatalf("model.UpdateProduct(): %v", err)
	}
	t.Log(pr)

}

// func TestGetUsers(t *testing.T) {
// 	model, teardown := setup(t)
// 	defer teardown()

// 	ctx := context.Background()

// 	prs, err := model.GetUsers(ctx, "firstname", "ASC", 2, "1bb7faa4-0435-4cf5-978d-3ee76327c32a")
// 	if err != nil {
// 		t.Fatalf("model.GetUser(): %v", err)
// 	}
// 	if err != nil {
// 		t.Errorf("model.GetCategories(ctx): %v", err)
// 	}

// 	for i, v := range prs.Rset.([]*model.User) {
// 		fmt.Println(i, v)
// 	}
// }

func TestGetAddressByUUID(t *testing.T) {
	model, teardown := setup(t)
	defer teardown()

	ctx := context.Background()

	t.Run("get valid address", func(t *testing.T) {
		addr, err := model.GetAddressByUUID(ctx, "f35f10c4-e9fe-4283-b2e6-8be1f9eed875")
		if err != nil {
			t.Fail()
		}
		t.Log(addr)
	})

	t.Run("get missing address", func(t *testing.T) {
		addr, err := model.GetAddressByUUID(ctx, "f35f10c4-e9fe-4283-b2e6-8be1f9eed876")
		if model.IsNotExist(err) {
			if ne, ok := err.(*ResourceError); ok {
				assert.Equal(t, "GetAddressByUUID", ne.Op)
				assert.Equal(t, "address", ne.Resource)
				assert.Equal(t, "f35f10c4-e9fe-4283-b2e6-8be1f9eed876", ne.UUID)
				assert.Equal(t, ErrNotExist, ne.Err)
				return
			}
		}
		t.Log(addr)
		t.Fail()
	})

	t.Run("get invalid format addresss", func(t *testing.T) {
		addr, err := model.GetAddressByUUID(ctx, "f35f10c4-e9fe-4283-b2e6-8be1f9eed87x")
		if err != nil {
			if IsInvalidText(err) {
				if ite, ok := err.(*ResourceError); ok {
					assert.Equal(t, "GetAddressByUUID", ite.Op)
					assert.Equal(t, "address", ite.Resource)
					assert.Equal(t, "f35f10c4-e9fe-4283-b2e6-8be1f9eed87x", ite.UUID)
					assert.Equal(t, ErrInvalidText, ite.Err)
				}
				return
			}
		}
		t.Log(addr)
		t.Fail()
	})
}

func TestCreateProductCategoryAssoc(t *testing.T) {
	model, teardown := setup(t)
	defer teardown()

	ctx := context.Background()
	cp, err := model.CreateProductCategoryAssocs(ctx, "a/c/f/j/m", "WATER-SKU")
	if err != nil {
		t.Errorf("create category product assoc: %v", err)
	}
	t.Log(cp)
}

func TestDeleteProductCategoryAssoc(t *testing.T) {
	model, teardown := setup(t)
	defer teardown()

	ctx := context.Background()
	err := model.DeleteProductCatalogAssoc(ctx, "a/c/f/j/m", "WATER-SKU")
	if err != nil {
		t.Errorf("delete category product assoc: %v", err)
	}

}

func TestGetCategoryByPath(t *testing.T) {
	model, teardown := setup(t)
	defer teardown()

	ctx := context.Background()
	ns, err := model.GetCatagoryByPath(ctx, "a/c/f/j")
	if err != nil {
		t.Errorf("get category by path: %v", err)
	}
	t.Log(ns)
}

func TestGetCategories(t *testing.T) {
	model, teardown := setup(t)
	defer teardown()

	ctx := context.Background()
	nodes, err := model.GetCategories(ctx)
	if err != nil {
		t.Errorf("model.GetCategories(ctx): %v", err)
	}

	assert.Equal(t, len(nodes), 14, "should be 14 nodes in the set")

	expected := map[int]struct {
		segment string
		path    string
		lft     int
		rgt     int
		depth   int
	}{
		0:  {segment: "a", path: "a", lft: 1, rgt: 28, depth: 0},
		1:  {segment: "b", path: "a/b", lft: 2, rgt: 5, depth: 1},
		2:  {segment: "e", path: "a/b/e", lft: 3, rgt: 4, depth: 2},
		3:  {segment: "c", path: "a/c", lft: 6, rgt: 19, depth: 1},
		4:  {segment: "f", path: "a/c/f", lft: 7, rgt: 16, depth: 2},
		5:  {segment: "i", path: "a/c/f/i", lft: 8, rgt: 9, depth: 3},
		6:  {segment: "j", path: "a/c/f/j", lft: 10, rgt: 15, depth: 3},
		7:  {segment: "m", path: "a/c/f/j/m", lft: 11, rgt: 12, depth: 4},
		8:  {segment: "n", path: "a/c/f/j/n", lft: 13, rgt: 14, depth: 4},
		9:  {segment: "g", path: "a/c/g", lft: 17, rgt: 18, depth: 2},
		10: {segment: "d", path: "a/d", lft: 20, rgt: 27, depth: 1},
		11: {segment: "h", path: "a/d/h", lft: 21, rgt: 26, depth: 2},
		12: {segment: "k", path: "a/d/h/k", lft: 22, rgt: 23, depth: 3},
		13: {segment: "l", path: "a/d/h/l", lft: 24, rgt: 25, depth: 3},
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

	t.Run("AddProductToCart", func(t *testing.T) {
		_, err := model.AddProductToCart(ctx, *uuid, "default", "WATER", 1)
		if err != nil {
			t.Errorf("AddItemToCart(...): %v", err)
		}
	})
}

func TestCreateImageEntry(t *testing.T) {
	model, teardown := setup(t)
	defer teardown()

	ctx := context.Background()
	cpis := []CreateImage{
		{ // 0
			W:     800,
			H:     600,
			Path:  "products/WATER/images/originals/front_view.jpg",
			Typ:   "image/jpeg",
			Ori:   true,
			Pri:   10,
			Size:  345345,
			Q:     100,
			GSURL: "gs://test-data-spycameracctv.appspot.com/products/WATER/images/originals/front_view.jpg",
			Data:  nil,
		},
		{ // 1
			W:     700,
			H:     300,
			Path:  "products/TV/images/originals/side_view.jpg",
			Typ:   "image/jpeg",
			Ori:   true,
			Pri:   10,
			Size:  12345,
			Q:     90,
			GSURL: "gs://test-data-spycameracctv.appspot.com/products/TV/images/originals/side_view.jpg",
			Data:  nil,
		},
		{ // 2
			W:     700,
			H:     300,
			Path:  "products/TV/images/originals/rear_view.jpg",
			Typ:   "image/jpeg",
			Ori:   true,
			Pri:   20,
			Size:  23456,
			Q:     95,
			GSURL: "gs://test-data-spycameracctv.appspot.com/products/TV/images/originals/rear_view.jpg",
			Data:  nil,
		},
	}
	pis := make([]*Image, 3)
	t.Run("CreateProductImages", func(t *testing.T) {
		var err error
		for i, c := range cpis {
			pis[i], err = model.CreateImageEntry(ctx, &c)
			if err != nil {
				t.Fatalf("CreateImageEntry(ctx, %v): %s", c, err)
			}
			assert.Equal(t, int(cpis[i].W), pis[i].W)
			assert.Equal(t, int(cpis[i].H), pis[i].H)
			assert.Equal(t, cpis[i].Path, pis[i].Path)
			assert.Equal(t, cpis[i].Typ, pis[i].Typ)
			assert.Equal(t, true, pis[i].Ori)
			assert.Equal(t, false, pis[i].Up)
			assert.Equal(t, int(cpis[i].Pri), pis[i].Pri)
			assert.Equal(t, int(cpis[i].Size), pis[i].Size)
			assert.Equal(t, int(cpis[i].Q), pis[i].Q)
			assert.Equal(t, cpis[i].GSURL, pis[i].GSURL)
		}
	})

	t.Run("ConfirmImageUpload", func(t *testing.T) {
		cp, err := model.ConfirmImageUploaded(ctx, pis[0].UUID)
		if err != nil {
			t.Fatalf("ConfirmImageUploaded(ctx): %s", err)
		}
		assert.Equal(t, int(1), cp.productID)
		assert.Equal(t, pis[0].UUID, cp.UUID)
		assert.Equal(t, int(800), cp.W)
		assert.Equal(t, int(600), cp.H)
		assert.Equal(t, "products/WATER/images/originals/front_view.jpg", pis[0].Path)
		assert.Equal(t, "image/jpeg", cp.Typ)
		assert.Equal(t, true, cp.Ori)
		assert.Equal(t, true, cp.Up)
		assert.Equal(t, int(10), cp.Pri)
		assert.Equal(t, int(345345), cp.Size)
		assert.Equal(t, "gs://test-data-spycameracctv.appspot.com/products/WATER/images/originals/front_view.jpg", cp.GSURL)
	})

	t.Run("GetImageEntries", func(t *testing.T) {
		images, err := model.GetImagesBySKU(ctx, "TV")
		if err != nil {
			t.Fatalf("GetImageEntries(ctx, %q): %s", "TV", err)
		}
		for j, p := range images {
			idx := j + 1
			assert.Equal(t, int(pis[idx].ProductID), p.productID)
			assert.Equal(t, pis[idx].UUID, p.UUID)
			assert.Equal(t, int(pis[idx].W), p.W)
			assert.Equal(t, int(pis[idx].H), p.H)
			assert.Equal(t, pis[idx].Path, p.Path)
			assert.Equal(t, pis[idx].Typ, p.Typ)
			assert.Equal(t, int(pis[idx].Pri), p.Pri)
			assert.Equal(t, int(pis[idx].Size), p.Size)
			assert.Equal(t, pis[idx].GSURL, p.GSURL)
			assert.Equal(t, pis[idx].Created, p.Created)
			assert.Equal(t, pis[idx].Modified, p.Modified)
		}
	})

	t.Run("DeleteImageEntry", func(t *testing.T) {
		for _, p := range pis {
			count, err := model.DeleteProductImage(ctx, p.UUID)
			if err != nil {
				t.Fatalf("DeleteImageEntry(ctx, %v): %s", p.UUID, err)
			}
			assert.Equal(t, int64(1), count)
		}
	})
}
