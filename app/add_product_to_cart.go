package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

type addProductToCartRequestBody struct {
	CartID    string `json:"cart_id"`
	ProductID string `json:"product_id"`
	Qty       int    `json:"qty"`
}

func validateAddProductRequestBody(request *addProductToCartRequestBody) (bool, string) {
	if request.Qty < 1 {
		return false, "qty must be set to at least 1"
	}
	return true, ""
}

// AddProductToCartHandler creates a handler to add a product to a given cart
func (a *App) AddProductToCartHandler() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: AddProductToCartHandler started")

		request := addProductToCartRequestBody{}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(struct {
				Status  int    `json:"status"`
				Code    string `json:"code"`
				Message string `json:"message"`
			}{
				http.StatusBadRequest,
				ErrCodeBadRequest,
				"bad request",
			})
			return
		}

		ok, message := validateAddProductRequestBody(&request)
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(struct {
				Status  int    `json:"status"`
				Code    string `json:"code"`
				Message string `json:"message"`
			}{
				http.StatusBadRequest,
				ErrCodeBadRequest,
				message,
			})
			return
		}

		userID := ctx.Value("ecom_uid").(string)
		product, err := a.Service.AddProductToCart(ctx, userID, request.CartID, request.ProductID, request.Qty)
		if err != nil {
			if err == service.ErrCartNotFound {
				w.WriteHeader(http.StatusNotFound) // 404 Not Found
				return
			} else if err == service.ErrUserNotFound {
				w.WriteHeader(http.StatusConflict) // 409 Conflict
				json.NewEncoder(w).Encode(struct {
					Status  int    `json:"status"`
					Code    string `json:"code"`
					Message string `json:"message"`
				}{
					http.StatusConflict,
					ErrCodeUserNotFound,
					"The userID inside the JWT did not match any user in the system",
				})
				return
			} else if err == service.ErrProductNotFound {
				w.WriteHeader(http.StatusConflict) // 409 Conflict
				json.NewEncoder(w).Encode(struct {
					Status  int    `json:"status"`
					Code    string `json:"code"`
					Message string `json:"message"`
				}{
					http.StatusConflict,
					ErrCodeProductNotFound,
					"failed to add product with given id to the cart as the product cannot be found",
				})
				return
			} else if err == service.ErrDefaultPriceListNotFound {
				contextLogger.Error("ErrDefaultPriceListNotFound")
				w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
				return
			} else if err == service.ErrCartProductExists {
				w.WriteHeader(http.StatusConflict) // 409 Conflict
				json.NewEncoder(w).Encode(struct {
					Status  int    `json:"status"`
					Code    string `json:"code"`
					Message string `json:"message"`
				}{
					http.StatusConflict,
					ErrCodeCartProductExists,
					"cart product already in the cart",
				})
				return
			}

			contextLogger.Errorf("service AddProductToCart(cartID=%q, productUD=%q, qty=%d) failed with error: %v", request.CartID, request.ProductID, request.Qty, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusCreated) // 201 Created
		json.NewEncoder(w).Encode(&product)
	}
}
