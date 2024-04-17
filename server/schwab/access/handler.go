package access

import (
	"encoding/json"
	"log"

	"github.com/wazupwiddat/postrack/server/config"
	"github.com/wazupwiddat/postrack/server/schwab"
	"github.com/wazupwiddat/postrack/server/user"
	schwabClient "github.com/wazupwiddat/schwab-api/client"
	"gorm.io/gorm"
)

type Request struct {
	Code string
	User *user.User
}

type Response struct {
	Id uint
}

func NewSchwabAccessToken(db *gorm.DB, cfg *config.Config, req *Request) (*Response, error) {
	// Delete all access tokens for user
	db.Unscoped().Delete(&schwab.SchwabAccess{}, "user_id = ?", req.User.ID)

	// HTTP POST Schwab Client
	sClient := schwabClient.NewSchwabClient(cfg.Schwab.ClientID, cfg.Schwab.ClientSecret, cfg.Schwab.AuthRedirect)
	accessToken, err := sClient.CreateAccessToken(req.Code)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	schwabAccess := &schwab.SchwabAccess{}
	schwabAccess.UserID = req.User.ID
	schwabAccess.AccessToken = accessToken.AccessToken
	schwabAccess.RefreshToken = accessToken.RefreshToken
	schwabAccess.ExpiresIn = accessToken.ExpiresIn
	schwabAccess.Scope = accessToken.Scope
	schwabAccess.TokenType = accessToken.TokenType
	schwabAccess.IDToken = accessToken.IDToken

	// Store access details
	id, err := schwab.Create(db, schwabAccess)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// show accounts and positions
	acct, err := sClient.GetAccounts(schwabAccess.AccessToken, true)
	if err != nil {
		log.Println(err)
	}

	b, _ := json.MarshalIndent(acct, "", "  ")
	log.Println(string(b))

	// Start Refresh Token process
	go RefreshAccessToken(db, cfg, req.User.ID)
	return &Response{Id: id}, err
}
