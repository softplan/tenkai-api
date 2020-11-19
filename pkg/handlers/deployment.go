package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/softplan/tenkai-api/pkg/global"
)

func (appContext *AppContext) listDeployments(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	resquestDeploymentID := vars["id"]
	
	keys := r.URL.Query()
	pageSize := 100

	environmentID := keys.Get("environment_id")
	pageString := keys.Get("page")
	pageSizeString := keys.Get("pageSize")

	if pageSizeString != "" {
		pageSizeAux, err := strconv.ParseUint(pageSizeString, 10, 32)
		if err == nil {
			pageSize = int(pageSizeAux)
		} else {
			http.Error(w, "pageSize must be a number", http.StatusBadRequest)
			return
		}
	}
	page, success := validatePageParam(pageString)

	if !success {
		http.Error(w, "Page must be a number", http.StatusBadRequest)
		return
	}

	deployments, err := appContext.Repositories.DeploymentDAO.ListDeployments(environmentID, resquestDeploymentID, page, pageSize)
	if err != nil {
		logListDeployments("error on db query - " + err.Error())
	}
	count, err := appContext.Repositories.DeploymentDAO.CountDeployments(environmentID, resquestDeploymentID)
	if err != nil {
		logListDeployments("error on db query - " + err.Error())
	}

	responseStruct := model.DeploymentResponse{
		Data:       deployments,
		Count:      count,
		TotalPages: getTotalPages(pageSize, int(count)),
	}

	responseJSON, _ := json.Marshal(responseStruct)

	w.Header().Set(global.ContentType, "application/json; charset=UTF-8")
	w.Write(responseJSON)
}

func validatePageParam(page string) (int, bool) {
	if page == "" {
		return 1, true
	}
	pageNumber, err := strconv.ParseUint(page, 10, 64)
	if err != nil {
		logListDeployments("page must be a positive integer")
		return -1, false
	}
	return int(pageNumber), true
}

func getTotalPages(pageSize int, count int) int {
	totalPages := int(count / pageSize)
	if count%pageSize > 0 {
		totalPages++
	}
	return totalPages
}

func logListDeployments(errorMessage string) {
	global.Logger.Error(
		global.AppFields{
			global.Function: "ListDeployments",
		},
		errorMessage,
	)
}
