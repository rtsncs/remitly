package database

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rtsncs/remitly-swift-api/models"
)

type Database struct {
	pool *pgxpool.Pool
}

func Connect(c context.Context) (Database, error) {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		return Database{}, fmt.Errorf("DATABASE_URL is not set")
	}
	return ConnectWithConnString(c, connStr)
}

func ConnectWithConnString(c context.Context, connStr string) (Database, error) {
	pool, err := pgxpool.New(c, connStr)
	if err != nil {
		return Database{}, fmt.Errorf("Unable to create database connection pool: %w", err)
	}
	if err = pool.Ping(c); err != nil {
		return Database{}, fmt.Errorf("Failed to ping database: %w", err)
	}

	db := Database{pool}
	if err = db.createTable(c); err != nil {
		return Database{}, fmt.Errorf("Failed to create table: %w", err)
	}

	return db, nil
}

func (db *Database) Close() {
	db.pool.Close()
}

func (db *Database) createTable(c context.Context) error {
	sql := `
	CREATE TABLE IF NOT EXISTS swift_codes (
		id SERIAL PRIMARY KEY,
		swift_code VARCHAR(11) UNIQUE NOT NULL,
		bank_name TEXT NOT NULL,
		address TEXT,
		country_iso2 CHAR(2) NOT NULL,
		country_name TEXT NOT NULL,
		is_headquarter BOOLEAN NOT NULL
	);
	`
	_, err := db.pool.Exec(c, sql)
	return err
}

func (db *Database) InsertCode(c context.Context, code models.SwiftCode) error {
	sql := `
	INSERT INTO swift_codes (
		swift_code,
		bank_name,
		address,
		country_iso2,
		country_name,
		is_headquarter
	) VALUES (
		$1, $2, $3, $4, $5, $6
	);
	`
	_, err := db.pool.Exec(c, sql, code.SwiftCode, code.BankName, code.Address, code.CountryISO2, code.CountryName, code.IsHeadquarter)
	return err
}

func (db *Database) GetByCode(c context.Context, code string) (models.SwiftCode, error) {
	sql := `
	SELECT
		swift_code,
		bank_name,
		address,
		country_iso2,
		country_name,
		is_headquarter
	FROM swift_codes
	WHERE swift_code = $1;
	`
	rows, err := db.pool.Query(c, sql, code)
	if err != nil {
		return models.SwiftCode{}, err
	}

	return pgx.CollectOneRow(rows, pgx.RowToStructByName[models.SwiftCode])
}

func (db *Database) GetBranches(c context.Context, headquaterCode string) ([]models.SwiftCode, error) {
	sql := `
	SELECT
		swift_code,
		bank_name,
		address,
		country_iso2,
		is_headquarter
	FROM swift_codes
	WHERE LEFT(swift_code, 8) = $1 AND NOT swift_code LIKE '%XXX';
	`
	rows, err := db.pool.Query(c, sql, headquaterCode[:8])
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, pgx.RowToStructByNameLax[models.SwiftCode])
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

func (db *Database) GetByCountryCode(c context.Context, countryCode string) ([]models.SwiftCode, error) {
	sql := `
	SELECT
		swift_code,
		bank_name,
		address,
		country_iso2,
		is_headquarter
	FROM swift_codes
	WHERE country_iso2 = $1;
	`
	rows, err := db.pool.Query(c, sql, countryCode)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, pgx.RowToStructByNameLax[models.SwiftCode])
}

func (db *Database) DeleteByCode(c context.Context, code string) (int64, error) {
	sql := `DELETE FROM swift_codes WHERE swift_code = $1;`
	tag, err := db.pool.Exec(c, sql, code)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
