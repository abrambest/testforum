package internal

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func CreateDB() error {
	db, err := sql.Open("sqlite3", "data.db")
	if err != nil {
		return err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS product(
		users_id INTEGER PRIMARY KEY AUTOINCREMENT,
		login TEXT NOT NULL,
		email TEXT,
		password TEXT
	);`)
	if err != nil {
		return err
	}
	DB = db
	return nil
}
