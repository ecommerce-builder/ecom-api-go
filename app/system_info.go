package app

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// SystemInfo contains the system version string and system environment data.
type SystemInfo struct {
	Version string    `json:"api_version"`
	Env     SystemEnv `json:"env"`
}

// SystemEnv contains the Postgres database, Google Firebase and Application data.
type SystemEnv struct {
	PG       PgSystemEnv       `json:"pg"`
	Goog     GoogSystemEnv     `json:"google"`
	Firebase FirebaseSystemEnv `json:"firebase"`
	Stripe   StripeSystemEnv   `json:"stripe"`
	App      ApplSystemEnv     `json:"app"`
}

// PgSystemEnv contains the environment and runtime settings
// for the Postgres database. `schema_version` is embedded in
// a postgres function setup by the initial schema create script
// outside of this app.
type PgSystemEnv struct {
	PgHost        string `json:"ECOM_PG_HOST"`
	PgPort        string `json:"ECOM_PG_PORT"`
	PgDatabase    string `json:"ECOM_PG_DATABASE"`
	PgUser        string `json:"ECOM_PG_USER"`
	PgSSLMode     string `json:"ECOM_PG_SSLMODE"`
	SchemaVersion string `json:"schema_version"`
}

// GoogSystemEnv contains the Google environment variables.
type GoogSystemEnv struct {
	GAEProjectID string `json:"ECOM_GAE_PROJECT_ID"`
}

// FirebaseSystemEnv contains the Firebase environment variables.
type FirebaseSystemEnv struct {
	APIKEY            string `json:"apiKey"`
	AuthDomain        string `json:"authDomain"`
	DatabaseURL       string `json:"databaseURL"`
	ProjectID         string `json:"projectId"`
	StorageBucket     string `json:"storageBucket"`
	MessagingSenderID string `json:"messagingSenderId"`
	AppID             string `json:"appId"`
}

// StripeSystemEnv contains the Stripe environment variables.
type StripeSystemEnv struct {
	StripeSuccessURL string `json:"ECOM_STRIPE_SUCCESS_URL"`
	StripeCancelURL  string `json:"ECOM_STRIPE_CANCEL_URL"`
}

// ApplSystemEnv contains the application port and root email address.
type ApplSystemEnv struct {
	AppPort                     string `json:"PORT"`
	AppRootEmail                string `json:"ECOM_APP_ROOT_EMAIL"`
	AppEnableStackDriverLogging bool   `json:"ECOM_APP_ENABLE_STACKDRIVER_LOGGING"`
	AppEndpoint                 string `json:"ECOM_APP_ENDPOINT"`
}

// SystemInfoHandler returns data about the API runtime
func (a *App) SystemInfoHandler(si SystemInfo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: SystemInfoHandler started")

		version, err := a.Service.GetSchemaVersion(ctx)
		if err != nil {
			contextLogger.Errorf("app: a.Service.GetSchemaVersion(ctx) failed: %+v", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		si.Env.PG.SchemaVersion = *version
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(si)
	}
}
