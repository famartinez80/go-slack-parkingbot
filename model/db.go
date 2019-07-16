package model

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

type DataStore interface {
	AllSpaces() ([]*Space, error)
	UpdateSpace(*Space) error
	SpaceByUser(string) (*Space, error)

	CreateUser(User) error
	FindUser(string) (*User, error)

	CreateVehicle(Vehicle) error
	FindVehicle(string) (*Vehicle, error)
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
