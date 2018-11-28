package app

import (
	"net/http"
	"strings"

	"github.com/gorilla/context"
	log "github.com/sirupsen/logrus"
)

// AuthenticateMiddleware provides authentication layer
func (a *App) AuthenticateMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debugf("%+v", r)
		log.Debug("AuthenticateMiddleware started")

		// missing Authorization header
		token, ok := r.Header["Authorization"]
		if !ok {
			log.Debug("Authorization header missing")
			w.WriteHeader(http.StatusUnauthorized) // 401 Unauthorized
			w.Header().Set("WWW-Authenticate", "Bearer")
			return
		}

		// Single Bearer token line only
		if len(token) != 1 {
			log.Error("using multiple Bearer tokens")
			w.WriteHeader(http.StatusUnauthorized) // 401 Unauthorized
			w.Header().Set("WWW-Authenticate", "Bearer")
			return
		}

		// Authorization: Bearer <jwt> format only
		pieces := strings.Split(token[0], " ")
		if len(pieces) != 2 || strings.ToLower(pieces[0]) != "bearer" {
			log.Errorf("invalid Authorization: Bearer <jwt> header. token=%s", token[0])
			w.WriteHeader(http.StatusUnauthorized) // 401 Unauthorized
			w.Header().Set("WWW-Authenticate", "Bearer")
			return
		}

		jwt := pieces[1]
		decodedToken, err := a.Service.Authenticate(ctx, jwt)
		if err != nil {
			log.Errorf("authenticating failure: jwt=%s", jwt)
			return
		}

		log.Info("authentication success")

		// store the decodedToken in the context
		context.Set(r, "decodedToken", decodedToken)

		next(w, r)

		log.Debug("AuthenticateMiddleware ended")
	}
}
