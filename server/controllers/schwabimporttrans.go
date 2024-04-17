package controllers

import (
	"net/http"
)

func (c Controller) HandleSchwabImportTrans(w http.ResponseWriter, r *http.Request) {
	// u, err := userFromRequestContext(r, c.db)
	// if err != nil {
	// 	log.Println(err)
	// 	http.Error(w, "Unable to find user", http.StatusUnauthorized)
	// 	return
	// }
	w.WriteHeader(http.StatusCreated)
}
