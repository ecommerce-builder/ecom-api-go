package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"bitbucket.org/andyfusniakteam/ecom-api-go"
	model "bitbucket.org/andyfusniakteam/ecom-api-go/model/postgres"
	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	firebase "firebase.google.com/go"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/api/option"
)

// dsn is the Data Source name. For PostgreSQL the format is "host=localhost port=5432 user=postgres password=secret dbname=mydatabase sslmode=disable". The sslmode is optional.
var dsn = os.Getenv("ECOM_DSN")

var credentialsJSON = os.Getenv("ECOM_CREDENTIALS_JSON")

func initLogging() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.TextFormatter{
		ForceColors: true,
	})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.DebugLevel)
}

func main() {
	initLogging()

	if dsn == "" {
		log.Fatal("missing DSN. Use export ECOM_DSN")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("failed to verify db connection: %v", err)
	}

	log.Infoln("established database connection")

	// build a Postgres model
	pgModel, _ := model.New(db)

	// build a Google Firebase App
	var fbApp *firebase.App
	if credentialsJSON != "" {
		opt := option.WithCredentialsJSON([]byte(credentialsJSON))
		fbApp, err = firebase.NewApp(context.Background(), nil, opt)
	} else {
		fbApp, err = firebase.NewApp(context.Background(), nil)
	}

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
	r.HandleFunc("/customers/{uid}", a.AuthenticateMiddleware(a.GetCustomerController())).Methods("GET")
	r.HandleFunc("/customers/{cid}/addresses", a.AuthenticateMiddleware(a.CreateAddressController())).Methods("POST")

	r.HandleFunc("/addresses/{aid}", a.AuthenticateMiddleware(a.GetAddressController())).Methods("GET")
	r.HandleFunc("/customers/{cid}/addresses", a.AuthenticateMiddleware(a.ListAddressesController())).Methods("GET")
	r.HandleFunc("/customers/{cid}/addresses/{aid}", a.AuthenticateMiddleware(a.UpdateAddressController())).Methods("PATCH")
	r.HandleFunc("/addresses/{aid}", a.AuthenticateMiddleware(a.DeleteAddressController())).Methods("DELETE")

	r.HandleFunc("/carts", a.AuthenticateMiddleware(a.CreateCartController())).Methods("POST")
	r.HandleFunc("/carts/{ctid}/items", a.AuthenticateMiddleware(a.AddItemToCartController())).Methods("POST")
	r.HandleFunc("/carts/{ctid}/items", a.AuthenticateMiddleware(a.GetCartItemsController())).Methods("GET")
	r.HandleFunc("/carts/{ctid}/items/{sku}", a.AuthenticateMiddleware(a.UpdateCartItemController())).Methods("PATCH")
	r.HandleFunc("/carts/{ctid}/items/{sku}", a.AuthenticateMiddleware(a.DeleteCartItemController())).Methods("DELETE")
	r.HandleFunc("/carts/{ctid}/items", a.AuthenticateMiddleware(a.EmptyCartItemsController())).Methods("DELETE")

	handler := cors.Default().Handler(r)
	log.Fatal(http.ListenAndServe(":8080", handler))
}
