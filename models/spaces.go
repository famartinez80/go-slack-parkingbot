package models

type Spaces struct {
	IdSpace     int
	NumberSpace string
	Available   int
	IdUser      *string
}

func (db *DB) AllSpaces() ([]*Spaces, error) {

	rows, err := db.Query("SELECT * FROM parking.spaces WHERE available = 1;")
	if err != nil {
		return nil, err
	}

	spc := make([]*Spaces, 0)
	for rows.Next() {
		sp := new(Spaces)
		err := rows.Scan(&sp.IdSpace, &sp.NumberSpace, &sp.Available, &sp.IdUser)
		if err != nil {
			return nil, err
		}
		spc = append(spc, sp)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return spc, nil
}

func (db *DB) UpdateSpace(available int, idUser, numberSpace string) error {
	update, err := db.Prepare("UPDATE parking.spaces SET available = ?, idUser = ?  WHERE numberSpace=?")
	if err != nil {
		return err
	}

	_, err = update.Exec(available, idUser, numberSpace)
	if err != nil {
		return err
	}

	return nil
}

//func (db *DB) spaceByUser(idUser string) {
//	rows, err := db.Query("SELECT * FROM parking.spaces WHERE available = 1;")
//	if err != nil {
//		return nil, err
//	}

//	defer rows.Close()
//}
