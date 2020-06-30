package repository

import (
	"github.com/jinzhu/gorm"
	model2 "github.com/softplan/tenkai-api/pkg/dbms/model"
)

//NotesDAOInterface NotesDAOInterface
type NotesDAOInterface interface {
	CreateNotes(notes model2.Notes) (int, error)
	EditNotes(notes model2.Notes) error
	GetByID(ID int) (*model2.Notes, error)
	GetByServiceName(serviceName string) (*model2.Notes, error)
}

//NotesDAOImpl NotesDAOImpl
type NotesDAOImpl struct {
	Db *gorm.DB
}

//CreateNotes - Create a new notes
func (dao NotesDAOImpl) CreateNotes(notes model2.Notes) (int, error) {
	if err := dao.Db.Create(&notes).Error; err != nil {
		return -1, err
	}
	return int(notes.ID), nil
}

// EditNotes - Updates an existing notes
func (dao NotesDAOImpl) EditNotes(notes model2.Notes) error {
	return dao.Db.Save(&notes).Error
}

//GetByID - Get Notes By Id
func (dao NotesDAOImpl) GetByID(ID int) (*model2.Notes, error) {
	var result model2.Notes
	if err := dao.Db.First(&result, ID).Error; err != nil {
		return nil, err
	}
	return &result, nil
}

//GetByServiceName - Get Notes By ServicesName
func (dao NotesDAOImpl) GetByServiceName(serviceName string) (*model2.Notes, error) {
	var result model2.Notes
	if err := dao.Db.Where(model2.Notes{ServiceName: serviceName}).First(&result).Error; err != nil {
		return nil, err
	}
	return &result, nil
}
