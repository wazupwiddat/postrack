package remove

import (
	"github.com/wazupwiddat/postrack/server/stock"
	"github.com/wazupwiddat/postrack/server/user"
	"gorm.io/gorm"
)

type Request struct {
	User   *user.User
	Symbol string
}

func RemoveStock(db *gorm.DB, req *Request) error {
	err := stock.DeleteStock(db, &stock.Stock{
		UserID: req.User.ID,
		Symbol: req.Symbol,
	})
	if err != nil {
		return err
	}
	return nil
}
