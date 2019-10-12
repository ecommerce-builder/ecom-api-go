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

// ErrUserExists is returned when attempting to create a new user that
// already exists.
var ErrUserExists = errors.New("service: user exists")

// ErrUserInUse is returned when attempting to delete
// a user that has previously placed orders
var ErrUserInUse = errors.New("service: user in use")

// User details
type User struct {
	Object      string    `json:"object"`
	ID          string    `json:"id"`
	UID         string    `json:"uid"`
	Role        string    `json:"role"`
	PriceListID string    `json:"price_list_id"`
	Email       string    `json:"email"`
	Firstname   string    `json:"firstname"`
	Lastname    string    `json:"lastname"`
	Created     time.Time `json:"created"`
	Modified    time.Time `json:"modified"`
}

// PaginationQuery holds query
type PaginationQuery struct {
	OrderBy    string
	OrderDir   string
	Limit      int
	StartAfter string
	EndBefore  string
}

// PaginationContext holds the context for the currently retrieved set.
type PaginationContext struct {
	Total      int    `json:"total"`
	FirstID    string `json:"first_id"`
	LastID     string `json:"last_id"`
	SetFirstID string `json:"set_first_id"`
	SetLastID  string `json:"set_last_id"`
}

// PaginationResultSet contains the set retrieved from the last fetch
// including the context information.
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

// CreateUser attempts creates a new user returning the newly created user
// or nil with error `ErrUserExists` if that user already exists.
func (s *Service) CreateUser(ctx context.Context, role, email, password, firstname, lastname string) (*User, error) {
	contextLogger := log.WithContext(ctx)
	contextLogger.Debugf("s.CreateUser(ctx, role=%q, email=%q, password=%q, firstname=%q, lastname=%q) started", role, email, "*****", firstname, lastname)

	authClient, err := s.fbApp.Auth(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "s.fbApp.Auth(ctx) failed")
	}

	// check if the user already exists
	_, err = authClient.GetUserByEmail(ctx, email)
	if err != nil {
		if !auth.IsUserNotFound(err) {
			return nil, errors.Wrapf(err, "authClient.GetUserByEmail(ctx, email=%q) failed", email)
		}
	} else {
		contextLogger.Infof("user with email=%q already exists", email)
		return nil, ErrUserExists
	}

	user := (&auth.UserToCreate{}).
		Email(email).
		EmailVerified(false).
		Password(password).
		DisplayName(fmt.Sprintf("%s %s", firstname, lastname)).
		Disabled(false)
	userRecord, err := authClient.CreateUser(ctx, user)
	if err != nil {
		return nil, errors.Wrap(err, "authClient.CreateUser(ctx, user) failed")
	}
	contextLogger.Infof("firebase auth user (uid=%q) created", userRecord.UID)

	c, err := s.model.CreateUser(ctx, userRecord.UID, role, email, firstname, lastname)
	if err != nil {
		return nil, errors.Wrapf(err, "service: s.model.CreateUser(ctx, uid=%q, role=%q, email=%q, firstname=%q, lastname=%q) failed: %v", userRecord.UID, role, email, firstname, lastname, err)
	}

	// Set the custom claims for this user
	err = authClient.SetCustomUserClaims(ctx, c.UID, map[string]interface{}{
		"ecom_uid":  c.UUID,
		"ecom_role": role,
	})
	if err != nil {
		return nil, fmt.Errorf("set custom claims for uid=%s uuid=%s role=%s failed: %v", c.UID, c.UUID, role, err)
	}
	contextLogger.Infof("firebase custom claims set ecom_uid=%q ecom_role=%q", c.UUID, role)

	ac := User{
		Object:      "user",
		ID:          c.UUID,
		UID:         c.UID,
		Role:        c.Role,
		PriceListID: c.PriceListUUID,
		Email:       c.Email,
		Firstname:   c.Firstname,
		Lastname:    c.Lastname,
		Created:     c.Created,
		Modified:    c.Modified,
	}

	if err := s.PublishTopicEvent(ctx, EventUserCreated, &ac); err != nil {
		return nil, errors.Wrapf(err, "service: s.PublishTopicEvent(ctx, event=%q, data=%v) failed", EventUserCreated, ac)
	}
	return &ac, nil
}

// GetUsers gets users with pagination.
func (s *Service) GetUsers(ctx context.Context, pq *PaginationQuery) (*PaginationResultSet, error) {
	q := &postgres.PaginationQuery{
		OrderBy:    pq.OrderBy,
		OrderDir:   pq.OrderDir,
		Limit:      pq.Limit,
		StartAfter: pq.StartAfter,
		EndBefore:  pq.EndBefore,
	}
	prs, err := s.model.GetUsers(ctx, q)
	if err != nil {
		return nil, err
	}

	users := make([]*User, 0)
	for _, row := range prs.RSet.([]*postgres.UsrJoinRow) {
		c := User{
			Object:      "user",
			ID:          row.UUID,
			UID:         row.UID,
			Role:        row.Role,
			PriceListID: row.PriceListUUID,
			Email:       row.Email,
			Firstname:   row.Firstname,
			Lastname:    row.Lastname,
			Created:     row.Created,
			Modified:    row.Modified,
		}
		users = append(users, &c)
	}

	aprs := &PaginationResultSet{
		RContext: PaginationContext{
			Total:   prs.RContext.Total,
			FirstID: prs.RContext.FirstUUID,
			LastID:  prs.RContext.LastUUID,
		},
		RSet: users,
	}
	return aprs, nil
}

// GetUser retrieves a user by user ID.
func (s *Service) GetUser(ctx context.Context, userID string) (*User, error) {
	contextLogger := log.WithContext(ctx)
	contextLogger.Debugf("service: GetUser(ctx, userID=%q)", userID)

	row, err := s.model.GetUserByUUID(ctx, userID)
	if err == postgres.ErrUserNotFound {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, errors.Wrapf(err, "service: s.model.GetUserByUUID(ctx, userUUID=%q) failed", userID)
	}
	contextLogger.Debugf("service: s.model.GetUserByUUID(ctx, userUUID=%s) returned %v", userID, row)

	user := User{
		Object:      "user",
		ID:          row.UUID,
		UID:         row.UID,
		Role:        row.Role,
		PriceListID: row.PriceListUUID,
		Email:       row.Email,
		Firstname:   row.Firstname,
		Lastname:    row.Lastname,
		Created:     row.Created,
		Modified:    row.Modified,
	}
	return &user, nil
}

// DeleteUser attempts to delete a user.
func (s *Service) DeleteUser(ctx context.Context, userID string) error {
	contextLogger := log.WithContext(ctx)

	authClient, err := s.fbApp.Auth(ctx)
	if err != nil {
		return errors.Wrap(err, "service: failed to get firebase auth client")
	}

	uid, err := s.model.DeleteUserByUUID(ctx, userID)
	if err == postgres.ErrUserNotFound {
		return ErrUserNotFound
	}
	if err == postgres.ErrUserInUse {
		return ErrUserInUse
	}
	if err != nil {
		return errors.Wrapf(err, "service: s.model.DeleteUserByUUID(ctx, userUUID=%q) failed", userID)
	}
	contextLogger.Infof("service: deleted user userID=%q from the ecom system returning uid=%q", userID, uid)

	if err = authClient.DeleteUser(ctx, uid); err != nil {
		return errors.Wrapf(err, "service: firebase auth delete user failed for uid=%q", uid)
	}
	contextLogger.Infof("service: delete firebase auth user uid=%q", uid)
	return nil
}
