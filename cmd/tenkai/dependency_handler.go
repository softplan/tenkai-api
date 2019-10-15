package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/softplan/tenkai-api/dbms/model"
	service_tenkai "github.com/softplan/tenkai-api/service/tenkai"
	"github.com/softplan/tenkai-api/util"
	"log"
	"net/http"
	"strconv"
)

func (appContext *appContext) newDependency(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var payload model.Dependency

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	if err := appContext.database.CreateDependency(payload); err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	w.WriteHeader(http.StatusCreated)

}

func (appContext *appContext) deleteDependency(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	sl := vars["id"]
	id, _ := strconv.Atoi(sl)

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	if err := appContext.database.DeleteDependency(id); err != nil {
		log.Println("Error deleting environment: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

}

func (appContext *appContext) listDependencies(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	releaseID, ok := r.URL.Query()["releaseId"]
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	id, _ := strconv.Atoi(releaseID[0])
	dependencyResult := &model.DependencyResult{}
	var err error

	if dependencyResult.Dependencies, err = appContext.database.ListDependencies(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data, _ := json.Marshal(dependencyResult)
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}

func (appContext *appContext) analyse(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var payload model.DepAnalyseRequest
	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var analyse model.DepAnalyse

	err := service_tenkai.Analyse(appContext.environmentDAO, appContext.database, payload, &analyse)
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
