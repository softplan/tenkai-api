package main

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"github.com/softplan/tenkai-api/dbms/model"
	"github.com/softplan/tenkai-api/util"
	"net/http"
	"strconv"
)

func (appContext *appContext) listDockerRepositories(w http.ResponseWriter, r *http.Request) {

	principal := util.GetPrincipal(r)
	if !contains(principal.Roles, TenkaiAdmin) {
		http.Error(w, errors.New("Acccess Denied").Error(), http.StatusUnauthorized)
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	result := &model.ListDockerRepositoryResponse{}
	var err error
	if result.Repositories, err = appContext.database.ListDockerRepos(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, _ := json.Marshal(result)
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}

func (appContext *appContext) newDockerRepository(w http.ResponseWriter, r *http.Request) {

	principal := util.GetPrincipal(r)
	if !contains(principal.Roles, TenkaiAdmin) {
		http.Error(w, errors.New("Acccess Denied").Error(), http.StatusUnauthorized)
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var payload model.DockerRepo

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, err := appContext.database.CreateDockerRepo(payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

}

func (appContext *appContext) deleteDockerRepository(w http.ResponseWriter, r *http.Request) {

	principal := util.GetPrincipal(r)
	if !contains(principal.Roles, TenkaiAdmin) {
		http.Error(w, errors.New("Acccess Denied").Error(), http.StatusUnauthorized)
	}

	vars := mux.Vars(r)
	sl := vars["id"]
	id, _ := strconv.Atoi(sl)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if err := appContext.database.DeleteDockerRepo(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)

}
