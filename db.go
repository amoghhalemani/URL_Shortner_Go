package main

import (
	"database/sql"
)

// intializing the database connection
func InitDB(connStr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}

// function to check for if table exists in db
func CreateTable(DB *sql.DB) error {
	_, err := DB.Exec("CREATE TABLE IF NOT EXISTS urls (short_url TEXT, original_url TEXT, expires_at TIMESTAMP)")
	return err
}

// function to create indexes
func CreateIndex(DB *sql.DB) error {
	_, err := DB.Exec("CREATE INDEX IF NOT EXISTS idx_short ON urls(short_url)")
	return err
}
