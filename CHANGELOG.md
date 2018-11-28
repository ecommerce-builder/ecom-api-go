# CHANGELOG
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
