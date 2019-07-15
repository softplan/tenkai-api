package main

import (
	"encoding/json"
	"github.com/softplan/tenkai-api/util"
	"net/http"

	"github.com/softplan/tenkai-api/dbms/model"
)

func (appContext *appContext) saveVariableValues(w http.ResponseWriter, r *http.Request) {

	var payload model.VariableData

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		w.WriteHeader(422)
		if err := json.NewEncoder(w).Encode(err); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	for _, item := range payload.Data {
		if err := appContext.database.CreateVariable(item); err != nil {
			if err := json.NewEncoder(w).Encode(err); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}
	}
	w.WriteHeader(http.StatusCreated)
}

func (appContext *appContext) getVariablesByEnvironmentAndScope(w http.ResponseWriter, r *http.Request) {

	type Payload struct {
		EnvironmentID int    `json:"environmentId"`
		Scope         string `json:"scope"`
	}

	var payload Payload

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		w.WriteHeader(422)
		if err := json.NewEncoder(w).Encode(err); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	variableResult := &model.VariablesResult{}

	var err error
	if variableResult.Variables, err = appContext.database.GetAllVariablesByEnvironmentAndScope(payload.EnvironmentID, payload.Scope); err != nil {
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
