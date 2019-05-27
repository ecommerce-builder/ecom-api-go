# CHANGELOG
## v0.36.0 (Mon 27 May 2019)
+ `OpAddImage` POST /products/:sku/images
+ `OpGetImage` GET /images/:uuid
+ `OpListProductImages` GET /products/:sku/images
+ `OpDeleteImage` DELETE /images/:uuid
+ `OpDeleteAllProductImages` DELETE /products/:sku/images

## v0.35.0 (Mon 27 May 2019)
+ Operation constants all prefixed with `Op`.
+ `OpGetTierPricing` GET /products/:sku/tiers/:ref/pricing
+ `OpListPricingBySKU` GET /products/:sku/pricing
+ `OpListPricingByTier` GET /products/tiers/:ref/pricing
+ `OpUpdateTierPricing` PUT /products/:sku/tiers/:ref/pricing
+ `OpDeleteTierPricing` DELETE /products/:sku/tiers/:ref/pricing
+ Operations requiring `RoleAdmin` return 403 Forbidden if `admin` is not passed in claim.

## v0.34.1 (Wed 22 May 2019)
+ Fixes OpUpdateCatalog to batch update the catalog table.
+ Code style tidy.

## v0.34.0 (Mon 20 May 2019)
+ OpDeleteAdmin to remove an administrator from the DB and Firebase.

## v0.33.1 (Mon 20 May 2019)
+ Fix missing role from signin-with-dev-key response.

## v0.32.0 (Mon 20 May 2019)
+ Remove table name prefixes from UUIDs.
+ Fixes missing CustomerUUID using a SQL join in GetCustomerDevKeyByDevKey.
+ OpSignInWithDevKey now returns both token and customer record.

## v0.31.0 (Sun 19 May 2019)
+ Generate root account includes local `customers` and `customer_devkey` tables.
+ Renamed CustomerUUID field to UUID to match database column.

## v0.30.0 (Fri 17 May 2019)
+ OpSysConfig (HTTP GET /config) implemented returns Google API Key.
+ Modified OpSystemInfo handler to include Google API Key.
+ `ECOM_GOOGLE_WEB_API_KEY` mandatory env var.
+ Uses no-cache middleware for non-protected routes.

## v0.29.0 (Tue 14 May 2019)
+ OpListProducts returns a list of all product SKUs.

## v0.28.5 (Mon 13 May 2019)
+ Service layer returns `map[string]*Assoc` and controller converts to slice.

## v0.28.4 (Mon 13 May 2019)
+ Match version string to tag.

## v0.28.3 (Mon 13 May 2019)
+ OpGetCatalogProductAssocs group by path.

## v0.28.2 (Fri 10 May 2019)
+ Replaces broken build `v0.28.1` fixing non-used fmt import.

## v0.28.1 (Fri 10 May 2019)
+ OpUpdateCatalogProductAssocs deletes existing associations at start of tx.
+ OpGetCatalog returns an empty JSON object {} if the catalog is empty.
+ Remove spurious fmt.Print

## v0.28.0 (Fri 10 May 2019)
+ Prevent adding catalog product assocs if catalog is empty.
+ Prevent purging catalog if catalog product assocs exist.
+ Rename operations to use the word Update in place of Replace.

## v0.27.0 (Thu, 9 May 2019)
+ OpGetCatalog returns list of product SKUs for each leaf node.

## v0.26.0 (Thu, 9 May 2019)
+ OpPurgeCatalogProductAssocs

## v0.25.0 (Wed, 8 May 2019)
+ Catalog API now uses JSON hierarchical structure, not nested set
+ OpPurgeCatalog deletes entire catalog
+ OpReplaceCatalog to load new catalog

## v0.24.0 (Tue, 30 Apr 2019)
+ Catalog Product Associations model and test code

## v0.23.0 (Sun, 28 Apr 2019)
+ Improve error handling for model
+ Remove convoluted interface layer from app
+ Remove error return for service.NewService

## v0.22.0 (Thu, 18 Apr 2019)
+ Product storing uses Postgres JSONB column (required scheme 0.7.0 and above)

## v0.21.0 (Fri, 12 Apr 2019)
+ OpProductExists implemented using HTTP HEAD on `/products/:sku`

## v0.20.0 (Fri, 12 Apr 2019)
+ Product API features
+ Interface pattern removed from the model layer

## v0.19.0 (Tue, 9 Apr 2019)
+ Add role field to customers

## v0.18.0 (Wed, 3 Apr 2019)
+ Developer keys

## v0.17.0 (Thu, 21 Feb 2019)
+ Expanded Authorization
+ GetCatalogProduct and UpdateCatalogProductAssocs handlers
+ Option to PASS Google service account credentials as base64 string
+ SystemInfo handler

## v0.16.0 (Wed, 13 Feb 2019)
+ CreateAdmin handler
+ `Op<X>` consts and `Role<X>` for compile-time checking

