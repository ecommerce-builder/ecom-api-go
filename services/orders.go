package services

import "bitbucket.org/andyfusniakteam/ecom-api-go/models"

type Order struct {
}

func PlaceOrder(o Order) {
	models.CreateOrder()
}
