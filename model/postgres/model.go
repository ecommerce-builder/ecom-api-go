package postgres

import "time"

// CartItem structure holds the details individual cart item
type CartItem struct {
	ID        int
	CartUUID  string
	Sku       string
	Qty       int
	UnitPrice float64
	Created   time.Time
	Modified  time.Time
}

// Customer details
type Customer struct {
	ID           int
	CustomerUUID string
	UID          string
	Role         string
	Email        string
	Firstname    string
	Lastname     string
	Created      time.Time
	Modified     time.Time
}

// PaginationResultSet contains both the underlying result set as well as
// context about the data including Total; the total number of rows in
// the table, First; set to true if this result set represents the first
// page, Last; set to true if this result set represents the last page of
// results.
type PaginationResultSet struct {
	RContext struct {
		Total               int
		FirstUUID, LastUUID string
	}
	RSet interface{}
}

type PaginationQuery struct {
	OrderBy    string
	OrderDir   string
	Limit      int
	StartAfter string
}

// Address contains address information for a Customer
type Address struct {
	ID          int
	AddrUUID    string
	CustomerID  int
	Typ         string
	ContactName string
	Addr1       string
	Addr2       *string
	City        string
	County      *string
	Postcode    string
	Country     string
	Created     time.Time
	Modified    time.Time
}

type CustomerDevKey struct {
	ID           int       `json:"id"`
	UUID         string    `json:"uuid"`
	Key          string    `json:"key"`
	Hash         string    `json:"hash"`
	CustomerID   int       `json:"customer_id"`
	CustomerUUID string    `json:"customer_uuid"`
	Created      time.Time `json:"created"`
	Modified     time.Time `json:"modified"`
}

// CatalogProductAssoc maps products to leaf nodes in the catalogue hierarchy
type CatalogProductAssoc struct {
	ID        int
	CatalogID int
	ProductID int
	Path      string `json:"path"`
	SKU       string `json:"sku"`
	Pri       int    `json:"pri"`
	Created   time.Time
	Modified  time.Time
}

type CreateProductImage struct {
	SKU   string
	W     uint
	H     uint
	Path  string
	Typ   string
	Ori   bool
	Pri   uint
	Size  uint
	Q     uint
	GSURL string
	Data  interface{}
}

type ProductImage struct {
	ID        uint
	ProductID uint
	UUID      string
	SKU       string
	W         uint
	H         uint
	Path      string
	Typ       string
	Ori       bool
	Up        bool
	Pri       uint
	Size      uint
	Q         uint
	GSURL     string
	Data      interface{}
	Created   time.Time
	Modified  time.Time
}
