package controllers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/wazupwiddat/postrack/server/schwab/access"
)

type SchwabAccess struct {
	Code    string
	Session string
}

func (c Controller) HandleSchwabAccess(w http.ResponseWriter, r *http.Request) {
	var req SchwabAccess
	json.NewDecoder(r.Body).Decode(&req)

	// validate the request
	if req.Code == "" || req.Session == "" {
		http.Error(w, "Authorization Code and Session are required", http.StatusBadRequest)
		return
	}

	u, err := userFromRequestContext(r, c.db)
	if err != nil {
		log.Println(err)
		http.Error(w, "Unable to find user", http.StatusUnauthorized)
		return
	}

	access.NewSchwabAccessToken(c.db, c.cfg, &access.Request{
		Code: req.Code,
		User: u,
	})

	w.WriteHeader(http.StatusCreated)
}
