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

func (appContext *appContext) newSolutionChart(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var payload model.SolutionChart

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		w.WriteHeader(422)
		if err := json.NewEncoder(w).Encode(err); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if err := appContext.database.CreateSolutionChart(payload); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

}

func (appContext *appContext) deleteSolutionChart(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	sl := vars["id"]
	id, _ := strconv.Atoi(sl)

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	if err := appContext.database.DeleteSolutionChart(id); err != nil {
		log.Println("Error deleting environment: ", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

}

func (appContext *appContext) listSolutionCharts(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	ids, ok := r.URL.Query()["solutionId"]
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	id, _ := strconv.Atoi(ids[0])
	result := &model.SolutionChartResult{}
	var err error

	if result.List, err = appContext.database.ListSolutionChart(id); err != nil {
		if err := json.NewEncoder(w).Encode(err); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	data, _ := json.Marshal(result)
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}

