package model

import "github.com/jinzhu/gorm"

type Vehicle struct {
	ID     uint    `json:"id"`
	Type   string  `json:"type"`
	Brand  string  `json:"brand"`
	Color  string  `json:"color"`
	IdUser *string `json:"id_user"`
}

func (db *DB) CreateVehicle(v Vehicle) error {

	err := db.Create(&v).Error
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) FindVehicle(id string) (*Vehicle, error) {

	var veh Vehicle
	err := db.Where("id = ?", id).Find(&veh).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	return &veh, nil
}
