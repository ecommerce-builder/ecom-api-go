package firebase

import (
	"database/sql"
	"flag"
	"fmt"
	"testing"

	"bitbucket.org/andyfusniakteam/ecom-api-go/model/postgres"
)

var ecomPgPassword = flag.String("pgpass", "5oNnH50f1bz91pAF", "Set the postgres password")

func setup(t *testing.T) (*postgres.PgModel, func()) {
	dsn := fmt.Sprintf("host=postgres.open24seven.co.uk port=5432 user=postgres password=%s dbname=ecom_dev sslmode=disable connect_timeout=10", *ecomPgPassword)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
		return nil, func() {}
	}
	err = db.Ping()
	if err != nil {
		t.Fatalf("failed to verify db connection: %v", err)
		return nil, func() {}
	}
	model := postgres.NewPgModel(db)
	return model, func() {
		if err := db.Close(); err != nil {
			t.Errorf("db.Close(): %s", err)
		}
	}
}
