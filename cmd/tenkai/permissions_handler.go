package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/softplan/tenkai-api/util"
)

func (appContext *appContext) newEnvironmentPermission(w http.ResponseWriter, r *http.Request) {

	principal := util.GetPrincipal(r)
	if !contains(principal.Roles, TenkaiAdmin) {
		http.Error(w, errors.New("Acccess Defined").Error(), http.StatusUnauthorized)
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["userID"])

	if err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	environmentID, err := strconv.Atoi(vars["environmentId"])
	if err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	if err := appContext.repositories.userDAO.AssociateEnvironmentUser(userID, environmentID); err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	w.WriteHeader(http.StatusCreated)

}
