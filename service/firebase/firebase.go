package firebase

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"

	"bitbucket.org/andyfusniakteam/ecom-api-go/app"
	"bitbucket.org/andyfusniakteam/ecom-api-go/model/postgres"
	"bitbucket.org/andyfusniakteam/ecom-api-go/utils/nestedset"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"github.com/btcsuite/btcutil/base58"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

// Service firebase implementation
type Service struct {
	model *postgres.PgModel
	fbApp *firebase.App
}

func NewService(model *postgres.PgModel, fbApp *firebase.App) (*Service, error) {
	s := Service{model, fbApp}
	return &s, nil
}

// Auth accepts a JSON Web Token, usually passed from the HTTP client and returns a auth.Token if valid or nil if
func (s *Service) Authenticate(ctx context.Context, jwt string) (*auth.Token, error) {
	authClient, err := s.fbApp.Auth(ctx)
	if err != nil {
		return nil, err
	}

	token, err := authClient.VerifyIDToken(ctx, jwt)
	if err != nil {
		return nil, err
	}
	return token, nil
}

// CreateCart generates a new random UUID to be used for subseqent cart calls
func (s *Service) CreateCart(ctx context.Context) (*string, error) {
	log.Debug("s.CreateCart() started")

	strptr, err := s.model.CreateCart(ctx)
	if err != nil {
		return nil, err
	}

	log.Debugf("s.CreateCart() returned %s", *strptr)
	return strptr, nil
}

