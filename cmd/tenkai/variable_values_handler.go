package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"github.com/softplan/tenkai-api/audit"
	"github.com/softplan/tenkai-api/dbms/model"
	"github.com/softplan/tenkai-api/global"
	helmapi "github.com/softplan/tenkai-api/service/helm"
	"github.com/softplan/tenkai-api/util"
	"net/http"
	"strconv"
	"strings"
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

		targetEnvironment, err := appContext.environmentDAO.GetByID(int(item.EnvironmentID))
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

func (appContext *appContext) getVariablesNotUsed(w http.ResponseWriter, r *http.Request) {

	type responseResult struct {
		ID    int    `json:"id"`
		Scope string `json:"scope"`
		Name  string `json:"name"`
		Value string `json:"value"`
	}

	vars := mux.Vars(r)

	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	//Retrieve all variables
	variableResult := &model.VariablesResult{}
	if variableResult.Variables, err = appContext.database.GetAllVariablesByEnvironment(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//Retrieve all helm release in environment
	//Locate Environment
	environment, err := appContext.environmentDAO.GetByID(id)
	if err != nil {
		http.Error(w, err.Error(), 501)
		return
	}
	kubeConfig := global.KubeConfigBasePath + environment.Group + "_" + environment.Name
	helmReleases, err := appContext.helmServiceAPI.ListHelmDeployments(kubeConfig, environment.Namespace)

	result := make([]responseResult, 0)
	for _, e := range variableResult.Variables {
		if e.Scope != "global" {
			if !scopeRunning(helmReleases, e.Scope, environment.Namespace) {
				result = append(result, responseResult{ID: int(e.ID), Scope: e.Scope, Name: e.Name, Value: e.Value})
			}
		}
	}

	data, _ := json.Marshal(result)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}

func scopeRunning(helmList *helmapi.HelmListResult, scope string, namespace string) bool {

	result := false

	if strings.Contains(scope, "gcm") {
		scope = scope + "-" + namespace
	} else {
		scope = strings.ReplaceAll(scope, "saj6/", "")
	}

	for _, e := range helmList.Releases {

		searchable := e.Chart
		if strings.Contains(scope, "gcm") {
			searchable = e.Name
		}

		if strings.Contains(searchable, scope) {
			result = true
			break
		}
	}
	return result
}
