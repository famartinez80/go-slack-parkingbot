package models

import "github.com/jinzhu/gorm"

type Spaces struct {
	ID          uint    `json:"id"`
	NumberSpace string  `json:"numbers_pace"`
	Available   int     `json:"available"`
	IdUser      *string `json:"id_user"`
}

func (db *DB) AllSpaces() ([]*Spaces, error) {

	var spc []*Spaces
	err := db.Where("available = ?", 1).Find(&spc).Error
	if err != nil {
		return nil, err
	}
	return spc, nil
}

func (db *DB) UpdateSpace(space *Spaces) error {

	err := db.Save(&space).Error
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) SpaceByUser(idUser string) (*Spaces, error) {

	var spc Spaces
	err := db.Where("id_user = ?", idUser).Find(&spc).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	return &spc, nil
}
