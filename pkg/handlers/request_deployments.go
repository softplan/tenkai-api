package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/softplan/tenkai-api/pkg/global"
)

func (appContext *AppContext) listRequestDeployments(w http.ResponseWriter, r *http.Request) {
	keys := r.URL.Query()
	pageSize := 100

	startDate := keys.Get("start_date")
	endDate := keys.Get("end_date")
	pageString := keys.Get("page")
	pageSizeString := keys.Get("pageSize")
	environmentID := keys.Get("environment_id")

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

	deployments, err := appContext.Repositories.RequestDeploymentDAO.ListRequestDeployments(startDate, endDate, environmentID, -1, page, pageSize)
	if err != nil {
		logListRequestDeployments("error on db query - " + err.Error())
	}
	count, err := appContext.Repositories.RequestDeploymentDAO.CountRequestDeployments(startDate, endDate, environmentID)
	if err != nil {
		logListRequestDeployments("error on db query - " + err.Error())
	}

	responseStruct := model.ResponseDeploymentResponse{
		Data:       deployments,
		Count:      count,
		TotalPages: getTotalPages(pageSize, int(count)),
	}

	responseJSON, _ := json.Marshal(responseStruct)

	w.Header().Set(global.ContentType, "application/json; charset=UTF-8")
	w.Write(responseJSON)
}

func logListRequestDeployments(errorMessage string) {
	global.Logger.Error(
		global.AppFields{
			global.Function: "ListRequestDeployments",
		},
		errorMessage,
	)
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
