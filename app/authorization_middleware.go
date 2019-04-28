package app

import (
	"crypto/subtle"
	"fmt"
	"net/http"

	"firebase.google.com/go/auth"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

func unauthorized(w http.ResponseWriter) {
	// w.Header().Set("Content-Length", "0")
	w.WriteHeader(http.StatusUnauthorized) // 401 Unauthorized
}

// Authorization provides authorization middleware
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
				unauthorized(w)
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
		// Operations that don't require any special authorization
		case OpCreateCart, OpAddItemToCart, OpGetCartItems, OpUpdateCartItem, OpDeleteCartItem, OpEmptyCartItems, OpGetCatalog, OpSignInWithDevKey, OpProductExists, OpGetProduct:
			next.ServeHTTP(w, r)
			return
		case OpListCustomers, OpCreateProduct, OpUpdateProduct, OpDeleteProduct:
			if role == RoleAdmin {
				next.ServeHTTP(w, r)
				return
			}
			unauthorized(w)
			return
		case OpCreateCustomer:
			// Only anonymous users can create a new customer account
			if role == RoleShopper {
				next.ServeHTTP(w, r)
				return
			}
			unauthorized(w)
			return
		case OpCreateAddress, OpGetCustomer, OpGetCustomersAddresses, OpUpdateAddress, OpGenerateCustomerDevKey, OpListCustomersDevKeys:
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
			unauthorized(w)
			return
		case OpSystemInfo:
			if role == RoleAdmin {
				next.ServeHTTP(w, r)
				return
			}
			unauthorized(w)
			return
		case OpDeleteCustomerDevKey:
			if role == RoleAdmin {
				uuid := chi.URLParam(r, "uuid")
				fmt.Println(uuid)
			}
			unauthorized(w)
			return
		case OpGetAddress, OpDeleteAddress:
			// The customer UUID is not in the route so we ask the service layer for the  resource owner's customer UUID
			if role == RoleShopper {
				unauthorized(w)
				return
			}

			uuid := chi.URLParam(r, "uuid")
			ocuuid, err := a.Service.GetAddressOwner(ctx, uuid)
			if err != nil {
				log.Errorf("a.Service.GetAddressOwner(%s) error: %v", uuid, err)
				return
			}
			if ocuuid == nil {
				log.Errorf("a.Service.GetAddressOwner(%s) returned nil", uuid)
				w.WriteHeader(http.StatusUnauthorized) // 401 Unauthorized
				return
			}

			if subtle.ConstantTimeCompare([]byte(cuuid), []byte(*ocuuid)) == 1 {
				next.ServeHTTP(w, r)
				return
			}
			unauthorized(w)
			return
		default:
			log.Infof("(default) authorization declined for operation %s", op)
			unauthorized(w)
			return
		}
	}

	return http.HandlerFunc(fn)
}
