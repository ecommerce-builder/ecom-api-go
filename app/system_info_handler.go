package app

import (
	"encoding/json"
	"net/http"
)

// SystemInfo contains the system version string and system environment data.
type SystemInfo struct {
	Version string    `json:"api_version"`
	Env     SystemEnv `json:"env"`
}

// SystemEnv contains the Postgres database, Google Firebase and Application data.
type SystemEnv struct {
	PG   PgSystemEnv   `json:"pg"`
	Goog GoogSystemEnv `json:"google"`
	App  AppSystemEnv  `json:"app"`
}

// PgSystemEnv contains the environment settings for the Postgres database.
type PgSystemEnv struct {
	PgHost     string `json:"ECOM_PG_HOST"`
	PgPort     string `json:"ECOM_PG_PORT"`
	PgDatabase string `json:"ECOM_PG_DATABASE"`
	PgUser     string `json:"ECOM_PG_USER"`
	PgSSLMode  string `json:"ECOM_PG_SSLMODE"`
}

// GoogSystemEnv contains the Google Firebase environment variables.
type GoogSystemEnv struct {
	GoogProjectID string `json:"ECOM_GOOGLE_PROJECT_ID"`
	WebAPIKey     string `json:"ECOM_GOOGLE_WEB_API_KEY"`
}

// AppSystemEnv contains the application port and root email address.
type AppSystemEnv struct {
	AppPort      string `json:"PORT"`
	AppRootEmail string `json:"ECOM_APP_ROOT_EMAIL"`
}

// SystemInfoHandler returns data about the API runtime
func (app *App) SystemInfoHandler(si SystemInfo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(si)
	}
}
