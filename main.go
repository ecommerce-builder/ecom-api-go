package main

import (
	"log"
	"net/http"

	controllers "bitbucket.org/andyfusniakteam/ecom-api-go/controllers"
	"bitbucket.org/andyfusniakteam/ecom-api-go/models"
	"bitbucket.org/andyfusniakteam/ecom-api-go/services"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func main() {
	db, err := services.ConnectDb()
	if err != nil {
		panic(err)
	}
	controllers.DB = db
	models.DB = db

	r := mux.NewRouter()

	r.HandleFunc("/customers", controllers.CreateCustomer).Methods("POST")
	r.HandleFunc("/customers/{cid}", controllers.GetCustomer).Methods("GET")
	r.HandleFunc("/customers/{cid}/addresses", controllers.CreateAddress).Methods("POST")
	r.HandleFunc("/addresses/{aid}", controllers.DeleteAddress).Methods("DELETE")
	r.HandleFunc("/addresses/{aid}", controllers.GetAddress).Methods("GET")
	r.HandleFunc("/customers/{cid}/addresses", controllers.ListAddresses).Methods("GET")
	r.HandleFunc("/customers/{cid}/addresses/{aid}", controllers.UpdateAddress).Methods("PATCH")

	log.Fatal(http.ListenAndServe(":8080", r))
}
