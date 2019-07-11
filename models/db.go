package models

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

type DataStore interface {
	AllSpaces() ([]*Spaces, error)
	UpdateSpace(available bool, idSpace int) error
}

type DB struct {
	*sql.DB
}

func InitDB(dataSourceName string) (*DB, error) {
	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		return nil, err
	}

	return &DB{db}, nil
}
