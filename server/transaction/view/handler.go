package view

import (
	"github.com/wazupwiddat/postrack/server/transaction"
	"github.com/wazupwiddat/postrack/server/user"
	"gorm.io/gorm"
)

type Request struct {
	User     *user.User
	Page     int
	PageSize int
}

type Response struct {
	Transactions []transaction.Transaction
}

func View(db *gorm.DB, req *Request) (*Response, error) {
	trans, err := transaction.FindAllByUserPaginated(db, req.User, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	return &Response{
		Transactions: trans,
	}, nil
}
