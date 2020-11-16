package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/softplan/tenkai-api/pkg/global"
)

func (appContext *AppContext) listDeployments(w http.ResponseWriter, r *http.Request) {
	keys := r.URL.Query()
	pageSize := 100

	startDate := keys.Get("start_date")
	endDate := keys.Get("end_date")
	userID := keys.Get("user_id")
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

	if errorMessage, success := validateRequiredParams(startDate, endDate); !success {
		http.Error(w, errorMessage, http.StatusBadRequest)
		return
	}
	page, success := validatePageParam(pageString)

	if !success {
		http.Error(w, "Page must be a number", http.StatusBadRequest)
		return
	}

	deployments, err := appContext.Repositories.DeploymentDAO.ListDeployments(startDate, endDate, userID, environmentID, page, pageSize)
	if err != nil {
		logListDeployments("error on db query - " + err.Error())
	}
	count, err := appContext.Repositories.DeploymentDAO.CountDeployments(startDate, endDate, userID, environmentID)
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

func validateRequiredParams(startDate, endDate string) (string, bool) {
	if startDate == "" {
		logListDeployments("Parameter start_date is required")
		return "Parameter start_date is required", false
	} else if _, err := time.Parse("2006-01-02", startDate); err != nil {
		logListDeployments("Parameter start_date is required with format YYYY-MM-DD - " + err.Error())
		return "Parameter start_date is required with format YYYY-MM-DD", false
	} else if endDate == "" {
		logListDeployments("Parameter end_date is required")
		return "Parameter end_date is required", false
	} else if _, err := time.Parse("2006-01-02", endDate); err != nil {
		logListDeployments("Parameter end_date is required with format YYYY-MM-DD - " + err.Error())
		return "Parameter end_date is required with format YYYY-MM-DD", false
	}
	return "", true
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
