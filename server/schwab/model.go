package schwab

import "gorm.io/gorm"

type SchwabAccess struct {
	gorm.Model
	ID           uint   `gorm:"primary_key" json:"-"`
	UserID       uint   `gorm:"index"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `gorm:"size:10" json:"token_type"`
	Scope        string `gorm:"size:30" json:"scope"`
	RefreshToken string `gorm:"size:100" json:"refresh_token"` // valid for 7 days
	AccessToken  string `gorm:"size:100" json:"access_token"`  // valid for 30 minutes
	IDToken      string `gorm:"size:400" json:"id_token"`      // JWT
}
