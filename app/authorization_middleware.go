package app

import (
	"context"
	"crypto/subtle"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"firebase.google.com/go/auth"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// Authorization provides authorization middleware
func (a *App) Authorization(op string, next http.HandlerFunc) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Debugf("authorization started for %s", op)
		decodedToken := ctx.Value(ecomDecodedTokenKey).(*auth.Token)

		// Get the user ID and customer role from the JWT
		var cid, role string
		if val, ok := decodedToken.Claims["ecom_role"]; ok {
			role = val.(string)
		}

		if role == "" {
			role = RoleShopper
			log.Debug("authorization: setting role to RoleShopper (anon)")
		} else {
			if val, ok := decodedToken.Claims["ecom_uid"]; ok {
				cid = val.(string)
			}

			if role != RoleCustomer && role != RoleAdmin && role != RoleSuperUser {
				clientError(w, http.StatusForbidden, "auth/forbidden",
					"request forbidden") // 403
				return
			}
		}

		ctx2 := context.WithValue(ctx, ecomUIDKey, cid)

		// superuser has all privileges. The JWT containing the claims is cryptographically
		// signed with a claim of "root" so we give maximum privilege.
		if role == RoleSuperUser {
			contextLogger.Infof("%s authorized for RoleSuperUser", op)
			next.ServeHTTP(w, r.WithContext(ctx2))
			return
		}

		contextLogger.Infof("%s role %s, JWT cid %s", op, role, cid)

		// at this point the role is set to either "anon", "customer" or "admin"
		switch op {
		// Operations that don't require any special authorization
		case OpCreateCart, OpAddProductToCart, OpGetCartProducts, OpUpdateCartProduct,
			OpDeleteCartProduct, OpEmptyCartProducts, OpGetCategories, OpGetCategoriesTree, OpSignInWithDevKey,
			OpGetProduct, OpListProducts, OpGetProductCategoryRelations,
			OpGetTierPricing, OpMapPricingByTier, OpGetImage,
			OpListProductImages, OpPlaceOrder, OpStripeCheckout, OpGetPriceList,
			OpListInventory, OpGetInventory, OpGetShippingTariff, OpListShippingTariffs,
			OpGetProductSetItems, OpGetOffer, OpListOffers, OpApplyCouponToCart, OpUnapplyCouponFromCart,
			OpGetCartCoupon, OpListCartCoupons, OpGetProductToProductAssocGroup,
			OpListProductToProductAssocGroups,
			OpGetProductToProductAssoc, OpListProductToProductAssocs:
			next.ServeHTTP(w, r.WithContext(ctx2))
			return
		// Operations that required at least RoleAdmin privileges
		case OpListUsers, OpDeleteUser,
			OpCreateProduct, OpUpdateProduct, OpDeleteProduct, OpDeleteCategories,
			OpUpdateProductCategoryRelations, OpSystemInfo,
			OpAddProductCategoryRelations, OpUpdateProductPrices,
			OpDeleteProductCategoryRelations, OpDeleteTierPricing,
			OpAddImage, OpDeleteImage, OpDeleteAllProductImages,
			OpCreatePriceList, OpListPriceLists, OpUpdatePriceList, OpDeletePriceList,
			OpCreatePromoRule, OpDeletePromoRule, OpGetPromoRule, OpListPromoRules,
			OpUpdateInventory, OpBatchUpdateInventory,
			OpUpdateCategoriesTree,
			OpCreateShippingTariff, OpUpdateShippingTariff, OpDeleteShippingTariff,
			OpActivateOffer, OpDeactivateOffer,
			OpCreateCoupon, OpGetCoupon, OpListCoupons, OpUpdateCoupon, OpDeleteCoupon,
			OpCreateProductToProductAssocGroup,
			OpDeleteProductToProductAssocGroup, OpDeleteProductToProductAssoc,
			OpBatchUpdateProductToProductAssocs, OpCreateWebhook, OpGetWebhook, OpListWebhooks,
			OpUpdateWebhook, OpDeleteWebhook, OpGetOrder:
			if role == RoleAdmin {
				next.ServeHTTP(w, r.WithContext(ctx2))
				return
			}
			w.WriteHeader(http.StatusForbidden) // 403 Forbidden
			return
		case OpCreateUser:
			// Only anonymous users can create a new user account
			if role == RoleShopper {
				next.ServeHTTP(w, r.WithContext(ctx2))
				return
			}
			clientError(w, http.StatusForbidden, "auth/forbidden",
				"request forbidden") // 403
			return
		// both admins and shoppers can get prices, but shoppers are only allowed a price_list_id
		// matching their user account
		case OpGetProductPrices:
			if role == RoleAdmin {
				next.ServeHTTP(w, r.WithContext(ctx2))
				return
			}

			priceListID := r.URL.Query().Get("price_list_id")
			if priceListID != "" {
				contextLogger.Infof("auth: OpGetProductPrices received cid=%q query param price_list_id set to %q", cid, priceListID)
			}
			valid, err := a.Service.UserCanAccessPriceList(ctx, cid, priceListID)
			if err != nil {
				if err == service.ErrUserNotFound {
					contextLogger.Errorf("a.Service.UserCanAccessPriceList(ctx, cid=%q, priceListID=%q) error: %v", cid, priceListID, err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				contextLogger.Errorf("a.Service.UserCanAccessPriceList(ctx, cid=%q, priceListID=%q) error: %v", cid, priceListID, err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if valid {
				next.ServeHTTP(w, r.WithContext(ctx2))
				return
			}

			contextLogger.Errorf("a.Service.UserCanAccessPriceList(ctx, cid=%q, priceListID=%q) error: %v", cid, priceListID, err)

			// 403 Forbidden
			clientError(w, http.StatusForbidden, "auth/forbidden",
				"forbidden access to prices with the given price list") // 403
			return
		case OpCreateAddress, OpGetUser, OpGetUsersAddresses, OpUpdateAddress, OpGenerateUserDevKey, OpListUsersDevKeys:
			// Check the JWT Claim's user UUID and safely compare it to the user UUID in the route
			// Anonymous signin results in automatic rejection. These operations are reserved for customer role.
			if role == RoleAdmin {
				next.ServeHTTP(w, r.WithContext(ctx2))
				return
			}

			if role == RoleCustomer {
				contextLogger.Debugf("URL id %s", chi.URLParam(r, "id"))
				if subtle.ConstantTimeCompare([]byte(cid), []byte(chi.URLParam(r, "id"))) == 1 {
					next.ServeHTTP(w, r.WithContext(ctx2))
					return
				}
			}

			// RoleShopper
			clientError(w, http.StatusForbidden, "auth/forbidden",
				"request forbidden") // 403
			return
		case OpDeleteUserDevKey:
			if role == RoleAdmin {
				next.ServeHTTP(w, r.WithContext(ctx2))
				return
			}
			clientError(w, http.StatusForbidden, "auth/forbidden",
				"request forbidden") // 403
			return
		case OpGetAddress, OpDeleteAddress:
			// The user ID is not in the route so we ask the service layer for the
			// resource owner's user ID
			if role == RoleShopper {
				clientError(w, http.StatusForbidden, "auth/forbidden",
					"request forbidden") // 403
				return
			}

			if role == RoleAdmin {
				next.ServeHTTP(w, r.WithContext(ctx2))
				return
			}

			id := chi.URLParam(r, "id")
			ocid, err := a.Service.GetAddressOwner(ctx, id)
			if err != nil {
				contextLogger.Errorf("a.Service.GetAddressOwner(%s) error: %v", id, err)
				clientError(w, http.StatusForbidden, "auth/forbidden",
					"request forbidden") // 403
				return
			}
			if ocid == nil {
				contextLogger.Errorf("a.Service.GetAddressOwner(%s) returned nil", id)
				clientError(w, http.StatusForbidden, "auth/forbidden",
					"request forbidden") // 403
				return
			}

			if subtle.ConstantTimeCompare([]byte(cid), []byte(*ocid)) == 1 {
				next.ServeHTTP(w, r.WithContext(ctx2))
				return
			}
			clientError(w, http.StatusForbidden, "auth/forbidden",
				"request forbidden") // 403
			return
		default:
			contextLogger.Infof("(default) authorization declined for %s", op)
			clientError(w, http.StatusForbidden, "auth/forbidden",
				"request forbidden") // 403
			return
		}
	}

	return http.HandlerFunc(fn)
}
