package dbms

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDatabase(t *testing.T) {
	database := Database{}
	database.Connect("xpto", true)
	assert.NotNil(t, database.Db)
}
