package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/softplan/tenkai-api/dbms/model"
	"github.com/softplan/tenkai-api/global"
	helmapi "github.com/softplan/tenkai-api/service/helm"
	"net/http"
	"strconv"
)

func (appContext *appContext) pods(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	vars := mux.Vars(r)
	id := vars["id"]

	idI, _ := strconv.ParseInt(id, 10, 64)

	environment, err := appContext.database.GetByID(int(idI))
	if err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	kubeConfig := global.KubeConfigBasePath + environment.Group + "_" + environment.Name

	pods, err := helmapi.GetPods(kubeConfig, environment.Namespace)
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
