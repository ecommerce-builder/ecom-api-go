# CHANGELOG
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
