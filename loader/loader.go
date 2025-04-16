package loader

import (
	"context"
	"log"
	"strings"

	"github.com/rtsncs/remitly-swift-api/database"
	"github.com/rtsncs/remitly-swift-api/models"
	"github.com/xuri/excelize/v2"
)

func LoadFromFile(path string) {
	c := context.Background()
	db := database.Connect(c)
	defer db.Close()

	f, err := excelize.OpenFile(path)
	if err != nil {
		log.Fatalf("Failed to open file: %v\n", err)
	}
	defer f.Close()
	log.Printf("Parsing file: %s\n", path)

	rows, err := f.GetRows("Sheet1")
	if err != nil {
		log.Fatalf("Failed to get rows: %v\n", err)
	}

	inserted := 0
	for i, row := range rows[1:] {
		if len(row) < 7 {
			log.Printf("Invalid row #%d %v: row too short ", i, row)
			continue
		}
		code := models.SwiftCode{
			CountryISO2: strings.ToUpper(row[0]),
			Code:        row[1],
			Name:        row[3],
			Address:     row[4],
			CountryName: strings.ToUpper(row[6]),
			Headquarter: strings.HasSuffix(row[1], "XXX"),
		}
		if err := code.Validate(); err != nil {
			log.Printf("Invalid row #%d %v: %v\n", i+2, row, err)
			continue
		}
		if err := db.InsertCode(c, code); err != nil {
			log.Printf("Failed to insert row #%d %v: %v\n", i, row, err)
		} else {
			inserted++
		}
	}

	log.Printf("Inserted %d out of %d codes\n", inserted, len(rows)-1)
}
