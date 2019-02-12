# Ecom API


## Architecture

The project is divided into distict layers; app, service and model.

model/model.go   model package; defines the model structs and interfaces
model/postgres   postgres package; is an implementation of the model

app/service.go   app package; defines the service structs and interfaces
service/firebase firebase package; is an implementation of the service

### model

The models is the bottom most layer and is responsible for calling the pg
library with appropriate SQL queries.


### service
The services layer is an intermediary layer between the controllers and the
models. It calls the models layer to perform work and uses Pub/Sub to
communicate results to external systems.


### app
Each API endpoint exposes an Operation. Operations have a unique name
identifying them, such as `CreateCustomer`, `CreateAddress`, 
`AddItemToCart`.


## API
### Carts

#### CreateCart
Creates a new shopping cart returning a unique cart UUID to be used for all
subseqent requests.
```
POST /carts
```


#### AddItemToCart
Add a single item to a given cart.
```
POST /carts/{cart_uuid}/items
{
  "sku": "drill",
  "qty": 2
}
```


#### UpdateCartItem
Update an individual item in a given cart.

##### Request
```
POST /carts/{cart_uuid}/items/{sku}
{
  "qty": 3
}
```
##### Response
Returns `201 Created` on success.


#### DeleteCartItem
Delete an individual item from a given cart.

##### Request
```
DELETE /carts/{cart_uuid}/items/{sku}
```
##### Response
Returns `204 No Content` if succesfully deleted, or `404 Not Found` if the
item is not in the cart.

#### EmptyCartItems
Empties the entire shopping cart of all items.
##### Request
```
DELETE /carts/{cart_uuid}/items
```
##### Response
Returns `204 No Content` if the cart is successfully emptied.

### Customer

#### CreateCustomer
POST /customers

#### GetCustomer
GET /customers/{customer_uuid}

### Addresses

#### CreateAddress
POST /customers/{customer_uuid}/addresses

#### GetAddress
GET /addresses/{addr_uuid}

#### ListAddresses
GET /customers/{customer_uuid}/addresses

#### UpdateAddress
PATCH /addresses/{addr_uuid}

#### DeleteAddress
DELETE /addresses/{addr_uuid}

