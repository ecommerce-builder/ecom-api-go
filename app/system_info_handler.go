package app

import (
	"encoding/json"
	"net/http"
)

type SystemInfo struct {
	Version string    `json:"api_version"`
	Env     SystemEnv `json:"env"`
}

type SystemEnv struct {
	PG   PgSystemEnv   `json:"pg"`
	Goog GoogSystemEnv `json:"google"`
	App  AppSystemEnv  `json:"app"`
}

type PgSystemEnv struct {
	PgHost     string `json:"ECOM_PG_HOST"`
	PgPort     string `json:"ECOM_PG_PORT"`
	PgDatabase string `json:"ECOM_PG_DATABASE"`
	PgUser     string `json:"ECOM_PG_USER"`
	PgSSLMode  string `json:"ECOM_PG_SSLMODE"`
}

type GoogSystemEnv struct {
	GoogProjectID string `json:"ECOM_GOOGLE_PROJECT_ID"`
}

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
