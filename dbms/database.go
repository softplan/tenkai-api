package dbms

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	mocket "github.com/selvatico/go-mocket"
	"github.com/softplan/tenkai-api/dbms/model"
)

//Database Structure
type Database struct {
	Db *gorm.DB
}

//Connect - Connect to a database
func (database *Database) Connect(dbmsURI string, local bool) {
	var err error

	if local {
		database.Db, err = gorm.Open("sqlite3", "/tmp/tekai.db")
	} else {
		database.Db, err = gorm.Open("postgres", dbmsURI)
	}

	if err != nil {
		panic("failed to connect database")
	}

	database.Db.AutoMigrate(&model.Environment{})
	database.Db.AutoMigrate(&model.Variable{})
	database.Db.AutoMigrate(&model.Release{})
	database.Db.AutoMigrate(&model.Dependency{}) //.AddForeignKey("release_id", "release(id)", "CASCADE", "RESTRICT")
	database.Db.AutoMigrate(&model.Solution{})
	database.Db.AutoMigrate(&model.SolutionChart{}) //.AddForeignKey("solution_id", "solution(id)", "CASCADE", "RESTRICT")
	database.Db.AutoMigrate(&model.User{})
	database.Db.AutoMigrate(&model.ConfigMap{})

}

//MockConnect - Mock a database connection
func (database *Database) MockConnect() {
	mocket.Catcher.Register() // Safe register. Allowed multiple calls to save
	mocket.Catcher.Logging = true
	// GORM
	db, _ := gorm.Open(mocket.DriverName, "connection_string") // Can be any connection string
	database.Db = db
}
