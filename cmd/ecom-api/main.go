package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"bitbucket.org/andyfusniakteam/ecom-api-go/app"
	model "bitbucket.org/andyfusniakteam/ecom-api-go/model/postgres"
	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"cloud.google.com/go/profiler"
	"cloud.google.com/go/pubsub"
	firebase "firebase.google.com/go"
	_ "firebase.google.com/go/auth"
	stackdriver "github.com/andyfusniak/stackdriver-gae-logrus-plugin"
	stackdm "github.com/andyfusniak/stackdriver-gae-logrus-plugin/middleware"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
	stripe "github.com/stripe/stripe-go"

	log "github.com/sirupsen/logrus"

	"golang.org/x/sys/unix"
	"google.golang.org/api/option"
)

var version = "v0.64.0"

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
	gaeProjectIDEnv   = os.Getenv("ECOM_GAE_PROJECT_ID")
	fbPublicConfigEnv = os.Getenv("ECOM_FIREBASE_PUBLIC_CONFIG")
	fbCredentialsEnv  = os.Getenv("ECOM_FIREBASE_PRIVATE_CREDENTIALS")
	pubSubPushToken   = os.Getenv("ECOM_GOOGLE_PUBSUB_PUSH_TOKEN")

	// Stripe settings (optional)
	stripeSecretKey     = os.Getenv("ECOM_STRIPE_SECRET_KEY")
	stripeSigningSecret = os.Getenv("ECOM_STRIPE_SIGNING_SECRET")
	stripeSuccessURL    = os.Getenv("ECOM_STRIPE_SUCCESS_URL")
	stripeCancelURL     = os.Getenv("ECOM_STRIPE_CANCEL_URL")

	//
	// Application settings
	//
	port                        = os.Getenv("PORT")
	tlsModeFlag                 = os.Getenv("ECOM_APP_TLS_MODE")
	tlsCertFile                 = os.Getenv("ECOM_APP_TLS_CERT")
	tlsKeyFile                  = os.Getenv("ECOM_APP_TLS_KEY")
	rootEmail                   = os.Getenv("ECOM_APP_ROOT_EMAIL")
	rootPassword                = os.Getenv("ECOM_APP_ROOT_PASSWORD")
	maxOpenConnsEnv             = os.Getenv("ECOM_APP_MAX_OPEN_CONNS")
	maxIdleConnsEnv             = os.Getenv("ECOM_APP_MAX_IDLE_CONNS")
	connMaxLifetimeEnv          = os.Getenv("ECOM_APP_CONN_MAX_LIFETIME")
	enableStackDriverLoggingEnv = os.Getenv("ECOM_APP_ENABLE_STACKDRIVER_LOGGING")
	appEndpoint                 = os.Getenv("ECOM_APP_ENDPOINT")
)

var enableStackDriverLogging bool

const (
	// secretVolume points to the base path for the mounted drive or secret files using k8s or docker mount
	secretVolume = "/etc/secret-volume"

	// directory in the secret volume that holds all Service Account Credentials files
	sacDir = "service_account_credentials"

	// events-topic handles the initial event before publishing multiple messages
	// (one per registered webhook listening for those events) to broadcast-topic
	pubSubEventsTopic          = "ecom-api-events-topic"
	pubSubEventsSubscription   = "ecom-api-events-subscription"
	pubSubWebhooksTopic        = "ecom-api-broadcast-topic"
	pubSubWebhooksSubscription = "ecom-api-broadcast-subscription"
)

func initLogging() {
	if enableStackDriverLoggingEnv == "yes" ||
		enableStackDriverLoggingEnv == "on" ||
		enableStackDriverLoggingEnv == "True" {
		enableStackDriverLogging = true
	}

	if enableStackDriverLogging {
		// Log as JSON Stackdriver with entry threading
		// instead of the default ASCII formatter.
		formatter := stackdriver.GAEStandardFormatter(
			stackdriver.WithProjectID(gaeProjectIDEnv),
		)
		log.SetFormatter(formatter)

		// Profiler initialization, best done as early as possible.
		if err := profiler.Start(profiler.Config{
			// Service and ServiceVersion can be automatically inferred when running
			// on App Engine.
			// ProjectID must be set if not running on GCP.
		}); err != nil {
			log.Fatalf("main: failed to init profiler: %v", err)
		}
		log.Info("main: profiler initialised")
	} else {
		// Output logs with colour
		log.SetFormatter(&log.TextFormatter{
			ForceColors: true,
		})
	}

	// Output to stdout instead of the default stderr
	log.SetOutput(os.Stdout)

	// Log debug level severity or above.
	log.SetLevel(log.DebugLevel)
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func testDelayHandler(w http.ResponseWriter, r *http.Request) {
	time.Sleep(time.Second * 30)
	w.WriteHeader(http.StatusNoContent)
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
		log.Fatalf("main: failed to determine if %s %s exists: %v", title, path, err)
	}
	if !ex {
		log.Fatalf("main: cannot find %s %s. Check permissions.", title, path)
	}
	log.Infof("main: %s: %s", title, path)
}

