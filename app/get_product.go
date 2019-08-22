package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/go-chi/chi"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// ErrIncludeQueryParamParseError error
var ErrIncludeQueryParamParseError = errors.New("app: include query params failed to parse")

// ErrIncludeQueryContainUnacceptedValue error
var ErrIncludeQueryContainUnacceptedValue = errors.New("app: include query param contains invalid values")

func sliceContains(list []string, item string) bool {
	for _, i := range list {
		if item == i {
			return true
		}
	}
	return false
}

func parseIncludeQueryParam(include string, accepted []string) (invalid string, parsed []string, err error) {
	if include == "" {
		return "", nil, nil
	}

	r, err := regexp.Compile("^[a-z,]+$")
	if err != nil {
		return "", nil, errors.Wrapf(err, `app: regexp.Compile("^[a-z,]+$") failed`)
	}
	if !r.MatchString(include) {
		return "", nil, ErrIncludeQueryParamParseError
	}

	includes := strings.Split(include, ",")
	for _, i := range includes {
		fmt.Printf("%#v\n", i)
		if !sliceContains(accepted, i) {
			return i, nil, ErrIncludeQueryContainUnacceptedValue
		}
	}
	return "", includes, nil
}

// GetProductHandler returns a handler function that gets a product by SKU
// containing product and image data.
func (a *App) GetProductHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: GetProductHandler called")

		productID := chi.URLParam(r, "id")
		include := r.URL.Query().Get("include")
		unaccepted, includeList, err := parseIncludeQueryParam(include, []string{"images", "prices"})
		if err != nil {
			if err == ErrIncludeQueryParamParseError {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(struct {
					Status  int    `json:"status"`
					Code    string `json:"code"`
					Message string `json:"message"`
				}{
					http.StatusBadRequest,
					ErrCodeIncludeQueryParamParseError,
					"include query parameter not valid",
				})
				return

			} else if err == ErrIncludeQueryContainUnacceptedValue {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(struct {
					Status  int    `json:"status"`
					Code    string `json:"code"`
					Message string `json:"message"`
				}{
					http.StatusBadRequest,
					ErrCodeIncludeQueryParamParseError,
					fmt.Sprintf("%s is not an acceptable value for an include in this context", unaccepted),
				})
				return
			}
			contextLogger.Errorf("app: parseIncludeQueryParam(include=%q) error: %+v", include, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}

		// optional: include the image sub-resource
		var includeImages bool
		if sliceContains(includeList, "images") {
			includeImages = true
		}
		var includePrices bool
		if sliceContains(includeList, "prices") {
			includePrices = true
		}

		product, err := a.Service.GetProduct(ctx, productID, includeImages, includePrices)
		if err != nil {
			if err == service.ErrProductNotFound {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			contextLogger.Errorf("app: GetProduct(ctx, productID=%q) error: %+v", productID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(*product)
	}
}
