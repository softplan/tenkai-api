package handlers

import (
	"encoding/json"
	"errors"
	"github.com/softplan/tenkai-api/pkg/constraints"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/softplan/tenkai-api/pkg/util"
)

func (appContext *AppContext) newUser(w http.ResponseWriter, r *http.Request) {

	principal := util.GetPrincipal(r)
	if !util.Contains(principal.Roles, constraints.TenkaiAdmin) {
		http.Error(w, errors.New("Acccess Defined").Error(), http.StatusUnauthorized)
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var payload model.User

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	if err := appContext.Repositories.UserDAO.CreateUser(payload); err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	w.WriteHeader(http.StatusCreated)

}

func (appContext *AppContext) createOrUpdateUser(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var payload model.User

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	if err := appContext.Repositories.UserDAO.CreateOrUpdateUser(payload); err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	w.WriteHeader(http.StatusCreated)

}

func (appContext *AppContext) listUsers(w http.ResponseWriter, r *http.Request) {

	result := &model.UserResult{}
	var err error

	if result.Users, err = appContext.Repositories.UserDAO.ListAllUsers(); err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	data, _ := json.Marshal(result)
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}

func (appContext *AppContext) deleteUser(w http.ResponseWriter, r *http.Request) {

	principal := util.GetPrincipal(r)
	if !util.Contains(principal.Roles, constraints.TenkaiAdmin) {
		http.Error(w, errors.New("Acccess Denied").Error(), http.StatusUnauthorized)
	}

	vars := mux.Vars(r)
	sl := vars["id"]
	id, _ := strconv.Atoi(sl)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if err := appContext.Repositories.UserDAO.DeleteUser(id); err != nil {
		log.Println("Error deleting variable: ", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}