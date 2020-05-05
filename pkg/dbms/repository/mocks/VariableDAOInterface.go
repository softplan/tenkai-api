// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	model "github.com/softplan/tenkai-api/pkg/dbms/model"
	mock "github.com/stretchr/testify/mock"
)

// VariableDAOInterface is an autogenerated mock type for the VariableDAOInterface type
type VariableDAOInterface struct {
	mock.Mock
}

// CreateVariable provides a mock function with given fields: variable
func (_m *VariableDAOInterface) CreateVariable(variable model.Variable) (map[string]string, bool, error) {
	ret := _m.Called(variable)

	var r0 map[string]string
	if rf, ok := ret.Get(0).(func(model.Variable) map[string]string); ok {
		r0 = rf(variable)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]string)
		}
	}

	var r1 bool
	if rf, ok := ret.Get(1).(func(model.Variable) bool); ok {
		r1 = rf(variable)
	} else {
		r1 = ret.Get(1).(bool)
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(model.Variable) error); ok {
		r2 = rf(variable)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// CreateVariableWithDefaultValue provides a mock function with given fields: variable
func (_m *VariableDAOInterface) CreateVariableWithDefaultValue(variable model.Variable) (map[string]string, bool, error) {
	ret := _m.Called(variable)

	var r0 map[string]string
	if rf, ok := ret.Get(0).(func(model.Variable) map[string]string); ok {
		r0 = rf(variable)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]string)
		}
	}

	var r1 bool
	if rf, ok := ret.Get(1).(func(model.Variable) bool); ok {
		r1 = rf(variable)
	} else {
		r1 = ret.Get(1).(bool)
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(model.Variable) error); ok {
		r2 = rf(variable)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// DeleteVariable provides a mock function with given fields: id
func (_m *VariableDAOInterface) DeleteVariable(id int) error {
	ret := _m.Called(id)

	var r0 error
	if rf, ok := ret.Get(0).(func(int) error); ok {
		r0 = rf(id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteVariableByEnvironmentID provides a mock function with given fields: envID
func (_m *VariableDAOInterface) DeleteVariableByEnvironmentID(envID int) error {
	ret := _m.Called(envID)

	var r0 error
	if rf, ok := ret.Get(0).(func(int) error); ok {
		r0 = rf(envID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// EditVariable provides a mock function with given fields: data
func (_m *VariableDAOInterface) EditVariable(data model.Variable) error {
	ret := _m.Called(data)

	var r0 error
	if rf, ok := ret.Get(0).(func(model.Variable) error); ok {
		r0 = rf(data)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetAllVariablesByEnvironment provides a mock function with given fields: envID
func (_m *VariableDAOInterface) GetAllVariablesByEnvironment(envID int) ([]model.Variable, error) {
	ret := _m.Called(envID)

	var r0 []model.Variable
	if rf, ok := ret.Get(0).(func(int) []model.Variable); ok {
		r0 = rf(envID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]model.Variable)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(int) error); ok {
		r1 = rf(envID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetAllVariablesByEnvironmentAndScope provides a mock function with given fields: envID, scope
func (_m *VariableDAOInterface) GetAllVariablesByEnvironmentAndScope(envID int, scope string) ([]model.Variable, error) {
	ret := _m.Called(envID, scope)

	var r0 []model.Variable
	if rf, ok := ret.Get(0).(func(int, string) []model.Variable); ok {
		r0 = rf(envID, scope)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]model.Variable)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(int, string) error); ok {
		r1 = rf(envID, scope)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetByID provides a mock function with given fields: id
func (_m *VariableDAOInterface) GetByID(id uint) (*model.Variable, error) {
	ret := _m.Called(id)

	var r0 *model.Variable
	if rf, ok := ret.Get(0).(func(uint) *model.Variable); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Variable)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(uint) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetVarImageTagByEnvAndScope provides a mock function with given fields: envID, scope
func (_m *VariableDAOInterface) GetVarImageTagByEnvAndScope(envID int, scope string) (model.Variable, error) {
	ret := _m.Called(envID, scope)

	var r0 model.Variable
	if rf, ok := ret.Get(0).(func(int, string) model.Variable); ok {
		r0 = rf(envID, scope)
	} else {
		r0 = ret.Get(0).(model.Variable)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(int, string) error); ok {
		r1 = rf(envID, scope)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
