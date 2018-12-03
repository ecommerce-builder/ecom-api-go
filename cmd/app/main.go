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
	"github.com/go-chi/chi"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/api/option"
)

var version string

// dsn is the Data Source name. For PostgreSQL the format is "host=localhost port=5432 user=postgres password=secret dbname=mydatabase sslmode=disable". The sslmode is optional.
var dsn = os.Getenv("ECOM_DSN")

var credentialsJSON = os.Getenv("ECOM_CREDENTIALS_JSON")

func initLogging() {
	// Output logs with colour
	log.SetFormatter(&log.TextFormatter{
		ForceColors: true,
	})

	// Output to stdout instead of the default stderr
	log.SetOutput(os.Stdout)

	// Log debug level severity or above.
	log.SetLevel(log.DebugLevel)
}

func main() {
	initLogging()

	if dsn == "" {
		log.Fatal("missing DSN. Use export ECOM_DSN")
	}

	if credentialsJSON == "" {
		log.Fatal("missing service account credentials. Use export ECOM_CREDENTIALS_JSON")
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
	ctx := context.Background()
	opt := option.WithCredentialsJSON([]byte(credentialsJSON))
	fbApp, err = firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalf("%v", fmt.Errorf("failed to initialise Firebase app: %v", err))
	}

	fbAuthClient, err = fbApp.Auth(ctx)

	// build a Firebase service injecting in the model and firebase app as dependencies
	fbSrv, _ := service.New(pgModel, fbApp, fbAuthClient)
	a := app.App{
		Service: fbSrv,
	}

	r := chi.NewRouter()

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedHeaders: []string{"Authorization", "Content-Type"},
		AllowedMethods: []string{"GET", "POST", "PATCH", "DELETE"},
		//AllowCredentials: true,
		// Enable Debugging for testing, consider disabling in production
		Debug: false,
	})
	r.Use(c.Handler)

	//r.Use(middleware.Logger)
	r.Use(a.AuthenticateMiddleware)

	// version info
	r.Get("/info", infoHandler)

	// Customer and address management API
	r.Post("/customers", a.Authorization("CreateCustomer", a.CreateCustomerController()))
	r.Get("/customers/{cuuid}", a.Authorization("GetCustomer", a.GetCustomerHandler()))
	r.Post("/customers/{cuuid}/addresses", a.Authorization("CreateAddress", a.CreateAddressController()))

	r.Get("/addresses/{auuid}", a.Authorization("GetAddress", a.GetAddressController()))
	r.Get("/customers/{cuuid}/addresses", a.Authorization("GetCustomersAddresses", a.ListAddressesController()))
	r.Patch("/customers/{cuuid}/addresses/{auuid}", a.Authorization("UpdateAddress", a.UpdateAddressController()))
	r.Delete("/addresses/{auuid}", a.Authorization("DeleteAddress", a.DeleteAddressController()))

	// carts API
	r.Post("/carts", a.Authorization("CreateCart", a.CreateCartController()))
	r.Post("/carts/{ctid}/items", a.Authorization("AddItemToCart", a.AddItemToCartController()))
	r.Get("/carts/{ctid}/items", a.Authorization("GetCartItems", a.GetCartItemsController()))
	r.Patch("/carts/{ctid}/items/{sku}", a.Authorization("UpdateCartItem", a.UpdateCartItemController()))
	r.Delete("/carts/{ctid}/items/{sku}", a.Authorization("DeleteCartItem", a.DeleteCartItemController()))
	r.Delete("/carts/{ctid}/items", a.Authorization("EmptyCartItems", a.EmptyCartItemsController()))

	log.Info("Server starting on port 9000")
	log.Fatal(http.ListenAndServe(":9000", r))
}

func infoHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK) // 200 OK
	fmt.Fprintf(w, "Hello, world!\n")
}
