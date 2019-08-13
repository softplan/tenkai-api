package dbms

import (
	"github.com/jinzhu/gorm"
	mocket "github.com/selvatico/go-mocket"
	"github.com/softplan/tenkai-api/dbms/model"
	"testing"
)

func SetupTests() *gorm.DB {

	mocket.Catcher.Register() // Safe register. Allowed multiple calls to save
	mocket.Catcher.Logging = true
	// GORM
	db, _ := gorm.Open(mocket.DriverName, "connection_string") // Can be any connection string

	return db
}

func TestGetDependencies(t *testing.T) {

	db := SetupTests()
	database := Database{Db: db}
	var expected []model.Dependency
	dependencies, err := database.GetDependencies("alfa", "1.0")
	if err != nil {
		t.Errorf("Error getting dependencies: got %v want %v", dependencies, expected)
	}

}
