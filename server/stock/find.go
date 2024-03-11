package stock

import (
	"errors"

	"github.com/wazupwiddat/postrack/server/user"
	"gorm.io/gorm"
)

type SymbolNotExistsError struct{}

func (*SymbolNotExistsError) Error() string {
	return "symbol not exists"
}

type StockIDDoesNotExistError struct{}

func (*StockIDDoesNotExistError) Error() string {
	return "stock by id does not exist"
}

func FindByID(db *gorm.DB, id uint) (*Stock, error) {
	var user Stock
	res := db.Find(&user, &Stock{ID: id})
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, &StockIDDoesNotExistError{}
	}
	return &user, nil
}

func FindBySymbol(db *gorm.DB, symbol string) (*Stock, error) {
	var user Stock
	res := db.Find(&user, &Stock{Symbol: symbol})
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, &SymbolNotExistsError{}
	}
	return &user, nil
}

func FindAllByUser(db *gorm.DB, u *user.User) ([]Stock, error) {
	var stocks []Stock
	res := db.Find(&stocks, &Stock{UserID: u.ID})
	if res.Error != nil {
		return nil, res.Error
	}
	return stocks, nil
}
