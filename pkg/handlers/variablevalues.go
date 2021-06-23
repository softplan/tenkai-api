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

	logFields := global.AppFields{global.Function: "saveVariableValues"}

	global.Logger.Info(logFields, "Entering saveVariableValues method!")

	isAdmin := false
	principal := util.GetPrincipal(r)
	if util.Contains(principal.Roles, constraints.TenkaiAdmin) {
		isAdmin = true
	}

	var payload model.VariableData

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		global.Logger.Error(logFields, "Error util.UnmarshalPayload")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//Avoid processing of empty payload
	if len(payload.Data) == 0 {
		global.Logger.Info(logFields, "payload.Data is empty")
		w.WriteHeader(http.StatusCreated)
		return
	}

	firstVar := payload.Data[0]
	targetEnvironment, err := appContext.Repositories.EnvironmentDAO.GetByID(int(firstVar.EnvironmentID))
	if err != nil {
		global.Logger.Error(logFields, "Error appContext.Repositories.EnvironmentDAO.GetByID")
		http.Error(w, "Environment not found", http.StatusBadRequest)
		return
	}

	//If not admin, verify authorization of user for specific environment
	hasSaveVariablesRole := false
	if !isAdmin {
		hasSaveVariablesRole, _ = appContext.hasEnvironmentRole(principal, targetEnvironment.ID, "ACTION_SAVE_VARIABLES")
		if !hasSaveVariablesRole {

			//Allow only save TAG
			auth := payloadHasOnlyTag(payload)
			if !auth {
				global.Logger.Error(logFields, "Error payloadHasOnlyTag(payload)")
				http.Error(w, errors.New(global.AccessDenied).Error(), http.StatusUnauthorized)
				return
			}
		}
	}

	cacheVars := make(map[string]map[string]interface{})

	for _, item := range payload.Data {

		global.Logger.Info(logFields, "Item: "+item.Scope+" => "+item.Name)

		has, err := appContext.hasAccess(principal.Email, int(targetEnvironment.ID))
		if err != nil || !has {
			global.Logger.Error(logFields, "Error appContext.hasAccess")
			http.Error(w, errors.New("Access Denied in environment "+targetEnvironment.Namespace).Error(), http.StatusUnauthorized)
			return
		}

		if err := appContext.loadChartVars(cacheVars, item); err != nil {
			global.Logger.Error(logFields, "Error appContext.loadChartVars")
			http.Error(w, "Helm chart does not exist", http.StatusBadRequest)
			return
		}

		var updated bool
		var auditValues map[string]string
		if auditValues, updated, err = appContext.Repositories.VariableDAO.CreateVariable(item); err != nil {
			global.Logger.Error(logFields, "Error appContext.Repositories.VariableDAO.CreateVariable")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		appContext.audit(updated, auditValues, targetEnvironment, principal, r)
	}

	if hasSaveVariablesRole || isAdmin {
		// Save variables with default values specified in values.yaml
		if err := appContext.saveVariablesWithDefaultValue(cacheVars, firstVar, targetEnvironment, r, principal); err != nil {
			global.Logger.Error(logFields, "Error appContext.saveVariablesWithDefaultValue")
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}

	global.Logger.Info(logFields, "Exiting saveVariableValues method!")

	w.WriteHeader(http.StatusCreated)
}

func payloadHasOnlyTag(payload model.VariableData) bool {
	result := true
	for _, e := range payload.Data {
		if e.Name != "image.tag" {
			result = false
			break
		}
	}
	return result
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
		ScopeVersion  string `json:"scopeVersion"`
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

	if payload.ScopeVersion != "" {
		chartVars, err := appContext.getHelmChartAppVars(payload.Scope, payload.ScopeVersion)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		variableResult.Variables = filterChartVars(chartVars, variableResult.Variables)
	}

	data, _ := json.Marshal(variableResult)

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func filterChartVars(chartVars map[string]interface{}, databaseVars []model.Variable) (list []model.Variable) {
	for _, databaseVar := range databaseVars {
		for chartVarKey := range chartVars {
			if databaseVar.Name == chartVarKey {
				list = append(list, databaseVar)
			}
		}
	}
	return
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

func (appContext *AppContext) listVariablesNew(w http.ResponseWriter, r *http.Request) {
	logFields := global.AppFields{global.Function: "listVariables"}
	global.Logger.Info(logFields, "Request received")

	type Payload struct {
		Repo          string `json:"repo"`
		ChartName     string `json:"chartName"`
		ChartVersion  string `json:"chartVersion"`
		EnvironmentID int    `json:"environmentId"`
	}

	var payload Payload
	payload.Repo = r.URL.Query().Get("repo")
	payload.ChartName = r.URL.Query().Get("chartName")
	payload.ChartVersion = r.URL.Query().Get("chartVersion")

	if payload.Repo == "" || payload.ChartName == "" || payload.ChartVersion == "" {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		global.Logger.Error(logFields, "Invalid payload")
		return
	}
	environmentID, err := strconv.Atoi(r.URL.Query().Get("environmentId"))
	if err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		global.Logger.Error(logFields, "Invalid payload")
		return
	}
	payload.EnvironmentID = environmentID

	chartName := fmt.Sprintf("%s/%s", payload.Repo, payload.ChartName)

	principal := util.GetPrincipal(r)

	if has, err := appContext.hasAccess(principal.Email, payload.EnvironmentID); err != nil || !has {
		global.Logger.Error(logFields, "Error appContext.hasAccess - "+err.Error())
		http.Error(w, global.AccessDenied, http.StatusUnauthorized)
		return
	}

	variableResult := &model.VariablesResult{}

	if variableResult.Variables, err = appContext.Repositories.VariableDAO.GetAllVariablesByEnvironmentAndScope(payload.EnvironmentID, chartName); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	chartVars, err := appContext.getHelmChartAppVars(chartName, payload.ChartVersion)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	chartVarsConverted := convertChartVars(chartVars, payload.ChartName, payload.ChartVersion, payload.EnvironmentID)
	vars := findNewVars(chartVarsConverted, variableResult.Variables)

	data, _ := json.Marshal(vars)

	w.Header().Add(global.ContentType, global.JSONContentType)
	w.Write(data)
}

func convertChartVars(chartVars map[string]interface{}, chartName, chartVersion string, environmentID int) []model.NewVariable {
	list := make([]model.NewVariable, 0)
	for key, value := range chartVars {
		list = append(list, model.NewVariable{
			Scope:         chartName,
			ChartVersion:  chartVersion,
			Name:          key,
			Value:         fmt.Sprintf("%v", value),
			Secret:        false,
			Description:   "",
			EnvironmentID: environmentID,
		})
	}
	return list
}

func findNewVars(chartVars []model.NewVariable, databaseVars []model.Variable) []model.NewVariable {
	list := make([]model.NewVariable, 0)
	for _, chartVar := range chartVars {
		if exists, dbVar := existsInDatabase(chartVar.Name, databaseVars); exists {
			chartVar.Value = dbVar.Value
			chartVar.New = false
		} else {
			chartVar.New = true
		}
		list = append(list, chartVar)
	}
	return list
}

func existsInDatabase(varName string, databaseVars []model.Variable) (bool, model.Variable) {
	if len(databaseVars) == 0 {
		return false, model.Variable{}
	}

	for _, dv := range databaseVars {
		if dv.Name == varName {
			return true, dv
		}
	}
	return false, model.Variable{}
}

func (appContext *AppContext) validateNewVariablesBeforeInstall(w http.ResponseWriter, r *http.Request) {
	logFields := global.AppFields{global.Function: "listVariables"}
	logger := global.Logger

	logger.Info(logFields, "Request received")

	type Payload struct {
		Charts       []model.Chart
		Environments []int
	}

	type Response struct {
		EnvironmentID int `json:"environmentId"`
		Charts        []model.Chart
	}

	var payload Payload
	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, "Invalid payload - "+err.Error(), http.StatusBadRequest)
		logger.Error(logFields, "Error unmarshaling payload - "+err.Error())
		return
	}

	envs := make([]string, 0)
	for _, envID := range payload.Environments {
		envs = append(envs, strconv.Itoa(envID))
		if _, err := appContext.Repositories.EnvironmentDAO.GetByID(envID); err != nil {
			logger.Error(logFields, fmt.Sprintf("Invalida environment id: %d", envID))
			return
		}
	}

	charts := make([]string, 0)
	for _, chart := range payload.Charts {
		fullname := fmt.Sprintf("%s/%s", chart.Repo, chart.Name)
		charts = append(charts, fullname)
	}

	rawVariables, err := appContext.Repositories.VariableDAO.GetAllVariablesByEnvironmentsAndScopes(envs, charts)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		logger.Error(logFields, err.Error())
		return
	}

	variablesDatabase := formatAllDatabaseVariables(rawVariables)
	variablesTemplate := make([]model.VariablesDefault, 0)

	for _, chart := range payload.Charts {
		//get chart vars
		chartFullname := fmt.Sprintf("%s/%s", chart.Repo, chart.Name)
		chartVars, err := appContext.getHelmChartAppVars(chartFullname, chart.Version)
		if err != nil {
			http.Error(w, "Invalid chart - "+chartFullname, http.StatusBadRequest)
			logger.Error(logFields, err.Error())
			return
		}
		variablesTemplate = append(variablesTemplate, model.VariablesDefault{Chart: chartFullname, Variables: chartVars})
	}

	result := compare(variablesTemplate, variablesDatabase, payload.Environments)

	data, _ := json.Marshal(result)

	w.Header().Add(global.ContentType, global.JSONContentType)
	w.Write(data)
}

