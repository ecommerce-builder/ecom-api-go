package firebase

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"strings"
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
	CartUUID  string    `json:"uuid"`
	Sku       string    `json:"sku"`
	Qty       int       `json:"qty"`
	UnitPrice float64   `json:"unit_price"`
	Created   time.Time `json:"created"`
	Modified  time.Time `json:"modified"`
}

// Customer details
type Customer struct {
	UUID      string    `json:"uuid"`
	UID       string    `json:"uid"`
	Role      string    `json:"role"`
	Email     string    `json:"email"`
	Firstname string    `json:"firstname"`
	Lastname  string    `json:"lastname"`
	Created   time.Time `json:"created"`
	Modified  time.Time `json:"modified"`
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
	AddrUUID    string    `json:"uuid"`
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
	Path string      `json:"path" yaml:"path"`
	Name string      `json:"name" yaml:"name"`
	Data ProductData `json:"data" yaml:"data"`
}

type ProductCreate struct {
	SKU  string      `json:"sku" yaml:"sku"`
	EAN  string      `json:"ean" yaml:"ean"`
	Path string      `json:"path" yaml:"path"`
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
	SKU      string                     `json:"sku" yaml:"sku,omitempty"`
	EAN      string                     `json:"ean" yaml:"ean"`
	Path     string                     `json:"path" yaml:"path"`
	Name     string                     `json:"name" yaml:"name"`
	Data     ProductData                `json:"data" yaml:"data"`
	Images   []*Image                   `json:"images" yaml:"images"`
	Pricing  map[string]*ProductPricing `json:"pricing" yaml:"pricing"`
	Created  time.Time                  `json:"created,omitempty"`
	Modified time.Time                  `json:"modified,omitempty"`
}

func NewService(model *postgres.PgModel, fbApp *firebase.App) *Service {
	return &Service{model, fbApp}
}

// Authenticate accepts a JSON Web Token, usually passed from the HTTP client and returns a auth.Token if valid or nil if
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

// ListAdmins returns a list of administrators
func (s *Service) ListAdmins(ctx context.Context) ([]*Customer, error) {
	admins, err := s.model.GetAllAdmins(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "GetAllAdmins(ctx) failed")
	}
	adms := make([]*Customer, 0, 8)
	for _, c := range admins {
		customer := &Customer{
			UUID:      c.UUID,
			UID:       c.UID,
			Role:      c.Role,
			Email:     c.Email,
			Firstname: c.Firstname,
			Lastname:  c.Lastname,
			Created:   c.Created,
			Modified:  c.Modified,
		}
		adms = append(adms, customer)
	}
	return adms, nil
}

// DeleteAdmin deletes an administrator.
func (s *Service) DeleteAdmin(ctx context.Context, uuid string) error {
	authClient, err := s.fbApp.Auth(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to firebase auth client")
	}
	admin, err := s.model.GetAdmin(ctx, uuid)
	if err != nil {
		return errors.Wrapf(err, "get admin failed for uuid=%q", uuid)
	}
	if admin.Role != "admin" {
		return errors.Errorf("customer record with uuid=%q does not have role admin or could not be found", uuid)
	}
	if err = s.model.DeleteAdminByUUID(ctx, uuid); err != nil {
		return errors.Wrapf(err, "delete customer by uuid for uuid=%q failed", uuid)
	}
	if err = authClient.DeleteUser(ctx, admin.UID); err != nil {
		return errors.Wrapf(err, "firebase auth delete user failed for uid=%q", admin.UID)
	}
	return nil
}

