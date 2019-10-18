package handlers

import (
	"errors"
	"github.com/softplan/tenkai-api/pkg/constraints"
	"github.com/softplan/tenkai-api/pkg/global"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/softplan/tenkai-api/pkg/util"
)

func (appContext *AppContext) newEnvironmentPermission(w http.ResponseWriter, r *http.Request) {

	principal := util.GetPrincipal(r)
	if !util.Contains(principal.Roles, constraints.TenkaiAdmin) {
		http.Error(w, errors.New(global.AccessDenied).Error(), http.StatusUnauthorized)
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["userID"])

	if err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	environmentID, err := strconv.Atoi(vars["environmentId"])
	if err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	if err := appContext.Repositories.UserDAO.AssociateEnvironmentUser(userID, environmentID); err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	w.WriteHeader(http.StatusCreated)

}
