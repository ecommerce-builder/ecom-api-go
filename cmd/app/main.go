package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"bitbucket.org/andyfusniakteam/ecom-api-go"
	model "bitbucket.org/andyfusniakteam/ecom-api-go/model/postgres"
	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"firebase.google.com/go"
	"firebase.google.com/go/auth"
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
	var fbAuthClient *auth.Client

	if credentialsJSON != "" {
		ctx := context.Background()
		opt := option.WithCredentialsJSON([]byte(credentialsJSON))
		fbApp, err = firebase.NewApp(ctx, nil, opt)
		if err != nil {
			log.Fatalf("%v", fmt.Errorf("failed to initialise Firebase app: %v", err))
		}

		fbAuthClient, err = fbApp.Auth(ctx)
	} else {
		ctx := context.Background()
		fbApp, err = firebase.NewApp(ctx, nil)
		if err != nil {
			log.Fatalf("%v", fmt.Errorf("failed to initialise Firebase app: %v", err))
		}

		fbAuthClient, err = fbApp.Auth(ctx)
	}

	// build a Firebase service injecting in the model and firebase app as dependencies
	fbSrv, _ := service.New(pgModel, fbApp, fbAuthClient)

	a := app.App{
		Service: fbSrv,
	}

	r := mux.NewRouter()

	r.HandleFunc("/", indexHandler).Methods("GET")
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


	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedHeaders: []string{"Authorization", "Content-Type"},
		AllowedMethods: []string{"POST", "PATCH", "DELETE"},
		//AllowCredentials: true,
		// Enable Debugging for testing, consider disabling in production
		Debug: true,
	})

	//c := cors.Default()

	handler := c.Handler(r)
	log.Fatal(http.ListenAndServe(":9000", handler))
	log.Info("Server started on port 9000")
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	//w.WriteHeader(http.StatusOK) // 200 OK
	fmt.Fprintf(w, "Hello, world!\n")
}
