package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"bitbucket.org/andyfusniakteam/ecom-api-go"
	model "bitbucket.org/andyfusniakteam/ecom-api-go/model/postgres"
	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"cloud.google.com/go/pubsub"
	"firebase.google.com/go"
	"firebase.google.com/go/auth"
	"github.com/go-chi/chi"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/option"
)

var version string

// dsn is the Data Source name. For PostgreSQL the format is "host=localhost port=5432 user=postgres password=secret dbname=mydatabase sslmode=disable". The sslmode is optional.
var dsn = os.Getenv("ECOM_DSN")
var credentialsFile = os.Getenv("ECOM_CREDENTIALS_FILE")
var projectID = os.Getenv("ECOM_PROJECT_ID")
var port = os.Getenv("ECOM_PORT")
var tlsModeFlag = os.Getenv("ECOM_TLS_MODE")
var tlsCertFile = os.Getenv("ECOM_TLS_CERT_FILE")
var tlsKeyFile = os.Getenv(("ECOM_TLS_KEY_FILE"))

const (
	// secretVolume points to the base path for the mounted drive or secret files using k8s or docker mount
	secretVolume = "/etc/secret-volume"

	// directory in the secret volume that holds all Service Account Credentials files
	sacDir      = "service_account_credentials"
	pubsubTopic = "ecom-api"
)

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

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte("ok"))
}

func infoHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK) // 200 OK
	fmt.Fprintf(w, fmt.Sprintf(`{"app_version":"%s"}`, version))
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func main() {
	initLogging()

	log.Infof("ecom-api started NEW ONE")

	// 1. Data Source Name
	if dsn == "" {
		log.Fatal("missing DSN. Use export ECOM_DSN")
	}

	// 2. Service Account Credentials
	if credentialsFile == "" {
		log.Fatal("missing service account credentials file. Use export ECOM_CREDENTIALS_FILE=/path/to/your/service-account-file")
	}
	// if the credentialsFile is a relative pathname, make it relative to the secretVolume/sacDir root
	// i.e. /etc/secret-volume/service_account_credentials/<file>
	if !filepath.IsAbs(credentialsFile) {
		log.Debugf("credentialsFile is a relative pathname so building absolute pathname")
		credentialsFile = filepath.Join(secretVolume, sacDir, credentialsFile)
	}
	ex, err := exists(credentialsFile)
	if err != nil {
		log.Fatalf("failed to determine if Service Account Credentials File %s exists: %v", credentialsFile, err)
	}
	if !ex {
		log.Fatalf("cannot find Service Account Credentials File %s. Check permissions.", credentialsFile)
	}
	log.Infof("Service Account Credentials File: %s", credentialsFile)

	// 3. Google Project ID
	if projectID == "" {
		log.Fatal("missing project ID. Use export ECOM_PROJECT_ID")
	}

	// 4. Server Port
	if port == "" {
		port = "8080"
		log.Infof("no port specified using default port %s", port)
	} else {
		log.Infof("environment specifies port %s to be used", port)
	}

	ex, err = exists(secretVolume)
	if err != nil {
		log.Fatalf("failed to determine if secret volume %s exists: %v", secretVolume, err)
	}
	if !ex {
		log.Fatalf("cannot find secret volume %s. Have you mounted it?", secretVolume)
	}

	// TLS Mode defaults to false unless ECOM_TLS_MODE is set to enable
	tlsMode := false
	if tlsModeFlag == "enable" || tlsModeFlag == "enabled" {
		tlsMode = true
		log.Info("TLS Mode enabled")

		// Ensure the TLS Certificate and Key files exist
		if tlsCertFile == "" {
			log.Fatal("TLS_MODE is enabled so you must set the cert file. Use export ECOM_TLS_CERT_FILE=/path/to/your/cert.pem")
		}

		// if the tlsCertFile is a relative pathname, make it relative to the secretVolume root
		if !filepath.IsAbs(tlsCertFile) {
			log.Debugf("tlsCertFile is a relative pathname so building absolute pathname")
			tlsCertFile = filepath.Join(secretVolume, tlsCertFile)
		}
		ex, err := exists(tlsCertFile)
		if err != nil {
			log.Fatalf("failed to determine if TLS Cert File %s exists: %v", tlsCertFile, err)
		}
		if !ex {
			log.Fatalf("cannot find TLS Cert File %s. Check permissions.", tlsCertFile)
		}
		log.Infof("TLS Certificate File: %s", tlsCertFile)

		if tlsKeyFile == "" {
			log.Fatal("TLS_MODE is enabled so you must set the key file. Use export ECOM_TLS_KEY_FILE=/path/to/your/key.pem")
		}
		if !filepath.IsAbs(tlsKeyFile) {
			log.Debugf("tlsKeyFile is a relative pathname so building absolute pathname")
			tlsKeyFile = filepath.Join(secretVolume, tlsKeyFile)
		}
		ex, err = exists(tlsKeyFile)
		if err != nil {
			log.Fatalf("failed to determine if TLS Key File %s exists: %v", tlsKeyFile, err)
		}
		if !ex {
			log.Fatalf("cannot find TLS Cert File %s. Check permissions.", tlsKeyFile)
		}
		log.Infof("TLS Key File: %s", tlsKeyFile)
	}

	// connect to postgres
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

	// PubSub
	ctx := context.Background()
	opt := option.WithCredentialsFile(credentialsFile)
	log.Infof("initializing Google PubSub client for project %s", projectID)
	psClient, err := pubsub.NewClient(ctx, projectID, opt)
	defer psClient.Close()
	if err != nil {
		log.Fatal(err)
	}

	topic := psClient.Topic(pubsubTopic)
	defer topic.Stop()
	log.Infof("created pubsub topic %v", topic.String())

	// Broadcast startup message
	pr := topic.Publish(ctx, &pubsub.Message{
		Data: []byte("hello world"),
	})

	serverID, err := pr.Get(ctx)

	fmt.Println(serverID, err)

	// build a Google Firebase App
	var fbApp *firebase.App
	var fbAuthClient *auth.Client
	fbApp, err = firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalf("%v", fmt.Errorf("failed to initialise Firebase app: %v", err))
	}

	fbAuthClient, err = fbApp.Auth(ctx)
	log.Info("Firebase Auth initialised")

	// build a Firebase service injecting in the model and firebase app as dependencies
	fbSrv, _ := service.New(pgModel, fbApp, fbAuthClient, psClient)
	a := app.App{
		Service: fbSrv,
	}

	r := chi.NewRouter()

	// protected routes
	r.Group(func(r chi.Router) {
		c := cors.New(cors.Options{
			AllowedOrigins: []string{"*"},
			AllowedHeaders: []string{"Authorization", "Content-Type"},
			AllowedMethods: []string{"GET", "POST", "PATCH", "DELETE"},
			//AllowCredentials: true,
			// Enable Debugging for testing, consider disabling in production
			Debug: false,
		})
		r.Use(c.Handler)
		r.Use(a.AuthenticateMiddleware)

		// Customer and address management API
		r.Route("/customers", func(r chi.Router) {
			r.Post("/", a.Authorization("CreateCustomer", a.CreateCustomerController()))
			r.Get("/{cuuid}", a.Authorization("GetCustomer", a.GetCustomerHandler()))
			r.Post("/{cuuid}/addresses", a.Authorization("CreateAddress", a.CreateAddressController()))
			r.Get("/{cuuid}/addresses", a.Authorization("GetCustomersAddresses", a.ListAddressesController()))
			r.Patch("/{cuuid}/addresses/{auuid}", a.Authorization("UpdateAddress", a.UpdateAddressController()))
		})

		r.Route("/addresses", func(r chi.Router) {
			r.Get("/{auuid}", a.Authorization("GetAddress", a.GetAddressController()))
			r.Delete("/{auuid}", a.Authorization("DeleteAddress", a.DeleteAddressController()))
		})

		r.Route("/carts", func(r chi.Router) {
			r.Post("/", a.Authorization("CreateCart", a.CreateCartController()))
			r.Post("/{ctid}/items", a.Authorization("AddItemToCart", a.AddItemToCartController()))
			r.Get("/{ctid}/items", a.Authorization("GetCartItems", a.GetCartItemsController()))
			r.Patch("/{ctid}/items/{sku}", a.Authorization("UpdateCartItem", a.UpdateCartItemController()))
			r.Delete("/{ctid}/items/{sku}", a.Authorization("DeleteCartItem", a.DeleteCartItemController()))
			r.Delete("/{ctid}/items", a.Authorization("EmptyCartItems", a.EmptyCartItemsController()))
		})
	})

	// public routes including GET / for Google Kuberenetes default healthcheck
	r.Group(func(r chi.Router) {
		// r.Use(middleware.NoCache)

		// version info
		r.Get("/", healthCheckHandler)
		r.Get("/healthz", healthCheckHandler)
		r.Get("/info", infoHandler)
	})

	//
	if tlsMode {
		log.Infof("Server listening on HTTPS port %s", port)
		log.Fatal(http.ListenAndServeTLS(fmt.Sprintf(":%s", port), tlsCertFile, tlsKeyFile, r))
	} else {
		log.Infof("Server listening on HTTP port %s", port)
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), r))
	}
}
