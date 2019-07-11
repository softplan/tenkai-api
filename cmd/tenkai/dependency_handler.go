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

func (appContext *appContext) newDependency(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var payload model.Dependency

	body, error := util.GetHttpBody(r)
	if error != nil {
		w.WriteHeader(422)
		if err := json.NewEncoder(w).Encode(error); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		w.WriteHeader(422)
		if err := json.NewEncoder(w).Encode(err); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if err := appContext.database.CreateDependency(payload); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)


}

func (appContext *appContext) deleteDependency(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	sl:= vars["id"]
	id, _ := strconv.Atoi(sl)

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	if err := appContext.database.DeleteDependency(id); err != nil {
		log.Println("Error deleting environment: ", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

}

func (appContext *appContext) listDependencies(w http.ResponseWriter, r *http.Request) {

	releaseId, ok := r.URL.Query()["releaseId"]

	if ok {
		id, _ := strconv.Atoi(releaseId[0])
		dependencyResult := &model.DependencyResult{}
		var err error
		if dependencyResult.Dependencies, err = appContext.database.ListDependencies(id); err == nil {
			data, _ := json.Marshal(dependencyResult)
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
