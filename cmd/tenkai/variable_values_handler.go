package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/softplan/tenkai-api/audit"
	"github.com/softplan/tenkai-api/dbms/model"
	"github.com/softplan/tenkai-api/util"
	"net/http"
)

func (appContext *appContext) saveVariableValues(w http.ResponseWriter, r *http.Request) {

	principal := util.GetPrincipal(r)
	if !contains(principal.Roles, TenkaiVariablesSave) {
		http.Error(w, errors.New("Access Denied").Error(), http.StatusUnauthorized)
		return
	}

	var payload model.VariableData

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, item := range payload.Data {

		targetEnvironment, err := appContext.database.GetByID(int(item.EnvironmentID))
		if err != nil {
			http.Error(w, err.Error(), 501)
			return
		}

		has, err := appContext.hasAccess(principal.Email, int(targetEnvironment.ID))
		if err != nil || !has {
			http.Error(w, errors.New("Access Denied in environment "+targetEnvironment.Namespace).Error(), http.StatusUnauthorized)
			return
		}

		var auditValues map[string]string
		var updated bool
		if auditValues, updated, err = appContext.database.CreateVariable(item); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if updated {
			auditValues["environment"] = targetEnvironment.Name
			audit.DoAudit(r.Context(), appContext.elk, principal.Email, "saveVariable", auditValues)
		}

	}
	w.WriteHeader(http.StatusCreated)
}

func (appContext *appContext) getVariablesByEnvironmentAndScope(w http.ResponseWriter, r *http.Request) {

	principal := util.GetPrincipal(r)

	type Payload struct {
		EnvironmentID int    `json:"environmentId"`
		Scope         string `json:"scope"`
	}

	var payload Payload

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	has, failed := appContext.hasAccess(principal.Email, payload.EnvironmentID)
	if failed != nil || !has {
		http.Error(w, errors.New("Access Denied in this environment").Error(), http.StatusUnauthorized)
		return
	}

	variableResult := &model.VariablesResult{}

	var err error
	if variableResult.Variables, err = appContext.database.GetAllVariablesByEnvironmentAndScope(payload.EnvironmentID, payload.Scope); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for i, e := range variableResult.Variables {
		if e.Secret {
			byteValues, _ := hex.DecodeString(e.Value)
			value, err := util.Decrypt(byteValues, appContext.configuration.App.Passkey)
			if err == nil {
				variableResult.Variables[i].Value = string(value)
			}
		}
	}

	data, _ := json.Marshal(variableResult)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}
