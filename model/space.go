package model

import "github.com/jinzhu/gorm"

type Space struct {
	ID          uint    `json:"id"`
	NumberSpace string  `json:"numbers_pace"`
	Available   int     `json:"available"`
	BlockID     int     `json:"block_id"`
	IdUser      *string `json:"id_user"`
}

func (db *DB) AllSpaces() ([]*Space, error) {

	var spc []*Space
	err := db.Where("available = ?", 1).Find(&spc).Error
	if err != nil {
		return nil, err
	}
	return spc, nil
}

func (db *DB) UpdateSpace(s *Space) error {

	err := db.Save(&s).Error
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) SpaceByUser(id string) (*Space, error) {

	var spc Space
	err := db.Where("id_user = ?", id).Find(&spc).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	return &spc, nil
}
