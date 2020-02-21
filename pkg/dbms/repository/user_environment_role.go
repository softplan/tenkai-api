package repository

import (
	"github.com/jinzhu/gorm"
	model2 "github.com/softplan/tenkai-api/pkg/dbms/model"
)

//UserEnvironmentRoleDAOInterface - UserEnvironmentRoleDAOInterface
type UserEnvironmentRoleDAOInterface interface {
	CreateOrUpdate(so model2.UserEnvironmentRole) error
	GetRoleByUserAndEnvironment(user model2.User, envID uint) (*model2.SecurityOperation, error)
}

//UserEnvironmentRoleDAOImpl UserEnvironmentRoleDAOImpl
type UserEnvironmentRoleDAOImpl struct {
	Db *gorm.DB
}

//CreateOrUpdate - Create or update a security operation
func (dao UserEnvironmentRoleDAOImpl) CreateOrUpdate(so model2.UserEnvironmentRole) error {
	loadSO, err := dao.isEdit(so)
	if err != nil {
		return err
	}
	if loadSO != nil {
		return dao.edit(so, loadSO)
	}
	return dao.create(so)
}

func (dao UserEnvironmentRoleDAOImpl) isEdit(so model2.UserEnvironmentRole) (*model2.UserEnvironmentRole, error) {
	var loadSO model2.UserEnvironmentRole
	if err := dao.Db.Where(model2.UserEnvironmentRole{UserID: so.UserID, EnvironmentID: so.EnvironmentID}).First(&loadSO).Error; err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			return nil, err
		}
		return nil, nil
	}
	return &loadSO, nil
}

func (dao UserEnvironmentRoleDAOImpl) edit(so model2.UserEnvironmentRole, loadSo *model2.UserEnvironmentRole) error {
	loadSo.SecurityOperationID = so.SecurityOperationID
	if err := dao.Db.Save(&so).Error; err != nil {
		return err
	}
	return nil
}

func (dao UserEnvironmentRoleDAOImpl) create(so model2.UserEnvironmentRole) error {
	if err := dao.Db.Create(&so).Error; err != nil {
		return err
	}
	return nil
}

//GetRoleByUserAndEnvironment - GetRoleByUserAndEnvironment
func (dao UserEnvironmentRoleDAOImpl) GetRoleByUserAndEnvironment(user model2.User,
	envID uint) (*model2.SecurityOperation, error) {
	var userEnvironmentRole model2.UserEnvironmentRole
	if err := dao.Db.Where(model2.UserEnvironmentRole{UserID: user.ID, EnvironmentID: envID}).Find(&userEnvironmentRole).Error; err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			return nil, err
		}
		return nil, nil
	}
	var result model2.SecurityOperation
	if err := dao.Db.First(&result, userEnvironmentRole.SecurityOperationID).Error; err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			return nil, err
		}
		return nil, nil
	}
	return &result, nil
}
