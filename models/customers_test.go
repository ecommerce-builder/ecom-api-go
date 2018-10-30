package models

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

var db *sql.DB

func connectToDb() {
	var err error

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s ", os.Getenv("ECOM_DBHOST"), os.Getenv("ECOM_DBPORT"), os.Getenv("ECOM_DBUSER"), os.Getenv("ECOM_DBPASS"), os.Getenv("ECOM_DBNAME"))

	// default is to use SSL for DB connections
	if os.Getenv("ECOM_SSL") == "disable" {
		dsn = dsn + " sslmode=disable"
	}

	db, err = sql.Open("postgres", dsn)
	if err != nil {
		fmt.Println("Failed to open db", err)
	}
}

func TestCreateCutomerAndAddress(t *testing.T) {
	connectToDb()

	customer := CreateCustomer(db, "John", "Doe")

	addr := customer.CreateAddress("billing", "Mr Test", "123 Lex Road", "Joy Lane", "Cambridge", "Cambridgeshire", "GB", "CB25 0HD")

	fmt.Println(addr)
}
