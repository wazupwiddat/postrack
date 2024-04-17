package schwab

import (
	"gorm.io/gorm"
)

func Create(db *gorm.DB, a *SchwabAccess) (uint, error) {
	err := db.Create(a).Error
	if err != nil {
		return 0, db.Error
	}
	return a.ID, nil
}
