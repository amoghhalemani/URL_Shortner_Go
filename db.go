package main

import (
	"database/sql"
)

// intializing the database connection
func InitDB(filepath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", filepath)
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

	_, err := DB.Exec("CREATE TABLE IF NOT EXISTS urls (short TEXT, long TEXT)")
	return err
}
