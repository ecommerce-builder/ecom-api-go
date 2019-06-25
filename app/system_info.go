package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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
	ProjectID string `json:"ECOM_FIREBASE_PROJECT_ID"`
	WebAPIKey string `json:"ECOM_FIREBASE_WEB_API_KEY"`
}

// ApplSystemEnv contains the application port and root email address.
type ApplSystemEnv struct {
	AppPort      string `json:"PORT"`
	AppRootEmail string `json:"ECOM_APP_ROOT_EMAIL"`
}

// SystemInfoHandler returns data about the API runtime
func (app *App) SystemInfoHandler(si SystemInfo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		version, err := app.Service.GetSchemaVersion(r.Context())
		if err != nil {
			fmt.Fprintf(os.Stderr, "%+v", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
		}
		si.Env.PG.SchemaVersion = *version
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(si)
	}
}