func formatAllDatabaseVariables(variables []model.Variable) []model.VariablesByChartAndEnvironment {
	list := make([]model.VariablesByChartAndEnvironment, 0)
	for _, variable := range variables {
		environmentID := variable.EnvironmentID
		chartName := variable.Scope
		exists, index := existsInList(list, environmentID, chartName)
		if !exists {
			list = append(list, model.VariablesByChartAndEnvironment{EnvironmentID: environmentID, Chart: chartName})
			index = len(list) - 1
		}
		list[index].Variables = append(list[index].Variables, variable)
	}
	return list
}

func existsInList(list []model.VariablesByChartAndEnvironment, environmentID int, chartName string) (bool, int) {
	for index, item := range list {
		if item.EnvironmentID == environmentID && chartName == item.Chart {
			return true, index
		}
	}
	return false, -1
}

func compare(chartVars []model.VariablesDefault, databaseVars []model.VariablesByChartAndEnvironment, environments []int) []map[string]interface{} {
	returnable := make([]map[string]interface{}, 0)
	for _, env := range environments {
		failed := make([]string, 0)
		for _, chart := range chartVars {
			variables := getVariablesByEnvironmentAndScopeFromList(env, chart.Chart, databaseVars)
			if !validateOneChart(chart, variables) {
				failed = append(failed, chart.Chart)
			}
		}
		if len(failed) > 0 {
			returnable = append(returnable, map[string]interface{}{"environmentId": env, "charts": failed})
		}
	}
	return returnable
}

func getVariablesByEnvironmentAndScopeFromList(envID int, scope string, list []model.VariablesByChartAndEnvironment) []model.Variable {
	for _, item := range list {
		if item.Chart == scope && item.EnvironmentID == envID {
			return item.Variables
		}
	}
	return []model.Variable{}
}

func validateOneChart(chart model.VariablesDefault, dbVars []model.Variable) bool {
	for chartVarName := range chart.Variables {
		if !(chartVarName == "dateHour" || chartVarName == "version") {
			exists, _ := existsInDatabase(chartVarName, dbVars)
			if !exists {
				return false
			}
		}
	}
	return true
}
