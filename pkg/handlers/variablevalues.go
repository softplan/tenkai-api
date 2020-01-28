package handlers

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/softplan/tenkai-api/pkg/constraints"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/softplan/tenkai-api/pkg/global"
	helmapi "github.com/softplan/tenkai-api/pkg/service/_helm"
	"github.com/softplan/tenkai-api/pkg/util"
)

func (appContext *AppContext) saveVariableValues(w http.ResponseWriter, r *http.Request) {

	principal := util.GetPrincipal(r)
	if !util.Contains(principal.Roles, constraints.TenkaiVariablesSave) {
		http.Error(w, errors.New("Access Denied").Error(), http.StatusUnauthorized)
		return
	}

	var payload model.VariableData

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	firstVar := payload.Data[0]
	targetEnvironment, err := appContext.Repositories.EnvironmentDAO.GetByID(int(firstVar.EnvironmentID))
	if err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	cacheVars := make(map[string]map[string]interface{})

	for _, item := range payload.Data {

		has, err := appContext.hasAccess(principal.Email, int(targetEnvironment.ID))
		if err != nil || !has {
			http.Error(w, errors.New("Access Denied in environment "+targetEnvironment.Namespace).Error(), http.StatusUnauthorized)
			return
		}

		if err := appContext.loadChartVars(cacheVars, item); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var updated bool
		var auditValues map[string]string
		if auditValues, updated, err = appContext.Repositories.VariableDAO.CreateVariable(item); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		appContext.audit(updated, auditValues, targetEnvironment, principal, r)
	}

	// Save variables with default values specified in values.yaml
	if err := appContext.saveVariablesWithDefaultValue(cacheVars, firstVar, targetEnvironment, r, principal); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusCreated)
}

func (appContext *AppContext) loadChartVars(cacheVars map[string]map[string]interface{}, item model.Variable) error {
	if cacheVars[item.Scope] == nil {
		appVars, err := appContext.getHelmChartAppVars(item.Scope, item.ChartVersion)
		if err != nil {
			return err
		}
		cacheVars[item.Scope] = appVars
	}
	return nil
}

func (appContext *AppContext) saveVariablesWithDefaultValue(cacheVars map[string]map[string]interface{},
	firstVar model.Variable, targetEnvironment *model.Environment, r *http.Request, principal model.Principal) error {

	for chartName, charts := range cacheVars {
		for varName, varDefaultValue := range charts {
			defaultValue := fmt.Sprintf("%v", varDefaultValue)

			if strings.HasPrefix(defaultValue, "[map") || varName == "dateHour" || varName == "version" {
				continue
			}

			item := model.Variable{
				EnvironmentID: int(firstVar.EnvironmentID),
				Scope:         chartName,
				Name:          varName,
				Value:         defaultValue,
			}

			var err error
			var updated bool
			var auditValues map[string]string
			if auditValues, updated, err = appContext.Repositories.VariableDAO.CreateVariableWithDefaultValue(item); err != nil {
				return err
			}
			appContext.audit(updated, auditValues, targetEnvironment, principal, r)
		}
	}
	return nil
}

func (appContext *AppContext) audit(updated bool, auditValues map[string]string,
	targetEnvironment *model.Environment, principal model.Principal, r *http.Request) {

	if updated {
		auditValues["environment"] = targetEnvironment.Name
		appContext.Auditing.DoAudit(r.Context(), appContext.Elk, principal.Email, "saveVariable", auditValues)
	}
}

func (appContext *AppContext) getHelmChartAppVars(chart string, chartVersion string) (map[string]interface{}, error) {

	if strings.HasSuffix(chart, "-gcm") {
		var config model.ConfigMap
		var err error
		if config, err = appContext.Repositories.ConfigDAO.GetConfigByName("commonValuesConfigMapChart"); err != nil {
			return nil, err
		}
		chart = config.Value
	}

	chartVariables, err := appContext.HelmServiceAPI.GetTemplate(&appContext.Mutex, chart, chartVersion, "values")
	if err != nil {
		return nil, err
	}
	fmt.Println(string(chartVariables))

	var result map[string]interface{}
	json.Unmarshal(chartVariables, &result)
	app := result["app"].(map[string]interface{})

	return app, nil
}

func (appContext *AppContext) getVariablesByEnvironmentAndScope(w http.ResponseWriter, r *http.Request) {

	w.Header().Set(global.ContentType, global.JSONContentType)

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
	if variableResult.Variables, err = appContext.Repositories.VariableDAO.GetAllVariablesByEnvironmentAndScope(payload.EnvironmentID, payload.Scope); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	appContext.decodeSecrets(variableResult)

	data, _ := json.Marshal(variableResult)

	w.WriteHeader(http.StatusOK)
	w.Write(data)

}

func (appContext *AppContext) decodeSecrets(variableResult *model.VariablesResult) {
	for i, e := range variableResult.Variables {
		if e.Secret {
			byteValues, _ := hex.DecodeString(e.Value)
			value, err := util.Decrypt(byteValues, appContext.Configuration.App.Passkey)
			if err == nil {
				variableResult.Variables[i].Value = string(value)
			}
		}
	}
}

func (appContext *AppContext) getVariablesNotUsed(w http.ResponseWriter, r *http.Request) {

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
	if variableResult.Variables, err = appContext.Repositories.VariableDAO.GetAllVariablesByEnvironment(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//Retrieve all helm release in environment
	//Locate Environment
	environment, err := appContext.Repositories.EnvironmentDAO.GetByID(id)
	if err != nil {
		http.Error(w, err.Error(), 501)
		return
	}
	kubeConfig := appContext.ConventionInterface.GetKubeConfigFileName(environment.Group, environment.Name)
	helmReleases, err := appContext.HelmServiceAPI.ListHelmDeployments(kubeConfig, environment.Namespace)

	result := make([]responseResult, 0)
	for _, e := range variableResult.Variables {
		if e.Scope != "global" {
			if !scopeRunning(helmReleases, e.Scope, environment.Namespace) {
				result = append(result, responseResult{ID: int(e.ID), Scope: e.Scope, Name: e.Name, Value: e.Value})
			}
		}
	}

	data, _ := json.Marshal(result)
	w.Header().Set(global.ContentType, global.JSONContentType)
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}

func scopeRunning(helmList *helmapi.HelmListResult, scope string, namespace string) bool {
	result := false
	if strings.Contains(scope, "gcm") {
		scope = scope + "-" + namespace
	} else {
		i := strings.Index(scope, "/")
		if i > 0 {
			beforeBar := scope[:i+1]
			scope = strings.ReplaceAll(scope, beforeBar, "")
		}
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
