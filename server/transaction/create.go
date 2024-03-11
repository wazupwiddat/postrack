package transaction

import "gorm.io/gorm"

func Create(db *gorm.DB, t *Transaction) (uint, error) {
	err := db.Create(&t).Error
	if err != nil {
		return 0, err
	}
	return t.ID, nil
}

func CreateMany(db *gorm.DB, trans []Transaction) error {
	return db.Create(trans).Error
}