## v0.15.0 (Wed, 13 Feb 2019)
+ Create root superuser at startup if not exists. Uses `ECOM_APP_ROOT_EMAIL`
  and `ECOM_APP_ROOT_PASSWORD` environment variables. Mandatory.

## v0.14.0 (Tue, 12 Feb 2019)
+ Remove Google Pub/Sub
+ Pagination Result Query and Set management for Postgres pagination
+ Changed all reference to `tablename_uuid` to use generic uuid column names

## v0.13.0 (7 Feb 2019)
+ main tries 3 times to connect to PG with 5 seconds backoff periods.
+ data directory with README.md with instructions on how to generate init.sql
+ Docker Compose config added
+ Catalog and Catalog product associations with handler for HTTP GET

## v0.12.1 (22 Jan 2019)
+ Use v prefix (forgotton in last build)

## v0.12.0 (22 Jan 2019)
+ Prefix version tag with v from here on
+ `ECOM_PG_CONNECT_TIMEOUT` implemented and defaults to 10 seconds
+ `ListCustomers` operation for admin role
+ uses `go mod` for module management of project

## 0.11.0 (18 Dec 2018)
+ Git ignore all `ecom-api-*` executables and vendor/ dir
+ Add role param to CreateCustomer to create customers with admin roles

## 0.10.3 (7 Dec 2018)
+ Fix bug with deploy latest tag not being applied to current deploy
+ Build linux and Mac OS binaries at the same time as the rest of the deploy
+ Add version string to logging when app starts

## 0.10.2 (7 Dec 2018)
+ Remove all Kubernetes deploys and App Engine scripts

## 0.10.1 (7 Dec 2018)
+ Swap `ECOM_GOOGLE_CREDENTIALS_FILE`, `ECOM_APP_TLS_CERT_FILE` and `ECOM_APP_TLS_KEY_FILE` for `ECOM_GOOGLE_CREDENTIALS`, `ECOM_APP_TLS_CERT` and `ECOM_APP_TLS_KEY` respectively
## 0.10.0 (7 Dec 2018)
+ Drops `ECOM_DSN` env var single line string to use `ECOM_PG_*`
+ Use 3 categories of env vars `ECOM_APP_*`, `ECOM_GOOGLE_*` and `ECOM_PG_*` for App, Google and PostgreSQL config

## 0.9.3 (6 Dec 2018)
+ Remove COPY certificates from Dockerfile
+ Certs to be found on /etc/secret-volume mount point (not yet implemented for k8s)
+ `ECOM_CREDENTIALS_JSON` removed in place of `ECOM_CREDENTIALS_FILE`
+ `ECOM_TLS_MODE=enabled` to do SSL/TLS negociation at the go server level
+ `ECOM_TLS_CERT_FILE` and `ECOM_TLS_KEY_FILE` point to `cert.pem` and `key.pem` files
+ `ECOM_PORT` to override default 8080

## 0.7.2 (3 Dec 2018)
+ Fix CORS for HTTP GET requests
+ Extra debug on authz decorator middleware

## 0.7.1 (3 Dec 2018)
+ Adds missing authorization middleware
+ Adds missing changes to model interface

## 0.7.0 (3 Dec 2018)
+ Chi router

+ Authorization decorator pattern

## 0.6.1 (28 Nov 2018)
+ Fix broken 0.6.0 build

## 0.6.0 (28 Nov 2018)
+ Fix CORS
+ Authentication moved to global scope
+ Application logging using Logrus
+ API listens on port 9000
+ Adds Kubernetes deployment files. NodePort with Ingress

## 0.5.1 (19 Nov 2018)
+ Include missing file from 0.5.0

## 0.5.0 (19 Nov 2018)
+ Firebase Auth middleware
+ Pick runtime config from `ECOM_CREDENTIALS_JSON` and `ECOM_DSN` instead of individual `ECOM_DB*`
  environment variables
+ Kubernetes deploy config
+ Git ignore the alpine binary

## 0.4.0 (15 Nov 2018)
+ Interfaces used to create application domain and model domain
+ Model implementation provided for PostgreSQL in models/postgres
+ Service implemenation provided for Firebase in service/firebase
+ Explict Google cloud logging removed (will use Stdout for future)
+ Tested against ecom-js-client 1.1.0 integration tests

## 0.3.0 (7 Nov 2018)
+ Firebase Auth Go library calls CreateUser
+ Google Stackdriver logging
+ Address addr2 and county fields are optional

## 0.2.0 (1 Nov 2018)
+ GetCartItems operation returns all items in a a shopping cart
+ AddItemToCart uses the products_pricing table to lookup the price using the default tier

## 0.1.1 (1 Nov 2018)
+ Fixed typo on README extension.

## 0.1.0 (1 Nov 2018)
+ Initial API including model, service and controller layers
+ Cart API
+ Customer and address management API
