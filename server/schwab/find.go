package schwab

import (
	"gorm.io/gorm"
)

func FindAllByUser(db *gorm.DB, userId uint) ([]SchwabAccess, int64, error) {
	var tokens []SchwabAccess
	res := db.Find(&tokens, &SchwabAccess{UserID: userId})
	if res.Error != nil {
		return nil, 0, res.Error
	}
	return tokens, res.RowsAffected, nil
}
