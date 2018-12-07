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

// set at compile-time using -ldflags "-X main.version=$VERSION"
var version string

var (
	//
	// PostgreSQL Database settings
	//

	// Name of host to connect to.
	pghost = os.Getenv("ECOM_PG_HOST")

	// Port number to connect to at the server host, or socket file name extension
	// for Unix-domain connections.
	pgport = os.Getenv("ECOM_PG_PORT")

	// The database name. Defaults to be the same as the user name.
	pgdatabase = os.Getenv("ECOM_PG_DATABASE")

	// PostgreSQL user name to connect as. Defaults to be the same as the operating system name
	// of the user running the application.
	pguser = os.Getenv("ECOM_PG_USER")

	// Password to be used if the server demands password authentication.
	pgpassword = os.Getenv("ECOM_PG_PASSWORD")

	// This option determines whether or with what priority a secure SSL TCP/IP connection will be
	// negotiated with the server. There are six modes:
	// 	disable
	// 		only try a non-SSL connection
	//
	// 	allow
	//		first try a non-SSL connection; if that fails, try an SSL connection
	//
	// 	prefer (default)
	//		first try an SSL connection; if that fails, try a non-SSL connection
	//
	//	require
	//		only try an SSL connection. If a root CA file is present, verify the
	//		certificate in the same way as if verify-ca was specified
	//
	// 	verify-ca
	//		only try an SSL connection, and verify that the server certificate
	//		is issued by a trusted certificate authority (CA)
	//
	// 	verify-full
	//		only try an SSL connection, verify that the server certificate is issued by
	//		a trusted CA and that the server host name matches that in the certificate

	pgsslmode = os.Getenv("ECOM_PG_SSLMODE")

	// This parameter specifies the file name of the client SSL certificate, replacing the default
	// ~/.postgresql/postgresql.crt. This parameter is ignored if an SSL connection is not made.
	pgsslcert = os.Getenv("ECOM_PG_SSLCERT")

	// This parameter specifies the location for the secret key used for the client certificate. It can
	// either specify a file name that will be used instead of the default ~/.postgresql/postgresql.key, or
	// it can specify a key obtained from an external "engine" (engines are OpenSSL loadable modules). An
	// external engine specification should consist of a colon-separated engine name and an engine-specific
	// key identifier. This parameter is ignored if an SSL connection is not made.
	pgsslkey = os.Getenv("ECOM_PG_SSLKEY")

	// This parameter specifies the name of a file containing SSL certificate authority (CA)
	// certificate(s). If the file exists, the server's certificate will be verified to be
	// signed by one of these authorities.
	pgsslrootcert = os.Getenv("ECOM_PG_SSLROOTCERT")

	// Maximum wait for connection, in seconds (write as a decimal integer string). Zero or
	// not specified means wait indefinitely. It is not recommended to use a timeout of
	// less than 2 seconds.
	pgconnectTimeout = os.Getenv("ECOM_PG_CONNECT_TIMEOUT")

	//
	// Google settings
	//
	projectID       = os.Getenv("ECOM_GOOGLE_PROJECT_ID")
	credentialsFile = os.Getenv("ECOM_GOOGLE_CREDENTIALS_FILE")

	//
	// Application settings
	//
	port        = os.Getenv("ECOM_APP_PORT")
	tlsModeFlag = os.Getenv("ECOM_APP_TLS_MODE")
	tlsCertFile = os.Getenv("ECOM_APP_TLS_CERT_FILE")
	tlsKeyFile  = os.Getenv("ECOM_APP_TLS_KEY_FILE")
)

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

func mustHaveFile(path, title string) {
	ex, err := exists(path)
	if err != nil {
		log.Fatalf("failed to determine if %s %s exists: %v", title, path, err)
	}
	if !ex {
		log.Fatalf("cannot find %s %s. Check permissions.", title, path)
	}
	log.Infof("%s: %s", title, path)
}

