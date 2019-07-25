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

func (appContext *appContext) newSolution(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var payload model.Solution

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		w.WriteHeader(422)
		if err := json.NewEncoder(w).Encode(err); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if _, err := appContext.database.CreateSolution(payload); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

}

func (appContext *appContext) editSolution(w http.ResponseWriter, r *http.Request) {

	var payload model.Solution

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
		return
	}

	if err := appContext.database.EditSolution(payload); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
		return
	}

	w.WriteHeader(http.StatusOK)
}


func (appContext *appContext) deleteSolution(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	sl := vars["id"]
	id, _ := strconv.Atoi(sl)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if err := appContext.database.DeleteSolution(id); err != nil {
		log.Println("Error deleting environment: ", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)

}

func (appContext *appContext) listSolution(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	result := &model.SolutionResult{}
	var err error
	if result.List, err = appContext.database.ListSolutions(); err != nil {
		if err := json.NewEncoder(w).Encode(err); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	data, _ := json.Marshal(result)
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}

