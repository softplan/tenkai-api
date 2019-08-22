package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/softplan/tenkai-api/util"
	"net/http"
	"strconv"

	"strings"

	"github.com/gorilla/mux"
	"github.com/softplan/tenkai-api/dbms/model"
	"github.com/softplan/tenkai-api/global"
	"github.com/softplan/tenkai-api/service/helm"
)

func (appContext *appContext) listCharts(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	vars := mux.Vars(r)
	repo := vars["repo"]

	all, ok := r.URL.Query()["all"]
	allVersions := true
	if ok && len(all[0]) > 0 {
		allVersions = all[0] == "true"
	}

	searchTerms := []string{repo}
	searchResult := helmapi.SearchCharts(searchTerms, allVersions)
	result := &model.ChartsResult{Charts: *searchResult}

	data, _ := json.Marshal(result)

	w.WriteHeader(http.StatusOK)
	w.Write(data)

}

func (appContext *appContext) deleteHelmRelease(w http.ResponseWriter, r *http.Request) {

	principal := util.GetPrincipal(r)

	environmentIDs, ok := r.URL.Query()["environmentID"]
	if !ok || len(environmentIDs[0]) < 1 {
		http.Error(w, errors.New("param environmentID is required").Error(), 501)
		return
	}

	releasesName, ok := r.URL.Query()["releaseName"]
	if !ok || len(releasesName[0]) < 1 {
		http.Error(w, errors.New("param releasesName is required").Error(), 501)
		return
	}

	purges, ok := r.URL.Query()["purge"]
	if !ok || len(purges[0]) < 1 {
		http.Error(w, errors.New("param purges, is required").Error(), 501)
		return
	}

	//Locate Environment
	envID, _ := strconv.Atoi(environmentIDs[0])

	has, err := appContext.hasAccess(principal.Email, envID)
	if err != nil || !has {
		http.Error(w, errors.New("Access Denied in this environment").Error(), http.StatusUnauthorized)
		return
	}

	environment, err := appContext.database.GetByID(int(envID))
	if err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	kubeConfig := global.KubeConfigBasePath + environment.Group + "_" + environment.Name

	purge, _ := strconv.ParseBool(purges[0])
	err = helmapi.DeleteHelmRelease(kubeConfig, releasesName[0], purge)
	if err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	w.WriteHeader(http.StatusOK)

}

func (appContext *appContext) revision(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var payload model.GetRevisionRequest

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	//Locate Environment
	environment, err := appContext.database.GetByID(payload.EnvironmentID)
	if err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	kubeConfig := global.KubeConfigBasePath + environment.Group + "_" + environment.Name

	yaml, err := helmapi.Get(kubeConfig, payload.ReleaseName, payload.Revision)

	data, _ := json.Marshal(yaml)
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}

func (appContext *appContext) listReleaseHistory(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var payload model.HistoryRequest

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	//Locate Environment
	environment, err := appContext.database.GetByID(payload.EnvironmentID)
	if err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	kubeConfig := global.KubeConfigBasePath + environment.Group + "_" + environment.Name

	history, err := helmapi.GetHelmReleaseHistory(kubeConfig, payload.ReleaseName)

	data, _ := json.Marshal(history)
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}

func (appContext *appContext) listHelmDeploymentsByEnvironment(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	//Locate Environment
	environment, err := appContext.database.GetByID(id)
	if err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	kubeConfig := global.KubeConfigBasePath + environment.Group + "_" + environment.Name

	result, err := helmapi.ListHelmDeployments(kubeConfig, environment.Namespace)
	if err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	data, _ := json.Marshal(result)
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}

func (appContext *appContext) hasConfigMap(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var payload model.GetChartRequest

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	mutex := appContext.mutex

	result, err := helmapi.GetTemplate(&mutex, payload.ChartName, payload.ChartVersion, "deployment")

	w.WriteHeader(http.StatusOK)
	if err != nil {
		w.Write([]byte("{\"result\":\"false\"}"))
	} else {
		deployment := string(result)
		if strings.Index(deployment, "global-configmap") > 0 {
			w.Write([]byte("{\"result\":\"true\"}"))
		} else {
			w.Write([]byte("{\"result\":\"false\"}"))
		}
	}

}

