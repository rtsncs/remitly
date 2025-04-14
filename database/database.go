package database

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SwiftCode struct {
	Code        string `json:"swiftCode"`
	Name        string `json:"bankName"`
	Address     string `json:"address"`
	CountryISO2 string `json:"countryISO2"`
	CountryName string `json:"countryName,omitempty"`
	Headquarter bool   `json:"isHeadquarter"`
}

type SwiftCodeWithBranches struct {
	SwiftCode
	Branches []SwiftCode `json:"branches"`
}

type Database struct {
	pool *pgxpool.Pool
}

func Connect(c context.Context) Database {
	pool, err := pgxpool.New(c, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Unable to create database connection pool: %v\n", err)
	}
	if err = pool.Ping(c); err != nil {
		log.Fatalf("Failed to ping database: %v\n", err)
	}
	log.Println("Connected to database")

	db := Database{pool}
	if err = db.createTable(c); err != nil {
		log.Fatalf("Failed to create table: %v\n", err)
	}

	return db
}

func (db *Database) Close() {
	db.pool.Close()
}

func (db *Database) createTable(c context.Context) error {
	sql := `
	CREATE TABLE IF NOT EXISTS swift_codes (
		id SERIAL PRIMARY KEY,
		code VARCHAR(11) UNIQUE NOT NULL,
		name TEXT NOT NULL,
		address TEXT,
		country_iso2 CHAR(2) NOT NULL,
		country_name TEXT NOT NULL,
		headquarter BOOLEAN NOT NULL
	);
	`
	_, err := db.pool.Exec(c, sql)
	return err
}

func (db *Database) InsertCode(c context.Context, code SwiftCode) error {
	sql := `
	INSERT INTO swift_codes (
		code,
		name,
		address,
		country_iso2,
		country_name,
		headquarter
	) VALUES (
		$1, $2, $3, $4, $5, $6
	);
	`
	_, err := db.pool.Exec(c, sql, code.Code, code.Name, code.Address, code.CountryISO2, code.CountryName, code.Headquarter)
	if err != nil {
		return fmt.Errorf("insert failed: %v", err)
	}
	return nil
}

func (db *Database) GetByCode(c context.Context, code string) (SwiftCode, error) {
	sql := `
	SELECT
		code,
		name,
		address,
		country_iso2,
		country_name,
		headquarter
	FROM swift_codes
	WHERE code = $1;
	`
	rows, err := db.pool.Query(c, sql, code)
	if err != nil {
		return SwiftCode{}, err
	}

	return pgx.CollectOneRow(rows, pgx.RowToStructByName[SwiftCode])
}

func (db *Database) GetBranches(c context.Context, headquaterCode string) ([]SwiftCode, error) {
	sql := `
	SELECT
		code,
		name,
		address,
		country_iso2,
		country_name,
		headquarter
	FROM swift_codes
	WHERE LEFT(code, 8) = $1 AND NOT code LIKE '%XXX';
	`
	rows, err := db.pool.Query(c, sql, headquaterCode[:8])
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, pgx.RowToStructByName[SwiftCode])
}

func (db *Database) GetCountryName(c context.Context, countryCode string) (string, error) {
	sql := `
	SELECT country_name
	FROM swift_codes
	WHERE country_iso2 = $1
	LIMIT 1;
	`
	var name string
	err := db.pool.QueryRow(c, sql, countryCode).Scan(&name)
	return name, err
}

func (db *Database) GetByCountryCode(c context.Context, countryCode string) ([]SwiftCode, error) {
	sql := `
	SELECT
		code,
		name,
		address,
		country_iso2,
		headquarter
	FROM swift_codes
	WHERE country_iso2 = $1;
	`
	rows, err := db.pool.Query(c, sql, countryCode)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, pgx.RowToStructByNameLax[SwiftCode])
}
