package util

import (
	"bytes"
	"encoding/json"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestUnmarshall(t *testing.T) {
	var payload model.Solution
	payload.Name = "alfa"
	payload.Team = "teamx"
	payloadStr, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/solutions", bytes.NewBuffer(payloadStr))
	var result model.Solution
	UnmarshalPayload(req, &result)
	assert.Equal(t, payload, result)
}

func TestGetPrincipal(t *testing.T) {
	var payload model.Solution
	payload.Name = "alfa"
	payload.Team = "teamx"
	payloadStr, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/solutions", bytes.NewBuffer(payloadStr))
	roles := []string{"abacaxi"}
	principal := model.Principal{Name: "alfa", Email: "beta@alfa.com", Roles: roles}
	pSe, _ := json.Marshal(principal)
	req.Header.Set("principal", string(pSe))
	principal = GetPrincipal(req)
	assert.Equal(t, "beta@alfa.com", principal.Email)
	assert.Equal(t, 1, len(principal.Roles))
}
