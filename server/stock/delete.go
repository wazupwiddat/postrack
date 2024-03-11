package stock

import "gorm.io/gorm"

func DeleteStock(db *gorm.DB, condition *Stock) error {
	err := db.Delete(&Stock{}, condition).Error
	if err != nil {
		return err
	}
	return nil
}
