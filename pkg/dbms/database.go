//+build !test

package dbms

import (
	"github.com/jinzhu/gorm"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	model2 "github.com/softplan/tenkai-api/pkg/dbms/model"

	//postgres
	_ "github.com/jinzhu/gorm/dialects/postgres"
	//sqllite
	_ "github.com/jinzhu/gorm/dialects/sqlite"
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

	database.Db.AutoMigrate(&model2.Environment{})
	database.Db.AutoMigrate(&model2.Variable{})
	database.Db.AutoMigrate(&model2.Solution{})
	database.Db.AutoMigrate(&model2.SolutionChart{}) //.AddForeignKey("solution_id", "solution(id)", "CASCADE", "RESTRICT")
	database.Db.AutoMigrate(&model2.User{})
	database.Db.AutoMigrate(&model2.ConfigMap{})
	database.Db.AutoMigrate(&model2.Product{})
	database.Db.AutoMigrate(&model2.ProductVersion{})
	database.Db.AutoMigrate(&model2.ProductVersionService{})
	database.Db.AutoMigrate(&model2.ValueRule{})
	database.Db.AutoMigrate(&model2.VariableRule{})
	database.Db.AutoMigrate(&model2.CompareEnvsQuery{})
	database.Db.AutoMigrate(&model2.SecurityOperation{})
	database.Db.AutoMigrate(&model2.UserEnvironmentRole{})

	database.Db.Model(&model.ValueRule{}).
		AddForeignKey("variable_rule_id", "variable_rules(id)", "CASCADE", "CASCADE")
}
