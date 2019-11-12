package audit

import (
	"github.com/olivere/elastic"
	"github.com/softplan/tenkai-api/pkg/audit/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestGetElkClient(t *testing.T) {
	audit := AuditingImpl{}
	mockElk := &mocks.ElkInterface{}
	audit.Elk = mockElk
	elkClient := elastic.Client{}
	mockElk.On("NewClient", mock.Anything).Return(&elkClient, nil)
	_, e := audit.ElkClient("http://localhost:8080", "alfa", "beta")
	assert.Nil(t, e)
}

func TestGetAuditBuilder(t *testing.T) {
	a := AuditingBuilder()
	assert.NotNil(t, a)
}
