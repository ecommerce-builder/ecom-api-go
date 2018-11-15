package firebase

import (
	"context"

	"bitbucket.org/andyfusniakteam/ecomapi"
	"bitbucket.org/andyfusniakteam/ecomapi/model"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
)

type FirebaseService struct {
	model model.EcomModel
	fbApp *firebase.App
}

func New(model model.EcomModel, fbApp *firebase.App) (*FirebaseService, error) {
	s := FirebaseService{model, fbApp}
	return &s, nil
}

// CreateCart generates a new random UUID to be used for subseqent cart calls
func (s *FirebaseService) CreateCart() (*string, error) {
	strp, err := s.model.CreateCart()
	if err != nil {
		return nil, err
	}

	return strp, nil
}

// AddItemToCart adds a single item to a given cart
func (s *FirebaseService) AddItemToCart(cartUUID string, sku string, qty int) (*app.CartItem, error) {
	item, err := s.model.AddItemToCart(cartUUID, "default", sku, qty)
	if err != nil {
		return nil, err
	}

	sitem := app.CartItem{
		CartUUID:  item.CartUUID,
		Sku:       item.Sku,
		Qty:       item.Qty,
		UnitPrice: item.UnitPrice,
		Created:   item.Created,
		Modified:  item.Modified,
	}
	return &sitem, nil
}

// GetCartItems get all items in the given cart
func (s *FirebaseService) GetCartItems(cartUUID string) ([]*app.CartItem, error) {
	items, err := s.model.GetCartItems(cartUUID)
	if err != nil {
		return nil, err
	}

	results := make([]*app.CartItem, 0, 32)
	for _, v := range items {
		i := app.CartItem{
			CartUUID:  v.CartUUID,
			Sku:       v.Sku,
			Qty:       v.Qty,
			UnitPrice: v.UnitPrice,
			Created:   v.Created,
			Modified:  v.Modified,
		}
		results = append(results, &i)
	}
	return results, nil
}

// UpdateCartItem updates a single item's qty
func (s *FirebaseService) UpdateCartItem(cartUUID string, sku string, qty int) (*app.CartItem, error) {
	item, err := s.model.UpdateItemByCartUUID(cartUUID, sku, qty)
	if err != nil {
		return nil, err
	}

	sitem := app.CartItem{
		CartUUID:  item.CartUUID,
		Sku:       item.Sku,
		Qty:       item.Qty,
		UnitPrice: item.UnitPrice,
		Created:   item.Created,
		Modified:  item.Modified,
	}
	return &sitem, nil
}

// DeleteCartItem deletes a single cart item
func (s *FirebaseService) DeleteCartItem(cartUUID string, sku string) (count int64, err error) {
	count, err = s.model.DeleteCartItem(cartUUID, sku)
	if err != nil {
		return -1, err
	}
	return count, nil
}

// EmptyCartItems empties the cart of all items but not coupons
func (s *FirebaseService) EmptyCartItems(cartUUID string) (err error) {
	return s.model.EmptyCartItems(cartUUID)
}

// CreateCustomer creates a new customer
func (s *FirebaseService) CreateCustomer(email, password, firstname, lastname string) (*app.Customer, error) {
	ctx := context.Background()
	authClient, err := s.fbApp.Auth(ctx)
	if err != nil {
		return nil, err
	}

	user := (&auth.UserToCreate{}).
		Email(email).
		EmailVerified(false).
		Password(password).
		DisplayName(`${firstname} ${lastname}`).
		Disabled(false)

	userRecord, err := authClient.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	c, _ := s.model.CreateCustomer(userRecord.UID, email, firstname, lastname)

	ac := app.Customer{
		CustomerUUID: c.CustomerUUID,
		UID:          c.UID,
		Email:        c.Email,
		Firstname:    c.Firstname,
		Lastname:     c.Lastname,
		Created:      c.Created,
		Modified:     c.Modified,
	}
	return &ac, nil
}

// GetCustomer retrieves a customer by customer UUID
func (s *FirebaseService) GetCustomer(customerUUID string) (*app.Customer, error) {
	c, err := s.model.GetCustomerByUUID(customerUUID)
	if err != nil {
		return nil, err
	}

	ac := app.Customer{
		CustomerUUID: c.CustomerUUID,
		UID:          c.UID,
		Email:        c.Email,
		Firstname:    c.Firstname,
		Lastname:     c.Lastname,
		Created:      c.Created,
		Modified:     c.Modified,
	}
	return &ac, nil
}

// CreateAddress creates a new address for a customer
func (s *FirebaseService) CreateAddress(customerUUID, typ, contactName, addr1 string, addr2 *string, city string, county *string, postcode string, country string) (*app.Address, error) {
	customerID, err := s.model.GetCustomerIDByUUID(customerUUID)
	if err != nil {
		return nil, err
	}

	a, err := s.model.CreateAddress(customerID, typ, contactName, addr1, addr2, city, county, postcode, country)
	if err != nil {
		return nil, err
	}

	aa := app.Address{
		AddrUUID:    a.AddrUUID,
		Typ:         a.Typ,
		ContactName: a.ContactName,
		Addr1:       a.Addr1,
		Addr2:       a.Addr2,
		City:        a.City,
		County:      a.County,
		Postcode:    a.Postcode,
		Country:     a.Country,
		Created:     a.Created,
		Modified:    a.Modified,
	}
	return &aa, nil
}

// GetAddress gets an address by UUID
func (s *FirebaseService) GetAddress(addressUUID string) (*app.Address, error) {
	a, err := s.model.GetAddressByUUID(addressUUID)
	if err != nil {
		return nil, err
	}

	aa := app.Address{
		AddrUUID:    a.AddrUUID,
		Typ:         a.Typ,
		ContactName: a.ContactName,
		Addr1:       a.Addr1,
		Addr2:       a.Addr2,
		City:        a.City,
		County:      a.County,
		Postcode:    a.Postcode,
		Country:     a.Country,
		Created:     a.Created,
		Modified:    a.Modified,
	}
	return &aa, nil
}

// GetAddresses gets a slice of addresses for a given customer
func (s *FirebaseService) GetAddresses(customerUUID string) ([]*app.Address, error) {
	customerID, err := s.model.GetCustomerIDByUUID(customerUUID)
	if err != nil {
		return nil, err
	}

	al, err := s.model.GetAddresses(customerID)
	if err != nil {
		return nil, err
	}

	results := make([]*app.Address, 0, 32)
	for _, v := range al {
		i := app.Address{
			AddrUUID:    v.AddrUUID,
			Typ:         v.Typ,
			ContactName: v.ContactName,
			Addr1:       v.Addr1,
			Addr2:       v.Addr2,
			City:        v.City,
			County:      v.County,
			Postcode:    v.Postcode,
			Country:     v.Country,
			Created:     v.Created,
			Modified:    v.Modified,
		}
		results = append(results, &i)
	}

	return results, nil
}

// DeleteAddress deletes an address by uuid
func (s *FirebaseService) DeleteAddress(addrUUID string) error {
	err := s.model.DeleteAddressByUUID(addrUUID)
	if err != nil {
		return err
	}

	return nil
}
