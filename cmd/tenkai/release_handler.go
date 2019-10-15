package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/softplan/tenkai-api/dbms/model"
	"github.com/softplan/tenkai-api/util"
	"log"
	"net/http"
	"strconv"
)

func (appContext *appContext) newRelease(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var payload model.Release

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := appContext.repositories.releaseDAO.CreateRelease(payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

}

func (appContext *appContext) deleteRelease(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	sl := vars["id"]
	id, _ := strconv.Atoi(sl)

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	if err := appContext.repositories.releaseDAO.DeleteRelease(id); err != nil {
		log.Println("Error deleting environment: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

}

func (appContext *appContext) listReleases(w http.ResponseWriter, r *http.Request) {

	chartName, ok := r.URL.Query()["chartName"]
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	releaseResult := &model.ReleaseResult{}
	var err error
	if releaseResult.Releases, err = appContext.repositories.releaseDAO.ListRelease(chartName[0]); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, _ := json.Marshal(releaseResult)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}
