package schwab

import "gorm.io/gorm"

func Update(db *gorm.DB, a *SchwabAccess) (uint, error) {
	err := db.Save(a).Error
	if err != nil {
		return 0, db.Error
	}
	return a.ID, nil
}
