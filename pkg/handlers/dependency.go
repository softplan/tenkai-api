package handlers

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	service_tenkai "github.com/softplan/tenkai-api/pkg/service/tenkai"
	"github.com/softplan/tenkai-api/pkg/util"
	"log"
	"net/http"
	"strconv"
)

func (appContext *AppContext) newDependency(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var payload model.Dependency

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	if err := appContext.Repositories.DependencyDAO.CreateDependency(payload); err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	w.WriteHeader(http.StatusCreated)

}

func (appContext *AppContext) deleteDependency(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	sl := vars["id"]
	id, _ := strconv.Atoi(sl)

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	if err := appContext.Repositories.DependencyDAO.DeleteDependency(id); err != nil {
		log.Println("Error deleting environment: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

}

func (appContext *AppContext) listDependencies(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	releaseID, ok := r.URL.Query()["releaseId"]
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	id, _ := strconv.Atoi(releaseID[0])
	dependencyResult := &model.DependencyResult{}
	var err error

	if dependencyResult.Dependencies, err = appContext.Repositories.DependencyDAO.ListDependencies(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data, _ := json.Marshal(dependencyResult)
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}

func (appContext *AppContext) analyse(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var payload model.DepAnalyseRequest
	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var analyse model.DepAnalyse

	err := service_tenkai.Analyse(appContext.Repositories.EnvironmentDAO, appContext.HelmServiceAPI, appContext.Repositories.DependencyDAO, payload, &analyse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if analyse.Links == nil {
		analyse.Links = make([]model.DepLink, 0)
	}

	data, _ := json.Marshal(analyse)
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}
