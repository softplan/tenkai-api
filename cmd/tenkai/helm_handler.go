package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/softplan/tenkai-api/util"
	"net/http"

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

	helmapi.RepoUpdate()

	searchTerms := []string{repo}
	searchResult := helmapi.SearchCharts(searchTerms, allVersions)
	result := &model.ChartsResult{Charts: *searchResult}

	data, _ := json.Marshal(result)

	w.WriteHeader(http.StatusOK)
	w.Write(data)

}

func (appContext *appContext) listHelmDeployments(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	result := helmapi.ListHelmDeployments()
	data, _ := json.Marshal(result)
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}

func (appContext *appContext) getChartVariables(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	var payload model.GetChartRequest

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		w.WriteHeader(422)
		if err := json.NewEncoder(w).Encode(err); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	result, _ := helmapi.GetValues(payload.ChartName, payload.ChartVersion)

	w.WriteHeader(http.StatusOK)
	w.Write(result)

}

func (appContext *appContext) multipleInstall(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	var payload model.MultipleInstallPayload

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		w.WriteHeader(422)
		if err := json.NewEncoder(w).Encode(err); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	out := &bytes.Buffer{}

	for _, element := range payload.Deployables {
		err := appContext.simpleInstall(element.EnvironmentID, element.Chart, element.Name, out)
		if err != nil {
			if err := json.NewEncoder(w).Encode(err); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}
	}

	fmt.Println(out.String())
	w.WriteHeader(http.StatusOK)

}

func (appContext *appContext) install(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	var payload model.InstallPayload

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		w.WriteHeader(422)
		if err := json.NewEncoder(w).Encode(err); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	out := &bytes.Buffer{}

	//TODO Verify if chart exists
	err := appContext.simpleInstall(payload.EnvironmentID, payload.Chart, payload.Name, out)
	if err != nil {
		if err := json.NewEncoder(w).Encode(err); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	fmt.Println(out.String())
	w.WriteHeader(http.StatusOK)

}

func (appContext *appContext) simpleInstall(envId int, chart string, name string, out *bytes.Buffer) error {

	//Locate Environment
	environment, err := appContext.database.GetByID(envId)

	//TODO - VERIFY IF CONFIG FILE EXISTS !!! This is the cause of  u.client.ReleaseHistory fail sometimes.

	variables, err := appContext.database.GetAllVariablesByEnvironmentAndScope(envId, chart)
	globalVariables := appContext.getGlobalVariables(int(environment.ID))

	var args []string
	for _, item := range variables {
		if len(item.Name) > 0 {
			args = append(args, normalizeVariableName(item.Name)+"="+replace(item.Value, *environment, globalVariables))
		}
	}
	//Add Default Gateway
	if len(environment.Gateway) > 0 {
		args = append(args, "istio.virtualservices.gateways[0]="+environment.Gateway)
	}
	if err != nil {
		return err
	} else {
		name := name + "-" + environment.Namespace
		kubeConfig := global.KUBECONFIG_BASE_PATH + environment.Group + "_" + environment.Name
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
	if strings.Index(value, "istio.") > -1 || (strings.Index(value, "image.")) > -1 {
		return value
	} else {
		return "app." + value
	}
}

func (appContext *appContext) getGlobalVariables(id int) []model.Variable {
	variables, _ := appContext.database.GetAllVariablesByEnvironmentAndScope(id, "global")
	return variables
}
