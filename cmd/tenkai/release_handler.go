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

	body, error := util.GetHttpBody(r)
	if error != nil {
		w.WriteHeader(422)
		if err := json.NewEncoder(w).Encode(error); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		log.Fatalln("Error unmarshalling data", err)
		w.WriteHeader(422)
		if err := json.NewEncoder(w).Encode(err); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if err := appContext.database.CreateRelease(payload); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)


}

func (appContext *appContext) deleteRelease(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	sl:= vars["id"]
	id, _ := strconv.Atoi(sl)

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	if err := appContext.database.DeleteRelease(id); err != nil {
		log.Println("Error deleting environment: ", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

}

func (appContext *appContext) listReleases(w http.ResponseWriter, r *http.Request) {
	chartName, ok := r.URL.Query()["chartName"]
	if ok {
		releaseResult := &model.ReleaseResult{}
		var err error
		if releaseResult.Releases, err = appContext.database.ListRelease(chartName[0]); err == nil {
			data, _ := json.Marshal(releaseResult)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(http.StatusOK)
			w.Write(data)
		} else {
			if err := json.NewEncoder(w).Encode(err); err != nil {
				log.Fatalln("Error unmarshalling data", err)
				w.WriteHeader(http.StatusInternalServerError)
			}
		}
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
	return
}
