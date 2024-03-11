package controllers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/wazupwiddat/postrack/server/transaction/summary"
)

func (c Controller) HandleSummary(w http.ResponseWriter, r *http.Request) {
	u, err := userFromRequestContext(r, c.db)
	if err != nil {
		log.Println(err)
		http.Error(w, "Unable to find user", http.StatusUnauthorized)
		return
	}

	response, err := summary.Summary(c.db, &summary.Request{User: u})
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(response)
}
