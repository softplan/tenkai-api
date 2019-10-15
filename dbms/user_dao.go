package dbms

import (
	"github.com/jinzhu/gorm"
	"github.com/softplan/tenkai-api/dbms/model"
)

//UserDAOInterface UserDAOInterface
type UserDAOInterface interface {
	CreateUser(user model.User) error
	DeleteUser(id int) error
	AssociateEnvironmentUser(userID int, environmentID int) error
	ListAllUsers() ([]model.User, error)
	CreateOrUpdateUser(user model.User) error
}

//UserDAOImpl UserDAOImpl
type UserDAOImpl struct {
	Db *gorm.DB
}

//CreateUser - Creates a new user
func (dao *UserDAOImpl) CreateUser(user model.User) error {
	if err := dao.Db.Create(&user).Error; err != nil {
		return err
	}
	return nil
}

//DeleteUser - Delete user
func (dao *UserDAOImpl) DeleteUser(id int) error {

	var user model.User

	if err := dao.Db.First(&user, id).Error; err != nil {
		return err
	}

	//Remove all associations
	if err := dao.Db.Model(&user).Association("Environments").Clear().Error; err != nil {
		return err
	}

	if err := dao.Db.Unscoped().Delete(&user).Error; err != nil {
		return err
	}

	return nil
}

//AssociateEnvironmentUser - Associate an environment with a user
func (dao *UserDAOImpl) AssociateEnvironmentUser(userID int, environmentID int) error {
	var user model.User
	var environment model.Environment

	if err := dao.Db.First(&user, userID).Error; err != nil {
		return err
	}

	if err := dao.Db.First(&environment, environmentID).Error; err != nil {
		return err
	}
	if err := dao.Db.Model(&user).Association("Environments").Append(&environment).Error; err == nil {
		return err
	}

	return nil
}

//ListAllUsers - List all users
func (dao *UserDAOImpl) ListAllUsers() ([]model.User, error) {
	users := make([]model.User, 0)
	if err := dao.Db.Preload("Environments").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

//CreateOrUpdateUser - Create or update a user
func (dao *UserDAOImpl) CreateOrUpdateUser(user model.User) error {

	var loadUser model.User
	edit := true

	if err := dao.Db.Where(model.User{Email: user.Email}).First(&loadUser).Error; err != nil {
		edit = false
		if !gorm.IsRecordNotFoundError(err) {
			return err
		}
	}

	if edit {

		//Remove all associations
		if err := dao.Db.Model(&loadUser).Association("Environments").Clear().Error; err != nil {
			return err
		}

		//Associate Envs
		for _, element := range user.Environments {
			var environment model.Environment
			if err := dao.Db.First(&environment, element.ID).Error; err != nil {
				return err
			}
			if err := dao.Db.Model(&loadUser).Association("Environments").Append(&environment).Error; err != nil {
				return err
			}
		}

	} else {
		//Create User
		envsToAssociate := user.Environments

		user.Environments = nil

		if err := dao.Db.Create(&user).Error; err != nil {
			return err
		}

		//Associate Envs
		for _, element := range envsToAssociate {
			var environment model.Environment
			if err := dao.Db.First(&environment, element.ID).Error; err != nil {
				return err
			}
			if err := dao.Db.Model(&user).Association("Environments").Append(&environment).Error; err != nil {
				return err
			}
		}

	}

	return nil
}
