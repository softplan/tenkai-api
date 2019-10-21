package handlers

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/softplan/tenkai-api/pkg/global"
	"github.com/softplan/tenkai-api/pkg/util"
	"log"
	"net/http"
	"strconv"
)

func (appContext *AppContext) newRelease(w http.ResponseWriter, r *http.Request) {

	w.Header().Set(global.ContentType, global.JSONContentType)

	var payload model.Release

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := appContext.Repositories.ReleaseDAO.CreateRelease(payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

}

func (appContext *AppContext) deleteRelease(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	sl := vars["id"]
	id, _ := strconv.Atoi(sl)

	w.Header().Set(global.ContentType, global.JSONContentType)

	if err := appContext.Repositories.ReleaseDAO.DeleteRelease(id); err != nil {
		log.Println("Error deleting environment: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

}

func (appContext *AppContext) listReleases(w http.ResponseWriter, r *http.Request) {

	chartName, ok := r.URL.Query()["chartName"]
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	releaseResult := &model.ReleaseResult{}
	var err error
	if releaseResult.Releases, err = appContext.Repositories.ReleaseDAO.ListRelease(chartName[0]); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, _ := json.Marshal(releaseResult)
	w.Header().Set(global.ContentType, global.JSONContentType)
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}
