package dbms

import (
	"github.com/softplan/tenkai-api/dbms/model"
)

// EditVariable - Edit an existent variable
func (database *Database) EditVariable(data model.Variable) error {
	if err := database.Db.Save(&data).Error; err != nil {
		return err
	}
	return nil
}

//CreateVariable - Create a new environment
func (database *Database) CreateVariable(variable model.Variable) (map[string]string, bool, error) {

	auditValues := make(map[string]string)
	updated := false

	var variableEntity model.Variable
	//Verify if update
	if err := database.Db.Where(&model.Variable{
		EnvironmentID: variable.EnvironmentID,
		Scope:         variable.Scope,
		Name:          variable.Name}).First(&variableEntity).Error; err == nil {

		if variable.Value != variableEntity.Value {

			auditValues["variable_name"] = variableEntity.Name
			auditValues["variable_old_value"] = variableEntity.Value
			auditValues["variable_new_value"] = variable.Value
			auditValues["scope"] = variable.Scope

			variableEntity.Value = variable.Value
			if err := database.Db.Save(variableEntity).Error; err != nil {
				return auditValues, updated, err
			}
			updated = true
		}

	} else {

		if err := database.Db.Create(&variable).Error; err != nil {
			return auditValues, updated, err
		}
		updated = true

		auditValues["variable_name"] = variable.Name
		auditValues["variable_value"] = variable.Value
	}

	return auditValues, updated, nil
}

//GetAllVariablesByEnvironment - Retrieve all variables
func (database *Database) GetAllVariablesByEnvironment(envID int) ([]model.Variable, error) {

	variables := make([]model.Variable, 0)
	var env model.Environment
	var err error

	if err = database.Db.First(&env, envID).Error; err != nil {
		return nil, err
	}

	if err = database.Db.Model(&env).Order("scope").Order("name").Related(&variables).Error; err != nil {
		return nil, err
	}

	return variables, nil

}

//GetAllVariablesByEnvironmentAndScope - Retrieve all variables
func (database *Database) GetAllVariablesByEnvironmentAndScope(envID int, scope string) ([]model.Variable, error) {
	variables := make([]model.Variable, 0)

	if err := database.Db.Where(&model.Variable{EnvironmentID: envID,
		Scope: scope,
	}).Find(&variables).Error; err != nil {
		return nil, err
	}

	return variables, nil
}

//DeleteVariable - Delete environment
func (database *Database) DeleteVariable(id int) error {
	if err := database.Db.Unscoped().Delete(model.Variable{}, id).Error; err != nil {
		return err
	}
	return nil
}

//DeleteVariableByEnvironmentID - Delete environment
func (database *Database) DeleteVariableByEnvironmentID(envID int) error {
	if err := database.Db.Unscoped().Where(model.Variable{EnvironmentID: envID}).Delete(model.Variable{}).Error; err != nil {
		return err
	}
	return nil
}
