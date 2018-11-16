package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"bitbucket.org/andyfusniakteam/ecom-api-go"
	model "bitbucket.org/andyfusniakteam/ecom-api-go/model/postgres"
	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	firebase "firebase.google.com/go"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
	"golang.org/x/net/context"
)

// DSN is the Data Source name. For PostgreSQL the format is "host=localhost port=5432 user=postgres password=secret dbname=mydatabase sslmode=disable". The sslmode is optional.
var DSN = os.Getenv("ECOM_DSN")

// CredentialsJSON contains the JSON string of the service credentials needed for Firebase Auth
//var CredentialsJSON = os.Getenv("ECOM_CREDENTIALS_JSON")

func main() {
	if DSN == "" {
		log.Fatal("missing DSN. Use export ECOM_DSN")
	}

	// if CredentialsJSON == "" {
	// 	log.Fatal("missing credentials file. Use export ECOM_CREDENTIALS_JSON")
	// }

	db, err := sql.Open("postgres", DSN)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("failed to verify db connection: %v", err)
	}

	// build a Postgres model
	pgModel, _ := model.New(db)

	// build a Google Firebase App
	//opt := option.WithCredentialsJSON([]byte(CredentialsJSON))
	fbApp, err := firebase.NewApp(context.Background(), nil)
	if err != nil {
		log.Fatalf("%v", fmt.Errorf("failed to initialise Firebase app: %v", err))
	}

	// build a Firebase service injecting in the model and firebase app as dependencies
	fbSrv, _ := service.New(pgModel, fbApp)

	a := app.App{
		Service: fbSrv,
	}

	r := mux.NewRouter()

	// Customer and address management API
	r.HandleFunc("/customers", a.CreateCustomerController()).Methods("POST")
	r.HandleFunc("/customers/{cid}", a.GetCustomerController()).Methods("GET")
	r.HandleFunc("/customers/{cid}/addresses", a.CreateAddressController()).Methods("POST")
	r.HandleFunc("/addresses/{aid}", a.GetAddressController()).Methods("GET")
	r.HandleFunc("/customers/{cid}/addresses", a.ListAddressesController()).Methods("GET")
	r.HandleFunc("/customers/{cid}/addresses/{aid}", a.UpdateAddressController()).Methods("PATCH")
	r.HandleFunc("/addresses/{aid}", a.DeleteAddressController()).Methods("DELETE")

	r.HandleFunc("/carts", a.CreateCartController()).Methods("POST")
	r.HandleFunc("/carts/{ctid}/items", a.AddItemToCartController()).Methods("POST")
	r.HandleFunc("/carts/{ctid}/items", a.GetCartItemsController()).Methods("GET")
	r.HandleFunc("/carts/{ctid}/items/{sku}", a.UpdateCartItemController()).Methods("PATCH")
	r.HandleFunc("/carts/{ctid}/items/{sku}", a.DeleteCartItemController()).Methods("DELETE")
	r.HandleFunc("/carts/{ctid}/items", a.EmptyCartItemsController()).Methods("DELETE")

	handler := cors.Default().Handler(r)
	log.Fatal(http.ListenAndServe(":9000", handler))
}
