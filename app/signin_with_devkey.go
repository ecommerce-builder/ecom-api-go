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
		CustomToken string             `json:"custom_token"`
		Customer    *firebase.Customer `json:"customer"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: SignInWithDevKeyHandler started")

		if r.Body == nil {
			http.Error(w, "Please send a request body", 400)
			return
		}
		o := signInRequestBody{}
		err := json.NewDecoder(r.Body).Decode(&o)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		customToken, customer, err := a.Service.SignInWithDevKey(ctx, o.Key)
		if err != nil {
			if err == bcrypt.ErrMismatchedHashAndPassword {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			contextLogger.Errorf("service SignInWithDevKeyHandler(ctx, ...) error: %v\n", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		ctRes := signInResponseBody{
			CustomToken: customToken,
			Customer:    customer,
		}
		w.WriteHeader(http.StatusOK) // 200 Ok
		json.NewEncoder(w).Encode(ctRes)
	}
}
