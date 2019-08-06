package main

import (
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

func (appContext *appContext) newEnvironmentPermission(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	vars := mux.Vars(r)
	userId, err := strconv.Atoi(vars["userId"])

	if err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	environmentId, err := strconv.Atoi(vars["environmentId"])
	if err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	if err := appContext.database.AssociateEnvironmentUser(userId, environmentId); err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	w.WriteHeader(http.StatusCreated)

}
