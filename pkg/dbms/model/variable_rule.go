package model

import "github.com/jinzhu/gorm"

//VariableRule Structure
type VariableRule struct {
	gorm.Model
	Name       string       `json:"name"`
	ValueRules []*ValueRule `gorm:"foreignkey:VariableRuleID"`
}

//VariableRuleReponse struct
type VariableRuleReponse struct {
	List []VariableRule `json:"list"`
}
