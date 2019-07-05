package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/softplan/tenkai-api/dbms/model"
	"github.com/softplan/tenkai-api/global"
	helmapi "github.com/softplan/tenkai-api/service/helm"
	"github.com/softplan/tenkai-api/util"
	"log"
	"net/http"
)

func (appContext *appContext) listRepositories(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	result := &model.RepositoryResult{}

	repositories, error := helmapi.GetRepositories()
	if error == nil {
		result.Repositories = repositories
		data, _ := json.Marshal(result)
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}

}

func (appContext *appContext) newRepository(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var payload model.Repository

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

	if error := helmapi.AddRepository(payload); error != nil {

		logFields := global.AppFields{global.FUNCTION: "newRepository"}
		global.Logger.Error(logFields, "Error creating repository:" + error.Error())


		w.WriteHeader(512)
		if err := json.NewEncoder(w).Encode(error); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusOK)

}

func (appContext *appContext) deleteRepository(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	name:= vars["name"]

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if error :=  helmapi.RemoveRepository(name); error != nil {
		w.WriteHeader(512)
		if err := json.NewEncoder(w).Encode(error); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)

}