package main

import (
	"bytes"
	"github.com/softplan/tenkai-api/dbms/model"
	"github.com/softplan/tenkai-api/global"
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
	environment, err := appContext.environmentDAO.GetByID(payload.EnvironmentID)
	domain := strings.Replace(payload.Domain, "${NAMESPACE}", environment.Namespace, -1)
	serviceName := strings.Replace(payload.ServiceName, "${NAMESPACE}", environment.Namespace, -1)

	defaultReleaseName := strings.Replace(payload.DefaultReleaseName, "${NAMESPACE}", environment.Namespace, -1)
	headerReleaseName := strings.Replace(payload.HeaderReleaseName, "${NAMESPACE}", environment.Namespace, -1)

	//
	chart := "saj6/tenkai-canary"
	name := "canary-" + serviceName
	out := &bytes.Buffer{}

	kubeConfig := global.KubeConfigBasePath + environment.Group + "_" + environment.Name

	variables := make([]string, 1)
	variables = append(variables, "istio.virtualservices.hosts[0]="+domain)

	variables = append(variables, "istio.virtualservices.apiPath="+payload.ContextPath)

	appName := payload.ServiceName[:strings.Index(payload.ServiceName, "-")]

	variables = append(variables, "app.serviceName="+serviceName)
	variables = append(variables, "app.name="+appName)

	if payload.HeaderName != "" {
		variables = append(variables, "app.headerEnabled=true")
		variables = append(variables, "app.weightEnabled=false")
		variables = append(variables, "app.defaultReleaseName="+defaultReleaseName)
		variables = append(variables, "app.headerReleaseName="+headerReleaseName)
		variables = append(variables, "app.headers[0].name="+payload.HeaderName)
		variables = append(variables, "app.headers[0].value="+payload.HeaderValue)
	} else {
		variables = append(variables, "app.headerEnabled=false")
		variables = append(variables, "app.weightEnabled=true")
		for i, e := range payload.Releases {
			name := strings.Replace(e.Name, "${NAMESPACE}", environment.Namespace, -1)
			variables = append(variables, "app.releases["+strconv.Itoa(i)+"].name="+name)
			variables = append(variables, "app.releases["+strconv.Itoa(i)+"].value="+strconv.Itoa(e.Weight))
		}
	}

	//Retry 2 times (First time will fail because service already exists).
	//TODO - VERIFY HOW TO FIX IT
	err = appContext.helmServiceAPI.Upgrade(kubeConfig, name, chart, "", environment.Namespace, variables, out, false)
	if err != nil {
		err = appContext.helmServiceAPI.Upgrade(kubeConfig, name, chart, "", environment.Namespace, variables, out, false)
		if err != nil {
			http.Error(w, err.Error(), 501)
			return
		}
	}

	w.WriteHeader(http.StatusOK)

}
