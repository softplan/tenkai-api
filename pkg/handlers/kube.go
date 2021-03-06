package handlers

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"github.com/softplan/tenkai-api/pkg/constraints"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/softplan/tenkai-api/pkg/global"
	"github.com/softplan/tenkai-api/pkg/util"
	"net/http"
	"strconv"
)

func (appContext *AppContext) services(w http.ResponseWriter, r *http.Request) {

	w.Header().Set(global.ContentType, global.JSONContentType)

	vars := mux.Vars(r)
	id := vars["id"]

	idI, _ := strconv.ParseInt(id, 10, 64)

	environment, err := appContext.Repositories.EnvironmentDAO.GetByID(int(idI))
	if err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	kubeConfig := appContext.ConventionInterface.GetKubeConfigFileName(environment.Group, environment.Name)

	services, err := appContext.HelmServiceAPI.GetServices(kubeConfig, environment.Namespace)
	if err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	var result model.ServiceResult
	result.Services = services

	data, _ := json.Marshal(result)

	w.WriteHeader(http.StatusOK)
	w.Write(data)

}

func (appContext *AppContext) pods(w http.ResponseWriter, r *http.Request) {

	w.Header().Set(global.ContentType, global.JSONContentType)

	vars := mux.Vars(r)
	id := vars["id"]

	idI, _ := strconv.ParseInt(id, 10, 64)

	environment, err := appContext.Repositories.EnvironmentDAO.GetByID(int(idI))
	if err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	kubeConfig := appContext.ConventionInterface.GetKubeConfigFileName(environment.Group, environment.Name)

	pods, err := appContext.HelmServiceAPI.GetPods(kubeConfig, environment.Namespace)
	if err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	var result model.PodResult
	result.Pods = pods

	data, _ := json.Marshal(result)

	w.WriteHeader(http.StatusOK)
	w.Write(data)

}

func (appContext *AppContext) deletePod(w http.ResponseWriter, r *http.Request) {

	isAdmin := false
	principal := util.GetPrincipal(r)
	if util.Contains(principal.Roles, constraints.TenkaiAdmin) {
		isAdmin = true
	}

	environmentIDs, ok := r.URL.Query()["environmentID"]
	if !ok || len(environmentIDs[0]) < 1 {
		http.Error(w, errors.New("param environmentID is required").Error(), 501)
		return
	}

	podName, ok := r.URL.Query()["podName"]
	if !ok || len(podName[0]) < 1 {
		http.Error(w, errors.New("param podName is required").Error(), 501)
		return
	}

	//Locate Environment
	envID, _ := strconv.Atoi(environmentIDs[0])

	has, err := appContext.hasAccess(principal.Email, envID)
	if err != nil || !has {
		http.Error(w, errors.New("Access Denied in this environment").Error(), http.StatusUnauthorized)
		return
	}

	//If not admin, verify authorization of user for specific environment
	if !isAdmin {
		auth, _ := appContext.hasEnvironmentRole(principal, uint(envID), "ACTION_DELETE_POD")
		if !auth {
			http.Error(w, errors.New(global.AccessDenied).Error(), http.StatusUnauthorized)
			return
		}
	}

	environment, err := appContext.Repositories.EnvironmentDAO.GetByID(int(envID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	kubeConfig := appContext.ConventionInterface.GetKubeConfigFileName(environment.Group, environment.Name)

	err = appContext.HelmServiceAPI.DeletePod(kubeConfig, podName[0], environment.Namespace)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

}
