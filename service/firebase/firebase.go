package firebase

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

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

// CartItem structure holds the details individual cart item
type CartItem struct {
	CartUUID  string    `json:"cart_uuid"`
	Sku       string    `json:"sku"`
	Qty       int       `json:"qty"`
	UnitPrice float64   `json:"unit_price"`
	Created   time.Time `json:"created"`
	Modified  time.Time `json:"modified"`
}

// Customer details
type Customer struct {
	CustomerUUID string    `json:"customer_uuid"`
	UID          string    `json:"uid"`
	Role         string    `json:"role"`
	Email        string    `json:"email"`
	Firstname    string    `json:"firstname"`
	Lastname     string    `json:"lastname"`
	Created      time.Time `json:"created"`
	Modified     time.Time `json:"modified"`
}

type PaginationQuery struct {
	OrderBy    string
	OrderDir   string
	Limit      int
	StartAfter string
}

type PaginationContext struct {
	Total     int    `json:"total"`
	FirstUUID string `json:"first_uuid"`
	LastUUID  string `json:"last_uuid"`
}

type PaginationResultSet struct {
	RContext PaginationContext
	RSet     interface{}
}

// CatalogProductAssoc maps products to leaf nodes in the catalogue hierarchy
type CatalogProductAssoc struct {
	Path     string    `json:"path"`
	SKU      string    `json:"sku"`
	Pri      int       `json:"pri"`
	Created  time.Time `json:"created"`
	Modified time.Time `json:"modified"`
}

// CustomerDevKey struct holding the details of a customer Developer Key including its bcrypt hash.
type CustomerDevKey struct {
	UUID         string    `json:"uuid"`
	Key          string    `json:"key"`
	CustomerUUID string    `json:"customer_uuid"`
	Created      time.Time `json:"created"`
	Modified     time.Time `json:"modified"`
}

// Address contains address information for a Customer
type Address struct {
	AddrUUID    string    `json:"addr_uuid"`
	Typ         string    `json:"typ"`
	ContactName string    `json:"contact_name"`
	Addr1       string    `json:"addr1"`
	Addr2       *string   `json:"addr2,omitempty"`
	City        string    `json:"city"`
	County      *string   `json:"county,omitempty"`
	Postcode    string    `json:"postcode"`
	Country     string    `json:"country"`
	Created     time.Time `json:"created"`
	Modified    time.Time `json:"modified"`
}

type ProductUpdate struct {
	EAN  string      `json:"ean" yaml:"ean"`
	URL  string      `json:"url" yaml:"url"`
	Name string      `json:"name" yaml:"name"`
	Data ProductData `json:"data" yaml:"data"`
}

type ProductCreate struct {
	SKU  string      `json:"sku" yaml:"sku"`
	EAN  string      `json:"ean" yaml:"ean"`
	URL  string      `json:"url" yaml:"url"`
	Name string      `json:"name" yaml:"name"`
	Data ProductData `json:"data" yaml:"data"`
}

type ProductData struct {
	Summary string `json:"summary" yaml:"summary"`
	Desc    string `json:"description" yaml:"description"`
	Spec    string `json:"specification" yaml:"specification"`
}

// Product contains all the fields that comprise a product in the catalog.
type Product struct {
	SKU      string      `json:"sku" yaml:"sku,omitempty"`
	EAN      string      `json:"ean" yaml:"ean"`
	URL      string      `json:"url" yaml:"url"`
	Name     string      `json:"name" yaml:"name"`
	Data     ProductData `json:"data" yaml:"data"`
	Created  time.Time   `json:"created,omitempty"`
	Modified time.Time   `json:"modified,omitempty"`
}