// AddItemToCart adds a single item to a given cart
func (s *Service) AddItemToCart(ctx context.Context, cartUUID string, sku string, qty int) (*app.CartItem, error) {
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
func (s *Service) GetCartItems(ctx context.Context, cartUUID string) ([]*app.CartItem, error) {
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
func (s *Service) UpdateCartItem(ctx context.Context, cartUUID string, sku string, qty int) (*app.CartItem, error) {
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
func (s *Service) DeleteCartItem(ctx context.Context, cartUUID string, sku string) (count int64, err error) {
	count, err = s.model.DeleteCartItem(ctx, cartUUID, sku)
	if err != nil {
		return -1, err
	}
	return count, nil
}

// EmptyCartItems empties the cart of all items but not coupons
func (s *Service) EmptyCartItems(ctx context.Context, cartUUID string) (err error) {
	return s.model.EmptyCartItems(ctx, cartUUID)
}

// CreateRootIfNotExists create the root user if the root super admin does not exit.
func (s *Service) CreateRootIfNotExists(ctx context.Context, email, password string) error {
	authClient, err := s.fbApp.Auth(ctx)
	if err != nil {
		return err
	}

	userRecord, err := authClient.GetUserByEmail(ctx, email)
	if err != nil {
		if auth.IsUserNotFound(err) {
			user := (&auth.UserToCreate{}).
				Email(email).
				EmailVerified(false).
				Password(password).
				DisplayName("Root Superuser").
				Disabled(false)
			userRecord, err = authClient.CreateUser(ctx, user)
			if err != nil {
				return err
			}
			log.Infof("created root superuser email=%s", email)

			// Set the custom claims for the root user
			err = authClient.SetCustomUserClaims(ctx, userRecord.UID, map[string]interface{}{
				"role": "root",
			})
			if err != nil {
				return fmt.Errorf("failed to set custom claims for root user: %v", err)
			}
			log.Info("set custom claims for root superuser role=root")
			return nil
		}
		return err
	}

	log.Infof("root superuser email=%s already exists", email)
	return nil
}

// CreateCustomer creates a new customer
func (s *Service) CreateCustomer(ctx context.Context, role, email, password, firstname, lastname string) (*app.Customer, error) {
	log.Debugf("s.CreateCustomer(%s, %s, %s, %s, %s) started", role, email, "*****", firstname, lastname)

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
		Role              string `json:"role"`
		Email             string `json:"email"`
		DisplayName       string `json:"display_name"`
		CreationTimestamp int64  `json:"creation_timestamp"`
	}{
		UID:               userRecord.UID,
		Role:              role,
		Email:             userRecord.Email,
		DisplayName:       userRecord.DisplayName,
		CreationTimestamp: userRecord.UserMetadata.CreationTimestamp,
	}
	b, err := json.Marshal(&payload)
	if err != nil {
		log.Errorf("json.Marshal failed: %v", err)
	}
	log.Debugf("payload marshalled to string %s", b)

	c, err := s.model.CreateCustomer(ctx, userRecord.UID, role, email, firstname, lastname)
	if err != nil {
		return nil, fmt.Errorf("model.CreateCustomer(%s, %s, %s, %s, %s) failed: %v", userRecord.UID, role, email, firstname, lastname, err)
	}

	// Set the custom claims for this user
	err = authClient.SetCustomUserClaims(ctx, c.UID, map[string]interface{}{
		"cuuid": c.CustomerUUID,
		"role":  role,
	})
	if err != nil {
		return nil, fmt.Errorf("set custom claims for uid=%s customer_uuid=%s role=%s failed: %v", c.UID, c.CustomerUUID, role, err)
	}

	ac := app.Customer{
		CustomerUUID: c.CustomerUUID,
		UID:          c.UID,
		Role:         c.Role,
		Email:        c.Email,
		Firstname:    c.Firstname,
		Lastname:     c.Lastname,
		Created:      c.Created,
		Modified:     c.Modified,
	}

	log.Debugf("%+v", ac)
	return &ac, nil
}

// GetCustomers gets customers with pagination.
func (s *Service) GetCustomers(ctx context.Context, q *app.PaginationQuery) (*app.PaginationResultSet, error) {
	mq := &postgres.PaginationQuery{
		OrderBy:    q.OrderBy,
		OrderDir:   q.OrderDir,
		Limit:      q.Limit,
		StartAfter: q.StartAfter,
	}
	prs, err := s.model.GetCustomers(ctx, mq)
	if err != nil {
		return nil, err
	}

	results := make([]*app.Customer, 0)
	for _, v := range prs.RSet.([]*postgres.Customer) {
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

	aprs := &app.PaginationResultSet{
		RContext: app.PaginationContext{
			Total:     prs.RContext.Total,
			FirstUUID: prs.RContext.FirstUUID,
			LastUUID:  prs.RContext.LastUUID,
		},
		RSet: results,
	}
	return aprs, nil
}

// GetCustomer retrieves a customer by customer UUID
func (s *Service) GetCustomer(ctx context.Context, customerUUID string) (*app.Customer, error) {
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

func (s *Service) GetCustomerDevKey(ctx context.Context, uuid string) (*app.CustomerDevKey, error) {
	ak, err := s.model.GetCustomerDevKey(ctx, uuid)
	if err != nil {
		if err == sql.ErrNoRows {
		}
	}

	return &app.CustomerDevKey{
		UUID:         ak.UUID,
		Key:          ak.Key,
		CustomerUUID: ak.CustomerUUID,
		Created:      ak.Created,
		Modified:     ak.Modified,
	}, nil
}

// CreateProduct create a new product if the product SKU does not already exist.
func (s *Service) CreateProduct(ctx context.Context, pc *app.ProductCreate) (*app.Product, error) {
	p, err := s.model.CreateProduct(ctx, pc.SKU)
	if err != nil {
		return nil, errors.Wrapf(err, "create product %q failed", pc.SKU)
	}
	return &app.Product{
		SKU:      p.SKU,
		EAN:      p.EAN,
		URL:      p.URL,
		Name:     p.Name,
		Created:  p.Created,
		Modified: p.Modified,
	}, nil
}

// GetProduct gets a product given the SKU.
func (s *Service) GetProduct(ctx context.Context, sku string) (*app.Product, error) {
	p, err := s.model.GetProduct(ctx, sku)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, errors.Wrapf(err, "get product %q failed", sku)
	}
	return &app.Product{
		SKU:      p.SKU,
		EAN:      p.EAN,
		URL:      p.URL,
		Name:     p.Name,
		Created:  p.Created,
		Modified: p.Modified,
	}, nil
}

func marshalProduct(a *app.Product, m *postgres.Product) {
	a.SKU = m.SKU
	a.EAN = m.EAN
	a.URL = m.URL
	a.Name = m.Name
	a.Created = m.Created
	a.Modified = m.Modified
	return
}

// ProductExists return true if the given product exists.
func (s *Service) ProductExists(ctx context.Context, sku string) (bool, error) {
	exists, err := s.model.ProductExists(ctx, sku)
	if err != nil {
		return false, errors.Wrapf(err, "ProductExists(ctx, %q) failed", sku)
	}
	return exists, nil
}

// UpdateProduct updates a product by SKU.
func (s *Service) UpdateProduct(ctx context.Context, sku string, pu *app.ProductUpdate) (*app.Product, error) {
	update := &postgres.ProductUpdate{
		EAN:  pu.EAN,
		URL:  pu.URL,
		Name: pu.Name,
	}
	p, err := s.model.UpdateProduct(ctx, sku, update)
	if err != nil {
		return nil, errors.Wrapf(err, "update product sku=%q failed", sku)
	}
	ap := &app.Product{}
	marshalProduct(ap, p)
	return ap, nil
}

func (s *Service) DeleteProduct(ctx context.Context, sku string) error {
	err := s.model.DeleteProduct(ctx, sku)
	if err != nil {
		return errors.Wrapf(err, "delete product sku=%q failed", sku)
	}
	return nil
}

// SignInWithDevKey checks the apiKey hash using bcrypt.
func (s *Service) SignInWithDevKey(ctx context.Context, key string) (string, error) {
	ak, err := s.model.GetCustomerDevKeyByDevKey(ctx, key)
	if err != nil {
		fmt.Println(err)
		if err == sql.ErrNoRows {
			// if no key matches create a dummy apiKey struct
			// to ensure the compare hash happens. This mitigates against
			// timing attacks.
			ak = &postgres.CustomerDevKey{
				Key:  "none",
				Hash: "$2a$14$dRgjB9nBHoCs5txdVgN2EeVopE8rfZ7gLJNpLxw9GYq.u53FD00ny", // "nomatch"
			}
		} else {
			fmt.Println("here 2")
			return "", errors.Wrap(err, "s.model.GetCustomerDevKeyByDevKey(ctx, key)")
		}
	}

	err = bcrypt.CompareHashAndPassword([]byte(ak.Hash), []byte(ak.Key))
	if err != nil {
		return "", errors.Wrap(err, "bcrypt.CompareHashAndPassword([]byte(ak.Hash), []byte(ak.Key))")
	}

	customer, err := s.model.GetCustomerByID(ctx, ak.CustomerID)
	if err != nil {
		return "", errors.Wrapf(err, "s.model.GetCustomerByID(ctx, customerID=%q)", ak.CustomerID)
	}

	authClient, err := s.fbApp.Auth(ctx)
	if err != nil {
		return "", errors.Wrap(err, "s.fbApp.Auth(ctx)")
	}

	userRecord, err := authClient.GetUserByEmail(ctx, customer.Email)
	if err != nil {
		return "", errors.Wrapf(err, "authClient.GetUserByEmail(ctx, email=%q)", customer.Email)
	}

	token, err := authClient.CustomToken(ctx, userRecord.UID)
	if err != nil {
		return "", errors.Wrapf(err, "authClient.CustomToken(ctx, uid=%q)", userRecord.UID)
	}
	return token, nil
}

// ListCustomersDevAPIKeys gets all API Keys for a customer.
func (s *Service) ListCustomersDevKeys(ctx context.Context, uuid string) ([]*app.CustomerDevKey, error) {
	customerID, err := s.model.GetCustomerIDByUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}

	apks, err := s.model.GetCustomerDevKeys(ctx, customerID)
	if err != nil {
		return nil, err
	}
	apiKeys := make([]*app.CustomerDevKey, 0, len(apks))
	for _, ak := range apks {
		c := app.CustomerDevKey{
			UUID:         ak.UUID,
			Key:          ak.Key,
			CustomerUUID: uuid,
			Created:      ak.Created,
			Modified:     ak.Modified,
		}
		apiKeys = append(apiKeys, &c)
	}
	return apiKeys, nil
}

// GenerateCustomerAPIKey creates a new API Key for a customer
func (s *Service) GenerateCustomerDevKey(ctx context.Context, uuid string) (*app.CustomerDevKey, error) {
	customerID, err := s.model.GetCustomerIDByUUID(ctx, uuid)
	if err != nil {
		return nil, errors.Wrapf(err, "s.model.GetCustomerIDByUUID(ctx, %q)", uuid)
	}
	data := make([]byte, 32)
	_, err = rand.Read(data)
	if err != nil {
		return nil, errors.Wrap(err, "rand.Read(data)")
	}

	ak, err := s.model.CreateCustomerDevKey(ctx, customerID, base58.Encode(data))
	if err != nil {
		return nil, errors.Wrapf(err, "s.model.CreateCustomerDevKey(ctx, customerID=%q, ...)", customerID)
	}

	return &app.CustomerDevKey{
		Key:      ak.Key,
		Created:  ak.Created,
		Modified: ak.Modified,
	}, nil
}

// CreateAddress creates a new address for a customer
func (s *Service) CreateAddress(ctx context.Context, customerUUID, typ, contactName, addr1 string, addr2 *string, city string, county *string, postcode string, country string) (*app.Address, error) {
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
func (s *Service) GetAddress(ctx context.Context, uuid string) (*app.Address, error) {
	addr, err := s.model.GetAddressByUUID(ctx, uuid)
	if err != nil {
		if s.model.IsNotExist(err) {
			if ne, ok := err.(*ResourceError); ok {
				return nil, &ResourceError{
					Op:       "GetAddress",
					Resource: "address",
					UUID:     uuid,
					Err:      ne.Err,
				}
			}
		}
		return nil, err
	}

	aa := app.Address{
		AddrUUID:    addr.AddrUUID,
		Typ:         addr.Typ,
		ContactName: addr.ContactName,
		Addr1:       addr.Addr1,
		Addr2:       addr.Addr2,
		City:        addr.City,
		County:      addr.County,
		Postcode:    addr.Postcode,
		Country:     addr.Country,
		Created:     addr.Created,
		Modified:    addr.Modified,
	}
	return &aa, nil
}

func (s *Service) GetAddressOwner(ctx context.Context, uuid string) (*string, error) {
	customerUUID, err := s.model.GetAddressOwnerByUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}
	return customerUUID, nil
}

// GetAddresses gets a slice of addresses for a given customer
func (s *Service) GetAddresses(ctx context.Context, customerUUID string) ([]*app.Address, error) {
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
func (s *Service) DeleteAddress(ctx context.Context, addrUUID string) error {
	err := s.model.DeleteAddressByUUID(ctx, addrUUID)
	if err != nil {
		return err
	}

	return nil
}

// GetCatalog returns the catalog in nested set representation.
func (s *Service) GetCatalog(ctx context.Context) ([]*nestedset.NestedSetNode, error) {
	ns, err := s.model.GetCatalogNestedSet(ctx)
	if err != nil {
		return nil, err
	}
	return ns, nil
}

// GetCatalogProductAssocs returns the catalog product associations
func (s *Service) GetCatalogProductAssocs(ctx context.Context) ([]*app.CatalogProductAssoc, error) {
	cpo, err := s.model.GetCatalogProductAssocs(ctx)
	if err != nil {
		return nil, err
	}

	results := make([]*app.CatalogProductAssoc, 0, 32)
	for _, v := range cpo {
		i := app.CatalogProductAssoc{
			CatalogID: v.CatalogID,
			ProductID: v.ProductID,
			Path:      v.Path,
			SKU:       v.SKU,
			Pri:       v.Pri,
			Created:   v.Created,
			Modified:  v.Modified,
		}
		results = append(results, &i)
	}
	return results, nil
}

// UpdateCatalogProductAssocs updates the catalog product associations
func (s *Service) UpdateCatalogProductAssocs(ctx context.Context, cpo []*postgres.CatalogProductAssoc) error {
	err := s.model.UpdateCatalogProductAssocs(ctx, cpo)
	if err != nil {
		return err
	}
	return nil
}
