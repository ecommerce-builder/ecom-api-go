package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

type userResponseBody struct {
	Object string `json:"object"`
	*service.User
}

// CreateUserHandler creates a new user record
func (a *App) CreateUserHandler() http.HandlerFunc {
	type createUserRequestBody struct {
		Email     string `json:"email"`
		Password  string `json:"password"`
		Firstname string `json:"firstname"`
		Lastname  string `json:"lastname"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: CreateUserHandler called")

		if r.Body == nil {
			w.WriteHeader(http.StatusBadRequest) // 400 Bad Request
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, "missing request body")
			return
		}
		o := createUserRequestBody{}
		err := json.NewDecoder(r.Body).Decode(&o)
		if err != nil {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, err.Error())
			return
		}
		defer r.Body.Close()

		user, err := a.Service.CreateUser(ctx, "customer", o.Email, o.Password, o.Firstname, o.Lastname)
		if err != nil {
			contextLogger.Errorf("app: CreateUserHandler: failed Service.CreateUser(ctx, %q, %s, %s, %s, %s) with error: %v", "customer", o.Email, "*****", o.Firstname, o.Lastname, err)
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}

		res := userResponseBody{
			Object: "user",
			User:   user,
		}
		w.WriteHeader(http.StatusCreated) // 201 Created
		json.NewEncoder(w).Encode(res)
	}
}
