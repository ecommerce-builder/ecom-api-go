package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	controllers "bitbucket.org/andyfusniakteam/ecom-api-go/controllers"
	"bitbucket.org/andyfusniakteam/ecom-api-go/models"
	"bitbucket.org/andyfusniakteam/ecom-api-go/services"
	"cloud.google.com/go/logging"
	firebase "firebase.google.com/go"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
	"golang.org/x/net/context"
	"google.golang.org/api/option"
)

const (
	// LogName is the name of the log on Stack driver
	LogName = "ecom-api"
)

// ProjectID contains the Google project ID e.g. "ecom-test-bf262"
var ProjectID = os.Getenv("ECOM_PROJECT_ID")

// CredentialsFile contains the filename of the service credentials needed for Firebase Auth
var CredentialsFile = os.Getenv("ECOM_CREDENTIALS_FILE")

func main() {
	if ProjectID == "" {
		fmt.Fprintf(os.Stderr, "no ProjectID is set. Use export ECOM_PROJECT_ID=<your_project>\n")
		os.Exit(1)
	}

	if CredentialsFile == "" {
		fmt.Fprintf(os.Stderr, "missing credentials file. Use export ECOM_CREDENTIALS_FILE\n")
		os.Exit(1)
	}

	// Logging
	// Create a Client
	ctx := context.Background()
	client, err := logging.NewClient(ctx, ProjectID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "errror getting a logging client: %v", err)
		os.Exit(1)
	}
	defer client.Close()

	// Initialise a logger
	lg := client.Logger(LogName)
	controllers.Lg = lg
	services.Lg = lg
	lg.Log(logging.Entry{Payload: "API Started", Severity: logging.Info})

	// Database
	db, err := services.ConnectDb()
	if err != nil {
		panic(err)
	}
	controllers.DB = db
	models.DB = db

	// Firebase

	opt := option.WithCredentialsFile(CredentialsFile)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", fmt.Errorf("error initializing app: %v\n", err))
		os.Exit(1)
	}
	services.App = app

	r := mux.NewRouter()

	// Customer and address management API
	r.HandleFunc("/customers", controllers.CreateCustomer).Methods("POST")
	r.HandleFunc("/customers/{cid}", controllers.GetCustomer).Methods("GET")
	r.HandleFunc("/customers/{cid}/addresses", controllers.CreateAddress).Methods("POST")
	r.HandleFunc("/addresses/{aid}", controllers.GetAddress).Methods("GET")
	r.HandleFunc("/customers/{cid}/addresses", controllers.ListAddresses).Methods("GET")
	r.HandleFunc("/customers/{cid}/addresses/{aid}", controllers.UpdateAddress).Methods("PATCH")
	r.HandleFunc("/addresses/{aid}", controllers.DeleteAddress).Methods("DELETE")

	r.HandleFunc("/carts", controllers.CreateCart).Methods("POST")
	r.HandleFunc("/carts/{ctid}/items", controllers.AddItemToCart).Methods("POST")
	r.HandleFunc("/carts/{ctid}/items", controllers.GetCartItems).Methods("GET")
	r.HandleFunc("/carts/{ctid}/items/{sku}", controllers.UpdateCartItem).Methods("PATCH")
	r.HandleFunc("/carts/{ctid}/items/{sku}", controllers.DeleteCartItem).Methods("DELETE")
	r.HandleFunc("/carts/{ctid}/items", controllers.EmptyCartItems).Methods("DELETE")

	handler := cors.Default().Handler(r)
	log.Fatal(http.ListenAndServe(":9000", handler))
}
