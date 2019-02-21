package app

import (
	"crypto/subtle"
	"net/http"

	"firebase.google.com/go/auth"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// AuthorizationMiddleware provides authorization layer
func (a *App) Authorization(op string, next http.HandlerFunc) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log.Debugf("AuthorizationMiddleware started for operation %s", op)
		decodedToken := ctx.Value("ecomDecodedToken").(*auth.Token)

		// Get the customer UUID and customer role from the JWT
		var cuuid, role string
		if val, ok := decodedToken.Claims["role"]; ok {
			role = val.(string)
		}

		if role == "" {
			role = RoleShopper
		} else {
			if val, ok := decodedToken.Claims["cuuid"]; ok {
				cuuid = val.(string)
			}

			if role != RoleCustomer && role != RoleAdmin && role != RoleSuperUser {
				w.WriteHeader(http.StatusUnauthorized) // 401 Unauthorized
				return
			}
		}

		// superuser has all privileges. The JWT containing the claims is cryptographically
		// signed with a claim of "root" so we give maximum privilege.
		if role == RoleSuperUser {
			next.ServeHTTP(w, r)
			return
		}

		log.Debugf("role %s, op %s, JWT cuuid %s", role, op, cuuid)

		// at this point the role is set to either "anon", "customer" or "admin"
		switch op {
		case OpCreateCart, OpAddItemToCart, OpGetCartItems, OpUpdateCartItem, OpDeleteCartItem, OpEmptyCartItems, OpGetCatalog:
			// Cart and Catalog operations don't require any special authorization
			next.ServeHTTP(w, r)
			return
		case OpCreateCustomer:
			// Only anonymous users can create a new customer account
			if role == RoleShopper {
				next.ServeHTTP(w, r)
				return
			}

			w.WriteHeader(http.StatusUnauthorized) // 401 Unauthorized
			return
		case OpCreateAddress, OpGetCustomer, OpGetCustomersAddresses, OpUpdateAddress:
			// Check the JWT Claim's customer UUID and safely compare it to the customer UUID in the route
			// Anonymous signin results in automatic rejection. These operations are reserved for customers.
			if role == RoleAdmin {
				next.ServeHTTP(w, r)
				return
			}

			if role == RoleCustomer {
				log.Debugf("URL cuuid %s", chi.URLParam(r, "cuuid"))
				if subtle.ConstantTimeCompare([]byte(cuuid), []byte(chi.URLParam(r, "cuuid"))) == 1 {
					next.ServeHTTP(w, r)
					return
				}
			}

			// RoleShopper
			w.WriteHeader(http.StatusUnauthorized) // 401 Unauthorized
			return
		case OpSystemInfo:
			if role == RoleAdmin {
				next.ServeHTTP(w, r)
				return
			}
			w.WriteHeader(http.StatusUnauthorized) // 401 Unauthorized
			return
		case OpGetAddress, OpDeleteAddress:
			// The customer UUID is not in the route so we ask the service layer for the  resource owner's customer UUID
			if role == RoleShopper {
				w.WriteHeader(http.StatusUnauthorized) // 401 Unauthorized
				return
			}

			auuid := chi.URLParam(r, "auuid")
			ocuuid, err := a.Service.GetAddressOwner(ctx, auuid)
			if err != nil {
				log.Errorf("a.Service.GetAddressOwner(%s) error: %v", auuid, err)
				return
			}
			if ocuuid == nil {
				log.Errorf("a.Service.GetAddressOwner(%s) returned nil", auuid)
				w.WriteHeader(http.StatusUnauthorized) // 401 Unauthorized
				return
			}

			if subtle.ConstantTimeCompare([]byte(cuuid), []byte(*ocuuid)) == 1 {
				next.ServeHTTP(w, r)
				return
			}
			w.WriteHeader(http.StatusUnauthorized) // 401 Unauthorized
			return
		default:
			log.Infof("authorization declined for operation %s", op)
			w.WriteHeader(http.StatusUnauthorized) // 401 Unauthorized
			return
		}
	}

	return http.HandlerFunc(fn)
}
