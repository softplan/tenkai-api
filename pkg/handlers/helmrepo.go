package handlers

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"github.com/softplan/tenkai-api/pkg/constraints"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/softplan/tenkai-api/pkg/global"
	"github.com/softplan/tenkai-api/pkg/util"
	"net/http"
)

func (appContext *AppContext) repoUpdate(w http.ResponseWriter, r *http.Request) {

	appContext.HelmServiceAPI.RepoUpdate()
}

func (appContext *AppContext) listRepositories(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	result := &model.RepositoryResult{}

	repositories, err := appContext.HelmServiceAPI.GetRepositories()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result.Repositories = repositories
	data, _ := json.Marshal(result)
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}

func (appContext *AppContext) newRepository(w http.ResponseWriter, r *http.Request) {

	principal := util.GetPrincipal(r)
	if !util.Contains(principal.Roles, constraints.TenkaiAdmin) {
		http.Error(w, errors.New(global.AccessDenied).Error(), http.StatusUnauthorized)
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var payload model.Repository

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := appContext.HelmServiceAPI.AddRepository(payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)

}

func (appContext *AppContext) setDefaultRepo(w http.ResponseWriter, r *http.Request) {

	principal := util.GetPrincipal(r)

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var payload model.DefaultRepoRequest

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	configMap := model.ConfigMap{Name: "DEFAULT_REPO_" + principal.Email, Value: payload.Reponame}

	if _, err := appContext.Repositories.ConfigDAO.CreateOrUpdateConfig(configMap); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)

}

func (appContext *AppContext) getDefaultRepo(w http.ResponseWriter, r *http.Request) {
	principal := util.GetPrincipal(r)

	var config model.ConfigMap
	var err error
	if config, err = appContext.Repositories.ConfigDAO.GetConfigByName("DEFAULT_REPO_" + principal.Email); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data, _ := json.Marshal(config)
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}

func (appContext *AppContext) deleteRepository(w http.ResponseWriter, r *http.Request) {

	principal := util.GetPrincipal(r)
	if !util.Contains(principal.Roles, constraints.TenkaiAdmin) {
		http.Error(w, errors.New(global.AccessDenied).Error(), http.StatusUnauthorized)
	}

	vars := mux.Vars(r)
	name := vars["name"]
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if err := appContext.HelmServiceAPI.RemoveRepository(name); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
