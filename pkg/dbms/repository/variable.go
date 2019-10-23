package repository

import (
	"github.com/jinzhu/gorm"
	model2 "github.com/softplan/tenkai-api/pkg/dbms/model"
)

//VariableDAOInterface VariableDAOInterface
type VariableDAOInterface interface {
	EditVariable(data model2.Variable) error
	CreateVariable(variable model2.Variable) (map[string]string, bool, error)
	GetAllVariablesByEnvironment(envID int) ([]model2.Variable, error)
	GetAllVariablesByEnvironmentAndScope(envID int, scope string) ([]model2.Variable, error)
	DeleteVariable(id int) error
	DeleteVariableByEnvironmentID(envID int) error
}

//VariableDAOImpl VariableDAOImpl
type VariableDAOImpl struct {
	Db *gorm.DB
}

// EditVariable - Edit an existent variable
func (dao VariableDAOImpl) EditVariable(data model2.Variable) error {
	return dao.Db.Save(&data).Error
}

//CreateVariable - Create a new environment
func (dao VariableDAOImpl) CreateVariable(variable model2.Variable) (map[string]string, bool, error) {

	auditValues := make(map[string]string)
	updated := false

	var variableEntity model2.Variable
	//Verify if update
	if err := dao.Db.Where(&model2.Variable{
		EnvironmentID: variable.EnvironmentID,
		Scope:         variable.Scope,
		Name:          variable.Name}).First(&variableEntity).Error; err == nil {

		if variable.Value != variableEntity.Value {

			auditValues["variable_name"] = variableEntity.Name
			auditValues["variable_old_value"] = variableEntity.Value
			auditValues["variable_new_value"] = variable.Value
			auditValues["scope"] = variable.Scope

			variableEntity.Value = variable.Value
			if err := dao.Db.Save(variableEntity).Error; err != nil {
				return auditValues, updated, err
			}
			updated = true
		}

	} else {

		if err := dao.Db.Create(&variable).Error; err != nil {
			return auditValues, updated, err
		}
		updated = true

		auditValues["variable_name"] = variable.Name
		auditValues["variable_value"] = variable.Value
	}

	return auditValues, updated, nil
}

//GetAllVariablesByEnvironment - Retrieve all variables
func (dao VariableDAOImpl) GetAllVariablesByEnvironment(envID int) ([]model2.Variable, error) {

	variables := make([]model2.Variable, 0)
	var env model2.Environment
	var err error

	if err = dao.Db.First(&env, envID).Error; err != nil {
		return nil, err
	}

	if err = dao.Db.Model(&env).Order("scope").Order("name").Related(&variables).Error; err != nil {
		return nil, err
	}

	return variables, nil

}

//GetAllVariablesByEnvironmentAndScope - Retrieve all variables
func (dao VariableDAOImpl) GetAllVariablesByEnvironmentAndScope(envID int, scope string) ([]model2.Variable, error) {
	variables := make([]model2.Variable, 0)

	if err := dao.Db.Where(&model2.Variable{EnvironmentID: envID,
		Scope: scope,
	}).Find(&variables).Error; err != nil {
		return nil, err
	}

	return variables, nil
}

//DeleteVariable - Delete environment
func (dao VariableDAOImpl) DeleteVariable(id int) error {
	return dao.Db.Unscoped().Delete(model2.Variable{}, id).Error
}

//DeleteVariableByEnvironmentID - Delete environment
func (dao VariableDAOImpl) DeleteVariableByEnvironmentID(envID int) error {
	return dao.Db.Unscoped().Where(model2.Variable{EnvironmentID: envID}).Delete(model2.Variable{}).Error
}
