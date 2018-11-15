package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"bitbucket.org/andyfusniakteam/ecomapi"
	model "bitbucket.org/andyfusniakteam/ecomapi/model/postgres"
	service "bitbucket.org/andyfusniakteam/ecomapi/service/firebase"
	firebase "firebase.google.com/go"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
	"golang.org/x/net/context"
	"google.golang.org/api/option"
)

// CredentialsFile contains the filename of the service credentials needed for Firebase Auth
var CredentialsFile = os.Getenv("ECOM_CREDENTIALS_FILE")

func main() {
	if CredentialsFile == "" {
		fmt.Fprintf(os.Stderr, "missing credentials file. Use export ECOM_CREDENTIALS_FILE\n")
		os.Exit(1)
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s ", os.Getenv("ECOM_DBHOST"), os.Getenv("ECOM_DBPORT"), os.Getenv("ECOM_DBUSER"), os.Getenv("ECOM_DBPASS"), os.Getenv("ECOM_DBNAME"))

	// default is to use SSL for DB connections
	if os.Getenv("ECOM_SSL") == "disable" {
		dsn = dsn + " sslmode=disable"
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open db: %v", err)
		os.Exit(1)
	}

	err = db.Ping()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to verify db connection: %v", err)
		os.Exit(1)
	}

	// build a Postgres model
	pgModel, _ := model.New(db)

	// build a Google Firebase App
	opt := option.WithCredentialsFile(CredentialsFile)
	fbApp, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", fmt.Errorf("error initializing app: %v", err))
		os.Exit(1)
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
