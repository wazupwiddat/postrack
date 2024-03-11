package createnew

import (
	"github.com/wazupwiddat/postrack/server/stock"
	"github.com/wazupwiddat/postrack/server/user"
	"gorm.io/gorm"
)

type Request struct {
	User   *user.User
	Symbol string
	Name   string
}

type Response struct {
	Stock *stock.Stock
}

func CreateNewStock(db *gorm.DB, req *Request) (*Response, error) {
	s := &stock.Stock{
		Symbol: req.Symbol,
	}
	_, err := stock.Create(db, s)
	if err != nil {
		return nil, err
	}
	return &Response{
		Stock: s,
	}, nil
}