func (appContext *appContext) getChartVariables(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var payload model.GetChartRequest

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	result, err := helmapi.GetTemplate(&appContext.mutex, payload.ChartName, payload.ChartVersion, "values")

	if err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(result)

}

func (appContext *appContext) multipleInstall(w http.ResponseWriter, r *http.Request) {

	principal := util.GetPrincipal(r)
	if !contains(principal.Roles, TenkaiHelmUpgrade) {
		http.Error(w, errors.New("Access Denied").Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var payload model.MultipleInstallPayload

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	out := &bytes.Buffer{}

	for _, element := range payload.Deployables {
		err := appContext.simpleInstall(element.EnvironmentID, element.Chart, element.Name, out)
		if err != nil {
			http.Error(w, err.Error(), 501)
			return
		}
	}

	fmt.Println(out.String())
	w.WriteHeader(http.StatusOK)

}

func (appContext *appContext) install(w http.ResponseWriter, r *http.Request) {

	principal := util.GetPrincipal(r)
	if !contains(principal.Roles, TenkaiHelmUpgrade) {
		http.Error(w, errors.New("Access Denied").Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var payload model.InstallPayload

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	out := &bytes.Buffer{}

	//TODO Verify if chart exists
	err := appContext.simpleInstall(payload.EnvironmentID, payload.Chart, payload.Name, out)
	if err != nil {
		fmt.Println(out.String())
		http.Error(w, err.Error(), 501)
		return
	}

	w.WriteHeader(http.StatusOK)

}

func (appContext *appContext) simpleInstall(envID int, chart string, name string, out *bytes.Buffer) error {

	//Locate Environment
	environment, err := appContext.database.GetByID(envID)

	//TODO - VERIFY IF CONFIG FILE EXISTS !!! This is the cause of  u.client.ReleaseHistory fail sometimes.

	searchTerm := chart
	if strings.Index(name, "configmap") > -1 {
		searchTerm = name
	}
	variables, err := appContext.database.GetAllVariablesByEnvironmentAndScope(envID, searchTerm)
	globalVariables := appContext.getGlobalVariables(int(environment.ID))

	var args []string
	for _, item := range variables {
		if len(item.Name) > 0 && len(item.Value) > 0 {
			args = append(args, normalizeVariableName(item.Name)+"="+replace(item.Value, *environment, globalVariables))
		}
	}
	//Add Default Gateway
	if len(environment.Gateway) > 0 {
		args = append(args, "istio.virtualservices.gateways[0]="+environment.Gateway)
	}

	if err == nil {
		name := name + "-" + environment.Namespace
		kubeConfig := global.KubeConfigBasePath + environment.Group + "_" + environment.Name
		err := helmapi.Upgrade(kubeConfig, name, chart, environment.Namespace, args, out)
		if err != nil {
			return err
		}
	}

	return nil
}

func replace(value string, environment model.Environment, variables []model.Variable) string {
	newValue := strings.Replace(value, "${NAMESPACE}", environment.Namespace, -1)
	keywords := util.GetReplacebleKeyName(newValue)
	for _, keyword := range keywords {
		for _, element := range variables {
			if element.Name == keyword {
				newValue = strings.Replace(newValue, "${"+element.Name+"}", element.Value, -1)
				break
			}
		}
	}
	return newValue
}

func normalizeVariableName(value string) string {
	if strings.Index(value, "istio.") > -1 || (strings.Index(value, "image.")) > -1 || (strings.Index(value, "service.")) > -1 {
		return value
	}
	return "app." + value
}

func (appContext *appContext) getGlobalVariables(id int) []model.Variable {
	variables, _ := appContext.database.GetAllVariablesByEnvironmentAndScope(id, "global")
	return variables
}
