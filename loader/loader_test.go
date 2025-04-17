package loader

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/rtsncs/remitly-swift-api/database"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/xuri/excelize/v2"
)

var db database.Database

func TestLoadFromFile(t *testing.T) {
	c := context.Background()

	rows := [][]string{
		{"COUNTRY ISO2 CODE", "SWIFT CODE", "CODE TYPE", "NAME", "ADDRESS", "TOWN NAME", "COUNTRY NAME", "TIME ZONE"},
		{"AL", "AAISALTRXXX", "BIC11", "UNITED BANK OF ALBANIA SH.A", "HYRJA 3 RR. DRITAN HOXHA ND. 11 TIRANA, TIRANA, 1023", "TIRANA", "ALBANIA", "Europe/Tirane"},
		{"US", "BANKUS00XXX", "", "Test Bank HQ", "123 Wall St", "", "United States"},
		{"US", "BANKUS00NYC", "", "Test Bank NYC", "456 Broadway", "", "United States"},
		{"US", "BANKUS00NYC", "", "Test Bank NYC", "456 Broadway", "", "United States"},
		{"PL", "AAISALTR123", "BIC11", "UNITED BANK OF ALBANIA SH.A", "HYRJA 3 RR. DRITAN HOXHA ND. 11 TIRANA, TIRANA, 1023", "TIRANA", "ALBANIA", "Europe/Tirane"},
		{"US", "BADROW", ""},
	}

	file := excelize.NewFile()
	sheet := file.GetSheetName(0)

	for i, row := range rows {
		cell := fmt.Sprintf("A%d", i+1)
		assert.NoError(t, file.SetSheetRow(sheet, cell, &row))
	}

	tmpFile, err := os.CreateTemp("", "swiftcodes-*.xlsx")
	assert.NoError(t, err, "failed to create temp file")
	t.Cleanup(func() { os.Remove(tmpFile.Name()) })

	assert.NoError(t, file.SaveAs(tmpFile.Name()), "failed to save temp file")

	var logBuf bytes.Buffer
	originalOutput := log.Writer()
	log.SetOutput(&logBuf)
	t.Cleanup(func() { log.SetOutput(originalOutput) })

	LoadFromFileWithDatabase(tmpFile.Name(), db)

	logs := logBuf.String()
	assert.Contains(t, logs, fmt.Sprintf("Total rows: %d; Inserted %d; Failed: %d", len(rows)-1, 3, 3))

	bank, err := db.GetByCode(c, "AAISALTRXXX")
	assert.NoError(t, err)
	assert.Equal(t, "UNITED BANK OF ALBANIA SH.A", bank.Name)

	hq, err := db.GetByCode(c, "BANKUS00XXX")
	assert.NoError(t, err)
	assert.True(t, hq.Headquarter)

	branches, err := db.GetBranches(c, "BANKUS00XXX")
	assert.NoError(t, err)
	assert.Len(t, branches, 1)
	assert.Equal(t, "BANKUS00NYC", branches[0].Code)
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

	db, err = database.ConnectWithConnString(c, connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v\n", err)
	}
	defer db.Close()

	m.Run()
}
