package constraints

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRoles(t *testing.T) {
	assert.Equal(t, TenkaiPromote, "tenkai-promote")
	assert.Equal(t, TenkaiAdmin, "tenkai-admin")
	assert.Equal(t, TenkaiVariablesSave, "tenkai-variables-save")
	assert.Equal(t, TenkaiVariablesDelete, "tenkai-variables-delete")
	assert.Equal(t, TenkaiHelmUpgrade, "tenkai-helm-upgrade")
}