func main() {
	initLogging()

	log.Infof("app started")

	// 1. Data Source Name
	// dsn is the Data Source name. For PostgreSQL the format is "host=localhost port=5432 user=postgres password=secret dbname=mydatabase sslmode=disable". The sslmode is optional.
	if pghost == "" {
		log.Fatal("postgres host not set. Use ECOM_PG_HOST")
	}

	if pgport == "" {
		log.Info("using default port of 5432 for postgres because ECOM_PG_PORT is not set")
		pgport = "5432"
	}

	if pguser == "" {
		log.Info("using default user postgres because ECOM_PG_USER is not set")
		pguser = "postgres"
	}

	if pgdatabase == "" {
		log.Fatal("ECOM_PG_DATABASE not set.")
	}

	if pgpassword == "" {
		log.Fatal("ECOM_PG_PASSWORD not set. You must set a password")
	}

	if pgsslmode == "" {
		if pgsslkey != "" || pgsslrootcert != "" || pgsslcert != "" {
			log.Fatal("ECOM_PG_SSLMODE was not set, but one or more of ECOM_PG_SSLCERT, ECOM_PG_SSLKEY, ECOM_PG_SSLROOTCERT environment variables were set implying you intended to connect to postgres securely?")
		}
		log.Infof("using postgres sslmode=disable because ECOM_PG_SSLMODE was not set")
		pgsslmode = "disable"
	}

	var dsn string
	if pgsslmode == "disable" {
		log.Infof("postgres running with sslmode=disable")
		dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", pghost, pgport, pguser, pgpassword, pgdatabase, pgsslmode)
		log.Infof("postgres dsn: host=%s port=%s user=%s password=**** dbname=%s sslmode=%s", pghost, pgport, pguser, pgdatabase)
	} else {
		// Ensure that the ECOM_PG_SSLCERT, ECOM_PG_SSLROOTCERT and ECOM_PG_SSLKEY are all
		// referenced using absolute paths.
		if pgsslcert == "" {
			log.Fatal("missing PostgreSQL SSL certificate file. Use export ECOM_PG_SSLCERT")
		}
		if !filepath.IsAbs(pgsslcert) {
			log.Fatalf("ECOM_PG_SSLCERT should use an absolute path to certificate file to avoid ambiguity")
		}
		mustHaveFile(pgsslcert, "client certificate file")

		if pgsslrootcert == "" {
			log.Fatal("missing PostgreSQL SSL root certificate file. Use export ECOM_PG_SSLROOTCERT")
		}
		if !filepath.IsAbs(pgsslrootcert) {
			log.Fatalf("ECOM_PG_SSLROOTCERT should use an absolute path to root certificate file to avoid ambiguity")
		}
		mustHaveFile(pgsslrootcert, "ssl root certificate")

		if pgsslkey == "" {
			log.Fatal("missing PostgreSQL SSL key certificate file. Use export ECOM_PG_SSLKEY")
		}
		if !filepath.IsAbs(pgsslkey) {
			log.Fatalf("ECOM_PG_SSLKEY should use an absolute path to key certificate file to avoid ambiguity")
		}
		mustHaveFile(pgsslkey, "ssl key file")

		dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s sslcert=%s sslrootcert=%s sslkey=%s", pghost, pgport, pguser, pgpassword, pgdatabase, pgsslmode, pgsslcert, pgsslrootcert, pgsslkey)
		log.Infof("postgres dsn: host=%s port=%s user=%s password=***** dbname=%s sslmode=%s sslcert=%s sslrootcert=%s sslkey=%s", pghost, pgport, pguser, pgdatabase, pgsslmode, pgsslcert, pgsslrootcert, pgsslkey)
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
	mustHaveFile(credentialsFile, "service account credentials file")

	// 3. Google Project ID
	if projectID == "" {
		log.Fatal("missing project ID. Use export ECOM_GOOGLE_PROJECT_ID")
	}

	// 4. Server Port
	if port == "" {
		port = "8080"
		log.Infof("no application port specified using default port %s", port)
	} else {
		log.Infof("environment variable ECOM_APP_PORT specifies port %s to be used", port)
	}

	// ensure that we have access to the secret volume
	// Google Kubernetes Engine attaches secrets to /etc/secret-volume mount points
	// Google Compute Engine attaches a persistent disk containing the necessary assets
	// Assets include PostgreSQL .pem files, Google Firebase service account keys and
	// TLS/SSL certificate files for HTTPS termination (see ECOM_APP_TLS_MODE=enable).
	ex, err := exists(secretVolume)
	if err != nil {
		log.Fatalf("failed to determine if secret volume %s exists: %v", secretVolume, err)
	}
	if !ex {
		log.Fatalf("cannot find secret volume %s. Have you mounted it?", secretVolume)
	}

	// TLS Mode defaults to false unless ECOM_APP_TLS_MODE is set to enable
	// tlsMode will be used to determine whether to provide negociation for SSL
	// (see the bottom of main func)
	tlsMode := false
	if tlsModeFlag == "enable" || tlsModeFlag == "enabled" {
		tlsMode = true
		log.Info("ECOM_APP_TLS_MODE enabled")

		// Ensure the TLS Certificate and Key files exist
		if tlsCertFile == "" {
			log.Fatal("ECOM_APP_TLS_MODE is enabled so you must set the cert file. Use export ECOM_APP_TLS_CERT_FILE=/path/to/your/cert.pem")
		}

		// if the tlsCertFile is a relative pathname, make it relative to the secretVolume root
		if !filepath.IsAbs(tlsCertFile) {
			log.Debugf("tlsCertFile is a relative pathname so building absolute pathname")
			tlsCertFile = filepath.Join(secretVolume, tlsCertFile)
		}
		mustHaveFile(tlsCertFile, "TLS Cert File")

		if tlsKeyFile == "" {
			log.Fatal("ECOM_APP_TLS_MODE is enabled so you must set the key file. Use export ECOM_APP_TLS_KEY_FILE=/path/to/your/key.pem")
		}
		if !filepath.IsAbs(tlsKeyFile) {
			log.Debugf("tlsKeyFile is a relative pathname so building absolute pathname")
			tlsKeyFile = filepath.Join(secretVolume, tlsKeyFile)
		}
		mustHaveFile(tlsKeyFile, "TLS Key File")
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
	log.Info("firebase auth initialised")

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

	// tlsMode determines whether to serve HTTPS traffic directly
	// if tlsMode is false, you can provide HTTPS via GKE layer 7 load balancer
	// using an Ingress.
	if tlsMode {
		log.Infof("server listening on HTTPS port %s", port)
		log.Fatal(http.ListenAndServeTLS(fmt.Sprintf(":%s", port), tlsCertFile, tlsKeyFile, r))
	} else {
		log.Infof("server listening on HTTP port %s", port)
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), r))
	}
}