func main() {
	// wave goodbye on the way out the door
	defer func() {
		log.Infof("main: goodbye from ecom-api version %s", version)
	}()

	initLogging()

	log.Infof("main: hello from ecom-api version %s", version)
	log.Infof("main: built with %s for %s %s", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	log.Infof("main: running process id %d", os.Getpid())
	// 1. Data Source Name
	// dsn is the Data Source name. For PostgreSQL the format is "host=localhost port=5432 user=postgres password=secret dbname=mydatabase sslmode=disable". The sslmode is optional.
	if pghost == "" {
		log.Fatal("main: postgres host not set. Use ECOM_PG_HOST")
	}
	log.Infof("main: ECOM_PG_HOST set to %s", pghost)

	if pgport == "" {
		log.Info("main: using default port=5432 for postgres because ECOM_PG_PORT is not set")
		pgport = "5432"
	}
	log.Infof("main: ECOM_PG_PORT set to %s", pgport)

	if pguser == "" {
		log.Info("main: using default user=postgres because ECOM_PG_USER is not set")
		pguser = "postgres"
	}
	log.Infof("main: ECOM_PG_USER set to %s", pguser)

	if pgdatabase == "" {
		log.Fatal("main: ECOM_PG_DATABASE not set.")
	}
	log.Infof("main: ECOM_PG_DATABASE set to %s", pgdatabase)

	if pgpassword == "" {
		log.Fatal("main: ECOM_PG_PASSWORD not set. You must set a password")
	}

	if pgsslmode == "" {
		if pgsslkey != "" || pgsslrootcert != "" || pgsslcert != "" {
			log.Fatal("main: ECOM_PG_SSLMODE is not set, but one or more of ECOM_PG_SSLCERT, ECOM_PG_SSLKEY, ECOM_PG_SSLROOTCERT environment variables were set implying you intended to connect to postgres securely?")
		}
		log.Infof("main: using postgres sslmode=disable because ECOM_PG_SSLMODE is not set")
		pgsslmode = "disable"
	}

	if pgconnectTimeout == "" {
		log.Infof("main: using postgres connect_timeout=10 because ECOM_PG_CONNECT_TIMEOUT is not set")
		pgconnectTimeout = "10"
	}

	var dsn string
	if pgsslmode == "disable" {
		log.Infof("main: postgres running with sslmode=disable")
		dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s connect_timeout=%s", pghost, pgport, pguser, pgpassword, pgdatabase, pgsslmode, pgconnectTimeout)
		log.Infof("main: postgres dsn: host=%s port=%s user=%s password=**** dbname=%s sslmode=%s connect_timeout=%s", pghost, pgport, pguser, pgdatabase, pgsslmode, pgconnectTimeout)
	} else {
		// Ensure that the ECOM_PG_SSLCERT, ECOM_PG_SSLROOTCERT and ECOM_PG_SSLKEY are all
		// referenced using absolute paths.
		if pgsslcert == "" {
			log.Fatal("main: missing PostgreSQL SSL certificate file. Use export ECOM_PG_SSLCERT")
		}
		if !filepath.IsAbs(pgsslcert) {
			log.Fatalf("main: ECOM_PG_SSLCERT should use an absolute path to certificate file to avoid ambiguity")
		}
		mustHaveFile(pgsslcert, "client certificate file")

		if pgsslrootcert == "" {
			log.Fatal("main: missing PostgreSQL SSL root certificate file. Use export ECOM_PG_SSLROOTCERT")
		}
		if !filepath.IsAbs(pgsslrootcert) {
			log.Fatalf("main: ECOM_PG_SSLROOTCERT should use an absolute path to root certificate file to avoid ambiguity")
		}
		mustHaveFile(pgsslrootcert, "ssl root certificate")

		if pgsslkey == "" {
			log.Fatal("main: missing PostgreSQL SSL key certificate file. Use export ECOM_PG_SSLKEY")
		}
		if !filepath.IsAbs(pgsslkey) {
			log.Fatalf("main: ECOM_PG_SSLKEY should use an absolute path to key certificate file to avoid ambiguity")
		}
		mustHaveFile(pgsslkey, "ssl key file")

		dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s sslcert=%s sslrootcert=%s sslkey=%s connect_timeout=%s", pghost, pgport, pguser, pgpassword, pgdatabase, pgsslmode, pgsslcert, pgsslrootcert, pgsslkey, pgconnectTimeout)
		log.Infof("main: postgres dsn: host=%s port=%s user=%s password=***** dbname=%s sslmode=%s sslcert=%s sslrootcert=%s sslkey=%s connect_timeout=%s", pghost, pgport, pguser, pgdatabase, pgsslmode, pgsslcert, pgsslrootcert, pgsslkey, pgconnectTimeout)
	}

	// 2. Service Account Credentials
	if fbCredentialsEnv == "" {
		log.Fatal("main: missing service account credentials. Use export ECOM_FIREBASE_PRIVATE_CREDENTIALS=/path/to/your/service-account-file or ECOM_FIREBASE_PRIVATE_CREDENTIALS=<base64-json-file>")
	}
	// if the credentials is a relative pathname, make it relative to the secretVolume/sacDir root
	// i.e. /etc/secret-volume/service_account_credentials/<file>
	if fbCredentialsEnv[0] == '/' {
		if !filepath.IsAbs(fbCredentialsEnv) {
			log.Debugf("main: credentials is a relative pathname so building absolute pathname")
			fbCredentialsEnv = filepath.Join(secretVolume, sacDir, fbCredentialsEnv)
		}
		mustHaveFile(fbCredentialsEnv, "service account credentials")
	}

	// 3. GAE ProjectID, Google Project ID (Firebase) and Web API Key (Firebase)
	if gaeProjectIDEnv == "" {
		log.Fatal("main: missing GAE project ID. Use export ECOM_GAE_PROJECT_ID")
	}

	if fbPublicConfigEnv == "" {
		log.Fatal("main: missing public Firebase Config.  Use ECOM_FIREBASE_PUBLIC_CONFIG=<base64-json-string>")
	}
	// log.Infof("Web API Key set to %s", fbWebAPIKey)
	decoded, err := base64.StdEncoding.DecodeString(fbPublicConfigEnv)
	if err != nil {
		log.Fatalf("main: decode error: %v", err)
	}
	type firebaseConfig struct {
		Firebase app.FirebaseSystemEnv `json:"firebaseConfig"`
	}

	var fbConfig firebaseConfig
	if err := json.NewDecoder(bytes.NewReader(decoded)).Decode(&fbConfig); err != nil {
		log.Fatalf("main: %v", err)
	}

	if fbConfig.Firebase.APIKEY == "" {
		log.Fatal("main: Firebase Config not loaded. Check base64 encoded ECOM_FIREBASE_PUBLIC_CONFIG.")
	}

	log.Infof("main: firebase apiKey set to %s", fbConfig.Firebase.APIKEY)
	log.Infof("main: firebase projectID set to %s", fbConfig.Firebase.ProjectID)

	// ECOM_GOOGLE_PUBSUB_PUSH_TOKEN must be set.
	if pubSubPushToken == "" {
		log.Fatalf("main: missing google pubsub events push token. Use ECOM_GOOGLE_PUBSUB_PUSH_TOKEN to set this to a secret string token")
	}

	// ECOM_APP_ENDPOINT must be set to an absolute, secure URL endpoint.
	// It is used to derive the events and webhook cloud pubsub URL endpoints.
	if appEndpoint == "" {
		log.Fatalf("main: missing app endpoint. Use ECOM_APP_ENDPOINT to set this to an absolute secure URL")
	}
	u, err := url.Parse(appEndpoint)
	if err != nil {
		log.Fatalf("main: failed to url.Parse(%q)", appEndpoint)
	}
	if u.Scheme != "https" {
		log.Fatalf("main: ECOM_APP_ENDPOINT must be set to a secure URL - got %s", appEndpoint)
	}
	if !u.IsAbs() {
		log.Fatalf("main: ECOM_APP_ENDPOINT must be set to an absolute URL - got %s", appEndpoint)
	}
	log.Infof("main: ECOM_APP_ENDPOINT set to %s", appEndpoint)

	pubSubEventsPushEndpoint, err := url.Parse(appEndpoint + "/private-pubsub-events")
	if err != nil {
		log.Fatalf("main: url.Parse(rawurl=%q) failed: %+v", appEndpoint, err)
	}
	q := pubSubEventsPushEndpoint.Query()
	q.Set("token", pubSubPushToken)
	pubSubEventsPushEndpoint.RawQuery = q.Encode()
	log.Infof("main: using %s/private-pubsub-events?token=***** as the events push endpoint", appEndpoint)

	pubSubBroadcastPushEndpoint, err := url.Parse(appEndpoint + "/private-pubsub-broadcast")
	if err != nil {
		log.Fatalf("main: url.Parse(rawurl=%q) failed: %+v", appEndpoint, err)
	}
	q = pubSubBroadcastPushEndpoint.Query()
	q.Set("token", pubSubPushToken)
	pubSubBroadcastPushEndpoint.RawQuery = q.Encode()
	log.Infof("main: using %s/private-pubsub-broadcast?token=***** as the pubsub broadcast push endpoint", appEndpoint)

	// 4. Stripe Secret Key and Signing Key.
	if stripeSecretKey == "" && stripeSigningSecret == "" {
		log.Warn("main: ECOM_STRIPE_SECRET_KEY and ECOM_STRIPE_SIGNING_SECRET are not set. This service will not process Stripe payments")
	} else {
		if stripeSecretKey == "" {
			log.Fatal("main: ECOM_STRIPE_SIGNING_SECRET must be set since ECOM_STRIPE_SECRET_KEY has been set")
		}

		if stripeSigningSecret == "" {
			log.Fatal("main: ECOM_STRIPE_SECRET_KEY must be set since ECOM_STRIPE_SIGNING_SECRET has been set")
		}

		if stripeSecretKey != "" {
			stripe.Key = stripeSecretKey
			log.Info("main: Stripe Secret Key set in Stripe library")
		}
	}

	if stripeSuccessURL == "" {
		stripeSuccessURL = "https://example.com/success"
		log.Info("main: ECOM_STRIPE_SUCCESS_URL is not set, so using default value")
	}
	log.Infof("main: ECOM_STRIPE_SUCCESS_URL set to %s", stripeSuccessURL)

	if stripeCancelURL == "" {
		stripeCancelURL = "https://example.com/cancel"
		log.Info("main: ECOM_STRIPE_CANCEL_URL is not set, so using default value")
	}
	log.Infof("main: ECOM_STRIPE_CANCEL_URL set to %s", stripeCancelURL)

	// 5. Server Port
	if port == "" {
		port = "8080"
		log.Infof("main: HTTP Port not specified. Using default port %s", port)
	} else {
		log.Infof("main: environment variable PORT specifies port %s to be used", port)
	}

	// ensure that we have access to the secret volume
	// Google Kubernetes Engine attaches secrets to /etc/secret-volume mount points
	// Google Compute Engine attaches a persistent disk containing the necessary assets
	// Assets include PostgreSQL .pem files, Google Firebase service account keys and
	// TLS/SSL certificate files for HTTPS termination (see ECOM_APP_TLS_MODE=enable).
	if fbCredentialsEnv[0] == '/' {
		ex, err := exists(secretVolume)
		if err != nil {
			log.Fatalf("main: failed to determine if secret volume %s exists: %v", secretVolume, err)
		}
		if !ex {
			log.Fatalf("main: cannot find secret volume %s. Have you mounted it?", secretVolume)
		}
		log.Infof("main: found secret volume %s", secretVolume)
	}

	// TLS Mode defaults to false unless ECOM_APP_TLS_MODE is set to enable
	// tlsMode will be used to determine whether to provide negociation for SSL
	// (see the bottom of main func)
	tlsMode := false
	if tlsModeFlag == "enable" || tlsModeFlag == "enabled" {
		tlsMode = true
		log.Info("main: ECOM_APP_TLS_MODE enabled")

		// Ensure the TLS Certificate and Key files exist
		if tlsCertFile == "" {
			log.Fatal("main: ECOM_APP_TLS_MODE is enabled so you must set the cert file. Use export ECOM_APP_TLS_CERT=/path/to/your/cert.pem")
		}

		// if the tlsCertFile is a relative pathname, make it relative to the secretVolume root
		if !filepath.IsAbs(tlsCertFile) {
			log.Debugf("main: tlsCertFile is a relative pathname so building absolute pathname")
			tlsCertFile = filepath.Join(secretVolume, tlsCertFile)
		}
		mustHaveFile(tlsCertFile, "TLS Cert File")

		if tlsKeyFile == "" {
			log.Fatal("main: ECOM_APP_TLS_MODE is enabled so you must set the key file. Use export ECOM_APP_TLS_KEY=/path/to/your/key.pem")
		}
		if !filepath.IsAbs(tlsKeyFile) {
			log.Debugf("main: tlsKeyFile is a relative pathname so building absolute pathname")
			tlsKeyFile = filepath.Join(secretVolume, tlsKeyFile)
		}
		mustHaveFile(tlsKeyFile, "TLS Key File")
	}

	// 5. Root Credentials
	if rootEmail == "" {
		log.Fatal("main: app root email not set. Use ECOM_APP_ROOT_EMAIL")
	}
	if rootPassword == "" {
		log.Fatal("main: app root password not set. Use ECOM_APP_ROOT_PASSWORD")
	}

	// 6. Connection pooling
	var maxOpenConns int
	if maxOpenConnsEnv != "" {
		var err error
		maxOpenConns, err = strconv.Atoi(maxOpenConnsEnv)
		if err != nil {
			log.Fatal("main: app failed to read value in ECOM_APP_MAX_OPEN_CONNS")
		}
	} else {
		log.Info("main: ECOM_APP_MAX_OPEN_CONNS is not set. Using the default of unlimited")
	}

	var maxIdleConns int
	if maxIdleConnsEnv != "" {
		var err error
		maxIdleConns, err = strconv.Atoi(maxIdleConnsEnv)
		if err != nil {
			log.Fatal("main: app failed to read value in ECOM_APP_MAX_IDLE_CONNS")
		}
		// There is no point in ever having any more idle connections than the
		// maximum allowed open connections, because if you could instantaneously
		// grab all the allowed open connections, the remain idle connections
		// would always remain idle. It's like having a bridge with four lanes,
		// but only ever allowing three vehicles to drive across it at once.
		// https://stackoverflow.com/questions/31952791/setmaxopenconns-and-setmaxidleconns/31952911#31952911
		if maxIdleConns > maxOpenConns {
			log.Fatal("main: app maxIdleConns exceeds maxOpenConns. Check both ECOM_APP_MAX_OPEN_CONNS and ECOM_APP_MAX_IDLE_CONNS")
		}
	}

	var connMaxLifetime int
	if connMaxLifetimeEnv != "" {
		var err error
		connMaxLifetime, err = strconv.Atoi(connMaxLifetimeEnv)
		if err != nil {
			log.Fatal("main: app failed to read value in ECOM_APP_CONN_MAX_LIFETIME")
		}
	}

	// connect to postgres
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("main: failed to open db: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Warn("main: failed to close the database")
		} else {
			log.Info("main: database closed")
		}
	}()

	if maxOpenConns > 0 {
		db.SetMaxOpenConns(maxOpenConns)
		log.Infof("main: max open connections set to %d", maxOpenConns)
	}
	if maxIdleConns > 0 {
		db.SetMaxIdleConns(maxIdleConns)
		log.Infof("main: max idle connections set to %d", maxIdleConns)
	}
	if connMaxLifetime > 0 {
		db.SetConnMaxLifetime(time.Minute * time.Duration(connMaxLifetime))
		log.Infof("main: max conn max lifetime set to %d minutes", connMaxLifetime)
	}

	attempt := 0
	for attempt < maxDbConnectAttempts {
		err = db.Ping()
		if err != nil {
			attempt++
			if attempt >= maxDbConnectAttempts {
				log.Fatalf("main: attempt %d/%d failed to verify db connection: %v", attempt, maxDbConnectAttempts, err)
			}
			log.Warnf("main: attempt %d/%d, failed to verify db connection: %v", attempt, maxDbConnectAttempts, err)
			time.Sleep(5 * time.Second)
		} else {
			break
		}
	}
	log.Infoln("main: established a database connection")

	// build a Postgres model
	pgModel := model.NewPgModel(db)

	// build a Google Firebase App
	var fbApp *firebase.App

	var opt option.ClientOption
	if fbCredentialsEnv[0] == '/' {
		opt = option.WithCredentialsFile(fbCredentialsEnv)
	} else {
		decoded, err := base64.StdEncoding.DecodeString(fbCredentialsEnv)
		if err != nil {
			log.Fatalf("main:decode error: %v", err)
		}
		opt = option.WithCredentialsJSON(decoded)
	}
	ctx := context.Background()
	fbApp, err = firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalf("main: failed to initialise Firebase app: %v", err)
	}

	// Google Pub Sub
	pubSubClient, err := pubsub.NewClient(ctx, gaeProjectIDEnv, opt)
	if err != nil {
		log.Fatalf("main: pubsub.NewClient(ctx, gaeProjectIDEnv) failed: %+v", err)
	}
	log.Infof("main: initializing Google PubSub client for project %s",
		gaeProjectIDEnv)

	// var eventsTopic *pubsub.Topic
	// var webhooksTopic *pubsub.Topic

	// Create the topics and push subscriptions if they don't exist.
	eventsTopic, err := service.CreateTopicAndSubscription(ctx, pubSubClient,
		pubSubEventsTopic, pubSubEventsSubscription, pubSubEventsPushEndpoint.String())
	if err != nil {
		log.Fatalf("main: failed to create cloud pubsub topic and subscription for events topic: %+v", err)
	}

	whBroadcastTopic, err := service.CreateTopicAndSubscription(ctx, pubSubClient,
		pubSubWebhooksTopic, pubSubWebhooksSubscription, pubSubBroadcastPushEndpoint.String())
	if err != nil {
		log.Fatalf("main: failed to create cloud pubsub topic and subscription for broadcast topic: %+v", err)
	}

	// build a Firebase service injecting in the model and firebase app as dependencies
	fbSrv := service.NewService(pgModel, fbApp, eventsTopic, whBroadcastTopic)

	// ensure the root user has been created
	err = fbSrv.CreateRootIfNotExists(ctx, rootEmail, rootPassword)
	if err != nil {
		log.Fatalf("main: failed to create root credentials if not exists: %+v", err)
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
				GAEProjectID: gaeProjectIDEnv,
			},
			Firebase: fbConfig.Firebase,
			Stripe: app.StripeSystemEnv{
				StripeSuccessURL: stripeSuccessURL,
				StripeCancelURL:  stripeCancelURL,
			},
			App: app.ApplSystemEnv{
				AppPort:                     port,
				AppRootEmail:                rootEmail,
				AppEnableStackDriverLogging: enableStackDriverLogging,
				AppEndpoint:                 appEndpoint,
			},
		},
	}

	a := app.App{
		Service: fbSrv,
	}
	r := chi.NewRouter()
	r.Use(stackdm.XCloudTraceContext)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedHeaders:   []string{"Authorization", "Content-Type", "Accept", "X-Keyed-By"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowCredentials: true,
		// Enable Debugging for testing, consider disabling in production
		Debug: false,
	})
	r.Use(c.Handler)

	// protected routes
	r.Group(func(r chi.Router) {
		r.Use(a.AuthenticateMiddleware)
		r.Use(stackdm.XCloudTraceContext)

		// Users
		r.Route("/users", func(r chi.Router) {
			r.Post("/", a.Authorization(app.OpCreateUser, a.CreateUserHandler()))
			r.Get("/{id}", a.Authorization(app.OpGetUser, a.GetUserHandler()))
			r.Get("/", a.Authorization(app.OpListUsers, a.ListUsersHandler()))
			r.Delete("/{id}", a.Authorization(app.OpDeleteUser, a.DeleteUserHandler()))
		})

		// Addresses
		r.Route("/addresses", func(r chi.Router) {
			r.Post("/", a.Authorization(app.OpCreateAddress, a.CreateAddressHandler()))
			r.Get("/{id}", a.Authorization(app.OpGetAddress, a.GetAddressHandler()))
			r.Patch("/{id}", a.Authorization(app.OpUpdateAddress, a.UpdateAddressHandler()))
			r.Delete("/{id}", a.Authorization(app.OpDeleteAddress, a.DeleteAddressHandler()))
			r.Get("/", a.Authorization(app.OpGetUsersAddresses, a.ListAddressesHandler()))
		})

		// Developer Keys
		r.Route("/developer-keys", func(r chi.Router) {
			r.Post("/", a.Authorization(app.OpGenerateUserDevKey, a.GenerateUserDevKeyHandler()))
			r.Get("/", a.Authorization(app.OpListUsersDevKeys, a.ListUsersDevKeysHandler()))
			r.Delete("/{id}", a.Authorization(app.OpDeleteUserDevKey, a.DeleteUserDevKeyHandler()))
		})

		// price list resource operation all return 501 Not Implemented
		// apart from OpListPricingTiers
		r.Route("/price-lists", func(r chi.Router) {
			r.Post("/", a.Authorization(app.OpCreatePriceList, a.CreatePriceListHandler()))
			r.Get("/{id}", a.Authorization(app.OpGetPriceList, a.GetPriceListHandler()))
			r.Get("/", a.Authorization(app.OpListPriceLists, a.ListPriceListsHandler()))
			r.Put("/{id}", a.Authorization(app.OpUpdatePriceList, a.UpdatePriceListHandler()))
			r.Delete("/{id}", a.Authorization(app.OpDeletePriceList, a.DeletePriceListHandler()))
		})

		// Inventory
		r.Route("/inventory", func(r chi.Router) {
			r.Get("/", a.Authorization(app.OpListInventory, a.ListInventoryHandler()))
			r.Get("/{id}", a.Authorization(app.OpGetInventory, a.GetInventoryHandler()))
			r.Patch("/{id}", a.Authorization(app.OpUpdateInventory, a.UpdateInventoryHandler()))
		})

		r.Route("/inventory:batch-update", func(r chi.Router) {
			r.Patch("/", a.Authorization(app.OpBatchUpdateInventory, a.BatchUpdateInventoryHandler()))
		})

		// Promo Rules
		r.Route("/promo-rules", func(r chi.Router) {
			r.Post("/", a.Authorization(app.OpCreatePromoRule, a.CreatePromoRuleHandler()))
			r.Get("/{id}", a.Authorization(app.OpGetPromoRule, a.GetPromoRuleHandler()))
			r.Get("/", a.Authorization(app.OpListPromoRules, a.ListPromoRulesHandler()))
			r.Delete("/{id}", a.Authorization(app.OpDeletePromoRule, a.DeletePromoRuleHandler()))
		})

		// Offers
		r.Route("/offers", func(r chi.Router) {
			r.Post("/", a.Authorization(app.OpActivateOffer, a.ActivateOfferHandler()))
			r.Get("/{id}", a.Authorization(app.OpGetOffer, a.GetOfferHandler()))
			r.Get("/", a.Authorization(app.OpListOffers, a.ListOffersHandler()))
			r.Delete("/{id}", a.Authorization(app.OpDeactivateOffer, a.DeactivateOfferHandler()))
		})

		// Coupons
		r.Route("/coupons", func(r chi.Router) {
			r.Post("/", a.Authorization(app.OpCreateCoupon, a.CreateCouponHandler()))
			r.Get("/{id}", a.Authorization(app.OpGetCoupon, a.GetCouponHandler()))
			r.Get("/", a.Authorization(app.OpListCoupons, a.ListCouponsHandler()))
			r.Patch("/{id}", a.Authorization(app.OpUpdateCoupon, a.UpdateCouponHandler()))
			r.Delete("/{id}", a.Authorization(app.OpDeleteCoupon, a.DeleteCouponHandler()))
		})

		// Product Set Items
		r.Route("/", func(r chi.Router) {
			r.Get("/{id}", a.Authorization(app.OpGetProductSetItems, a.GetProductSetItemsHandler()))
		})

		// Products, images and prices
		r.Route("/products", func(r chi.Router) {
			r.Post("/", a.Authorization((app.OpCreateProduct), a.CreateProductHandler()))
			r.Put("/{id}", a.Authorization(app.OpUpdateProduct, a.UpdateProductHandler()))
			r.Get("/", a.Authorization(app.OpListProducts, a.ListProductsHandler()))
			r.Get("/{id}", a.Authorization(app.OpGetProduct, a.GetProductHandler()))
			r.Delete("/{id}", a.Authorization(app.OpDeleteProduct, a.DeleteProductHandler()))
		})

		r.Route("/images", func(r chi.Router) {
			r.Post("/", a.Authorization(app.OpAddImage, a.AddImageHandler()))
			r.Get("/{id}", a.Authorization(app.OpGetImage, a.GetImageHandler()))
			r.Get("/", a.Authorization(app.OpListProductImages, a.ListProductImagesHandler()))
			r.Delete("/{id}", a.Authorization(app.OpDeleteImage, a.DeleteImageHandler()))
			r.Delete("/", a.Authorization(app.OpDeleteAllProductImages, a.DeleteAllProductImagesHandler()))
		})

		r.Route("/prices", func(r chi.Router) {
			r.Get("/", a.Authorization(app.OpGetProductPrices, a.GetProductPrices()))
			r.Put("/", a.Authorization(app.OpUpdateProductPrices, a.UpdateProductPricesHandler()))
		})

		// Product to product associations groups
		r.Route("/products-assocs-groups", func(r chi.Router) {
			r.Post("/", a.Authorization(app.OpCreateProductToProductAssocGroup, a.CreatePPAssocGroupHandler()))
			r.Get("/{id}", a.Authorization(app.OpGetProductToProductAssocGroup, a.GetPPAssocGroupHandler()))
			r.Get("/", a.Authorization(app.OpListProductToProductAssocGroups, a.ListPPAssocGroupsHandler()))
			r.Delete("/{id}", a.Authorization(app.OpDeleteProductToProductAssocGroup, a.DeletePPAssocGroupHandler()))
		})

		// Product to product assocations
		r.Route("/products-assocs", func(r chi.Router) {
			r.Get("/{id}", a.Authorization(app.OpGetProductToProductAssoc, a.GetPPAssocHandler()))
			r.Get("/", a.Authorization(app.OpListProductToProductAssocs, a.ListPPAssocsHandler()))
			r.Delete("/{id}", a.Authorization(app.OpDeleteProductToProductAssoc, a.DeletePPAssocHandler()))
		})

		r.Route("/products-assocs:batch-update", func(r chi.Router) {
			r.Post("/", a.Authorization(app.OpBatchUpdateProductToProductAssocs, a.BatchUpdatePPAssocsHandler()))
		})

		// Shipping Tariffs
		r.Route("/shipping-tariffs", func(r chi.Router) {
			r.Post("/", a.Authorization(app.OpCreateShippingTariff, a.CreateShippingTariffHandler()))
			r.Get("/{id}", a.Authorization(app.OpGetShippingTariff, a.GetShippingTariffHandler()))
			r.Get("/", a.Authorization(app.OpListShippingTariffs, a.ListShippingTariffsHandler()))
			r.Put("/{id}", a.Authorization(app.OpUpdateShippingTariff, a.UpdateShippingTariffHandler()))
			r.Delete("/{id}", a.Authorization(app.OpDeleteShippingTariff, a.DeleteShippingTariffHandler()))
		})

		// Carts
		r.Route("/carts", func(r chi.Router) {
			r.Post("/", a.Authorization(app.OpCreateCart, a.CreateCartHandler()))
		})

		// Cart Coupons
		r.Route("/carts-coupons", func(r chi.Router) {
			r.Post("/", a.Authorization(app.OpApplyCouponToCart, a.ApplyCartCouponHandler()))
			r.Get("/{id}", a.Authorization(app.OpGetCartCoupon, a.GetCartCouponHandler()))
			r.Get("/", a.Authorization(app.OpListCartCoupons, a.ListCartCouponsHandler()))
			r.Delete("/{id}", a.Authorization(app.OpUnapplyCouponFromCart, a.UnapplyCartCouponHandler()))
		})

		r.Route("/carts-products", func(r chi.Router) {
			r.Post("/", a.Authorization(app.OpAddProductToCart, a.AddProductToCartHandler()))
			r.Get("/", a.Authorization(app.OpGetCartProducts, a.GetCartProductsHandler()))
			r.Patch("/{id}", a.Authorization(app.OpUpdateCartProduct, a.UpdateCartProductHandler()))
			r.Delete("/{id}", a.Authorization(app.OpDeleteCartProduct, a.DeleteCartProductHandler()))
			r.Delete("/", a.Authorization(app.OpEmptyCartProducts, a.EmptyCartProductsHandler()))
		})

		// Categories
		r.Route("/categories", func(r chi.Router) {
			r.Get("/", a.Authorization(app.OpGetCategories, a.GetCategoriesHandler()))
			r.Delete("/", a.Authorization(app.OpDeleteCategories, a.DeleteCategoriesHandler()))
		})

		r.Route("/categories-tree", func(r chi.Router) {
			r.Put("/", a.Authorization(app.OpUpdateCategoriesTree, a.UpdateCategoriesTreeHandler()))
			r.Get("/", a.Authorization(app.OpGetCategoriesTree, a.GetCategoriesTreeHandler()))
		})

		// Product - Category relationships
		r.Route("/products-categories", func(r chi.Router) {
			r.Post("/", a.Authorization(app.OpAddProductCategoryRelations, a.AddProductCategoryHandler()))
			r.Get("/{id}", a.Authorization(app.OpGetProductCategoryRelations, a.GetProductCategoryHandler()))
			r.Delete("/{id}", a.Authorization(app.OpDeleteProductCategoryRelations, a.DeleteProductCategoryHandler()))
			r.Put("/", a.Authorization(app.OpUpdateProductCategoryRelations, a.UpdateProductsCategoriesHandler()))
			r.Get("/", a.Authorization(app.OpGetProductCategoryRelations, a.GetProductsCategoriesHandler()))
			r.Delete("/", a.Authorization(app.OpDeleteProductCategoryRelations, a.DeleteProductsCategoriesHandler()))
		})

		// Orders
		r.Route("/orders", func(r chi.Router) {
			r.Post("/", a.Authorization(app.OpPlaceOrder, a.PlaceOrderHandler()))
			r.Get("/{id}", a.Authorization(app.OpGetOrder, a.GetOrderHandler()))
			r.Get("/", a.Authorization(app.OpListOrders, a.ListOrdersHandler()))
			r.Post("/{id}/stripecheckout", a.Authorization(app.OpStripeCheckout, a.StripeCheckoutHandler(stripeSuccessURL, stripeCancelURL)))
		})

		r.Route("/sysinfo", func(r chi.Router) {
			r.Get("/", a.Authorization(app.OpSystemInfo, a.SystemInfoHandler(si)))
		})

		// Webhooks
		r.Route("/webhooks", func(r chi.Router) {
			r.Post("/", a.Authorization(app.OpCreateWebhook, a.CreateWebhookHandler()))
			r.Get("/{id}", a.Authorization(app.OpGetWebhook, a.GetWebhookHandler()))
			r.Get("/", a.Authorization(app.OpListWebhooks, a.ListWebhooksHandler()))
			r.Patch("/{id}", a.Authorization(app.OpUpdateWebhook, a.UpdateWebhookHandler()))
			r.Delete("/{id}", a.Authorization(app.OpDeleteWebhook, a.DeleteWebhookHandler()))
		})
	})

	// public routes including GET / for Google Kuberenetes default healthcheck
	r.Group(func(r chi.Router) {
		r.Use(middleware.NoCache)

		// version info
		r.Get("/", healthCheckHandler)
		r.Get("/healthz", healthCheckHandler)
		r.Get("/config", a.ConfigHandler(si.Env.Firebase))
		r.Route("/stripe-webhook", func(r chi.Router) {
			r.Post("/", a.StripeWebhookHandler(stripeSigningSecret))
		})

		r.Route("/private-pubsub-events", func(r chi.Router) {
			r.Post("/", a.PubSubEventHandler(pubSubPushToken))
		})
		r.Route("/private-pubsub-broadcast", func(r chi.Router) {
			r.Post("/", a.PubSubBroadcastHandler(pubSubPushToken))
		})
	})

	r.Route("/signin-with-devkey", func(r chi.Router) {
		r.Post("/", a.SignInWithDevKeyHandler())
	})

	r.Route("/testdelay", func(r chi.Router) {
		r.Get("/", testDelayHandler)
	})

	// Server setup with signal handling
	srv := &http.Server{Addr: fmt.Sprintf(":%s", port), Handler: r}
	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)

		// interrupt signal sent from terminal
		signal.Notify(sigint, os.Interrupt)
		// sigterm signal sent from kubernetes
		signal.Notify(sigint, syscall.SIGTERM)

		switch sig := <-sigint; sig {
		case unix.SIGINT:
			log.Infof("main: received signal SIGINT")
		case unix.SIGTERM:
			log.Infof("main: received signal SIGTERM")
		default:
			log.Errorf("main: received unexpected signal %d", sig)
		}

		log.Infof("main: gracefully shutting down the server...")
		// We received an interrupt signal, shut down.
		if err := srv.Shutdown(context.Background()); err != nil {
			// Error from closing listeners, or context timeout:
			log.Infof("main: HTTP server Shutdown: %v", err)
		}
		log.Infof("main: HTTP server shutdown complete")
		close(idleConnsClosed)
	}()

	// Broadcast startup message on the events topic
	fbSrv.PublishTopicEvent(ctx, service.EventServiceStarted, service.ServiceStartedEventData{
		Name: "started",
	})

	// tlsMode determines whether to serve HTTPS traffic directly.
	// If tlsMode is false, you can enable HTTPS with a GKE Layer 7 load balancer
	// using an Ingress.
	if tlsMode {
		log.Infof("main: server listening on HTTPS port %s", port)
		log.Fatal(srv.ListenAndServeTLS(tlsCertFile, tlsKeyFile))
	} else {
		log.Infof("main: server listening on HTTP port %s", port)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("main: %v", err)
		}
	}

	<-idleConnsClosed
}
