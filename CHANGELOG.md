# CHANGELOG

## v0.62.2 (Tue, 5 Nov 2019)
+ Fix request validation for `PATCH /address` operation `OpUpdateAddress`.
+ Improved logging in main and app level handlers for coupons and addresses.
+ `PromoRuleCode` attribute added to coupons.
+ Minor changes to Open API Def.
+ Update deps using `go get -v all`.

## v0.62.1 (Sun, 3 Nov 2019)
+ Fix sysinfo JSON response to include Stripe and endpoint.

## v0.62.0 (Sat, 2 Nov 2019)
+ Add Stackdriver Profiler when
+ Fix issue with missing `created` and `modified` values when creating a new promo rule.
+ Alter compiler options to use -gcflags=all='-N -l' for compute engine.
+ Update all dependencies using `go get -v all`.

## v0.61.4 (Thu, 31 Oct 2019)
+ Fixes issues with JSON responses for promo_rule objects.
+ [product_id attribute missing from HTTP GET /promo-rules/{id} response](https://bitbucket.org/andyfusniakteam/ecom-api-go/issues/20/product_id-attribute-missing-from-http-get)
+ [HTTP POST /promo-rules returns a blank promo_rule_code attribute](https://bitbucket.org/andyfusniakteam/ecom-api-go/issues/23/http-post-promo-rules-returns-a-blank)

## v0.61.3 (Tue, 22 Oct 2019)
+ Add site to the Open API Def YAML.
+ Ignore google*.html in git.
+ Tidy code

## v0.61.2 (Tue, 22 Oct 2019)
+ Return structured JSON for 401 Unauthorized and 403 Forbidden responses.

## v0.61.1 (Tue, 22 Oct 2019)
+ Fix authorization issues.
+ Remove Google vertification handler and replace with dedicated HTTP service `google-verifcation`.
+ Code tidy `app/create_webhook.go` and fix logging of pointer values.

## v0.61.0 (Mon, 21 Oct 2019)
+ Google Pub/Sub broadcasting of messages to handle HTTP POSTs to webhook endpoints.
+ `overselling` for inventory.
+ `ErrCodeWebhookPostFailed` error handling for HTTP POST failures.
+ Most handlers use pointers for the request body to detect missing attributes.
+ Remove nesting from Go error handling.
+ `ECOM_APP_ENDPOINT env var introduced.
+ `ECOM_GOOGLE_PUBSUB_PUSH_TOKEN` secret token for basic auth to the push endpoint.
+ Initial creating of topics and subscriptions to handle webhooks.
+ `/privte-pubsub-events` and `/privte-pubsub-broadcast` resources.
+ Webhooks calls implemented.
+ Makefile has two new deployments for test and live.
+ GAE now supports go113 so go.mod file is updated to use Go 1.13.3.
+ Open API Def file improved to near match API Service.
+ `offer_id` added to `price` table.


## v0.60.1 (Tue 1 Oct 2019)
+ Minor fixes to schema (`shipping_code` on `shipping_tariff` table made non unique)
+ Remove verbosity of schema create, drop and load shell scripts.
+ `OpGetPromoRule` GetPromoRuleHandler validation fix for URL param.
+ `OpGetPromoRule` fix 404 response for promo rule not found.

## v0.60.0 (Tue 1 Oct 2019)
+ Webhooks CRUDL for configuration only.
+ Google Pub/sub connection on startup.
+ Product to product group associations impemented.
+ Product to product associations implemented.
+ Coupons in the cart implemented.
+ Inventory implemented.
+ Shipping Tariffs implemented.
+ `offer_price` added to `price` table in schema.
+ Offer activation.
+ Cart products fixes and Open API def updates.
+ Shipping Tariffs spelling change.

## v0.58.1 (Fri 13 Sep 2019)
+ Fixes issues with the Postgres table creation script order in which the tables are created.

## v0.58.0 (Fri 13 Sep 2019)
+ ecom binary `ECOM_FIREBASE_PROJECT_ID` deprecated.
+ ecom binary `ECOM_FIREBASE_WEB_API_KEY` and `ECOM_FIREBASE_CREDENTIALS` become `ECOM_FIREBASE_PUBLIC_CONFIG` and `ECOM_FIREBASE_PRIVATE_CREDENTIALS` respectively.
+ README.md removes all documenation (held in separate repo and in the Open API Spec Definition).
+ `OpActivateOffer`
+ New handlers for product to category relations.
+ New handlers for adding products to carts.
+ Implements Shipping Tariffs.
+ Implements Price Lists.
+ Implements Coupons.
+ Implements Promo Rules.
+ Implements Inventory.
+ Implements Offers.
+ Implements Carts Products.
+ Improve Developer Keys
+ Implements Addresses.
+ Implements Prices.
+ `/users` has pagination skeleton but not yet working.
+ New `/products-categories` resource.
+ Implements `/products?include=prices,images` although it's not determined if it will be used in the SDK.
+ Merged `/admin` resource to universal `/users` resource.
+ `categories-tree` handles hierarchial data. `categories` handles flat nested set lists.
+ `clientError` and `serverError` functions introduced to remove boilerplate code.
+ Context values use named types to prevent golint errors.
+ Firebase Config is returned in the `/config` resource so the client doesn't need to hard code it.
+ Fixes CORS issue with `PUT` not being allowed.
+ Fixes golint warnings.
+ Postgres Schema changes.
+ Open API Definition file is not up to date with this release.

## v0.57.0 (Mon 5 Aug 2019)
+ Cart split into `carts` and `cart_items` tables.
+ Fix CORS bug with public routes.
+ Updated Open API Spec.

## v0.56.0 (Thu 1 Aug 2019)
+ Fix authorization on `OpDeleteCustomerDevKey`.
+ Apply `"object": "image"` to `OpGetImage` and `OpUpdateTierPricing`.
+ Improve OpenAPI `api.yaml` contract.

## v0.55.0 (Tue 30 Jul 2019)
+ All routes use `id` not `uuid`.

## v0.54.0 (Tue 30 Jul 2019)
+ Cart routes `uuid` becomes `id`.
+ Cart operations return the `name` of the product alongside the `sku` and `qty` etc.
+ Schema test data uses `DESK-SKU=2987083`, `DRILL-SKU=395833`, `PHONE-SKU=241583`, `TV-SKU=2066250` and `WATER-SKU=20417`.

## v0.53.0 (Mon 29 Jul 2019)
+ Stripe webhook implemented
+ Stripe checkout implemented
+ Alter schema for payments
+ Unit prices use 4 decimal places
+ Updated README.md
+ Record Stripe Payment Intent reference with orders
+ Change developer key JSON response

## v0.52.0 (Wed 10 Jul 2019)
+ Silence the Makefile targets.
+ Schema `order_items` table includes `currency`, `discount` and `tax_codes`.
+ Guest and Customer order placement.

## v0.50.0 (Mon 1 Jul 2019)
+ `OpPlaceOrder` basic skelton implemetation.
+ `object: "<name>"` wrapper for image, cart item, customer, devkey
+ `UUID` replaced with `ID` for service layer.
+ Firebase Auth customer claims uses `cid` not `cuuid`.
+ Error handling for `UpdateCartItem` service for missing cart and item combination.
+ Error handling for `AddItemToCart` service for existing item using `ErrCodeCartAlreadyExists`.
+ Contextual logging throughout app, service and model layers.
+ Updated `README.md`.
+ Open API `app.yaml` file for documenation.

## v0.49.3 (Tue 25 Jun 2019)
+ Uses `stackdriver-gae-logrus-plugin` `v0.1.3` with parse `X-Cloud-Trace-Context` bugfix.

## v0.49.2 (Tue 25 Jun 2019)
+ Update deps for `go.mod` including `v0.1.2` of stackdriver-gae-logrus-plugin
+ Contextual logging in Authorization middleware.

## v0.49.0 (Tue 25 Jun 2019)
+ Implements `stackdriver-gae-logrus-plugin` for structured logging with log entry threading.
+ `OpSysInfo` returns separate Firebase and Google config settings.

## v0.48.2 (Fri 21 Jun 2019)
+ Have to embed the version string into `main.go` as no way to pass it to GAE builds.

## v0.48.1 (Fri 21 Jun 2019)
+ Exeutable `version` string comes from `VERSION` file in root directory. Must use make to build.
+ Makefile includes `deploy-gae` to push latest version to GAE Standard.
+ Executable says 'hello from ecom-api version x.y.z' to match 'goodbye'.

## v0.48.0 (Fri 21 Jun 2019)
+ Gracefully shutdown the HTTP Service by trapping signals `SIGINT` (2) and `SIGTERM` (15).
+ Defer shutdown on Postgres.
+ Goodbye message sent upon final termination.

## v0.47.1 (Thu 20 Jun 2019)
+ Executable outputs runtime.Version(), GOOS and GOARCH upon startup.
+ Makefile to automate builds and Dockerfile modified to use alpine build from Makefile.

## v0.47.0 (Thu 20 Jun 2019)
+ Postgres schema `0.15.1` relocated to this repo. For previous history of changes see the [original repo](https://bitbucket.org/andyfusniakteam/ecom-schemas). Each future schema version will match the `ecom-api` version string.

## v0.46.3 (Wed 19 Jun 2019)
+ Bugfix `OpCreateAddress`, `OpGetCustomersAddresses` and `OpUpdateAddress` for `cuuid` -> `uuid` url param.

## v0.46.2 (Wed 19 Jun 2019)
+ Bugfix `OpGetCustomer` using `cuuid` url param instead of `uuid`.

## v0.46.1 (Tue 18 Jun 2019)
+ Remove `uuid` property from JSON array response in GET /carts/:uuid/items

## v0.46.0 (Thu 13 Jun 2019)
+ New env vars `ECOM_APP_MAX_OPEN_CONNS`, `ECOM_APP_MAX_IDLE_CONNS` and `ECOM_APP_CONN_MAX_LIFETIME` used to
  configure relational database connections.
+ New env var `ECOM_APP_ENABLE_STACKDRIVER_LOGGING` to switch formats in Google App Engine.

## v0.45.0 (Thu 6 Jun 2019)
+ Catalog associations `PUT /assocs` uses map[string]Assocs instead of structs.

## v0.44.1 (Tue 4 Jun 2019)
+ Fix bug with `OpGetCatalog` not returning the first product of each path.

## v0.44.0 (Tue 4 Jun 2019)
+ Drop `summary` and `meta.keywords` fields. Not needed.
+ Remove `marshalProduct` func. No longer used.

## v0.43.0 (Mon 3 Jun 2019)
+ `OpListPricingBySKU` renamed to `OpMapPricingBySKU`.
+ `OpListPricingByTier` renamed to `OpMapPricingByTier`.
+ `a.Service.ListPricingBySKU` renamed to `a.Service.PricingMapBySKU`.
+ `a.Service.ListPricingByTier` renamed to `a.Service.PricingMapByTier`.
+ Buxfix: field tag corrected from `in-the-box` to `in_the_box`.
+ Introduces two _named types_ `SKU` and `TierRef` improving readablity.

## v0.42.1 (Mon 3 Jun 2019)
+ Fix broken field tag on `OpProductApply` `in_the_box`.

## v0.42.0 (Mon 3 Jun 2019)
+ Requires schema 0.15.0
+ Adds `schema_version` to `OpSystemInfo` response.

## v0.41.0 (Mon 3 Jun 2019)
+ Requires schema 0.14.0
+ Uses `catagories` instead of `catalog` for tables.
+ catalog becomes a collective noun for all categories in the `categories` table.
+ Associations are held in the `categories_products` table and data structs are referred to as `CategoryProductAssoc`.
+ Resource becomes GET `/categories` instead of `/catalog`. This ensures future proofing GET `/catalog/x/yz`.

## v0.40.1 (Fri 31 May 2019)
+ Bugfix `OpGetCatalog` returns products with incorrect category path instead of product slug path.
+ Add .gitcloudignore to prevent unnecessary files uploading.

## v0.40.0 (Fri 31 May 2019)
+ Requires database scheme version 0.13.0
+ `product_pricing_tiers` table becomes `pricing_tiers`.
+ Replaced two operations `OpCreateProduct` and `OpUpdateProduct` with `OpReplaceProduct`.
+ New `ErrCodeDuplicateImagePath` return code to prevent duplicate image paths on `OpReplaceProduct`.
+ `Data` field becomes `Content`.
+ Skeleton validation added to `OpReplaceProduct`. Still has more to do.

## v0.39.0 (Wed 29 May 2019)
+ Unexported ids on postgres layer.
+ Add `path` and `name` to the OpGetCatalog products section.

## v0.38.0 (Wed 29 May 2019)
+ Sentinel values for JSON error response `code` property.
+ Renamed all application handler to remove the postfix.
+ Split the model layer into separate files.
+ Split the service layer into separate files.
+ Moved the tree code from utils/nested set to the service catalog.

## v0.37.1 (Tue 28 May 2019)
+ Product `url` becomes `path`.

## v0.37.0 (Tue 28 May 2019)
+ Not Implemented routes for pricing tiers
+ `OpCreateTier` POST /tiers
+ `OpGetTier` GET /tiers/:ref
+ `OpListTiers` GET /tiers
+ `OpUpdateTier` PUT /tiers/:ref
+ `OpDeleteTier` DELETE /tiers/:ref
+ Add missing HTTP Status code responses
+ Enforce all routing through Authorization middleware
+ Code tidy

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
