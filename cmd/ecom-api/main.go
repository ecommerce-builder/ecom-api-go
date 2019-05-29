package main

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"bitbucket.org/andyfusniakteam/ecom-api-go/app"
	model "bitbucket.org/andyfusniakteam/ecom-api-go/model/postgres"
	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	firebase "firebase.google.com/go"
	_ "firebase.google.com/go/auth"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
	lg "github.com/sirupsen/logrus"
	"google.golang.org/api/option"
)

// set at compile-time using -ldflags "-X main.version=$VERSION"
var version = "v0.38.0"

const maxDbConnectAttempts = 3

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
	projectID   = os.Getenv("ECOM_GOOGLE_PROJECT_ID")
	webAPIKey   = os.Getenv("ECOM_GOOGLE_WEB_API_KEY")
	credentials = os.Getenv("ECOM_GOOGLE_CREDENTIALS")

	//
	// Application settings
	//
	port         = os.Getenv("PORT")
	tlsModeFlag  = os.Getenv("ECOM_APP_TLS_MODE")
	tlsCertFile  = os.Getenv("ECOM_APP_TLS_CERT")
	tlsKeyFile   = os.Getenv("ECOM_APP_TLS_KEY")
	rootEmail    = os.Getenv("ECOM_APP_ROOT_EMAIL")
	rootPassword = os.Getenv("ECOM_APP_ROOT_PASSWORD")
)

const (
	// secretVolume points to the base path for the mounted drive or secret files using k8s or docker mount
	secretVolume = "/etc/secret-volume"

	// directory in the secret volume that holds all Service Account Credentials files
	sacDir = "service_account_credentials"
)

func initLogging() {
	// Output logs with colour
	lg.SetFormatter(&lg.TextFormatter{
		ForceColors: true,
	})

	//lg.SetFormatter(stackdriver.NewFormatter(
	//	stackdriver.WithService("your-service"),
	//	stackdriver.WithVersion("v0.1.0"),
	//))

	// Output to stdout instead of the default stderr
	lg.SetOutput(os.Stdout)

	// Log debug level severity or above.
	lg.SetLevel(lg.DebugLevel)
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte("ok"))
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
		lg.Fatalf("failed to determine if %s %s exists: %v", title, path, err)
	}
	if !ex {
		lg.Fatalf("cannot find %s %s. Check permissions.", title, path)
	}
	lg.Infof("%s: %s", title, path)
}

