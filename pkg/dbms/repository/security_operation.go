package repository

import (
	"github.com/jinzhu/gorm"
	model2 "github.com/softplan/tenkai-api/pkg/dbms/model"
)

//SecurityOperationDAOInterface - SecurityOperationDAOInterface
type SecurityOperationDAOInterface interface {
	List() ([]model2.SecurityOperation, error)
	CreateOrUpdate(so model2.SecurityOperation) error
	Delete(id int) error
}

//SecurityOperationDAOImpl SecurityOperationDAOImpl
type SecurityOperationDAOImpl struct {
	Db *gorm.DB
}

//CreateOrUpdate - Create or update a security operation
func (dao SecurityOperationDAOImpl) CreateOrUpdate(so model2.SecurityOperation) error {
	loadSO, err := dao.isEdit(so)
	if err != nil {
		return err
	}
	if loadSO != nil {
		return dao.edit(so, loadSO)
	}
	return dao.create(so)
}

func (dao SecurityOperationDAOImpl) isEdit(so model2.SecurityOperation) (*model2.SecurityOperation, error) {
	var loadSO model2.SecurityOperation
	if err := dao.Db.Where(model2.SecurityOperation{Name: so.Name}).First(&loadSO).Error; err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			return nil, err
		}
		return nil, nil
	}
	return &loadSO, nil
}

func (dao SecurityOperationDAOImpl) edit(so model2.SecurityOperation, loadSo *model2.SecurityOperation) error {
	loadSo.Policies = so.Policies
	if err := dao.Db.Save(&so).Error; err != nil {
		return err
	}
	return nil
}

func (dao SecurityOperationDAOImpl) create(so model2.SecurityOperation) error {
	if err := dao.Db.Create(&so).Error; err != nil {
		return err
	}
	return nil
}

//List - List
func (dao SecurityOperationDAOImpl) List() ([]model2.SecurityOperation, error) {
	oss := make([]model2.SecurityOperation, 0)
	if err := dao.Db.Find(&oss).Error; err != nil {
		return nil, err
	}
	return oss, nil
}

//Delete - Delete
func (dao SecurityOperationDAOImpl) Delete(id int) error {
	var item model2.SecurityOperation
	if err := dao.Db.First(&item, id).Error; err != nil {
		return err
	}
	if err := dao.Db.Unscoped().Delete(&item).Error; err != nil {
		return err
	}
	return nil
}
