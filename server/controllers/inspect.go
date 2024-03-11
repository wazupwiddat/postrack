package controllers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/wazupwiddat/postrack/server/transaction/inspect"
)

func (c Controller) HandleInspect(w http.ResponseWriter, r *http.Request) {
	u, err := userFromRequestContext(r, c.db)
	if err != nil {
		log.Println(err)
		http.Error(w, "Unable to find user", http.StatusUnauthorized)
		return
	}

	response, err := inspect.Inspect(c.db, &inspect.InspectRequest{User: u})
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(response)
}

func (c Controller) HandleInspectSymbol(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	symbol := params["symbol"]
	if symbol == "" {
		log.Println("Symbol must be present for inspecting symbol")
		http.Error(w, "Symbol must be present for inspecting symbol", http.StatusUnauthorized)
		return
	}
	u, err := userFromRequestContext(r, c.db)
	if err != nil {
		log.Println(err)
		http.Error(w, "Unable to find user", http.StatusUnauthorized)
		return
	}

	response, err := inspect.InspectSymbol(c.db, &inspect.InspectSymbolRequest{User: u, Symbol: symbol})
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(response)
}
