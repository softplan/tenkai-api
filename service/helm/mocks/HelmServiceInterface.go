// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"
import model "github.com/softplan/tenkai-api/dbms/model"

// HelmServiceInterface is an autogenerated mock type for the HelmServiceInterface type
type HelmServiceInterface struct {
	mock.Mock
}

// SearchCharts provides a mock function with given fields: searchTerms, allVersions
func (_m *HelmServiceInterface) SearchCharts(searchTerms []string, allVersions bool) *[]model.SearchResult {
	ret := _m.Called(searchTerms, allVersions)

	var r0 *[]model.SearchResult
	if rf, ok := ret.Get(0).(func([]string, bool) *[]model.SearchResult); ok {
		r0 = rf(searchTerms, allVersions)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*[]model.SearchResult)
		}
	}

	return r0
}
