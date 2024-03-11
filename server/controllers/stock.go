package controllers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/piquette/finance-go/quote"
	"github.com/wazupwiddat/postrack/server/stock/createnew"
	"github.com/wazupwiddat/postrack/server/stock/remove"
	"github.com/wazupwiddat/postrack/server/stock/simpleview"
)

func (c Controller) HandleStockSimpleView(w http.ResponseWriter, r *http.Request) {
	u, err := userFromRequestContext(r, c.db)
	if err != nil {
		log.Println(err)
		http.Error(w, "Unable to find user", http.StatusUnauthorized)
		return
	}

	response, err := simpleview.SimpleView(c.db, &simpleview.Request{User: u})
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(response)
}

func (c Controller) HandleStockAdd(w http.ResponseWriter, r *http.Request) {
	u, err := userFromRequestContext(r, c.db)
	if err != nil {
		log.Println(err)
		http.Error(w, "Unable to find user", http.StatusUnauthorized)
		return
	}
	params := mux.Vars(r)
	symbol := params["symbol"]
	if symbol == "" {
		http.Error(w, "Symbol must be present to add", http.StatusBadRequest)
		return
	}

	// Hydrate stock name
	name := symbol
	q, err := quote.Get(symbol)
	if err == nil {
		name = q.ShortName
	} else {
		log.Println(err)
	}

	response, err := createnew.CreateNewStock(c.db, &createnew.Request{
		User:   u,
		Symbol: symbol,
		Name:   name,
	})
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(response)
}

func (c Controller) HandleStockRemove(w http.ResponseWriter, r *http.Request) {
	u, err := userFromRequestContext(r, c.db)
	if err != nil {
		log.Println(err)
		http.Error(w, "Unable to find user", http.StatusUnauthorized)
		return
	}

	params := mux.Vars(r)
	symbol := params["symbol"]
	if symbol == "" {
		http.Error(w, "Symbol must be present to remove", http.StatusBadRequest)
		return
	}

	err = remove.RemoveStock(c.db, &remove.Request{User: u, Symbol: symbol})
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
