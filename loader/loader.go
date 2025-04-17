package loader

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/rtsncs/remitly-swift-api/database"
	"github.com/rtsncs/remitly-swift-api/models"
	"github.com/xuri/excelize/v2"
)

func LoadFromFile(path string) error {
	c := context.Background()
	db, err := database.Connect(c)
	if err != nil {
		return err
	}
	defer db.Close()
	return LoadFromFileWithDatabase(path, db)
}

func LoadFromFileWithDatabase(path string, db database.Database) error {
	c := context.Background()

	f, err := excelize.OpenFile(path)
	if err != nil {
		return fmt.Errorf("Failed to open file: %w", err)
	}
	defer f.Close()
	log.Printf("Parsing file: %s\n", path)

	sheets := f.GetSheetList()

	inserted, total, failed := 0, 0, 0
	for _, sheet := range sheets {
		log.Printf("Parsing sheet: %s\n", sheet)
		rows, err := f.GetRows(sheet)
		if err != nil {
			log.Printf("Failed to get rows: %v\n", err)
			continue
		}
		total += len(rows) - 1

		for i, row := range rows[1:] {
			printIndex := i + 2
			if len(row) < 7 {
				log.Printf("Invalid row #%d %v: row too short\n", printIndex, row)
				failed++
				continue
			}
			code := models.SwiftCode{
				CountryISO2:   strings.ToUpper(row[0]),
				SwiftCode:     row[1],
				BankName:      row[3],
				Address:       row[4],
				CountryName:   strings.ToUpper(row[6]),
				IsHeadquarter: strings.HasSuffix(row[1], "XXX"),
			}
			if err := code.Validate(); err != nil {
				log.Printf("Invalid row #%d %v: %v\n", printIndex, row, err)
				failed++
				continue
			}
			if err := db.InsertCode(c, code); err != nil {
				log.Printf("Failed to insert row #%d %v: %v\n", printIndex, row, err)
				failed++
			} else {
				inserted++
			}
		}
	}

	log.Printf("Total rows: %d; Inserted %d; Failed: %d\n", total, inserted, failed)
	return nil
}
