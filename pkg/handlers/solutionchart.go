package handlers

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/softplan/tenkai-api/pkg/global"
	"github.com/softplan/tenkai-api/pkg/util"
	"net/http"
	"strconv"
)

func (appContext *AppContext) newSolutionChart(w http.ResponseWriter, r *http.Request) {

	w.Header().Set(global.ContentType, global.JsonContentType)

	var payload model.SolutionChart

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := appContext.Repositories.SolutionChartDAO.CreateSolutionChart(payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

}

func (appContext *AppContext) deleteSolutionChart(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	sl := vars["id"]
	id, _ := strconv.Atoi(sl)

	w.Header().Set(global.ContentType, global.JsonContentType)

	if err := appContext.Repositories.SolutionChartDAO.DeleteSolutionChart(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

}

func (appContext *AppContext) listSolutionCharts(w http.ResponseWriter, r *http.Request) {

	w.Header().Set(global.ContentType, global.JsonContentType)

	ids, ok := r.URL.Query()["solutionId"]
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	id, _ := strconv.Atoi(ids[0])
	result := &model.SolutionChartResult{}
	var err error

	if result.List, err = appContext.Repositories.SolutionChartDAO.ListSolutionChart(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data, _ := json.Marshal(result)
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}
