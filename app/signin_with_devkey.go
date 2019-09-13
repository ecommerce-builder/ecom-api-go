package app

import (
	"encoding/json"
	"net/http"

	"bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

// SignInWithDevKeyHandler issues a Custom JWT when passed a valid Developer API Key.
func (a *App) SignInWithDevKeyHandler() http.HandlerFunc {
	type signInRequestBody struct {
		Key string `json:"key"`
	}
	type signInResponseBody struct {
		CustomToken string         `json:"custom_token"`
		Customer    *firebase.User `json:"user"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: SignInWithDevKeyHandler started")

		if r.Body == nil {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, "missing request body")
			return
		}
		o := signInRequestBody{}
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		err := dec.Decode(&o)
		if err != nil {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, err.Error())
			return
		}
		customToken, customer, err := a.Service.SignInWithDevKey(ctx, o.Key)
		if err != nil {
			if err == bcrypt.ErrMismatchedHashAndPassword {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			contextLogger.Errorf("app: SignInWithDevKeyHandler(ctx, ...) error: %v\n", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		ctRes := signInResponseBody{
			CustomToken: customToken,
			Customer:    customer,
		}
		w.WriteHeader(http.StatusCreated) // 201 Created
		json.NewEncoder(w).Encode(ctRes)
	}
}
