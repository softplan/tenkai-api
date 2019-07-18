package dbms

import (
	"github.com/jinzhu/gorm"
	//_ "github.com/jinzhu/gorm/dialects/sqlite"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/softplan/tenkai-api/dbms/model"
)

//Database Structure
type Database struct {
	Db *gorm.DB
}

//Connect to a database
func (database *Database) Connect(dbmsUri string) {
	var err error
	//database.Db, err = gorm.Open("sqlite3", "/tmp/tekai.db")
	database.Db, err = gorm.Open("postgres", dbmsUri)

	if err != nil {
		panic("failed to connect database")
	}

	database.Db.AutoMigrate(&model.Environment{})
	database.Db.AutoMigrate(&model.Variable{})
	database.Db.AutoMigrate(&model.Release{})
	database.Db.AutoMigrate(&model.Dependency{}) //.AddForeignKey("release_id", "release(id)", "CASCADE", "RESTRICT")

}
