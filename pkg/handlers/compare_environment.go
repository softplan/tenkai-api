package handlers

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/softplan/tenkai-api/pkg/constraints"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/softplan/tenkai-api/pkg/global"
	"github.com/softplan/tenkai-api/pkg/util"
)

func (appContext *AppContext) compareEnvironments(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(global.ContentType, global.JSONContentType)

	var payload model.CompareEnvironments
	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	principal := util.GetPrincipal(r)

	hasPermission, err := hasPermissionToCompare(principal, uint(payload.SourceEnvID), uint(payload.TargetEnvID), appContext)

	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
	}

	if !hasPermission {
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	if len(payload.OnlyFields) > 0 && len(payload.ExceptFields) > 0 {
		http.Error(w, "Choose only one kind of filter fields: only or except", http.StatusInternalServerError)
		return
	}

	if len(payload.OnlyCharts) > 0 && len(payload.ExceptCharts) > 0 {
		http.Error(w, "Choose only one kind of filter charts: only or except", http.StatusInternalServerError)
		return
	}

	var sourceVars []model.Variable
	if sourceVars, err = appContext.Repositories.VariableDAO.
		GetAllVariablesByEnvironment(payload.SourceEnvID); err != nil {

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var targetVars []model.Variable
	if targetVars, err = appContext.Repositories.VariableDAO.
		GetAllVariablesByEnvironment(payload.TargetEnvID); err != nil {

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var resp model.CompareEnvsResponse
	rmap := make(map[uint32]model.EnvironmentsDiff)

	appContext.compare(rmap, payload, toMap(sourceVars), toMap(targetVars), false)
	appContext.compare(rmap, payload, toMap(targetVars), toMap(sourceVars), true)

	for _, v := range rmap {
		appContext.applyFilters(payload, v, &resp)
	}

	sort.Slice(resp.List, func(i int, j int) bool {
		return resp.List[i].SourceScope < resp.List[j].SourceScope
	})

	data, _ := json.Marshal(resp)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (appContext *AppContext) saveCompareEnvQuery(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(global.ContentType, global.JSONContentType)

	var payload model.SaveCompareEnvQuery

	var e error
	if e := util.UnmarshalPayload(r, &payload); e != nil {
		http.Error(w, e.Error(), http.StatusInternalServerError)
		return
	}

	var b []byte
	if b, e = json.Marshal(payload.Data); e != nil {
		http.Error(w, e.Error(), http.StatusInternalServerError)
		return
	}

	var user model.User
	if user, e = appContext.Repositories.UserDAO.FindByEmail(payload.UserEmail); e != nil {
		http.Error(w, e.Error(), http.StatusInternalServerError)
		return
	}

	var env model.CompareEnvsQuery
	env.Name = payload.Name
	env.UserID = int(user.ID)
	env.Query = postgres.Jsonb{RawMessage: b}

	if payload.ID > 0 {
		env.ID = payload.ID
	}

	if _, err := appContext.Repositories.CompareEnvsQueryDAO.SaveCompareEnvsQuery(env); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (appContext *AppContext) deleteCompareEnvQuery(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sl := vars["id"]
	id, _ := strconv.Atoi(sl)
	w.Header().Set(global.ContentType, global.JSONContentType)
	if err := appContext.Repositories.CompareEnvsQueryDAO.DeleteCompareEnvQuery(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (appContext *AppContext) loadCompareEnvQueries(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(global.ContentType, global.JSONContentType)
	principal := util.GetPrincipal(r)

	var user model.User
	var e error
	if user, e = appContext.Repositories.UserDAO.FindByEmail(principal.Email); e != nil {
		http.Error(w, e.Error(), http.StatusInternalServerError)
		return
	}

	var result []model.CompareEnvsQuery
	if result, e = appContext.Repositories.CompareEnvsQueryDAO.GetByUser(int(user.ID)); e != nil {
		http.Error(w, e.Error(), http.StatusInternalServerError)
		return
	}

	data, _ := json.Marshal(result)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (appContext *AppContext) applyFilters(payload model.CompareEnvironments,
	v model.EnvironmentsDiff, resp *model.CompareEnvsResponse) {

	var fieldName string
	if v.SourceName != "" {
		fieldName = v.SourceName
	} else {
		fieldName = v.TargetName
	}

	if len(payload.CustomFields) == 0 {
		resp.List = append(resp.List, v)
	} else {
		filterMatch := false
		for _, filter := range payload.CustomFields {
			if fieldFilter(filter.FilterType)(fieldName, filter.FilterValue) {
				filterMatch = true
				break
			}
		}

		if filterMatch {
			resp.List = append(resp.List, v)
		}
	}
}

func fieldFilter(filterType string) myFn {
	m := make(map[string]myFn)

	m["Contains"] = fieldContains
	m["StartsWith"] = fieldStartsWith
	m["EndsWith"] = fieldEndsWith
	m["RegEx"] = fieldRegExp

	return m[filterType]
}

type myFn func(field string, value string) bool

func fieldStartsWith(field string, value string) bool {
	return strings.HasPrefix(field, value)
}

func fieldContains(field string, value string) bool {
	return strings.Contains(field, value)
}

func fieldEndsWith(field string, value string) bool {
	return strings.HasSuffix(field, value)
}

func fieldRegExp(field string, value string) bool {
	result, err := regexp.MatchString(value, field)

	if err != nil {
		return false
	}

	return result
}

func (appContext *AppContext) compare(rmap map[uint32]model.EnvironmentsDiff,
	filter model.CompareEnvironments, source map[string]map[string]model.Variable,
	target map[string]map[string]model.Variable, reverse bool) {

	for scope, srcVars := range source {
		if shouldIgnoreChart(filter, scope) {
			continue
		}

		iterateOverSourceVars(filter, scope, srcVars, target, reverse, rmap)
	}
}

func iterateOverSourceVars(filter model.CompareEnvironments, scope string,
	srcVars map[string]model.Variable, target map[string]map[string]model.Variable,
	reverse bool, rmap map[uint32]model.EnvironmentsDiff) {

	for srcVarName, srcValue := range srcVars {
		if shouldIgnoreVar(filter, srcVarName) {
			continue
		}
		if _, ok := target[scope][srcVarName]; !ok {
			addToResp(rmap, filter, scope, scope, srcVarName, "", srcValue.Value, "", fmt.Sprint(srcValue.ID), "", reverse)
			continue
		}
		iterateOverTargetVars(filter, scope, target, srcValue, reverse, rmap)
	}
}

func iterateOverTargetVars(filter model.CompareEnvironments, scope string,
	target map[string]map[string]model.Variable, srcVar model.Variable,
	reverse bool, rmap map[uint32]model.EnvironmentsDiff) {

	for tarVarName, tarValue := range target[scope] {
		if shouldIgnoreVar(filter, tarVarName) || srcVar.Name != tarVarName || srcVar.Value == tarValue.Value {
			continue
		} else {
			addToResp(rmap, filter, scope, scope, srcVar.Name, tarVarName, srcVar.Value, tarValue.Value, fmt.Sprint(srcVar.ID), fmt.Sprint(tarValue.ID), reverse)
		}
	}
}

func shouldIgnoreVar(filter model.CompareEnvironments, varName string) bool {
	if len(filter.ExceptFields) > 0 {
		for _, e := range filter.ExceptFields {
			if e == varName {
				return true
			}
		}
		return false
	}

	if len(filter.OnlyFields) > 0 {
		found := false
		for _, e := range filter.OnlyFields {
			if e == varName {
				found = true
				continue
			}
		}
		return !found
	}

	return false
}

func shouldIgnoreChart(filter model.CompareEnvironments, scope string) bool {
	if len(filter.ExceptCharts) > 0 {
		for _, e := range filter.ExceptCharts {
			if e == scope {
				return true
			}
		}
		return false
	}

	if len(filter.OnlyCharts) > 0 {
		found := false
		for _, e := range filter.OnlyCharts {
			if e == scope {
				found = true
				continue
			}
		}
		return !found
	}

	return false
}

func toMap(vars []model.Variable) map[string]map[string]model.Variable {
	sm := make(map[string]map[string]model.Variable)

	for _, e := range vars {
		if sm[e.Scope] == nil {
			sm[e.Scope] = map[string]model.Variable{e.Name: e}
		} else {
			sm[e.Scope][e.Name] = e
		}
	}

	return sm
}

func addToResp(rmap map[uint32]model.EnvironmentsDiff, filter model.CompareEnvironments,
	srcScope string, tarScope string, srcVarName string,
	tarVarName string, srcValue string, tarValue string, srcVarID string, tarVarID string, reverse bool) {

	var e model.EnvironmentsDiff
	e.SourceEnvID = filter.SourceEnvID
	e.TargetEnvID = filter.TargetEnvID

	if reverse {
		e.SourceScope = tarScope
		e.TargetScope = srcScope
		e.SourceName = tarVarName
		e.TargetName = srcVarName
		e.SourceValue = tarValue
		e.TargetValue = srcValue
		e.SourceVarID = tarVarID
		e.TargetVarID = srcVarID
	} else {
		e.SourceScope = srcScope
		e.TargetScope = tarScope
		e.SourceName = srcVarName
		e.TargetName = tarVarName
		e.SourceValue = srcValue
		e.TargetValue = tarValue
		e.SourceVarID = srcVarID
		e.TargetVarID = tarVarID
	}

	// Avoid duplicated entries
	h := hash(e)
	if _, ok := rmap[h]; !ok {
		rmap[h] = e
	}
}

func hash(e model.EnvironmentsDiff) uint32 {
	key := e.SourceScope + e.TargetScope + e.SourceName +
		e.TargetName + e.SourceValue + e.TargetValue

	h := fnv.New32a()
	h.Write([]byte(key))
	return h.Sum32()
}

func hasPermissionToCompare(principal model.Principal, sourceEnvID, targetEnvID uint, appContext *AppContext) (bool, error) {
	if util.Contains(principal.Roles, constraints.TenkaiAdmin) {
		return true, nil
	}
	hasPermissionSourceEnv, err := appContext.hasEnvironmentRole(principal, sourceEnvID, "ACTION_COMPARE_ENVS")
	if err != nil {
		return false, err
	}
	hasPermissionTargetEnv, err := appContext.hasEnvironmentRole(principal, targetEnvID, "ACTION_COMPARE_ENVS")
	if err != nil {
		return false, err
	}
	return hasPermissionSourceEnv && hasPermissionTargetEnv, nil
}
