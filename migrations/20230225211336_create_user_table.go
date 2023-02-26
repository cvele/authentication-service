package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upCreateUserTable, downCreateUserTable)
}

func upCreateUserTable(tx *sql.Tx) error {
	_, err := tx.Exec(`CREATE DATABASE IF NOT EXISTS authentication`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY,
		username TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL
	)`)
	return err
}

func downCreateUserTable(tx *sql.Tx) error {
	_, err := tx.Exec("DROP TABLE IF EXISTS users")
	if err != nil {
		return err
	}
	_, err = tx.Exec(`DROP DATABASE IF EXISTS authentication`)
	if err != nil {
		return err
	}
	return nil
}
