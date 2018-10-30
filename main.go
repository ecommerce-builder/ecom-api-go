package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	controllers "bitbucket.org/andyfusniakteam/ecom-api-go/controllers"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func main() {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s ", os.Getenv("ECOM_DBHOST"), os.Getenv("ECOM_DBPORT"), os.Getenv("ECOM_DBUSER"), os.Getenv("ECOM_DBPASS"), os.Getenv("ECOM_DBNAME"))

	// default is to use SSL for DB connections
	if os.Getenv("ECOM_SSL") == "disable" {
		dsn = dsn + " sslmode=disable"
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		fmt.Println("Failed to open db", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		fmt.Println("Failed to verify db connection", err)
	}

	r := mux.NewRouter()

	r.HandleFunc("/customers", controllers.CreateCustomer).Methods("POST")
	r.HandleFunc("/customers/{cid}", controllers.GetCustomer).Methods("GET")
	r.HandleFunc("/customers/{cid}/addresses", controllers.CreateAddress).Methods("POST")
	r.HandleFunc("/customers/{cid}/addresses/{aid}", controllers.DeleteAddress).Methods("DELETE")
	r.HandleFunc("/customers/{cid}/addresses/{aid}", controllers.GetAddress).Methods("GET")
	r.HandleFunc("/customers/{cid}/addresses", controllers.ListAddresses).Methods("GET")
	r.HandleFunc("/customers/{cid}/addresses/{aid}", controllers.UpdateAddress).Methods("PATCH")

	log.Fatal(http.ListenAndServe(":8080", r))
}
