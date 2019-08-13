package dbms

import (
	"github.com/softplan/tenkai-api/dbms/model"
	"testing"
)


func TestGetDependencies(t *testing.T) {
	database := Database{}
	database.MockConnect()
	var expected []model.Dependency
	dependencies, err := database.GetDependencies("alfa", "1.0")
	if err != nil {
		t.Errorf("Error getting dependencies: got %v want %v", dependencies, expected)
	}

}
