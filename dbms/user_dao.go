package dbms

import (
	"github.com/jinzhu/gorm"
	"github.com/softplan/tenkai-api/dbms/model"
)

func (database *Database) CreateUser(user model.User) error {
	if err := database.Db.Create(&user).Error; err != nil {
		return err
	}
	return nil
}

func (database *Database) DeleteUser(id int) error {

	var user model.User

	if err := database.Db.First(&user, id).Error; err != nil {
		return err
	}

	//Remove all associations
	if err := database.Db.Model(&user).Association("Environments").Clear().Error; err != nil {
		return err
	}

	if err := database.Db.Unscoped().Delete(&user).Error; err != nil {
		return err
	}

	return nil
}

func (database *Database) AssociateEnvironmentUser(userId int, environmentId int) error {
	var user model.User
	var environment model.Environment

	if err := database.Db.First(&user, userId).Error; err != nil {
		return err
	}

	if err := database.Db.First(&environment, environmentId).Error; err != nil {
		return err
	}
	if err := database.Db.Model(&user).Association("Environments").Append(&environment).Error; err == nil {
		return err
	}

	return nil
}

func (database *Database) ListAllUsers() ([]model.User, error) {
	users := make([]model.User, 0)
	if err := database.Db.Preload("Environments").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (database *Database) CreateOrUpdateUser(user model.User) error {

	var loadUser model.User
	var edit bool = true

	if err := database.Db.Where(model.User{Email: user.Email}).First(&loadUser).Error; err != nil {
		edit = false
		if !gorm.IsRecordNotFoundError(err) {
			return err
		}
	}

	if edit {

		//Remove all associations
		if err := database.Db.Model(&loadUser).Association("Environments").Clear().Error; err != nil {
			return err
		}

		//Associate Envs
		for _, element := range user.Environments {
			var environment model.Environment
			if err := database.Db.First(&environment, element.ID).Error; err != nil {
				return err
			}
			if err := database.Db.Model(&loadUser).Association("Environments").Append(&environment).Error; err != nil {
				return err
			}
		}


	} else {
		//Create User
		envsToAssociate := user.Environments

		user.Environments = nil

		if err := database.Db.Create(&user).Error; err != nil {
			return err
		}

		//Associate Envs
		for _, element := range envsToAssociate {
			var environment model.Environment
			if err := database.Db.First(&environment, element.ID).Error; err != nil {
				return err
			}
			if err := database.Db.Model(&user).Association("Environments").Append(&environment).Error; err != nil {
				return err
			}
		}

	}

	return nil
}


