package services

import (
	"database/sql"
	"fmt"
	"os"
)

// ConnectDb establishes a connection to the database
func ConnectDb() (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s ", os.Getenv("ECOM_DBHOST"), os.Getenv("ECOM_DBPORT"), os.Getenv("ECOM_DBUSER"), os.Getenv("ECOM_DBPASS"), os.Getenv("ECOM_DBNAME"))

	// default is to use SSL for DB connections
	if os.Getenv("ECOM_SSL") == "disable" {
		dsn = dsn + " sslmode=disable"
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		fmt.Println("Failed to open db", err)
	}

	err = db.Ping()
	if err != nil {
		fmt.Println("Failed to verify db connection", err)
	}

	return db, nil
}

// CloseDb closes the database connection
func CloseDb(db *sql.DB) {
	defer db.Close()
}
