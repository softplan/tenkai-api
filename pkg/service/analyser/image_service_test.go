package analyser

import (
	"github.com/pkg/errors"
	"github.com/softplan/tenkai-api/pkg/service/_helm/mocks"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestGetImageFromService(t *testing.T) {

	m := sync.Mutex{}
	api := mocks.HelmServiceInterface{}

	bytes := []byte("{\"image\":{\"repository\":\"alfa\"}}")
	api.On("GetValues", "abacaxi", "0").Return(bytes, nil)

	s, e := GetImageFromService(&api, "abacaxi", &m)
	assert.Nil(t, e)
	assert.Equal(t, s, "alfa")
}

func TestGetImageFromServiceError(t *testing.T) {
	m := sync.Mutex{}
	api := mocks.HelmServiceInterface{}
	api.On("GetValues", "abacaxi", "0").Return(nil, errors.New("error"))
	s, e := GetImageFromService(&api, "abacaxi", &m)
	assert.NotNil(t, e)
	assert.Equal(t, "", s)
}

func TestGetImageFromServiceErrorUnmarshal(t *testing.T) {

	m := sync.Mutex{}
	api := mocks.HelmServiceInterface{}

	bytes := []byte("[{\"image\":{\"repository\":\"alfa\"}}]")
	api.On("GetValues", "abacaxi", "0").Return(bytes, nil)

	s, e := GetImageFromService(&api, "abacaxi", &m)
	assert.NotNil(t, e)
	assert.Equal(t, "", s)
}
