package services

import (
	"bitbucket.org/andyfusniakteam/ecom-api-go/models"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"golang.org/x/net/context"
)

// App is a Firebase App used for handling authentication
var App *firebase.App

// CreateCustomer creates a new customer
func CreateCustomer(email, password, firstname, lastname string) (*models.Customer, error) {
	ctx := context.Background()
	authClient, err := App.Auth(ctx)
	if err != nil {
		panic(err)
	}

	user := (&auth.UserToCreate{}).
		Email(email).
		EmailVerified(false).
		Password(password).
		DisplayName(`${firstname} ${lastname}`).
		Disabled(false)

	userRecord, err := authClient.CreateUser(ctx, user)
	if err != nil {
		panic(err)
	}

	return models.CreateCustomer(userRecord.UID, email, firstname, lastname)
}

// GetCustomer retrieves a customer by customer UUID
func GetCustomer(customerUUID string) *models.Customer {
	return models.GetCustomerByUUID(customerUUID)
}

// CreateAddress creates a new address for a customer
func CreateAddress(customerUUID, typ, contactName, addr1 string, addr2 *string, city string, county *string, postcode string, country string) *models.Address {
	customerID, _ := models.GetCustomerIDByUUID(customerUUID)
	return models.CreateAddress(customerID, typ, contactName, addr1, addr2, city, county, postcode, country)
}

// GetAddress gets an address by UUID
func GetAddress(addressUUID string) *models.Address {
	return models.GetAddressByUUID(addressUUID)
}

// GetAddresses gets a slice of addresses for a given customer
func GetAddresses(customerUUID string) ([]models.Address, error) {
	customerID, _ := models.GetCustomerIDByUUID(customerUUID)

	return models.GetAddresses(customerID)
}

// DeleteAddress deletes an address by uuid
func DeleteAddress(addrUUID string) error {
	return models.DeleteAddressByUUID(addrUUID)
}

// UpdateAddress updates an address by uuid
// func UpdateAddress() (*models.Address, error) {
// 	return models.UpdateAddressByUUID(addrUUID)
// }
