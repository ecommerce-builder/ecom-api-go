# Ecom API

* [Configuration](#configuration)
  * [Environment Variable](#environment-variables)
    * [App](#env-app)
    * [Google](#env-google)
    * [Stripe](#env-stripe)
    * [Postgres](#env-postgres)
  * [File Structure](#file-structure)
* [Deployment](#deployment)
  * [Google Compute Engine](#gce)
* [App Architecture](#arch)
  * [Model](#arch-model)
  * [Service](#arch-service)
  * [App](#arch-app)
* [API](#api)
  * [CreateCart](#CreateCart)
  * [AddItemToCart](#AddItemToCart)
  * [GenerateCustomerDevKey](#GenerateCustomerDevKey)
  * [SignInWithDevKey](#SignInWithDevKey)
  * [PlaceOrder](#PlaceOrder)
  * [StripeCheckout](#StripeCheckout)

## <a name="configuration"></a>Configuration

The `ecom-api` executable accepts a configuration through the use of environment variables.


### <a name="environment-variables"></a>Environment Variables

Configuration is grouped into four groups; [App](#env-app), [Google](#env-google), [Stripe](#env-stripe) and [PostgreSQL](#env-postgres) prefixed with `ECOM_APP_`, `ECOM_GOOGLE_`, `ECOM_STRIPE_` and `ECOM_PG_` respectively. The Stripe configuration is optional and is only require if you intend to take payments using Stripe.


#### <a name="env-app"></a>App

| Env Var                  | Required | Default | Description |
| -------------            | -------- | ------- | ------------|
| **`PORT`**                   | Optional | 8080    | Usually set by the container. |
| **`ECOM_APP_TLS_MODE`**      | Optional | disable | To enable set the value to `enable`. If enabled you must set `ECOM_APP_TLS_CERT` and `ECOM_APP_TLS_KEY`. |
| **`ECOM_APP_TLS_CERT`**      | Depends  |         | Path to TLS certificate. e.g. `/etc/secret-volume/tls/api_spycameracctv_com/cert.pem`. Required if `ECOM_APP_TLS_MODE=enable`. |
| **`ECOM_APP_TLS_KEY`**       | Depends  |         | Path to TSL certificate key. e.g. `/etc/secret-volume/tls/api_spycameracctv_com/key.pem` |
| **`ECOM_APP_ROOT_EMAIL`**   | Optional |         |
| **`ECOM_APP_ROOT_PASSWORD`** | Required |         |
| **`ECOM_APP_ENABLE_STACKDRIVER_LOGGING`** | Optional | on | Accepts a value of `on` or `off` to switch the stack driver JSON formatted logging. |



#### <a name="env-google"></a>Google

| Env Var                         | Required | Default | Description |
| -------------                   | -------- | ------- | ------------|
| **`ECOM_GAE_PROJECT_ID`**       | Required |         | Set the value to the Google Project ID where the GAE App is running. For example, `open247-gae`. |
| **`ECOM_FIREBASE_PROJECT_ID`**  | Required |         | Set the value to the Firebase Project ID. e.g. `ecom-test-bf262`. |
| **`ECOM_FIREBASE_WEB_API_KEY`** | Required |         | Firebase Web API Key used to Authentication. For example, `AIzaSyDlPXxufb303i4e7fDV9fiURU05lQeX9Kc`. |
| **`ECOM_FIREBASE_CREDENTIALS`** | Required |         | Use either the filepath of the Firebase Service Account Credentials file or provide a Base64 encoded string. e.g. `/etc/secret-volume/service_account_credentials/ecom-test-fa3e406ce4fe.json` (or base64 encoded JSON string) |

#### <a name="env-stripe"></a>Stripe

The stripe settings are optional and only required if you are processing payments with the Stripe checkout. If you do not use Stripe, you do not need to pass either of the following environment variables to the ecom service. If you are using Stripe, make sure you pass both the secret key and signing key. The secret key is used for payments and the signing key is used to verify confirmations from Stripe webhooks.

| Env Var                       | Required | Default | Description |
| -------                       | -------- | ------- | ----------- |
| **`ECOM_STRIPE_SECRET_KEY`**  | Depends  |         | Found in the [API Keys](https://dashboard.stripe.com/test/apikeys) section of the Stripe Dashboard. |
| **`ECOM_STRIPE_SIGNING_KEY`** | Depends  |         | Found in the [Webhooks](https://dashboard.stripe.com/test/webhooks) section of the Stripe Dashboard. |
| **`ECOM_STRIPE_SUCCESS_URL`** | Depends  |         | URL to redirect to when checkout payment is successful. |
| **`ECOM_STRIPE_CANCEL_URL`**  | Depends  |         | URL to redirect to when checkout payment is cancelled. |

#### <a name="env-postgres"></a>Postgres

| Env Var                   | Required | Default | Description |
| -------------             | -------- | ------- | ------------|
| **`ECOM_PG_HOST`**            | Required |         | Example `35.234.153.166` |
| **`ECOM_PG_PORT`**            | optional | 5432	   | Postgres TCP/IP Port |
| **`ECOM_PG_DATABASE`**        | Required |         | Postgres database name. Example `ecom_dev` |
| **`ECOM_PG_USER`**            | Required |         | Postgres database user. Example `postgres` |
| **`ECOM_PG_PASSWORD`**        | Required |         | Postgres database password. |
| **`ECOM_PG_SSLMODE`**         | Optional | disable | Example `verify-ca`. |
| **`ECOM_PG_SSLCERT`**         | Optional |         | e.g. `/etc/secret-volume/pg/client-cert.pem`. |
| **`ECOM_PG_SSLKEY`**          | Optioanl |         | Example `/etc/secret-volume/pg/client-key.pem`. |
| **`ECOM_PG_SSLROOTCERT`**     | Optional |         | Example `/etc/secret-volume/pg/server-ca.pem`. |
| **`ECOM_PG_CONNECT_TIMEOUT`** | Optional | 10      | Postgres connection timeout in seconds. Example `5`. |


`ECOM_GOOGLE_PROJECT_ID` can be obtained from the Firebase Console. When you create a Firebase app, Google creates a Google Cloud project with the same project ID. The Firebase Auth is accessible via the Firebase control panel and is backed by a project of the same ID in the Google Cloud Console.

`ECOM_GOOGLE_CREDENTIALS` is the service account key file found in the Google Cloud Console, or Firebase console under the service account keys section.
`ECOM_PG_PORT` should use a private IP if running on GKE or GCE.

GAE sets `PORT` for each container it starts. Hard coding the port to 8080 causes an error message to appear in Stackdriver about nginx reverse proxy misconfiguration, resulting in poor performance.

### <a name="file-structure"></a>File Structure

Each VM running the ecom-api must have access to its own private disk mounted at `/etc/secret-volume`. This directory contains three directories `pg`, `service_account_credentials` and tls housing the PostgreSQL key files for SSL connections, Firebase Service Account files and SSL certificates respectively.

```
secret-volume/
├── pg
│   ├── client-cert.pem
│   ├── client-key.pem
│   └── server-ca.pem
├── service_account_credentials
│   ├── ecom-test-fa3e406ce4fe.json
│   └── test-spycameracctv-firebase-adminsdk-b06ml-46cd9030e2.json
└── tls
    ├── api_spycameracctv_com
    │   ├── cert.pem
    │   └── key.pem
    └── star_open24seven_co_uk
        ├── cert.pem
        └── key.pem
```


## <a name="deployment"></a>Deployment

### <a name="gce"></a>Google Compute Engine
GAE offers free SSL endpoints and provides a proxy method to securely connect to Postgres, so there is no need to deploy the secret volume containing either the `secrets/tls` and `secrets/pg` directories. The `ecom-api` executable accepts the service account credentials as a Base64 encoded string passed using the `ECOM_GOOGLE_CREDENTIALS` so there is often no need to package the service credentials file. The app can be configured entirely using environment variables.

See Google's documentation for [Deploying Containers on VMs and Managed Instance Groups](https://cloud.google.com/compute/docs/containers/deploying-containers).

Go 1.12 Standard Environment for App Engine recommends logging to stdout and stderr. See [Writing Application Logs](https://cloud.google.com/appengine/docs/standard/go112/writing-application-logs).

Other blog articles indicate that calling the Google Cloud API has higher latency than streaming to stdout from the service. This is better for cross platform compatibility as the Google Cloud Go library will lock to GCP only.

Container terminated by the container manager on signal 9.

`SIGKIL`

Appears that messages sent to stdout or stderr must have a terminating newline character to be seen in the Log Viewer.


``` go
fmt.Fprintf(os.Stdout, "Test message\n")
```

In your application code, look for the `X-Cloud-Trace-Context` HTTP header of incoming requests. Extract the trace identifier from the header.

Set the trace identifier in the LogEntry trace field of your app log entries. The expected format `isprojects/[PROJECT_ID]/traces/[TRACE_ID]`.

```
"X-Cloud-Trace-Context: TRACE_ID/SPAN_ID;o=TRACE_TRUE"
```

Where:
+ **`TRACE_ID`** is a 32-character hexadecimal value representing a 128-bit number. It should be unique between your requests, unless you intentionally want to bundle the requests together. You can use UUIDs.

+ **`SPAN_ID`** is the decimal representation of the (unsigned) span ID. It should be 0 for the first span in your trace. For subsequent requests, set `SPAN_ID` to the span ID of the parent request. See the description of `TraceSpan(REST, RPC)` for more information about nested traces.

+ **`TRACE_TRUE`** must be 1 to trace this request. Specify 0 to not trace the request.


```
$ gcloud compute instances list
```

#### <a name="updating-a-container"></a>Updating a container

Update container declaration on the the instance
Stops and restarts the instance to actuate

```
$ gcloud compute instances update-container test-spy --container-image gcr.io/cloud-marketplace/google/nginx1:1.13
```

#### Connecting to a container using SSH

```
cloud compute ssh --project [PROJECT_ID] --zone [ZONE] [INSTANCE_NAME]
```


Where:
+ `[PROJECT_ID]` is the ID of the project that contains the instance.
+ `[ZONE]` is the name of the zone in which the instance is located.
+ `[INSTANCE_NAME]` is the name of the instance.

```
$ gcloud compute ssh --project spycameracctv-d48ac --zone europe-west2-b spy-test-instance
```

Reserving a static external IP address

```
$ gcloud compute instances list
```


#### Create a new static IP address

```
$ gcloud compute addresses create test-spycameracctv-ip

For the following address:
 - [test-spycameracctv-ip]
choose a region or global:
 [1] global
 [2] region: asia-east1
 [3] region: asia-east2
 [4] region: asia-northeast1
 [5] region: asia-south1
 [6] region: asia-southeast1
 [7] region: australia-southeast1
 [8] region: europe-north1
 [9] region: europe-west1
 [10] region: europe-west2
 [11] region: europe-west3
 [12] region: europe-west4
 [13] region: northamerica-northeast1
 [14] region: southamerica-east1
 [15] region: us-central1
 [16] region: us-east1
 [17] region: us-east4
 [18] region: us-west1
 [19] region: us-west2
Please enter your numeric choice:  10

Created [https://www.googleapis.com/compute/v1/projects/spycameracctv-d48ac/regions/europe-west2/addresses/test-spycameracctv-ip].
```

I have created a reserved static external IP address (bottom line of this screenshot). It's given a name of test-spycameracctv-ip and has an IP address of `35.246.47.254` running in the `europe-west2` region.

The Route 53 DNS of `test-spycameracctv-api.open24seven.co.uk` points to that same IP address `35.246.47.254`.

Even if the VM is destroyed and replaced with another, the new VM will be attached to the `test-spycameracctv-ip` configuration each time (i.e. it will always be given the same IP address. This means you always connected via the API endpoint above but you make be connecting to different VMs over time.)

Postgres

```
docker run --name postgres-9.6.10 -d -p 5432:5432 \
-e POSTGRES_PASSWORD=postgres postgres:9.6.10
```

Run psql to connect to the database:

```
CREATE DATABASE ecom_dev WITH ENCODING 'UTF8';
\connect ecom_dev
```

To enable SQL logging first connect to the running container.

```
$ docker exec -it postgres-9.6.10 bash
```

Inside the container, first install an editor such as vim or nano and edit `/var/lib/postgresql/data/postgresql.conf`.

```
apt update; apt install vim nano
vim /var/lib/postgresql/data/postgresql.conf
```

See Stack Overflow Question on [How to log PostgreSQL queries?](https://stackoverflow.com/questions/722221/how-to-log-postgresql-queries)


## <a name="arch"></a>App Architecture

The project is divided into distict layers; app, service and model.

model/model.go   model package; defines the model structs and interfaces
model/postgres   postgres package; is an implementation of the model

app/service.go   app package; defines the service structs and interfaces
service/firebase firebase package; is an implementation of the service

### <a name="arch-model"></a>Model

The models is the bottom most layer and is responsible for calling the pg
library with appropriate SQL queries.


### <a name="arch-service"></a>Service
The services layer is an intermediary layer between the controllers and the
models. It calls the models layer to perform work and uses Pub/Sub to
communicate results to external systems.


### <a name="arch-app"></a>App
Each API endpoint exposes an Operation. Operations have a unique name
identifying them, such as `CreateCustomer`, `CreateAddress`,
`AddItemToCart`.


#### Google Compute Engine (GCE)
Pros:
+ Cheap
+ Direct control over instance.

Cons:
+ Have to manage terminating SSL traffic with Go app.
+ Scores B grade for SSL analysis.
+ Hard to redeploy container.
+ Manual config for Load Balancing.


#### Google App Engine Stanard (GAE)

Pros:
+ Built in Load Balancer.
+ Handles SSL
+ Custom domains with free SSL or upload own certs.
+ Can handle multiple apps???

Cons:

+ Learn Proprietary system.
+ Vendor lock-in using platform specifics like image, mail, crown etc.
+ Hard to predict billing.
+ Might not connect direct to private IP for Postgres (Hi latency).


#### Kubernates (K8S)

Pros:
+ Industry standard looks good on CV.
+ Fast deployment.
+ Easy to scale.
+ Good documentation.

Cons:
+ Steep learning curve
+ Minimal unit is a container
+ Expensive Load Balancer.
+ Complex YAML to maintain.



|    | GCE  | GAE Standard | K8S |
| -- | ---- | ------------ | --- |
| Scalability | Poor | Good | Excellent |
| Ease of deployment | Fairly complicated Bash script. | Small number of YAML files. | Lots of YAML configurations. |
| Speed of deployment | | Slow (3-4m) | Fast |
| Learning curve | Medium | Medium | Hard |
| Fit for projects | Small standalone apps. |Start-up Apps and APIs. | Large-scale projects. |
| Logging | Hard | Good | Several layers of login including Docker. |
| Complexity | Average. | | Complex. |
| Proprietary | | Risk of vendor lock-in if relying on Particular dependencies. | Open source |


API
---

To run the curl examples in this section `$JWT` must be set to a valid Javascript Web Token from Firebase Auth.
`$ENDPOINT` must be set to a valid API endpoint such as
```
https://open247-gae.appspot.com
```

## Authentication

Most requests require a JSON Web Token (JWT) aquired by signing into the Firebase Auth that corresponds to the API endpoint.

API resources are protected by levels of access:

* None (No JWT required)
* `RoleShopper` (JWT contains no custom `role` claim)
* `RoleCustomer` (JWT contains a custom claim `"role": "customer"`)
* `RoleAdmin` (JWT contains a custom claim `"role": "admin"`)
* `RoleSuperUser` (JWT contains a custom claim `"role": "root"`)


Set the `Authorization: Bearer $JWT` header before calling the API endpoints that require authentication.

### CreateCart

Creates a new shopping cart returning a unique cart ID to be used for all
subseqent requests. JWT must contain `RoleShopper`, `RoleCustomer` or `RoleSuperUser` privileges.

#### Request
Use an empty request body.

```http
POST https://open247-gae.appspot.com/carts
```

#### Curl Example
```bash
curl -X POST \
-H "Authorization: Bearer $JWT" $ENDPOINT/carts
```

#### Example Response

#### 201 Created
``` json
{
    "object": "cart",
    "id": "f83796a0-b1f2-4e5a-a207-19ea0956475f"
}
```

___


### AddItemToCart
Add a an item to the cart with the given ID. JWT must contain `RoleShopper`, `RoleCustomer` or `RoleSuperUser` privileges.

#### Request
```http
POST /carts/:id/items
```

```bash
curl -v -X POST \
-H "Authorization: Bearer $JWT" \
-H 'Accept: application/json' \
-H 'Content-Type: application/json' \
-d '{"sku":"DRILL-SKU","qty":2}' \
$ENDPOINT/carts/f83796a0-b1f2-4e5a-a207-19ea0956475f/item
```

### Request body
``` json
{
  "sku": "DRILL-SKU",
  "qty": 2
}
```

#### Example Response

#### 201 Created
``` json
{
    "object": "item",
    "sku": "DRILL-SKU",
    "qty": 2,
    "unit_price": 192900,
    "created": "2019-07-01T13:49:25.526664Z",
    "modified": "2019-07-01T13:49:25.526664Z"
}
```


#### 409 Conflict
``` json
{
    "status": 409,
    "code": "cart/cart-item-already-exists",
    "message": "cart item already exists in the cart"
}
```

___


#### UpdateCartItem
Update an individual item in a given cart.

##### Request
```http
PATCH /carts/:id/items/:sku
```

##### Request body
```bash
{{ENDPOINT}}/carts/:uuid/items/:sku
```

``` json
{
  "qty": 3
}
```

Example Response

#### 201 Created


___



#### DeleteCartItem
Delete an individual item from a given cart.

##### Request
``` http
DELETE /carts/:id/items/:sku
```
##### Response
Returns `204 No Content` if succesfully deleted, or `404 Not Found` if the
item is not in the cart.


___

### OpGetCartItems

``` http
GET /carts/:id/items
```

Example Response

### 200 OK

``` json
{
    "object": "list",
    "items": [
        {
            "object": "cart_item",
            "sku": "DESK-SKU",
            "qty": 1,
            "unit_price": 25482,
            "created": "2019-06-27T12:28:52.999335Z",
            "modified": "2019-07-01T14:02:11.91783Z"
        },
        {
            "object": "cart_item",
            "sku": "TV-SKU",
            "qty": 2,
            "unit_price": 14457,
            "created": "2019-07-01T13:49:25.526664Z",
            "modified": "2019-07-01T13:49:25.526664Z"
        }
    ]
}
```

___


#### EmptyCartItems
Empties the entire shopping cart of all items.
##### Request
```
DELETE /carts/:id/items
```
##### Response
Returns `204 No Content` if the cart is successfully emptied.

___


#### CreateCustomer
``` http
POST /customers
```

___


#### GetCustomer
``` http
GET /customers/:id
```

____


#### CreateAddress
``` http
POST /customers/:id/addresses
```

____


#### GetAddress
``` http
GET /addresses/:id
```
____


#### ListAddresses
``` http
GET /customers/:id/addresses
```

____


#### UpdateAddress
``` http
PATCH /addresses/:id
```

____


#### DeleteAddress
``` http
DELETE /addresses/:id
```

___

#### GenerateCustomerDevKey

Generates a new customer developer key. Requires `RoleAdmin` or `RoleSuperUser` privilege.

```http
POST /customers/:id/devkeys
```

```bash
curl -v -X POST \
-H "Authorization: Bearer $JWT" \
-H 'Accept: application/json' \
$ENDPOINT/customers/faaa916e-bf0d-4e56-8023-23a65f1030ac/devkeys
```

##### Response 201 Created
```http
{
  "object": "developer_key",
  "id": "4305b5ad-7c20-4072-9513-119ecfc54870",
  "key": "8JjVkbyNU3hMmQdApnfmtwR52xy4AR3BmhbnsKCPgris",
  "created": "2019-07-26T16:17:12.990307Z",
  "modified": "2019-07-26T16:17:12.990307Z"
}
```

___


#### SignInWithDevKey

```http
POST /signin-with-devkey
```

Signin with a secret developer key. After you receive the custom token in the response, pass it to Firebase Auth [signInWithCustomToken](https://firebase.google.com/docs/reference/node/firebase.auth.Auth#sign-inwith-custom-token) to sign in the user:

```javascript
firebase.auth().signInWithCustomToken(token).catch(function(error) {
  // Handle Errors here.
  var errorCode = error.code;
  var errorMessage = error.message;
  // ...
});
```

To sign out a user, call `signOut`:

```javascript
firebase.auth().signOut().then(function() {
  // Sign-out successful.
}).catch(function(error) {
  // An error happened.
});
```



```bash
curl -v -X POST \
-H 'Accept: application/json' \
-H 'Content-Type: application/json' \
-d '{"key": "8JjVkbyNU3hMmQdApnfmtwR52xy4AR3BmhbnsKCPgris"}' $ENDPOINT/signin-with-devkey
```

#### Response 200 OK
```http
{
  "custom_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJmaXJlYmFzZS1hZG1pbnNkay1xbGwzaEB0ZXN0LWRhdGEtb2FrLmlhbS5nc2VydmljZWFjY291bnQuY29tIiwiYXVkIjoiaHR0cHM6Ly9pZGVudGl0eXRvb2xraXQuZ29vZ2xlYXBpcy5jb20vZ29vZ2xlLmlkZW50aXR5LmlkZW50aXR5dG9vbGtpdC52MS5JZGVudGl0eVRvb2xraXQiLCJleHAiOjE1NjQxNjE3MDUsImlhdCI6MTU2NDE1ODEwNSwic3ViIjoiZmlyZWJhc2UtYWRtaW5zZGstcWxsM2hAdGVzdC1kYXRhLW9hay5pYW0uZ3NlcnZpY2VhY2NvdW50LmNvbSIsInVpZCI6IkFIeXJWTUZ5cEdOOVJ1VDF3R0VlcmtvcEdybjEifQ.lp0xbU2UB0YHABnC8slFYOa7UmifqnsjNK7zASVpKfsU9l6J1aux26BWw06tnwJ-a6PuTrUw6RudWwqkbxezV3kmu7AfTWh0F9Vr3styyJc-k6NKkp8H3NzG11Jz00s-7q1WZOucR8yHf3puxyvxkKsp7Clq_b61uVpeotNEdmtZ6lZ1jU3wYTN5U3DucyLAGvMYiDZI9OO1KhQcN8xfCGREarTfpthVbWBig428ISc1rOgP0nUawzqd5c0t3rhYrZS6PKoi7DvOuSdw3f3YCgfk1mBNwweNT_6RjkNN5dsz1mt7Ss_P6dMUo1NOaHRADpkFK6u42g3ZTc6P7jqmBg",
  "customer": {
    "id": "faaa916e-bf0d-4e56-8023-23a65f1030ac",
    "uid": "AHyrVMFypGN9RuT1wGEerkopGrn1",
    "role": "root",
    "email": "andy+root@andyfusniak.com",
    "firstname": "Andy",
    "lastname": "Fusniak",
    "created": "2019-07-26T15:51:36.411177Z",
    "modified": "2019-07-26T15:51:36.411177Z"
  }
}
```
___

#### PlaceOrder

``` http
POST /orders
```

___


#### StripeCheckout

```http
POST /orders/:id/stripecheckout
```

Initiates the Stripe checkout process recording an intent to pay against the order.
##### Request


```bash
curl -v -X POST \
-H "Authorization: Bearer $JWT" \
-H 'Accept: application/json' \
$ENDPOINT/orders/f76538c3-e94b-42e6-b1a5-ddbee2b1e3c9/stripecheckout
```

##### Response
Returns `200 OK`.

```http
{
  "object": "stripe_checkout_session",
  "checkout_session_id": "cs_test_7sKcYwUtBPWVonB6b5aa0UrhwTBrId78Wb9l0GTEduj3rCwmi33EJAEr"
}
```
