package model

import "github.com/jinzhu/gorm"

type User struct {
	ID        string `json:"id"`
	ChannelID string `json:"channel_id"`
	Mobile    string `json:"mobile"`
}

func (db *DB) CreateUser(u User) error {

	err := db.Create(&u).Error
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) FindUser(id string) (*User, error) {

	var usr User
	err := db.Where("id = ?", id).Find(&usr).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	return &usr, nil
}
