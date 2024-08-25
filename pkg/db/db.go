package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper"
)

type Dbtools struct {
	db *sql.DB
}

func NewdbConnection() (*sql.DB, error) {
	dbname := viper.GetString("app.db")
	db, err := sql.Open("sqlite3", dbname)
	if err != nil {

	}
	return db, err
}

// can be used to create,insert,update the table
func WriteOnTable(query string, db *sql.DB) {

	statement, err := db.Prepare(query)
	if err != nil {

	}
	_, err = statement.Exec()
	if err != nil {

	}
}
