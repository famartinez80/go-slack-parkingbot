package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

type DataStore interface {
	AllSpaces() ([]*Spaces, error)
	UpdateSpace(*Spaces) error
	SpaceByUser(idUser string) (*Spaces, error)
}

type DB struct {
	*gorm.DB
}

func InitDB(dataSourceName string) (*DB, error) {
	db, err := gorm.Open("mysql", dataSourceName)
	if err != nil {
		return nil, err
	}
	//
	//defer func() {
	//	err := db.Close()
	//	if err != nil {
	//		log.Printf("[ERROR] %s", err)
	//	}
	//}()

	return &DB{db}, nil
}
