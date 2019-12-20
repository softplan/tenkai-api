package handlers

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"log"

	"github.com/gorilla/mux"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/softplan/tenkai-api/pkg/global"
	"github.com/softplan/tenkai-api/pkg/util"
)

func (appContext *AppContext) validateVariables(w http.ResponseWriter, r *http.Request) {

	w.Header().Set(global.ContentType, global.JSONContentType)

	type Payload struct {
		EnvironmentID int    `json:"environmentId"`
		Scope         string `json:"scope"`
	}

	var payload Payload

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	vars, err := appContext.Repositories.VariableDAO.GetAllVariablesByEnvironmentAndScope(payload.EnvironmentID, payload.Scope)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	vrs, err := appContext.Repositories.VariableRuleDAO.ListVariableRules()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ivr, err := appContext.validate(vars, vrs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, _ := json.Marshal(ivr)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (appContext *AppContext) validateEnvironmentVariables(w http.ResponseWriter, r *http.Request) {

	w.Header().Set(global.ContentType, global.JSONContentType)

	v := mux.Vars(r)
	sl := v["envId"]
	envID, _ := strconv.Atoi(sl)

	vars, err := appContext.Repositories.VariableDAO.GetAllVariablesByEnvironment(envID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	vrs, err := appContext.Repositories.VariableRuleDAO.ListVariableRules()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ivr, err := appContext.validate(vars, vrs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, _ := json.Marshal(ivr)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (appContext *AppContext) validate(variables []model.Variable,
	varRules []model.VariableRule) (*model.InvalidVariablesResult, error) {

	result := &model.InvalidVariablesResult{}
	result.InvalidVariables = []model.InvalidVariable{}

	for _, varRule := range varRules {
		for _, valueRule := range varRule.ValueRules {
			for _, variable := range variables {

				if varRuleAppliesToVar(varRule.Name, variable.Name) {

					if valid := validationFn(valueRule.Type)(varRule, valueRule, variable); !valid {
						invalidVar := createResult(varRule, valueRule, variable)
						result.InvalidVariables = append(result.InvalidVariables, invalidVar)
					}

				}

			}
		}
	}

	return result, nil
}

func validationFn(validator string) fn {
	m := make(map[string]fn)

	m["NotEmpty"] = notEmpty
	m["StartsWith"] = startsWith
	m["EndsWith"] = endsWith
	m["RegEx"] = regEx

	return m[validator]
}

type fn func(model.VariableRule, *model.ValueRule, model.Variable) bool

func notEmpty(vrr model.VariableRule, vlr *model.ValueRule, v model.Variable) bool {
	result := len(v.Value) > 0
	logMsg(vrr, vlr, v, result)
	return result
}

func startsWith(vrr model.VariableRule, vlr *model.ValueRule, v model.Variable) bool {
	result := strings.HasPrefix(v.Value, vlr.Value)
	logMsg(vrr, vlr, v, result)
	return result
}

func endsWith(vrr model.VariableRule, vlr *model.ValueRule, v model.Variable) bool {
	result := strings.HasSuffix(v.Value, vlr.Value)
	logMsg(vrr, vlr, v, result)
	return result
}

func regEx(vrr model.VariableRule, vlr *model.ValueRule, v model.Variable) bool {
	result, err := regexp.MatchString(vlr.Value, v.Value)

	if err != nil {
		log.Println("Error in regexp match", err)
		return false
	}

	logMsg(vrr, vlr, v, result)
	return result
}

//varRuleAppliesToVar Validates only variables whose name matches the variableRule value.
func varRuleAppliesToVar(regex string, varName string) bool {
	result, err := regexp.MatchString(regex, varName)
	if err != nil {
		log.Println("Error in regexp match", err)
		return false
	}
	return result
}

func createResult(vrr model.VariableRule, vlr *model.ValueRule, v model.Variable) model.InvalidVariable {
	var iv model.InvalidVariable
	iv.Scope = v.Scope
	iv.Name = v.Name
	iv.Value = v.Value
	iv.VariableRule = vrr.Name
	iv.RuleType = vlr.Type
	iv.ValueRule = vlr.Value
	return iv
}

func logMsg(vrr model.VariableRule, vlr *model.ValueRule, v model.Variable, result bool) {
	log.Print("Variable ", v.Name, "='", v.Value, "' ", vlr.Type, " '", vlr.Value, "'? ", strconv.FormatBool(result))
}
