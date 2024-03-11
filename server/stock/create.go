package stock

import (
	"gorm.io/gorm"
)

func Create(db *gorm.DB, s *Stock) (uint, error) {
	err := db.Create(s).Error
	if err != nil {
		return 0, err
	}
	return s.ID, nil
}
