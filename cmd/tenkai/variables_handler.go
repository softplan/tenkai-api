package main

import (
	"encoding/json"
	"github.com/softplan/tenkai-api/util"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/softplan/tenkai-api/dbms/model"
)

func (appContext *appContext) deleteVariable(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sl := vars["id"]
	id, _ := strconv.Atoi(sl)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if err := appContext.database.DeleteVariable(id); err != nil {
		log.Println("Error deleting variable: ", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (appContext *appContext) editVariable(w http.ResponseWriter, r *http.Request) {

	var payload model.DataVariableElement

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		w.WriteHeader(422)
		if err := json.NewEncoder(w).Encode(err); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if err := appContext.database.EditVariable(payload.Data); err != nil {
		if err := json.NewEncoder(w).Encode(err); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)

}

func (appContext *appContext) addVariables(w http.ResponseWriter, r *http.Request) {

	var payload model.DataVariableElement

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		w.WriteHeader(422)
		if err := json.NewEncoder(w).Encode(err); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if err := appContext.database.EditVariable(payload.Data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

}

func (appContext *appContext) getVariables(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	sl := vars["envId"]
	id, _ := strconv.Atoi(sl)
	variableResult := &model.VariablesResult{}

	var err error
	if variableResult.Variables, err = appContext.database.GetAllVariablesByEnvironment(id); err != nil {
		if err := json.NewEncoder(w).Encode(err); err != nil {
			if err := json.NewEncoder(w).Encode(err); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}
	}

	data, _ := json.Marshal(variableResult)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}