func main() {
	initLogging()

	lg.Infof("app version %s started", version)

	// 1. Data Source Name
	// dsn is the Data Source name. For PostgreSQL the format is "host=localhost port=5432 user=postgres password=secret dbname=mydatabase sslmode=disable". The sslmode is optional.
	if pghost == "" {
		lg.Fatal("postgres host not set. Use ECOM_PG_HOST")
	}

	if pgport == "" {
		lg.Info("using default port=5432 for postgres because ECOM_PG_PORT is not set")
		pgport = "5432"
	}

	if pguser == "" {
		lg.Info("using default user=postgres because ECOM_PG_USER is not set")
		pguser = "postgres"
	}

	if pgdatabase == "" {
		lg.Fatal("ECOM_PG_DATABASE not set.")
	}

	if pgpassword == "" {
		lg.Fatal("ECOM_PG_PASSWORD not set. You must set a password")
	}

	if pgsslmode == "" {
		if pgsslkey != "" || pgsslrootcert != "" || pgsslcert != "" {
			lg.Fatal("ECOM_PG_SSLMODE is not set, but one or more of ECOM_PG_SSLCERT, ECOM_PG_SSLKEY, ECOM_PG_SSLROOTCERT environment variables were set implying you intended to connect to postgres securely?")
		}
		lg.Infof("using postgres sslmode=disable because ECOM_PG_SSLMODE is not set")
		pgsslmode = "disable"
	}

	if pgconnectTimeout == "" {
		lg.Infof("using postgres connect_timeout=10 because ECOM_PG_CONNECT_TIMEOUT is not set")
		pgconnectTimeout = "10"
	}

	var dsn string
	if pgsslmode == "disable" {
		lg.Infof("postgres running with sslmode=disable")
		dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s connect_timeout=%s", pghost, pgport, pguser, pgpassword, pgdatabase, pgsslmode, pgconnectTimeout)
		lg.Infof("postgres dsn: host=%s port=%s user=%s password=**** dbname=%s sslmode=%s connect_timeout=%s", pghost, pgport, pguser, pgdatabase, pgsslmode, pgconnectTimeout)
	} else {
		// Ensure that the ECOM_PG_SSLCERT, ECOM_PG_SSLROOTCERT and ECOM_PG_SSLKEY are all
		// referenced using absolute paths.
		if pgsslcert == "" {
			lg.Fatal("missing PostgreSQL SSL certificate file. Use export ECOM_PG_SSLCERT")
		}
		if !filepath.IsAbs(pgsslcert) {
			lg.Fatalf("ECOM_PG_SSLCERT should use an absolute path to certificate file to avoid ambiguity")
		}
		mustHaveFile(pgsslcert, "client certificate file")

		if pgsslrootcert == "" {
			lg.Fatal("missing PostgreSQL SSL root certificate file. Use export ECOM_PG_SSLROOTCERT")
		}
		if !filepath.IsAbs(pgsslrootcert) {
			lg.Fatalf("ECOM_PG_SSLROOTCERT should use an absolute path to root certificate file to avoid ambiguity")
		}
		mustHaveFile(pgsslrootcert, "ssl root certificate")

		if pgsslkey == "" {
			lg.Fatal("missing PostgreSQL SSL key certificate file. Use export ECOM_PG_SSLKEY")
		}
		if !filepath.IsAbs(pgsslkey) {
			lg.Fatalf("ECOM_PG_SSLKEY should use an absolute path to key certificate file to avoid ambiguity")
		}
		mustHaveFile(pgsslkey, "ssl key file")

		dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s sslcert=%s sslrootcert=%s sslkey=%s connect_timeout=%s", pghost, pgport, pguser, pgpassword, pgdatabase, pgsslmode, pgsslcert, pgsslrootcert, pgsslkey, pgconnectTimeout)
		lg.Infof("postgres dsn: host=%s port=%s user=%s password=***** dbname=%s sslmode=%s sslcert=%s sslrootcert=%s sslkey=%s connect_timeout=%s", pghost, pgport, pguser, pgdatabase, pgsslmode, pgsslcert, pgsslrootcert, pgsslkey, pgconnectTimeout)
	}

	// 2. Service Account Credentials
	if credentials == "" {
		lg.Fatal("missing service account credentials. Use export ECOM_GOOGLE_CREDENTIALS=/path/to/your/service-account-file or ECOM_GOOGLE_CREDENTIALS=<base64-json-file>")
	}
	// if the credentials is a relative pathname, make it relative to the secretVolume/sacDir root
	// i.e. /etc/secret-volume/service_account_credentials/<file>
	if credentials[0] == '/' {
		if !filepath.IsAbs(credentials) {
			lg.Debugf("credentials is a relative pathname so building absolute pathname")
			credentials = filepath.Join(secretVolume, sacDir, credentials)
		}
		mustHaveFile(credentials, "service account credentials")
	}

	// 3. Google Project ID and Web API Key
	if projectID == "" {
		lg.Fatal("missing project ID. Use export ECOM_GOOGLE_PROJECT_ID")
	}
	lg.Infof("project ID set to %s", projectID)
	if webAPIKey == "" {
		lg.Fatal("missing Web API Key. Use export ECOM_GOOGLE_WEB_API_KEY")
	}
	lg.Infof("Web API Key set to %s", webAPIKey)

	// 4. Server Port
	if port == "" {
		port = "8080"
		lg.Infof("no application port specified using default port %s", port)
	} else {
		lg.Infof("environment variable PORT specifies port %s to be used", port)
	}

	// ensure that we have access to the secret volume
	// Google Kubernetes Engine attaches secrets to /etc/secret-volume mount points
	// Google Compute Engine attaches a persistent disk containing the necessary assets
	// Assets include PostgreSQL .pem files, Google Firebase service account keys and
	// TLS/SSL certificate files for HTTPS termination (see ECOM_APP_TLS_MODE=enable).
	if credentials[0] == '/' {
		ex, err := exists(secretVolume)
		if err != nil {
			lg.Fatalf("failed to determine if secret volume %s exists: %v", secretVolume, err)
		}
		if !ex {
			lg.Fatalf("cannot find secret volume %s. Have you mounted it?", secretVolume)
		}
		lg.Infof("found secret volume %s", secretVolume)
	}

	// TLS Mode defaults to false unless ECOM_APP_TLS_MODE is set to enable
	// tlsMode will be used to determine whether to provide negociation for SSL
	// (see the bottom of main func)
	tlsMode := false
	if tlsModeFlag == "enable" || tlsModeFlag == "enabled" {
		tlsMode = true
		lg.Info("ECOM_APP_TLS_MODE enabled")

		// Ensure the TLS Certificate and Key files exist
		if tlsCertFile == "" {
			lg.Fatal("ECOM_APP_TLS_MODE is enabled so you must set the cert file. Use export ECOM_APP_TLS_CERT=/path/to/your/cert.pem")
		}

		// if the tlsCertFile is a relative pathname, make it relative to the secretVolume root
		if !filepath.IsAbs(tlsCertFile) {
			lg.Debugf("tlsCertFile is a relative pathname so building absolute pathname")
			tlsCertFile = filepath.Join(secretVolume, tlsCertFile)
		}
		mustHaveFile(tlsCertFile, "TLS Cert File")

		if tlsKeyFile == "" {
			lg.Fatal("ECOM_APP_TLS_MODE is enabled so you must set the key file. Use export ECOM_APP_TLS_KEY=/path/to/your/key.pem")
		}
		if !filepath.IsAbs(tlsKeyFile) {
			lg.Debugf("tlsKeyFile is a relative pathname so building absolute pathname")
			tlsKeyFile = filepath.Join(secretVolume, tlsKeyFile)
		}
		mustHaveFile(tlsKeyFile, "TLS Key File")
	}

	// 5. Root Credentials
	if rootEmail == "" {
		lg.Fatal("app root email not set. Use ECOM_APP_ROOT_EMAIL")
	}
	if rootPassword == "" {
		lg.Fatal("app root password not set. Use ECOM_APP_ROOT_PASSWORD")
	}

	// connect to postgres
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		lg.Fatalf("failed to open db: %v", err)
	}

	attempt := 0
	for attempt < maxDbConnectAttempts {
		err = db.Ping()
		if err != nil {
			attempt++
			if attempt >= maxDbConnectAttempts {
				lg.Fatalf("attempt %d/%d failed to verify db connection: %v", attempt, maxDbConnectAttempts, err)
			}
			lg.Warnf("attempt %d/%d, failed to verify db connection: %v", attempt, maxDbConnectAttempts, err)
			time.Sleep(5 * time.Second)
		} else {
			break
		}
	}

	lg.Infoln("established database connection")

	// build a Postgres model
	pgModel := model.NewPgModel(db)

	// build a Google Firebase App
	var fbApp *firebase.App

	ctx := context.Background()
	var opt option.ClientOption
	if credentials[0] == '/' {
		opt = option.WithCredentialsFile(credentials)
	} else {
		decoded, err := base64.StdEncoding.DecodeString(credentials)
		if err != nil {
			lg.Fatalf("decode error: %v", err)
		}
		opt = option.WithCredentialsJSON(decoded)
	}
	fbApp, err = firebase.NewApp(ctx, nil, opt)
	if err != nil {
		lg.Fatalf("%v", fmt.Errorf("failed to initialise Firebase app: %v", err))
	}

	// build a Firebase service injecting in the model and firebase app as dependencies
	fbSrv := service.NewService(pgModel, fbApp)

	// ensure the root user has been created
	err = fbSrv.CreateRootIfNotExists(ctx, rootEmail, rootPassword)
	if err != nil {
		lg.Fatalf("failed to create root credentials if not exists: %v", err)
	}

	// SystemInfo
	si := app.SystemInfo{
		Version: version,
		Env: app.SystemEnv{
			PG: app.PgSystemEnv{
				PgHost:     pghost,
				PgPort:     pgport,
				PgDatabase: pgdatabase,
				PgUser:     pguser,
				PgSSLMode:  pgsslmode,
			},
			Goog: app.GoogSystemEnv{
				GoogProjectID: projectID,
				WebAPIKey:     webAPIKey,
			},
			App: app.ApplSystemEnv{
				AppPort:      port,
				AppRootEmail: rootEmail,
			},
		},
	}

	a := app.App{
		Service: fbSrv,
	}
	r := chi.NewRouter()

	// protected routes
	r.Group(func(r chi.Router) {
		c := cors.New(cors.Options{
			AllowedOrigins:   []string{"*"},
			AllowedHeaders:   []string{"Authorization", "Content-Type", "Accept"},
			AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
			AllowCredentials: true,
			// Enable Debugging for testing, consider disabling in production
			Debug: false,
		})
		r.Use(c.Handler)
		r.Use(a.AuthenticateMiddleware)

		r.Route("/admins", func(r chi.Router) {
			r.Post("/", a.Authorization(app.OpCreateAdmin, a.CreateAdminHandler()))
			r.Get("/", a.Authorization(app.OpListAdmins, a.ListAdminsHandler()))
			r.Delete("/{uuid}", a.Authorization(app.OpDeleteAdmin, a.DeleteAdminHandler()))
		})

		// Customer and address management API
		r.Route("/customers", func(r chi.Router) {
			r.Post("/", a.Authorization(app.OpCreateCustomer, a.CreateCustomerHandler()))
			r.Get("/{uuid}", a.Authorization(app.OpGetCustomer, a.GetCustomerHandler()))
			r.Get("/", a.Authorization(app.OpListCustomers, a.ListCustomersHandler()))

			r.Get("/{uuid}/devkeys", a.Authorization(app.OpListCustomersDevKeys, a.ListCustomersDevKeysHandler()))
			r.Post("/{uuid}/devkeys", a.Authorization(app.OpGenerateCustomerDevKey, a.GenerateCustomerDevKeyHandler()))
			r.Post("/{cuuid}/addresses", a.Authorization(app.OpCreateAddress, a.CreateAddressHandler()))
			r.Get("/{cuuid}/addresses", a.Authorization(app.OpGetCustomersAddresses, a.ListAddressesHandler()))
			r.Patch("/{cuuid}/addresses/{auuid}", a.Authorization(app.OpUpdateAddress, a.UpdateAddressHandler()))
		})

		// tiers resource operation all return 501 Not Implemented
		r.Route("/tiers", func(r chi.Router) {
			r.Post("/", a.Authorization(app.OpCreateTier, a.NotImplementedHandler()))
			r.Get("/{ref}", a.Authorization(app.OpGetTier, a.NotImplementedHandler()))
			r.Get("/", a.Authorization(app.OpListTiers, a.NotImplementedHandler()))
			r.Put("/{ref}", a.Authorization(app.OpUpdateTier, a.NotImplementedHandler()))
			r.Delete("/{ref}", a.Authorization(app.OpDeleteTier, a.NotImplementedHandler()))
		})

		r.Route("/products/{sku}/images", func(r chi.Router) {
			r.Post("/", a.Authorization(app.OpAddImage, a.AddImageHandler()))
			r.Get("/", a.Authorization(app.OpListProductImages, a.ListProductImagesHandler()))
			r.Delete("/", a.Authorization(app.OpDeleteAllProductImages, a.DeleteAllProductImagesHandler()))
		})

		r.Route("/images", func(r chi.Router) {
			r.Get("/{uuid}", a.Authorization(app.OpGetImage, a.GetImageHandler()))
			r.Delete("/{uuid}", a.Authorization(app.OpDeleteImage, a.DeleteImageHandler()))
		})

		r.Route("/products/{sku}/tiers/{ref}/pricing", func(r chi.Router) {
			r.Put("/", a.Authorization(app.OpUpdateTierPricing, a.UpdateTierPricingHandler()))
			r.Get("/", a.Authorization(app.OpGetTierPricing, a.GetTierPricingHandler()))
			r.Delete("/", a.Authorization(app.OpDeleteTierPricing, a.DeleteTierPricingHandler()))
		})

		r.Route("/products", func(r chi.Router) {
			r.Post("/", a.Authorization(app.OpCreateProduct, a.CreateProductHandler()))
			r.Get("/", a.Authorization(app.OpListProducts, a.ListProductsHandler()))
			r.Get("/{sku}", a.Authorization(app.OpGetProduct, a.GetProductHandler()))
			r.Head("/{sku}", a.Authorization(app.OpProductExists, a.ProductExistsHandler()))
			r.Put("/{sku}", a.Authorization(app.OpUpdateProduct, a.UpdateProductHandler()))
			r.Delete("/{sku}", a.Authorization(app.OpDeleteProduct, a.DeleteProductHandler()))

			r.Get("/{sku}/pricing", a.Authorization(app.OpListPricingBySKU, a.ListPricingBySKUHandler()))
			r.Get("/tiers/{ref}/pricing", a.Authorization(app.OpListPricingByTier, a.ListPricingByTierHandler()))
		})

		r.Route("/devkeys", func(r chi.Router) {
			r.Delete("/{uuid}", a.Authorization(app.OpDeleteCustomerDevKey, a.DeleteCustomerDevKeyHandler()))
		})

		r.Route("/addresses", func(r chi.Router) {
			r.Get("/{uuid}", a.Authorization(app.OpGetAddress, a.GetAddressHandler()))
			r.Delete("/{uuid}", a.Authorization(app.OpDeleteAddress, a.DeleteAddressHandler()))
		})

		r.Route("/carts", func(r chi.Router) {
			r.Post("/", a.Authorization(app.OpCreateCart, a.CreateCartHandler()))
			r.Post("/{uuid}/items", a.Authorization(app.OpAddItemToCart, a.AddItemToCartHandler()))
			r.Get("/{uuid}/items", a.Authorization(app.OpGetCartItems, a.GetCartItemsHandler()))
			r.Patch("/{uuid}/items/{sku}", a.Authorization(app.OpUpdateCartItem, a.UpdateCartItemHandler()))
			r.Delete("/{uuid}/items/{sku}", a.Authorization(app.OpDeleteCartItem, a.DeleteCartItemHandler()))
			r.Delete("/{uuid}/items", a.Authorization(app.OpEmptyCartItems, a.EmptyCartItemsHandler()))
		})

		r.Route("/catalog", func(r chi.Router) {
			r.Put("/", a.Authorization(app.OpUpdateCatalog, a.UpdateCatalogHandler()))
			r.Get("/", a.Authorization(app.OpGetCatalog, a.GetCatalogHandler()))
			r.Delete("/", a.Authorization(app.OpPurgeCatalog, a.PurgeCatalogHandler()))
		})

		r.Route("/assocs", func(r chi.Router) {
			r.Put("/", a.Authorization(app.OpUpdateCatalogAssocs, a.UpdateCatalogProductAssocsHandler()))
			r.Get("/", a.Authorization(app.OpGetCatalogAssocs, a.GetCatalogAssocsHandler()))
			r.Delete("/", a.Authorization(app.OpPurgeCatalogAssocs, a.PurgeCatalogAssocsHandler()))
		})

		r.Route("/sysinfo", func(r chi.Router) {
			r.Get("/", a.Authorization(app.OpSystemInfo, a.SystemInfoHandler(si)))
		})
	})

	// public routes including GET / for Google Kuberenetes default healthcheck
	r.Group(func(r chi.Router) {
		r.Use(middleware.NoCache)

		// version info
		r.Get("/", healthCheckHandler)
		r.Get("/healthz", healthCheckHandler)
		r.Get("/config", a.ConfigHandler(si.Env.Goog))
	})

	r.Route("/signin-with-devkey", func(r chi.Router) {
		r.Post("/", a.SignInWithDevKeyHandler())
	})

	// tlsMode determines whether to serve HTTPS traffic directly.
	// If tlsMode is false, you can enable HTTPS with a GKE Layer 7 load balancer
	// using an Ingress.
	if tlsMode {
		lg.Infof("server listening on HTTPS port %s", port)
		lg.Fatal(http.ListenAndServeTLS(fmt.Sprintf(":%s", port), tlsCertFile, tlsKeyFile, r))
	} else {
		lg.Infof("server listening on HTTP port %s", port)
		lg.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), r))
	}
}
