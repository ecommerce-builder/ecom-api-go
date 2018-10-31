package services

import (
	"bitbucket.org/andyfusniakteam/ecom-api-go/models"
)

// CreateCustomer creates a new customer
func CreateCustomer(firstname string, lastname string) *models.Customer {
	return models.CreateCustomer(firstname, lastname)
}

// GetCustomer retrieves a customer by customer UUID
func GetCustomer(customerUUID string) *models.Customer {
	return models.GetCustomerByUUID(customerUUID)
}

// CreateAddress creates a new address for a customer
func CreateAddress(customerUUID string, typ string, contactName string, addr1 string,
	addr2 string, city string, county string, postcode string, country string) *models.Address {
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