// CreateCart generates a new random UUID to be used for subseqent cart calls
func (s *Service) CreateCart(ctx context.Context) (*string, error) {
	strptr, err := s.model.CreateCart(ctx)
	if err != nil {
		return nil, err
	}
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
		return errors.Wrap(err, "create root if not exists")
	}
	_, err = authClient.GetUserByEmail(ctx, email)
	if err != nil {
		if auth.IsUserNotFound(err) {
			customer, err := s.CreateCustomer(ctx, "root", email, password, "Super", "User")
			if err != nil {
				return errors.Wrap(err, "create customer for root user failed")
			}
			_, err = s.GenerateCustomerDevKey(ctx, customer.UUID)
			if err != nil {
				return errors.Wrap(err, "generate customer devkey failed")
			}
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

	c, err := s.model.CreateCustomer(ctx, userRecord.UID, role, email, firstname, lastname)
	if err != nil {
		return nil, fmt.Errorf("model.CreateCustomer(%s, %s, %s, %s, %s) failed: %v", userRecord.UID, role, email, firstname, lastname, err)
	}

	// Set the custom claims for this user
	err = authClient.SetCustomUserClaims(ctx, c.UID, map[string]interface{}{
		"cuuid": c.UUID,
		"role":  role,
	})
	if err != nil {
		return nil, fmt.Errorf("set custom claims for uid=%s uuid=%s role=%s failed: %v", c.UID, c.UUID, role, err)
	}

	ac := Customer{
		UUID:      c.UUID,
		UID:       c.UID,
		Role:      c.Role,
		Email:     c.Email,
		Firstname: c.Firstname,
		Lastname:  c.Lastname,
		Created:   c.Created,
		Modified:  c.Modified,
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
			UUID:      v.UUID,
			UID:       v.UID,
			Role:      v.Role,
			Email:     v.Email,
			Firstname: v.Firstname,
			Lastname:  v.Lastname,
			Created:   v.Created,
			Modified:  v.Modified,
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
		UUID:      c.UUID,
		UID:       c.UID,
		Email:     c.Email,
		Firstname: c.Firstname,
		Lastname:  c.Lastname,
		Created:   c.Created,
		Modified:  c.Modified,
	}
	return &ac, nil
}

// GetCustomerDevKey returns a CustomerDevKey for the customer with the given UUID.
func (s *Service) GetCustomerDevKey(ctx context.Context, uuid string) (*CustomerDevKey, error) {
	ak, err := s.model.GetCustomerDevKey(ctx, uuid)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
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
		Path: pc.Path,
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
		Path: p.Path,
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
		if err == postgres.ErrProductNotFound {
			return nil, err
		}
		return nil, errors.Wrapf(err, "model: GetProduct(ctx, %q) failed", sku)
	}
	images, err := s.ListProductImages(ctx, sku)
	if err != nil {
		return nil, errors.Wrapf(err, "service: ListProductImages(ctx, %q)", sku)
	}
	return &Product{
		SKU:  p.SKU,
		EAN:  p.EAN,
		Path: p.Path,
		Name: p.Name,
		Data: ProductData{
			Summary: p.Data.Summary,
			Desc:    p.Data.Desc,
			Spec:    p.Data.Spec,
		},
		Images:   images,
		Created:  p.Created,
		Modified: p.Modified,
	}, nil
}

// ListProducts returns a slice of all product SKUS.
func (s *Service) ListProducts(ctx context.Context) ([]string, error) {
	products, err := s.model.GetProducts(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "service: GetProduct")
	}
	skus := make([]string, 0, 256)
	for _, p := range products {
		skus = append(skus, p.SKU)
	}
	return skus, nil
}

