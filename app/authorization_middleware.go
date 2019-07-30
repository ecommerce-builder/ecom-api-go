package app

import (
	"crypto/subtle"
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
		contextLogger := log.WithContext(ctx)
		contextLogger.Debugf("authorization started for %s", op)
		decodedToken := ctx.Value("ecomDecodedToken").(*auth.Token)

		// Get the customer ID and customer role from the JWT
		var cid, role string
		if val, ok := decodedToken.Claims["role"]; ok {
			role = val.(string)
		}

		if role == "" {
			role = RoleShopper
		} else {
			if val, ok := decodedToken.Claims["cid"]; ok {
				cid = val.(string)
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

		contextLogger.Infof("%s role %s, JWT cid %s", op, role, cid)

		// at this point the role is set to either "anon", "customer" or "admin"
		switch op {
		// Operations that don't require any special authorization
		case OpCreateCart, OpAddItemToCart, OpGetCartItems, OpUpdateCartItem,
			OpDeleteCartItem, OpEmptyCartItems, OpGetCatalog, OpSignInWithDevKey,
			OpProductExists, OpGetProduct, OpListProducts, OpGetCatalogAssocs,
			OpGetTierPricing, OpMapPricingBySKU, OpMapPricingByTier, OpGetImage,
			OpListProductImages, OpPlaceOrder, OpStripeCheckout:
			next.ServeHTTP(w, r)
			return
		case OpListCustomers, OpReplaceProduct, OpDeleteProduct,
			OpPurgeCatalogAssocs, OpUpdateCatalogAssocs, OpSystemInfo,
			OpUpdateCatalog, OpPurgeCatalog, OpUpdateTierPricing, OpDeleteTierPricing,
			OpAddImage, OpDeleteImage, OpDeleteAllProductImages:
			if role == RoleAdmin {
				next.ServeHTTP(w, r)
				return
			}
			w.WriteHeader(http.StatusForbidden) // 403 Forbidden
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
				contextLogger.Debugf("URL id %s", chi.URLParam(r, "id"))
				if subtle.ConstantTimeCompare([]byte(cid), []byte(chi.URLParam(r, "id"))) == 1 {
					next.ServeHTTP(w, r)
					return
				}
			}

			// RoleShopper
			unauthorized(w)
			return
		case OpDeleteCustomerDevKey:
			if role == RoleAdmin {
				// id := chi.URLParam(r, "id")
			}
			unauthorized(w)
			return
		case OpGetAddress, OpDeleteAddress:
			// The customer ID is not in the route so we ask the service layer for the  resource owner's customer ID
			if role == RoleShopper {
				unauthorized(w)
				return
			}

			if role == RoleAdmin {
				next.ServeHTTP(w, r)
				return
			}

			id := chi.URLParam(r, "id")
			ocid, err := a.Service.GetAddressOwner(ctx, id)
			if err != nil {
				contextLogger.Errorf("a.Service.GetAddressOwner(%s) error: %v", id, err)
				w.WriteHeader(http.StatusUnauthorized) // 401 Unauthorized
				return
			}
			if ocid == nil {
				contextLogger.Errorf("a.Service.GetAddressOwner(%s) returned nil", id)
				w.WriteHeader(http.StatusUnauthorized) // 401 Unauthorized
				return
			}

			if subtle.ConstantTimeCompare([]byte(cid), []byte(*ocid)) == 1 {
				next.ServeHTTP(w, r)
				return
			}
			unauthorized(w)
			return
		default:
			contextLogger.Infof("(default) authorization declined for %s", op)
			unauthorized(w)
			return
		}
	}

	return http.HandlerFunc(fn)
}
