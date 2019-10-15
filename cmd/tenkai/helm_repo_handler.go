package main

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"github.com/softplan/tenkai-api/dbms/model"
	"github.com/softplan/tenkai-api/util"
	"net/http"
)

func (appContext *appContext) repoUpdate(w http.ResponseWriter, r *http.Request) {

	appContext.helmServiceAPI.RepoUpdate()
}

func (appContext *appContext) listRepositories(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	result := &model.RepositoryResult{}

	repositories, err := appContext.helmServiceAPI.GetRepositories()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result.Repositories = repositories
	data, _ := json.Marshal(result)
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}

func (appContext *appContext) newRepository(w http.ResponseWriter, r *http.Request) {

	principal := util.GetPrincipal(r)
	if !contains(principal.Roles, TenkaiAdmin) {
		http.Error(w, errors.New("Acccess Defined").Error(), http.StatusUnauthorized)
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var payload model.Repository

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := appContext.helmServiceAPI.AddRepository(payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)

}

func (appContext *appContext) setDefaultRepo(w http.ResponseWriter, r *http.Request) {

	principal := util.GetPrincipal(r)

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var payload model.DefaultRepoRequest

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	configMap := model.ConfigMap{Name: "DEFAULT_REPO_" + principal.Email, Value: payload.Reponame}

	if _, err := appContext.repositories.configDAO.CreateOrUpdateConfig(configMap); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)

}

func (appContext *appContext) getDefaultRepo(w http.ResponseWriter, r *http.Request) {
	principal := util.GetPrincipal(r)

	var config model.ConfigMap
	var err error
	if config, err = appContext.repositories.configDAO.GetConfigByName("DEFAULT_REPO_" + principal.Email); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data, _ := json.Marshal(config)
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}

func (appContext *appContext) deleteRepository(w http.ResponseWriter, r *http.Request) {

	principal := util.GetPrincipal(r)
	if !contains(principal.Roles, TenkaiAdmin) {
		http.Error(w, errors.New("Acccess Defined").Error(), http.StatusUnauthorized)
	}

	vars := mux.Vars(r)
	name := vars["name"]
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if err := appContext.helmServiceAPI.RemoveRepository(name); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
