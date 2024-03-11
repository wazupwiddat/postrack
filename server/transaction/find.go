package transaction

import (
	"github.com/wazupwiddat/postrack/server/user"
	"gorm.io/gorm"
)

func FindAllByUser(db *gorm.DB, u *user.User) ([]Transaction, error) {
	var transactions []Transaction
	res := db.Find(&transactions, &Transaction{UserID: u.ID})
	if res.Error != nil {
		return nil, res.Error
	}
	return transactions, nil
}
