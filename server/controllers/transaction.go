package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/wazupwiddat/postrack/server/transaction/view"
)

func (c Controller) HandleTransactionView(w http.ResponseWriter, r *http.Request) {
	u, err := userFromRequestContext(r, c.db)
	if err != nil {
		log.Println(err)
		http.Error(w, "Unable to find user", http.StatusUnauthorized)
		return
	}

	// Default pagination values
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize <= 0 {
		pageSize = 10 // Default page size
	}

	response, err := view.View(c.db, &view.Request{User: u, Page: page, PageSize: pageSize})
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(response)
}