func marshalProduct(a *Product, m *postgres.Product) {
	a.SKU = m.SKU
	a.EAN = m.EAN
	a.Path = m.Path
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
		Path: pu.Path,
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

// DeleteProduct deletes the product with the given SKU.
func (s *Service) DeleteProduct(ctx context.Context, sku string) error {
	err := s.model.DeleteProduct(ctx, sku)
	if err != nil {
		return errors.Wrapf(err, "delete product sku=%q failed", sku)
	}
	return nil
}

// ProductPricing contains pricing information for a single SKU and tier ref.
type ProductPricing struct {
	SKU       string    `json:"sku"`
	TierRef   string    `json:"tier_ref"`
	UnitPrice float64   `json:"unit_price"`
	Created   time.Time `json:"created"`
	Modified  time.Time `json:"modified"`
}

// GetTierPricing returns a ProductPricing for the product with the
// given SKU and tier ref.
func (s *Service) GetTierPricing(ctx context.Context, sku, ref string) (*ProductPricing, error) {
	p, err := s.model.GetProductPricingBySKUAndTier(ctx, sku, ref)
	if err != nil {
		return nil, errors.Wrap(err, "GetProductPricingBySKUAndTier failed")
	}
	pricing := ProductPricing{
		TierRef:   p.TierRef,
		SKU:       p.SKU,
		UnitPrice: p.UnitPrice,
		Created:   p.Created,
		Modified:  p.Modified,
	}
	fmt.Println(pricing)
	return &pricing, nil
}

// ProductTierPricing contains pricing information for all tiers of
// a given SKU.
type ProductTierPricing struct {
	UnitPrice float64 `json:"unit_price"`
}

// ListPricingBySKU returns a map of tier to ProductTierPricings.
func (s *Service) ListPricingBySKU(ctx context.Context, sku string) (map[string]*ProductTierPricing, error) {
	plist, err := s.model.GetProductPricingBySKU(ctx, sku)
	if err != nil {
		return nil, errors.Wrap(err, "GetProductPricingBySKU failed")
	}
	pmap := make(map[string]*ProductTierPricing)
	for _, p := range plist {
		ptp := ProductTierPricing{
			UnitPrice: p.UnitPrice,
		}
		if _, ok := pmap[p.TierRef]; !ok {
			pmap[p.TierRef] = &ptp
		}
	}
	return pmap, nil
}

// ProductSKUPricing contains pricing information for all SKUs of given tier.
type ProductSKUPricing struct {
	UnitPrice float64 `json:"unit_price"`
}

// ListPricingByTier returns a map of SKU to ProductSKUPricings.
func (s *Service) ListPricingByTier(ctx context.Context, ref string) (map[string]*ProductSKUPricing, error) {
	plist, err := s.model.GetProductPricingByTier(ctx, ref)
	if err != nil {
		return nil, errors.Wrap(err, "GetProductPricingByTier failed")
	}
	pmap := make(map[string]*ProductSKUPricing)
	for _, p := range plist {
		ptp := ProductSKUPricing{
			UnitPrice: p.UnitPrice,
		}
		if _, ok := pmap[p.SKU]; !ok {
			pmap[p.SKU] = &ptp
		}
	}
	return pmap, nil
}

// ErrTierPricingNotFound is where no tier pricing could be found
// for the given sku and tier ref.
var ErrTierPricingNotFound = errors.New("service: tier pricing not found")

// UpdateTierPricing updates the tier pricing for the given sku and tier ref.
// If the produt pricing is not found returns nil, nil.
func (s *Service) UpdateTierPricing(ctx context.Context, sku, ref string, unitPrice float64) (*ProductPricing, error) {
	p, err := s.model.GetProductPricingBySKUAndTier(ctx, sku, ref)
	if err != nil {
		return nil, errors.Wrapf(err, "GetProductPricingBySKUAndTier(ctx, %q, %q) failed", sku, ref)
	}
	if p == nil {
		return nil, ErrTierPricingNotFound
	}
	p, err = s.model.UpdateTierPricing(ctx, sku, ref, unitPrice)
	if err != nil {
		return nil, errors.Wrapf(err, "UpdateTierPricing(ctx, %q, %q, %d) failed", sku, ref, unitPrice)
	}
	pricing := ProductPricing{
		SKU:       p.SKU,
		TierRef:   p.TierRef,
		UnitPrice: p.UnitPrice,
		Created:   p.Created,
		Modified:  p.Modified,
	}
	return &pricing, nil
}

// DeleteTierPricing deletes a tier pricing by SKU and tier ref.
func (s *Service) DeleteTierPricing(ctx context.Context, sku, ref string) error {
	if err := s.model.DeleteProductPricingBySKUAndTier(ctx, sku, ref); err != nil {
		return errors.Wrapf(err, "DeleteProductPricingBySKUAndTier(ctx, %q, %q) failed", sku, ref)
	}
	return nil
}

// SignInWithDevKey checks the apiKey hash using bcrypt.
func (s *Service) SignInWithDevKey(ctx context.Context, key string) (customToken string, customer *Customer, err error) {
	ak, err := s.model.GetCustomerDevKeyByDevKey(ctx, key)
	if err != nil {
		if err == sql.ErrNoRows {
			// if no key matches create a dummy apiKey struct
			// to ensure the compare hash happens. This mitigates against
			// timing attacks.
			ak = &postgres.CustomerDevKeyFull{
				Key:  "none",
				Hash: "$2a$14$dRgjB9nBHoCs5txdVgN2EeVopE8rfZ7gLJNpLxw9GYq.u53FD00ny", // "nomatch"
			}
		} else {
			return "", nil, errors.Wrap(err, "s.model.GetCustomerDevKeyByDevKey(ctx, key)")
		}
	}

	err = bcrypt.CompareHashAndPassword([]byte(ak.Hash), []byte(ak.Key))
	if err != nil {
		return "", nil, errors.Wrap(err, "bcrypt.CompareHashAndPassword([]byte(ak.Hash), []byte(ak.Key))")
	}

	cust, err := s.model.GetCustomerByUUID(ctx, ak.CustomerUUID)
	if err != nil {
		return "", nil, errors.Wrapf(err, "s.model.GetCustomerByUUID(ctx, customerID=%q)", ak.CustomerUUID)
	}

	authClient, err := s.fbApp.Auth(ctx)
	if err != nil {
		return "", nil, errors.Wrap(err, "s.fbApp.Auth(ctx)")
	}

	userRecord, err := authClient.GetUserByEmail(ctx, cust.Email)
	if err != nil {
		return "", nil, errors.Wrapf(err, "authClient.GetUserByEmail(ctx, email=%q)", cust.Email)
	}

	token, err := authClient.CustomToken(ctx, userRecord.UID)
	if err != nil {
		return "", nil, errors.Wrapf(err, "authClient.CustomToken(ctx, uid=%q)", userRecord.UID)
	}
	customer = &Customer{
		UUID:      cust.UUID,
		UID:       cust.UID,
		Role:      cust.Role,
		Email:     cust.Email,
		Firstname: cust.Firstname,
		Lastname:  cust.Lastname,
		Created:   cust.Created,
		Modified:  cust.Modified,
	}
	return token, customer, nil
}

// ListCustomersDevKeys gets all API Keys for a customer.
func (s *Service) ListCustomersDevKeys(ctx context.Context, uuid string) ([]*CustomerDevKey, error) {
	customerID, err := s.model.GetCustomerIDByUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}

	rows, err := s.model.GetCustomerDevKeys(ctx, customerID)
	if err != nil {
		return nil, err
	}
	apiKeys := make([]*CustomerDevKey, 0, len(rows))
	for _, row := range rows {
		c := CustomerDevKey{
			UUID:         row.UUID,
			Key:          row.Key,
			CustomerUUID: uuid,
			Created:      row.Created,
			Modified:     row.Modified,
		}
		apiKeys = append(apiKeys, &c)
	}
	return apiKeys, nil
}

