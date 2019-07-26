package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/softplan/tenkai-api/dbms/model"
	"github.com/softplan/tenkai-api/global"
	helmapi "github.com/softplan/tenkai-api/service/helm"
	"github.com/softplan/tenkai-api/util"
	"net/http"
)



func (appContext *appContext) repoUpdate(w http.ResponseWriter, r *http.Request) {

	helmapi.RepoUpdate()
}

func (appContext *appContext) listRepositories(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	result := &model.RepositoryResult{}

	repositories, error := helmapi.GetRepositories()
	if error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(error)
		return
	}

	result.Repositories = repositories
	data, _ := json.Marshal(result)
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}

func (appContext *appContext) newRepository(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var payload model.Repository

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		w.WriteHeader(422)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
		return
	}

	if error := helmapi.AddRepository(payload); error != nil {
		logFields := global.AppFields{global.FUNCTION: "newRepository"}
		global.Logger.Error(logFields, "Error creating repository:"+error.Error())
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(error)
		return
	}
	w.WriteHeader(http.StatusOK)

}

func (appContext *appContext) deleteRepository(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if error := helmapi.RemoveRepository(name); error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(error)
	}
	w.WriteHeader(http.StatusOK)
}
