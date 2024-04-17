package access

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wazupwiddat/postrack/server/config"
	"github.com/wazupwiddat/postrack/server/schwab"
	schwabClient "github.com/wazupwiddat/schwab-api/client"
	"github.com/wazupwiddat/schwab-api/models"
	"gorm.io/gorm"
)

func RefreshAccessToken(db *gorm.DB, cfg *config.Config, userID uint) {
	exitCh := make(chan os.Signal, 1)
	signal.Notify(exitCh, os.Interrupt, syscall.SIGINT)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-exitCh
		log.Println("Cancel context: Cancel")
		cancel()
	}()

	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	for {
		if err := updateAccessTokenIfNeeded(db, cfg, userID); err != nil {
			fmt.Printf("failed refresh access token: %s", err)
		}
		select {
		case <-ctx.Done():
			log.Println("Cancel context: Done")
			os.Exit(0)
			return
		case <-ticker.C:
			continue
		}
	}
}

func updateAccessTokenIfNeeded(db *gorm.DB,
	cfg *config.Config, userID uint) error {
	tokens, rows, err := schwab.FindAllByUser(db, userID)
	if err != nil {
		log.Println(err)
		return err
	}

	if rows == 0 {
		log.Println("Something went wrong saving the Access Token;  Refresh Token will exit.")
		return err
	}

	refreshToken := tokens[0]
	if rows > 1 {
		log.Println("Somehow we have more than 1 Access Token;  Using the latest.")
		refreshToken = tokens[len(tokens)-1]
	}

	// check to see if token is about to expire
	expires := refreshToken.UpdatedAt.Add(time.Duration(refreshToken.ExpiresIn-90) * time.Second)
	if time.Now().After(expires) {
		// Refresh needed
		log.Println("Attempting to Refresh Access Token for User:", refreshToken.UserID)

		requestToken := models.SchwabAccess{}
		requestToken.AccessToken = refreshToken.AccessToken
		requestToken.ExpiresIn = refreshToken.ExpiresIn
		requestToken.RefreshToken = refreshToken.RefreshToken
		requestToken.IDToken = refreshToken.IDToken

		sClient := schwabClient.NewSchwabClient(cfg.Schwab.ClientID, cfg.Schwab.ClientSecret, cfg.Schwab.AuthRedirect)
		err := sClient.RefreshAccessToken(&requestToken)
		if err != nil {
			log.Println(err)
			return err
		}

		refreshToken.AccessToken = requestToken.AccessToken
		refreshToken.ExpiresIn = requestToken.ExpiresIn
		refreshToken.RefreshToken = requestToken.RefreshToken
		refreshToken.IDToken = requestToken.IDToken

		_, err = schwab.Update(db, &refreshToken)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}
