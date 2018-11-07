package models

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

var db *sql.DB

var dbhost = os.Getenv("ECOM_DBHOST")
var dbport = os.Getenv("ECOM_DBPORT")
var dbuser = os.Getenv("ECOM_DBUSER")
var dbpass = os.Getenv("ECOM_DBPASS")
var dbname = os.Getenv("ECOM_DBNAME")

func connectToDb() {
	if dbhost == "" || dbport == "" || dbuser == "" || dbpass == "" || dbname == "" {
		fmt.Fprintf(os.Stderr, "Make sure you set ECOM_DBHOST, ECOM_DBPORT, ECOM_DBUSER, ECOM_DBPASS, ECOM_DBNAME")
		os.Exit(1)
	}
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s ", dbhost, dbport, dbuser, dbpass, dbname)

	// default is to use SSL for DB connections
	if os.Getenv("ECOM_SSL") == "disable" {
		dsn = dsn + " sslmode=disable"
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open db", err)
        os.Exit(1)
	}

	DB = db
}

func TestCreateCutomerAndAddress(t *testing.T) {
	connectToDb()

	customer, _ := CreateCustomer("uuid1", "john@example.com", "John", "Doe")

	addr := CreateAddress(customer.GetID(), "billing", "Mr Test", "123 Lex Road", "Joy Lane", "Cambridge", "Cambridgeshire", "CB25 0HD", "UK")
}
