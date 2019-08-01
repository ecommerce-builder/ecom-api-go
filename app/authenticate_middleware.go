package app

import (
	"context"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

type decodedTokenString string

// AuthenticateMiddleware provides authentication layer
func (a *App) AuthenticateMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)

		// missing Authorization header
		token, ok := r.Header["Authorization"]
		if !ok {
			contextLogger.Debug("authorization header missing")
			w.WriteHeader(http.StatusUnauthorized) // 401 Unauthorized
			w.Header().Set("WWW-Authenticate", "Bearer")
			w.Header().Set("Content-Type", "application/json")
			return
		}

		// Single Bearer token line only
		if len(token) != 1 {
			contextLogger.Error("using multiple Bearer tokens")
			w.WriteHeader(http.StatusUnauthorized) // 401 Unauthorized
			w.Header().Set("WWW-Authenticate", "Bearer")
			return
		}

		// Authorization: Bearer <jwt> format only
		pieces := strings.Split(token[0], " ")
		if len(pieces) != 2 || strings.ToLower(pieces[0]) != "bearer" {
			contextLogger.Errorf("invalid Authorization: Bearer <jwt> header. token=%s", token[0])
			w.WriteHeader(http.StatusUnauthorized) // 401 Unauthorized
			w.Header().Set("WWW-Authenticate", "Bearer")
			return
		}

		jwt := pieces[1]
		decodedToken, err := a.Service.Authenticate(ctx, jwt)
		if err != nil {
			contextLogger.Errorf("authenticating failure: jwt=%s", jwt)
			w.WriteHeader(http.StatusUnauthorized) // 401 Unauthorized
			w.Header().Set("WWW-Authenticate", "Bearer")
			return
		}

		contextLogger.Info("authentication success")

		// store the decodedToken in the context
		ctx2 := context.WithValue(ctx, "ecomDecodedToken", decodedToken)
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r.WithContext(ctx2))
	}

	return http.HandlerFunc(fn)
}
