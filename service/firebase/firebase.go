package firebase

import (
	"context"
	"encoding/json"
	"fmt"

	"bitbucket.org/andyfusniakteam/ecom-api-go"
	"bitbucket.org/andyfusniakteam/ecom-api-go/model"
	"cloud.google.com/go/pubsub"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	log "github.com/sirupsen/logrus"
)

type FirebaseService struct {
	model        model.EcomModel
	fbApp        *firebase.App
	fbAuthClient *auth.Client
	pubsubClient *pubsub.Client
}

func New(model model.EcomModel, fbApp *firebase.App, fbAuthClient *auth.Client, pubsubClient *pubsub.Client) (*FirebaseService, error) {
	s := FirebaseService{model, fbApp, fbAuthClient, pubsubClient}
	return &s, nil
}

// Auth accepts a JSON Web Token, usually passed from the HTTP client and returns a auth.Token if valid or nil if
func (s *FirebaseService) Authenticate(ctx context.Context, jwt string) (*auth.Token, error) {
	token, err := s.fbAuthClient.VerifyIDToken(ctx, jwt)

	if err != nil {
		return nil, err
	}
	return token, nil
}

// CreateCart generates a new random UUID to be used for subseqent cart calls
func (s *FirebaseService) CreateCart(ctx context.Context) (*string, error) {
	log.Debug("s.CreateCart() started")

	strptr, err := s.model.CreateCart(ctx)
	if err != nil {
		return nil, err
	}

	log.Debugf("s.CreateCart() returned %s", *strptr)
	return strptr, nil
}

// AddItemToCart adds a single item to a given cart
func (s *FirebaseService) AddItemToCart(ctx context.Context, cartUUID string, sku string, qty int) (*app.CartItem, error) {
	log.Debugf("s.AddItemToCart(%s, %s, %d) started", cartUUID, sku, qty)

	item, err := s.model.AddItemToCart(ctx, cartUUID, "default", sku, qty)
	if err != nil {
		return nil, err
	}

	log.Debug(item)

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
func (s *FirebaseService) GetCartItems(ctx context.Context, cartUUID string) ([]*app.CartItem, error) {
	items, err := s.model.GetCartItems(ctx, cartUUID)
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
func (s *FirebaseService) UpdateCartItem(ctx context.Context, cartUUID string, sku string, qty int) (*app.CartItem, error) {
	item, err := s.model.UpdateItemByCartUUID(ctx, cartUUID, sku, qty)
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
func (s *FirebaseService) DeleteCartItem(ctx context.Context, cartUUID string, sku string) (count int64, err error) {
	count, err = s.model.DeleteCartItem(ctx, cartUUID, sku)
	if err != nil {
		return -1, err
	}
	return count, nil
}

// EmptyCartItems empties the cart of all items but not coupons
func (s *FirebaseService) EmptyCartItems(ctx context.Context, cartUUID string) (err error) {
	return s.model.EmptyCartItems(ctx, cartUUID)
}

// CreateCustomer creates a new customer
func (s *FirebaseService) CreateCustomer(ctx context.Context, role, email, password, firstname, lastname string) (*app.Customer, error) {
	log.Debugf("s.CreateCustomer(%s, %s, %s, %s) started", email, "*****", firstname, lastname)

	authClient, err := s.fbApp.Auth(ctx)
	if err != nil {
		return nil, err
	}

	user := (&auth.UserToCreate{}).
		Email(email).
		EmailVerified(false).
		Password(password).
		DisplayName(fmt.Sprintf("%s %s", firstname, lastname)).
		Disabled(false)

	userRecord, err := authClient.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	payload := struct {
		UID               string `json:"uid"`
		Email             string `json:"email"`
		DisplayName       string `json:"display_name"`
		CreationTimestamp int64  `json:"creation_timestamp"`
	}{
		UID:               userRecord.UID,
		Email:             userRecord.Email,
		DisplayName:       userRecord.DisplayName,
		CreationTimestamp: userRecord.UserMetadata.CreationTimestamp,
	}
	b, err := json.Marshal(&payload)
	if err != nil {
		log.Errorf("json.Marshal failed: %v", err)
	}
	log.Debugf("payload marshalled to string %s", b)

	msg := &pubsub.Message{
		Data: b,
		Attributes: map[string]string{
			"eventName": string(app.OpCreateCustomer),
		},
	}

	topic := s.pubsubClient.Topic("ecom-api")
	defer topic.Stop()

	if _, err := topic.Publish(ctx, msg).Get(ctx); err != nil {
		return nil, fmt.Errorf("Could not publish message: %v", err)
	}

	c, err := s.model.CreateCustomer(ctx, userRecord.UID, email, firstname, lastname)
	if err != nil {
		return nil, fmt.Errorf("model.CreateCustomer(%s, %s, %s, %s) failed: %v", userRecord.UID, email, firstname, lastname, err)
	}

	// Set the custom claims for this user
	err = authClient.SetCustomUserClaims(ctx, c.UID, map[string]interface{}{
		"cuuid": c.CustomerUUID,
		"role":  role,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to set custom claims for uid=%s customer_uuid=%s: %v", c.UID, c.CustomerUUID, err)
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

	log.Debugf("%+v", ac)
	return &ac, nil
}

func (s *FirebaseService) GetCustomers(ctx context.Context, size int, startsAfter string) ([]*app.Customer, error) {
	customers, err := s.model.GetCustomers(ctx, 1, size, startsAfter)
	if err != nil {
		return nil, err
	}

	results := make([]*app.Customer, 0, size)

	for _, v := range customers {
		c := app.Customer{
			CustomerUUID: v.CustomerUUID,
			UID:          v.UID,
			Email:        v.Email,
			Firstname:    v.Firstname,
			Lastname:     v.Lastname,
			Created:      v.Created,
			Modified:     v.Modified,
		}
		results = append(results, &c)
	}

	return results, nil
}

// GetCustomer retrieves a customer by customer UUID
func (s *FirebaseService) GetCustomer(ctx context.Context, customerUUID string) (*app.Customer, error) {
	c, err := s.model.GetCustomerByUUID(ctx, customerUUID)
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
func (s *FirebaseService) CreateAddress(ctx context.Context, customerUUID, typ, contactName, addr1 string, addr2 *string, city string, county *string, postcode string, country string) (*app.Address, error) {
	customerID, err := s.model.GetCustomerIDByUUID(ctx, customerUUID)
	if err != nil {
		return nil, err
	}

	a, err := s.model.CreateAddress(ctx, customerID, typ, contactName, addr1, addr2, city, county, postcode, country)
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
func (s *FirebaseService) GetAddress(ctx context.Context, addressUUID string) (*app.Address, error) {
	a, err := s.model.GetAddressByUUID(ctx, addressUUID)
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

func (s *FirebaseService) GetAddressOwner(ctx context.Context, addrUUID string) (*string, error) {
	customerUUID, err := s.model.GetAddressOwnerByUUID(ctx, addrUUID)
	if err != nil {
		return nil, err
	}

	return customerUUID, nil
}

// GetAddresses gets a slice of addresses for a given customer
func (s *FirebaseService) GetAddresses(ctx context.Context, customerUUID string) ([]*app.Address, error) {
	customerID, err := s.model.GetCustomerIDByUUID(ctx, customerUUID)
	if err != nil {
		return nil, err
	}

	al, err := s.model.GetAddresses(ctx, customerID)
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
func (s *FirebaseService) DeleteAddress(ctx context.Context, addrUUID string) error {
	err := s.model.DeleteAddressByUUID(ctx, addrUUID)
	if err != nil {
		return err
	}

	return nil
}
