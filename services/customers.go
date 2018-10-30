package services

import (
	"bitbucket.org/andyfusniakteam/ecom-api-go/models"
)

// CreateCustomer creates a new customer
func CreateCustomer(firstname string, lastname string) *models.Customer {
	return models.CreateCustomer(firstname, lastname)
}

// CreateAddress creates a new address for a customer
func CreateAddress() {
}
