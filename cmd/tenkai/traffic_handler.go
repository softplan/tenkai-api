package main

import (
	"bytes"
	"github.com/softplan/tenkai-api/dbms/model"
	"github.com/softplan/tenkai-api/global"
	helmapi "github.com/softplan/tenkai-api/service/helm"
	"github.com/softplan/tenkai-api/util"
	"net/http"
	"strconv"
	"strings"
)

func (appContext *appContext) deployTrafficRule(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var payload model.TrafficRequest

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	//Locate Environment
	environment, err := appContext.database.GetByID(payload.EnvironmentID)

	//
	chart := "saj6/tenkai-canary"
	name := "canary-" + payload.ServiceName
	out := &bytes.Buffer{}

	kubeConfig := global.KubeConfigBasePath + environment.Group + "_" + environment.Name

	variables := make([]string, 1)

	variables = append(variables, "istio.virtualservices.hosts[0]="+payload.Domain)
	variables = append(variables, "istio.virtualservices.apiPath="+payload.ContextPath)

	appName := payload.ServiceName[:strings.Index(payload.ServiceName,"-")]

	variables = append(variables, "app.serviceName="+payload.ServiceName)
	variables = append(variables, "app.name="+appName)

	if payload.HeaderName != "" {
		variables = append(variables, "app.headerEnabled=true")
		variables = append(variables, "app.weightEnabled=false")
		variables = append(variables, "app.defaultReleaseName="+payload.DefaultReleaseName)
		variables = append(variables, "app.headerReleaseName="+payload.HeaderReleaseName)
		variables = append(variables, "app.headers[0].name="+payload.HeaderName)
		variables = append(variables, "app.headers[0].value="+payload.HeaderValue)
	} else {
		variables = append(variables, "app.headerEnabled=false")
		variables = append(variables, "app.weightEnabled=true")
		for i, e := range payload.Releases {
			variables = append(variables, "app.releases["+strconv.Itoa(i)+"].name="+e.Name)
			variables = append(variables, "app.releases["+strconv.Itoa(i)+"].value="+ strconv.Itoa(e.Weight))
		}
	}

	err = helmapi.Upgrade(kubeConfig, name, chart, environment.Namespace, variables, out)
	if err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	w.WriteHeader(http.StatusOK)

}
