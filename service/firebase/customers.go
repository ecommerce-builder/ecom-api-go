package firebase

import (
	"context"
	"fmt"
	"time"

	"bitbucket.org/andyfusniakteam/ecom-api-go/model/postgres"
	"firebase.google.com/go/auth"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// ErrCustomerNotFound is returned when a customer does not exist in the database.
var ErrCustomerNotFound = errors.New("service: customer not found")

// Customer details
type Customer struct {
	ID        string    `json:"id"`
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
			_, err = s.GenerateCustomerDevKey(ctx, customer.ID)
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
	contextLogger := log.WithContext(ctx)
	contextLogger.Debugf("s.CreateCustomer(%s, %s, %s, %s, %s) started", role, email, "*****", firstname, lastname)
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
		"cid":  c.UUID,
		"role": role,
	})
	if err != nil {
		return nil, fmt.Errorf("set custom claims for uid=%s uuid=%s role=%s failed: %v", c.UID, c.UUID, role, err)
	}

	ac := Customer{
		ID:        c.UUID,
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
	for _, v := range prs.RSet.([]*postgres.CustomerRow) {
		c := Customer{
			ID:        v.UUID,
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
func (s *Service) GetCustomer(ctx context.Context, customerID string) (*Customer, error) {
	contextLogger := log.WithContext(ctx)
	contextLogger.Debugf("service: GetCustomer(ctx, customerID=%s)", customerID)
	c, err := s.model.GetCustomerByUUID(ctx, customerID)
	if err != nil {
		if err == postgres.ErrCustomerNotFound {
			return nil, ErrCustomerNotFound
		}
		return nil, err
	}
	contextLogger.Debugf("service: s.model.GetCustomerByUUID(ctx, customerUUID=%s) returned %v", customerID, c)

	ac := Customer{
		ID:        c.UUID,
		UID:       c.UID,
		Email:     c.Email,
		Firstname: c.Firstname,
		Lastname:  c.Lastname,
		Created:   c.Created,
		Modified:  c.Modified,
	}
	return &ac, nil
}
