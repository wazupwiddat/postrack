package simpleview

import (
	"github.com/wazupwiddat/postrack/server/stock"
	"github.com/wazupwiddat/postrack/server/user"
	"gorm.io/gorm"
)

type Request struct {
	User *user.User
}

type Response struct {
	Stocks []stock.Stock
}

func SimpleView(db *gorm.DB, req *Request) (*Response, error) {
	stocks, err := stock.FindAllByUser(db, req.User)
	if err != nil {
		return nil, err
	}
	return &Response{
		Stocks: stocks,
	}, nil
}
