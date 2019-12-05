package model

import "github.com/jinzhu/gorm"

//ValueRule Structure
type ValueRule struct {
	gorm.Model
	Type           string `json:"type"`
	Value          string `json:"value"`
	VariableRuleID uint
}

//ValueRuleReponse struct
type ValueRuleReponse struct {
	List []ValueRule `json:"list"`
}
