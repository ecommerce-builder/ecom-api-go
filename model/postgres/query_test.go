package postgres_test

import (
	"testing"

	"bitbucket.org/andyfusniakteam/ecom-api-go/model/postgres"
)

func TestNewQuery(t *testing.T) {
	q := postgres.NewQuery("usr", map[string]bool{
		"id":        true,
		"uuid":      false,
		"uid":       false,
		"email":     true,
		"firstname": true,
		"lastname":  true,
		"created":   true,
		"modified":  true,
	})

	q = q.Select([]string{"id", "uuid", "firstname", "lastname", "email"})
	q = q.OrderBy("firstname")
	q = q.OrderDir("asc")
	q = q.Limit(2)
	q = q.StartAfter("93b8ea3a-6d4b-4fdc-abb3-49009198775c")
}
