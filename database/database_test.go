package database

import (
	"context"
	"log"
	"testing"

	"github.com/rtsncs/remitly-swift-api/models"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var db Database

func TestInsertAndGetByCode(t *testing.T) {
	c := context.Background()

	code := models.SwiftCode{
		Code:        "TESTUS33XXX",
		Name:        "Test Bank HQ",
		Address:     "123 Wall Street",
		CountryISO2: "US",
		CountryName: "United States",
		Headquarter: true,
	}

	err := db.InsertCode(c, code)
	assert.NoError(t, err)

	fetched, err := db.GetByCode(c, code.Code)
	assert.NoError(t, err)
	assert.Equal(t, code.Code, fetched.Code)
	assert.Equal(t, code.Name, fetched.Name)
	assert.Equal(t, code.Headquarter, fetched.Headquarter)
}

func TestGetBranches(t *testing.T) {
	c := context.Background()

	hq := models.SwiftCode{
		Code:        "BANKUS12XXX",
		Name:        "Bank HQ",
		Address:     "1 HQ Street",
		CountryISO2: "US",
		CountryName: "United States",
		Headquarter: true,
	}
	branch1 := models.SwiftCode{
		Code:        "BANKUS12NYC",
		Name:        "Bank NYC",
		Address:     "2 NYC Ave",
		CountryISO2: "US",
		CountryName: "United States",
		Headquarter: false,
	}
	branch2 := models.SwiftCode{
		Code:        "BANKUS12CHI",
		Name:        "Bank Chicago",
		Address:     "3 CHI Blvd",
		CountryISO2: "US",
		CountryName: "United States",
		Headquarter: false,
	}

	_ = db.InsertCode(c, hq)
	_ = db.InsertCode(c, branch1)
	_ = db.InsertCode(c, branch2)

	branches, err := db.GetBranches(c, hq.Code)
	assert.NoError(t, err)
	assert.Len(t, branches, 2)
	assert.NotEqual(t, "XXX", branches[0].Code[len(branches[0].Code)-3:])
}

func TestGetCountryName(t *testing.T) {
	c := context.Background()

	code := models.SwiftCode{
		Code:        "TESTCA22XXX",
		Name:        "Test Canada Bank",
		Address:     "123 Maple Road",
		CountryISO2: "CA",
		CountryName: "Canada",
		Headquarter: true,
	}

	_ = db.InsertCode(c, code)

	countryName, err := db.GetCountryName(c, "CA")
	assert.NoError(t, err)
	assert.Equal(t, "Canada", countryName)
}

func TestGetByCountryCode(t *testing.T) {
	c := context.Background()

	code := models.SwiftCode{
		Code:        "DEUTDEFFXXX",
		Name:        "Deutsche Bank",
		Address:     "Berlin",
		CountryISO2: "DE",
		CountryName: "Germany",
		Headquarter: true,
	}

	_ = db.InsertCode(c, code)

	results, err := db.GetByCountryCode(c, "DE")
	assert.NoError(t, err)
	assert.NotEmpty(t, results)
	assert.Equal(t, "DE", results[0].CountryISO2)
}

func TestDeleteByCode(t *testing.T) {
	c := context.Background()

	code := models.SwiftCode{
		Code:        "DELETE01XXX",
		Name:        "To Be Deleted Bank",
		Address:     "Delete St",
		CountryISO2: "XX",
		CountryName: "Nowhere",
		Headquarter: true,
	}

	_ = db.InsertCode(c, code)

	affected, err := db.DeleteByCode(c, code.Code)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), affected)

	_, err = db.GetByCode(c, code.Code)
	assert.Error(t, err)
}

func TestMain(m *testing.M) {
	c := context.Background()

	pgContainer, err := postgres.Run(
		c,
		"postgres:17",
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
			wait.ForListeningPort("5432/tcp"),
		),
	)
	if err != nil {
		log.Fatalf("Failed to start container: %v\n", err)
	}
	defer func() {
		if err := testcontainers.TerminateContainer(pgContainer); err != nil {
			log.Printf("Failed to terminate container: %v\n", err)
		}
	}()

	connStr, err := pgContainer.ConnectionString(c)
	if err != nil {
		log.Fatalf("Failed to get connection string: %v\n", err)
	}

	db = ConnectWithConnString(c, connStr)
	defer db.Close()

	m.Run()
}
