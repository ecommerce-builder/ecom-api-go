package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"firebase.google.com/go/auth"
	log "github.com/sirupsen/logrus"
)

type createUserRequestBody struct {
	Role      *string `json:"role"`
	Email     *string `json:"email"`
	Password  *string `json:"password"`
	Firstname *string `json:"firstname"`
	Lastname  *string `json:"lastname"`
}

func validateCreateUserRequest(request *createUserRequestBody) (bool, string) {
	// role attribute
	role := request.Role
	if role == nil {
		return false, "attribute role must be set to customer or admin"
	}
	if *role != "customer" && *role != "admin" {
		return false, "attribute role value not valid - must be customer or admin"
	}

	// email attribute
	email := request.Email
	if email == nil {
		return false, "attribute email must be set"
	}
	// TODO: check for valid email format

	// password attribute
	password := request.Password
	if password == nil {
		return false, "attribute password must be set"
	}
	// TODO: check password suitabilty

	// firstname attribute
	firstname := request.Firstname
	if firstname == nil {
		return false, "attribute firstname must be set"
	}
	// TODO: check for valid format for firstname

	// lastname attribute
	lastname := request.Lastname
	if lastname == nil {
		return false, "attribute lastname must be set"
	}

	return true, ""
}

// CreateUserHandler creates a new user record
func (a *App) CreateUserHandler() http.HandlerFunc {
	type userResponseBody struct {
		Object string `json:"object"`
		*service.User
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: CreateUserHandler called")

		if r.Body == nil {
			// 400 Bad Request
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, "missing request body")
			return
		}
		request := createUserRequestBody{}
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		err := dec.Decode(&request)
		if err != nil {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, err.Error())
			return
		}
		defer r.Body.Close()

		// valid the request body
		valid, message := validateCreateUserRequest(&request)
		if !valid {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, message)
			return
		}

		// TODO: make sure this is the best place to do validation for this
		// an admin has elivated privileges so care must be taken.
		//
		// Only the super user can create an administrator
		if *request.Role == "admin" {
			decodedToken := ctx.Value(ecomDecodedTokenKey).(*auth.Token)

			// Get the role from the JWT
			// (not to be confused with the role for the new user)
			var role string
			if val, ok := decodedToken.Claims["ecom_role"]; ok {
				role = val.(string)
			}

			if role != "root" {
				contextLogger.Info("app: request.Role is admin but JWT does not super user access")
				clientError(w, http.StatusForbidden, ErrCodeCreateUserForbidden, "create admin forbidden")
				return
			}

		}

		user, err := a.Service.CreateUser(ctx, *request.Role, *request.Email, *request.Password, *request.Firstname, *request.Lastname)
		if err != nil {
			if err == service.ErrUserExists {
				clientError(w, http.StatusConflict, ErrCodeUserExists, "user with this email already exists")
				return
			}

			contextLogger.Errorf("app: CreateUserHandler: failed Service.CreateUser(ctx, %q, %s, %s, %s, %s) with error: %v", "customer", *request.Email, "*****", *request.Firstname, *request.Lastname, err)
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}

		contextLogger.Infof("user created (id=%q, uid=%q, priceListID=%q, role=%q, email=%q, firstname=%q, lastname=%q)", user.ID, user.UID, user.PriceListID, user.Role, user.Email, user.Firstname, user.Lastname)

		w.WriteHeader(http.StatusCreated) // 201 Created
		json.NewEncoder(w).Encode(&user)
	}
}