func NewService(model *postgres.PgModel, fbApp *firebase.App) *Service {
	return &Service{model, fbApp}
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
func (s *Service) AddItemToCart(ctx context.Context, cartUUID string, sku string, qty int) (*CartItem, error) {
	log.Debugf("s.AddItemToCart(%s, %s, %d) started", cartUUID, sku, qty)

	item, err := s.model.AddItemToCart(ctx, cartUUID, "default", sku, qty)
	if err != nil {
		return nil, err
	}

	log.Debug(item)

	sitem := CartItem{
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
func (s *Service) GetCartItems(ctx context.Context, cartUUID string) ([]*CartItem, error) {
	items, err := s.model.GetCartItems(ctx, cartUUID)
	if err != nil {
		return nil, err
	}

	results := make([]*CartItem, 0, 32)
	for _, v := range items {
		i := CartItem{
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
func (s *Service) UpdateCartItem(ctx context.Context, cartUUID string, sku string, qty int) (*CartItem, error) {
	item, err := s.model.UpdateItemByCartUUID(ctx, cartUUID, sku, qty)
	if err != nil {
		return nil, err
	}

	sitem := CartItem{
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
func (s *Service) CreateCustomer(ctx context.Context, role, email, password, firstname, lastname string) (*Customer, error) {
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

	ac := Customer{
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
func (s *Service) GetCustomers(ctx context.Context, pq *PaginationQuery) (*PaginationResultSet, error) {
	q := &postgres.PaginationQuery{
		OrderBy:    pq.OrderBy,
		OrderDir:   pq.OrderDir,
		Limit:      pq.Limit,
		StartAfter: pq.StartAfter,
	}
	prs, err := s.model.GetCustomers(ctx, q)
	if err != nil {
		return nil, err
	}

	results := make([]*Customer, 0)
	for _, v := range prs.RSet.([]*postgres.Customer) {
		c := Customer{
			CustomerUUID: v.CustomerUUID,
			UID:          v.UID,
			Role:         v.Role,
			Email:        v.Email,
			Firstname:    v.Firstname,
			Lastname:     v.Lastname,
			Created:      v.Created,
			Modified:     v.Modified,
		}
		results = append(results, &c)
	}

	aprs := &PaginationResultSet{
		RContext: PaginationContext{
			Total:     prs.RContext.Total,
			FirstUUID: prs.RContext.FirstUUID,
			LastUUID:  prs.RContext.LastUUID,
		},
		RSet: results,
	}
	return aprs, nil
}

// GetCustomer retrieves a customer by customer UUID
func (s *Service) GetCustomer(ctx context.Context, customerUUID string) (*Customer, error) {
	c, err := s.model.GetCustomerByUUID(ctx, customerUUID)
	if err != nil {
		return nil, err
	}

	ac := Customer{
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

func (s *Service) GetCustomerDevKey(ctx context.Context, uuid string) (*CustomerDevKey, error) {
	ak, err := s.model.GetCustomerDevKey(ctx, uuid)
	if err != nil {
		if err == sql.ErrNoRows {
		}
	}

	return &CustomerDevKey{
		UUID:         ak.UUID,
		Key:          ak.Key,
		CustomerUUID: ak.CustomerUUID,
		Created:      ak.Created,
		Modified:     ak.Modified,
	}, nil
}

// CreateProduct create a new product if the product SKU does not already exist.
func (s *Service) CreateProduct(ctx context.Context, pc *ProductCreate) (*Product, error) {
	pu := &postgres.ProductUpdate{
		EAN:  pc.EAN,
		URL:  pc.URL,
		Name: pc.Name,
		Data: postgres.ProductData{
			Summary: pc.Data.Summary,
			Desc:    pc.Data.Desc,
			Spec:    pc.Data.Spec,
		},
	}
	p, err := s.model.CreateProduct(ctx, pc.SKU, pu)
	if err != nil {
		return nil, errors.Wrapf(err, "create product %q failed", pc.SKU)
	}
	return &Product{
		SKU:  p.SKU,
		EAN:  p.EAN,
		URL:  p.URL,
		Name: p.Name,
		Data: ProductData{
			Summary: p.Data.Summary,
			Desc:    p.Data.Desc,
			Spec:    p.Data.Spec,
		},
		Created:  p.Created,
		Modified: p.Modified,
	}, nil
}

// difference returns the elements in `a` that aren't in `b`.
func difference(a, b []string) []string {
	mb := map[string]bool{}
	for _, x := range b {
		mb[x] = true
	}
	var ab []string
	for _, x := range a {
		if _, ok := mb[x]; !ok {
			ab = append(ab, x)
		}
	}
	return ab
}

// ProductsExist accepts a slice of product SKUs and divides them into
// two lists of those that can exist in the system and those that are
// missing.
func (s *Service) ProductsExist(ctx context.Context, skus []string) (exists, missing []string, err error) {
	exists, err = s.model.ProductsExist(ctx, skus)
	if err != nil {
		return nil, nil, errors.Wrap(err, "service: ProductsExist")
	}
	missing = difference(skus, exists)
	return exists, missing, nil
}

// GetProduct gets a product given the SKU.
func (s *Service) GetProduct(ctx context.Context, sku string) (*Product, error) {
	p, err := s.model.GetProduct(ctx, sku)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, errors.Wrapf(err, "get product %q failed", sku)
	}
	return &Product{
		SKU:  p.SKU,
		EAN:  p.EAN,
		URL:  p.URL,
		Name: p.Name,
		Data: ProductData{
			Summary: p.Data.Summary,
			Desc:    p.Data.Desc,
			Spec:    p.Data.Spec,
		},
		Created:  p.Created,
		Modified: p.Modified,
	}, nil
}

func marshalProduct(a *Product, m *postgres.Product) {
	a.SKU = m.SKU
	a.EAN = m.EAN
	a.URL = m.URL
	a.Name = m.Name
	a.Data.Summary = m.Data.Summary
	a.Data.Desc = m.Data.Desc
	a.Data.Spec = m.Data.Spec
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
func (s *Service) UpdateProduct(ctx context.Context, sku string, pu *ProductUpdate) (*Product, error) {
	update := &postgres.ProductUpdate{
		EAN:  pu.EAN,
		URL:  pu.URL,
		Name: pu.Name,
		Data: postgres.ProductData{
			Summary: pu.Data.Summary,
			Desc:    pu.Data.Desc,
			Spec:    pu.Data.Spec,
		},
	}
	p, err := s.model.UpdateProduct(ctx, sku, update)
	if err != nil {
		return nil, errors.Wrapf(err, "update product sku=%q failed", sku)
	}
	ap := &Product{}
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
		if err == sql.ErrNoRows {
			// if no key matches create a dummy apiKey struct
			// to ensure the compare hash happens. This mitigates against
			// timing attacks.
			ak = &postgres.CustomerDevKey{
				Key:  "none",
				Hash: "$2a$14$dRgjB9nBHoCs5txdVgN2EeVopE8rfZ7gLJNpLxw9GYq.u53FD00ny", // "nomatch"
			}
		} else {
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
func (s *Service) ListCustomersDevKeys(ctx context.Context, uuid string) ([]*CustomerDevKey, error) {
	customerID, err := s.model.GetCustomerIDByUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}

	apks, err := s.model.GetCustomerDevKeys(ctx, customerID)
	if err != nil {
		return nil, err
	}
	apiKeys := make([]*CustomerDevKey, 0, len(apks))
	for _, ak := range apks {
		c := CustomerDevKey{
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
func (s *Service) GenerateCustomerDevKey(ctx context.Context, uuid string) (*CustomerDevKey, error) {
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

	return &CustomerDevKey{
		Key:      ak.Key,
		Created:  ak.Created,
		Modified: ak.Modified,
	}, nil
}

// CreateAddress creates a new address for a customer
func (s *Service) CreateAddress(ctx context.Context, customerUUID, typ, contactName, addr1 string, addr2 *string, city string, county *string, postcode string, country string) (*Address, error) {
	customerID, err := s.model.GetCustomerIDByUUID(ctx, customerUUID)
	if err != nil {
		return nil, err
	}

	a, err := s.model.CreateAddress(ctx, customerID, typ, contactName, addr1, addr2, city, county, postcode, country)
	if err != nil {
		return nil, err
	}

	aa := Address{
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
func (s *Service) GetAddress(ctx context.Context, uuid string) (*Address, error) {
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

	aa := Address{
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
func (s *Service) GetAddresses(ctx context.Context, customerUUID string) ([]*Address, error) {
	customerID, err := s.model.GetCustomerIDByUUID(ctx, customerUUID)
	if err != nil {
		return nil, err
	}

	al, err := s.model.GetAddresses(ctx, customerID)
	if err != nil {
		return nil, err
	}

	results := make([]*Address, 0, 32)
	for _, v := range al {
		i := Address{
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

// ReplaceCatalog takes a root tree Node and converts it to a nested set
// representation before calling the model to persist the replacement
// catalog.
func (s *Service) ReplaceCatalog(ctx context.Context, root *nestedset.Node) error {
	root.GenerateNestedSet(1, 0, "")
	ns := make([]*nestedset.NestedSetNode, 0, 128)
	root.NestedSet(&ns)
	err := s.model.BatchCreateNestedSet(ctx, ns)
	if err != nil {
		return errors.Wrap(err, "service: replace catalog")
	}
	return nil
}

// GetCatalog returns the catalog as a hierarchy of nodes.
func (s *Service) GetCatalog(ctx context.Context) (*nestedset.Node, error) {
	ns, err := s.model.GetCatalogNestedSet(ctx)
	if err != nil {
		return nil, err
	}
	tree := nestedset.BuildTree(ns)
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

// BatchCreateCatalogProductAssocs creates a set of catalog product
// associations either completing with all or failing with none
// being added.
func (s *Service) BatchCreateCatalogProductAssocs(ctx context.Context, cpas map[string][]string) error {
	err := s.model.BatchCreateCatalogProductAssocs(ctx, cpas)
	if err != nil {
		return errors.Wrap(err, "service: BatchCreateCatalogProductAssocs")
	}
	return nil
}

// CreateCatalogProductAssocs associates an existing product to a catalog entry.
func (s *Service) CreateCatalogProductAssocs(ctx context.Context, path, sku string) (*CatalogProductAssoc, error) {
	cpa, err := s.model.CreateCatalogProductAssoc(ctx, path, sku)
	if err != nil {
		return nil, errors.Wrapf(err, "service: create catalog product assoc sku=%q", sku)
	}
	scpa := CatalogProductAssoc{
		Path:     cpa.Path,
		SKU:      cpa.SKU,
		Pri:      cpa.Pri,
		Created:  cpa.Created,
		Modified: cpa.Modified,
	}
	return &scpa, nil
}

// HasCatalogProductAssocs returns true if any catalog product associations
// exist.
func (s *Service) HasCatalogProductAssocs(ctx context.Context) (bool, error) {
	has, err := s.model.HasCatalogProductAssocs(ctx)
	if err != nil {
		return false, errors.Wrap(err, "service: has catalog product assocs")
	}
	return has, nil
}

// GetCatalogProductAssocs returns the catalog product associations
func (s *Service) GetCatalogProductAssocs(ctx context.Context) ([]*CatalogProductAssoc, error) {
	cpo, err := s.model.GetCatalogProductAssocs(ctx)
	if err != nil {
		return nil, err
	}
	results := make([]*CatalogProductAssoc, 0, 32)
	for _, v := range cpo {
		i := CatalogProductAssoc{
			Path:     v.Path,
			SKU:      v.SKU,
			Pri:      v.Pri,
			Created:  v.Created,
			Modified: v.Modified,
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

// DeleteCatalogProductAssocs delete all catalog product associations.
func (s *Service) DeleteCatalogProductAssocs(ctx context.Context) (affected int64, err error) {
	n, err := s.model.DeleteCatalogProductAssocs(ctx)
	if err != nil {
		return 0, errors.Wrapf(err, "delete catalog product assocs")
	}
	return n, nil
}
