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

// ErrUserNotFound is returned when a user does not exist in the database.
var ErrUserNotFound = errors.New("service: user not found")

// User details
type User struct {
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
	FirstUUID string `json:"first_id"`
	LastUUID  string `json:"last_id"`
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
			user, err := s.CreateUser(ctx, "root", email, password, "Super", "User")
			if err != nil {
				return errors.Wrap(err, "create user for root user failed")
			}
			_, err = s.GenerateUserDevKey(ctx, user.ID)
			if err != nil {
				return errors.Wrap(err, "generate user devkey failed")
			}
			return nil
		}
		return err
	}
	log.Infof("root superuser email=%s already exists in Firebase Auth system", email)
	return nil
}

// CreateUser creates a new user
func (s *Service) CreateUser(ctx context.Context, role, email, password, firstname, lastname string) (*User, error) {
	contextLogger := log.WithContext(ctx)
	contextLogger.Debugf("s.CreateUser(ctx, role=%q, email=%q, password=%q, firstname=%q, lastname=%q) started", role, email, "*****", firstname, lastname)
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

	c, err := s.model.CreateUser(ctx, userRecord.UID, role, email, firstname, lastname)
	if err != nil {
		return nil, fmt.Errorf("model.CreateUser(%s, %s, %s, %s, %s) failed: %v", userRecord.UID, role, email, firstname, lastname, err)
	}

	// Set the custom claims for this user
	err = authClient.SetCustomUserClaims(ctx, c.UID, map[string]interface{}{
		"ecom_uid":  c.UUID,
		"ecom_role": role,
	})
	if err != nil {
		return nil, fmt.Errorf("set custom claims for uid=%s uuid=%s role=%s failed: %v", c.UID, c.UUID, role, err)
	}

	ac := User{
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

// GetUsers gets users with pagination.
func (s *Service) GetUsers(ctx context.Context, pq *PaginationQuery) (*PaginationResultSet, error) {
	q := &postgres.PaginationQuery{
		OrderBy:    pq.OrderBy,
		OrderDir:   pq.OrderDir,
		Limit:      pq.Limit,
		StartAfter: pq.StartAfter,
	}
	prs, err := s.model.GetUsers(ctx, q)
	if err != nil {
		return nil, err
	}

	results := make([]*User, 0)
	for _, v := range prs.RSet.([]*postgres.UsrRow) {
		c := User{
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

// GetUser retrieves a user by user ID.
func (s *Service) GetUser(ctx context.Context, userID string) (*User, error) {
	contextLogger := log.WithContext(ctx)
	contextLogger.Debugf("service: GetUser(ctx, userID=%q)", userID)
	c, err := s.model.GetUserByUUID(ctx, userID)
	if err != nil {
		if err == postgres.ErrUserNotFound {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	contextLogger.Debugf("service: s.model.GetUserByUUID(ctx, userUUID=%s) returned %v", userID, c)

	ac := User{
		ID:        c.UUID,
		UID:       c.UID,
		Role:      c.Role,
		Email:     c.Email,
		Firstname: c.Firstname,
		Lastname:  c.Lastname,
		Created:   c.Created,
		Modified:  c.Modified,
	}
	return &ac, nil
}
