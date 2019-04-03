package model_test

import (
	"context"
	"testing"

	"bitbucket.org/andyfusniakteam/ecom-api-go/model/mockmodel"
	"github.com/stretchr/testify/assert"
)

func setup(t *testing.T) (*mockmodel.MockModel, func()) {
	model := mockmodel.NewMockModel()
	return model, func() {
	}
}

func TestCreateCart(t *testing.T) {
	m, teardown := setup(t)
	defer teardown()

	ctx := context.Background()
	var cartUUID string
	t.Run("CreateCart", func(t *testing.T) {
		cartUUID, err := m.CreateCart(ctx)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "8f3139be-1d10-4847-8190-b882ce97a6bb", *cartUUID)
	})

	t.Run("AddItemToCart", func(t *testing.T) {
		t.Log(m)
		item, err := m.AddItemToCart(ctx, cartUUID, "default", "WATER", 3)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("%v\n", item)
	})
}