// GenerateCustomerDevKey creates a new API Key for a customer
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

// GetAddressOwner returns the Customer that owns the address with the given UUID.
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

// UpdateCatalog takes a root tree Node and converts it to a nested set
// representation before calling the model to persist the replacement
// catalog.
func (s *Service) UpdateCatalog(ctx context.Context, root *nestedset.Node) error {
	root.GenerateNestedSet(1, 0, "")
	ns := make([]*nestedset.NestedSetNode, 0, 128)
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
func buildCategoryTree(nestedset []*nestedset.NestedSetNode, cmap map[string][]string) *Category {
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

// A SKU handles unique product SKUs.
type SKU string

// An AssocProduct holds details of a product in the context of an AssocSet.
type AssocProduct struct {
	SKU      string    `json:"sku"`
	Created  time.Time `json:"created"`
	Modified time.Time `json:"modified"`
}

// Assoc details a catalog association including products.
type Assoc struct {
	Path     string         `json:"path"`
	Products []AssocProduct `json:"products"`
}

// GetCatalogAssocs returns the catalog product associations
func (s *Service) GetCatalogAssocs(ctx context.Context) (map[string]*Assoc, error) {
	cpo, err := s.model.GetCatalogProductAssocs(ctx)
	if err != nil {
		return nil, err
	}
	assocs := make(map[string]*Assoc)
	for _, v := range cpo {
		if _, ok := assocs[v.Path]; !ok {
			assocs[v.Path] = &Assoc{
				Path:     v.Path,
				Products: make([]AssocProduct, 0),
			}
		}
		p := AssocProduct{
			SKU:      v.SKU,
			Created:  v.Created,
			Modified: v.Modified,
		}
		assocs[v.Path].Products = append(assocs[v.Path].Products, p)
	}
	return assocs, nil
}

// UpdateCatalogProductAssocs updates the catalog product associations
// func (s *Service) UpdateCatalogProductAssocs(ctx context.Context, cpo []*postgres.catalogProductAssoc) error {
// 	err := s.model.UpdateCatalogProductAssocs(ctx, cpo)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// DeleteCatalogAssocs delete all catalog product associations.
func (s *Service) DeleteCatalogAssocs(ctx context.Context) (affected int64, err error) {
	n, err := s.model.DeleteCatalogProductAssocs(ctx)
	if err != nil {
		return 0, errors.Wrapf(err, "service: delete catalog assocs")
	}
	return n, nil
}

// Image represents a product image.
type Image struct {
	UUID     string    `json:"uuid"`
	SKU      string    `json:"sku"`
	Path     string    `json:"path"`
	GSURL    string    `json:"gsurl"`
	Width    uint      `json:"width"`
	Height   uint      `json:"height"`
	Size     uint      `json:"size"`
	Created  time.Time `json:"created"`
	Modified time.Time `json:"modified"`
}

// CreateImageEntry creates a new image entry for a product with the given SKU.
func (s *Service) CreateImageEntry(ctx context.Context, sku, path string) (*Image, error) {
	pc := postgres.CreateProductImage{
		SKU:   sku,
		W:     99999999,
		H:     99999999,
		Path:  path,
		GSURL: fmt.Sprintf("%s%s", "gs://", path),
		Typ:   "image/jpeg",
		Ori:   true,
		Pri:   10,
		Size:  0,
		Q:     100,
	}
	pi, err := s.model.CreateImageEntry(ctx, &pc)
	if err != nil {
		return nil, errors.Wrapf(err, "service: create image sku=%q, path=%q, entry failed", sku, path)
	}
	image := Image{
		UUID:     pi.UUID,
		SKU:      pi.SKU,
		Path:     pi.Path,
		GSURL:    pi.GSURL,
		Width:    pi.W,
		Height:   pi.H,
		Size:     pi.Size,
		Created:  pi.Created,
		Modified: pi.Modified,
	}
	return &image, nil
}

// ImageUUIDExists returns true if the image with the given UUID
// exists in the database. Note: it does not check if it exists
// in Google storage.
func (s *Service) ImageUUIDExists(ctx context.Context, uuid string) (bool, error) {
	exists, err := s.model.ImageUUIDExists(ctx, uuid)
	if err != nil {
		return false, errors.Wrapf(err, "service: ImageUUIDExists(ctx, %q) failed", uuid)
	}
	return exists, nil
}

// ImagePathExists returns true if the image with the given path
// exists in the database. Note: it does not check if it exists
// in Google storage.
func (s *Service) ImagePathExists(ctx context.Context, path string) (bool, error) {
	exists, err := s.model.ImagePathExists(ctx, path)
	if err != nil {
		return false, errors.Wrapf(err, "service: ImagePathExists(ctx, %q) failed", path)
	}
	return exists, nil
}

// GetImage returns an image by the given UUID.
func (s *Service) GetImage(ctx context.Context, uuid string) (*Image, error) {
	pi, err := s.model.GetProductImageByUUID(ctx, uuid)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, errors.Wrapf(err, "service: GetProductImageByUUID(ctx, %q) failed", uuid)
	}
	image := Image{
		UUID:     pi.UUID,
		SKU:      pi.SKU,
		Path:     pi.Path,
		GSURL:    pi.GSURL,
		Width:    pi.W,
		Height:   pi.H,
		Size:     pi.Size,
		Created:  pi.Created,
		Modified: pi.Modified,
	}
	return &image, nil
}

// ListProductImages return a slice of Images.
func (s *Service) ListProductImages(ctx context.Context, sku string) ([]*Image, error) {
	pilist, err := s.model.GetImagesBySKU(ctx, sku)
	if err != nil {
		return nil, errors.Wrapf(err, "service: ListProductImages(ctx, %q) failed", sku)
	}
	images := make([]*Image, 0, 8)
	for _, pi := range pilist {
		image := Image{
			UUID:     pi.UUID,
			SKU:      pi.SKU,
			Path:     pi.Path,
			GSURL:    pi.GSURL,
			Width:    pi.W,
			Height:   pi.H,
			Size:     pi.Size,
			Created:  pi.Created,
			Modified: pi.Modified,
		}
		images = append(images, &image)
	}
	return images, nil
}

// DeleteImage delete the image with the given UUID.
func (s *Service) DeleteImage(ctx context.Context, uuid string) error {
	if _, err := s.model.DeleteProductImage(ctx, uuid); err != nil {
		return err
	}
	return nil
}

// DeleteAllProductImages deletes all images associated to the product
// with the given SKU.
func (s *Service) DeleteAllProductImages(ctx context.Context, sku string) error {
	if _, err := s.model.DeleteAllProductImages(ctx, sku); err != nil {
		return errors.Wrapf(err, "service: DeleteAllProductImages(ctx, %q)", sku)
	}
	return nil
}
