package repository

import (
	"github.com/jinzhu/gorm"
	model2 "github.com/softplan/tenkai-api/pkg/dbms/model"
)

//ProductDAOInterface ProductDAOInterface
type ProductDAOInterface interface {
	CreateProduct(e model2.Product) (int, error)
	CreateProductVersion(e model2.ProductVersion) (int, error)
	CreateProductVersionService(e model2.ProductVersionService) (int, error)
	EditProduct(e model2.Product) error
	EditProductVersion(e model2.ProductVersion) error
	EditProductVersionService(e model2.ProductVersionService) error
	DeleteProduct(id int) error
	DeleteProductVersion(id int) error
	DeleteProductVersionService(id int) error
	ListProducts() ([]model2.Product, error)
	FindProductByID(id int) (model2.Product, error)
	ListProductsVersions(id int) ([]model2.ProductVersion, error)
	ListProductVersionsServiceByID(id int) (*model2.ProductVersionService, error)
	ListProductsVersionServices(id int) ([]model2.ProductVersionService, error)
	CreateProductVersionCopying(payload model2.ProductVersion) (int, error)
	ListProductVersionsByID(id int) (*model2.ProductVersion, error)
}

//ProductDAOImpl ProductDAOImpl
type ProductDAOImpl struct {
	Db *gorm.DB
}

// CreateProductVersionCopying create a new version product version
func (dao ProductDAOImpl) CreateProductVersionCopying(payload model2.ProductVersion) (int, error) {
	id, err := dao.CreateProductVersion(payload)
	if err != nil {
		return -1, err
	}

	if payload.BaseRelease > 0 {
		list, err := dao.ListProductsVersionServices(payload.BaseRelease)
		if err != nil {
			return -1, err
		}
		var pvs *model2.ProductVersionService
		for _, l := range list {
			pvs = &model2.ProductVersionService{}
			pvs.ProductVersionID = id
			pvs.ServiceName = l.ServiceName
			pvs.DockerImageTag = l.DockerImageTag

			if _, err := dao.CreateProductVersionService(*pvs); err != nil {
				return -1, err
			}
		}
	}
	return id, nil
}

//CreateProduct - Create a new product
func (dao ProductDAOImpl) CreateProduct(e model2.Product) (int, error) {
	if err := dao.Db.Create(&e).Error; err != nil {
		return -1, err
	}
	return int(e.ID), nil
}

//CreateProductVersion - Create a new product version
func (dao ProductDAOImpl) CreateProductVersion(e model2.ProductVersion) (int, error) {
	if err := dao.Db.Create(&e).Error; err != nil {
		return -1, err
	}
	return int(e.ID), nil
}

//CreateProductVersionService - Create a new product version
func (dao ProductDAOImpl) CreateProductVersionService(e model2.ProductVersionService) (int, error) {
	if err := dao.Db.Create(&e).Error; err != nil {
		return -1, err
	}
	return int(e.ID), nil
}

//EditProduct - Updates an existing product
func (dao ProductDAOImpl) EditProduct(e model2.Product) error {
	return dao.Db.Save(&e).Error
}

//EditProductVersion - Updates an existing product version
func (dao ProductDAOImpl) EditProductVersion(e model2.ProductVersion) error {
	return dao.Db.Save(&e).Error
}

//EditProductVersionService - Updates an existing product version
func (dao ProductDAOImpl) EditProductVersionService(e model2.ProductVersionService) error {
	return dao.Db.Save(&e).Error
}

//DeleteProduct - Deletes a product
func (dao ProductDAOImpl) DeleteProduct(id int) error {
	return dao.Db.Unscoped().Delete(model2.Product{}, id).Error
}

//DeleteProductVersion - Deletes a productVersion
func (dao ProductDAOImpl) DeleteProductVersion(id int) error {
	return dao.Db.Unscoped().Delete(model2.ProductVersion{}, id).Error
}

//DeleteProductVersionService - Deletes a productVersionService
func (dao ProductDAOImpl) DeleteProductVersionService(id int) error {
	return dao.Db.Unscoped().Delete(model2.ProductVersionService{}, id).Error
}

//ListProducts - List products
func (dao ProductDAOImpl) ListProducts() ([]model2.Product, error) {
	list := make([]model2.Product, 0)
	if err := dao.Db.Find(&list).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return make([]model2.Product, 0), nil
		}
		return nil, err
	}
	return list, nil
}

//FindProductByID - FindProductByID
func (dao ProductDAOImpl) FindProductByID(id int) (model2.Product, error) {
	var result model2.Product
	if err := dao.Db.First(&result, id).Error; err != nil {
		return result, err
	}
	return result, nil
}

//ListProductVersionsServiceByID - ListProductVersionsServiceByID
func (dao ProductDAOImpl) ListProductVersionsServiceByID(id int) (*model2.ProductVersionService, error) {
	var result model2.ProductVersionService
	if err := dao.Db.First(&result, id).Error; err != nil {
		return nil, err
	}
	return &result, nil
}

//ListProductsVersions - List products versions
func (dao ProductDAOImpl) ListProductsVersions(id int) ([]model2.ProductVersion, error) {
	list := make([]model2.ProductVersion, 0)
	if err := dao.Db.Where(&model2.ProductVersion{ProductID: id}).Find(&list).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return make([]model2.ProductVersion, 0), nil
		}
		return nil, err
	}
	return list, nil
}

//ListProductsVersionServices - List products versions
func (dao ProductDAOImpl) ListProductsVersionServices(id int) ([]model2.ProductVersionService, error) {

	list := make([]model2.ProductVersionService, 0)

	var ID uint
	var serviceName string
	var dockerImageTag string
	var notes string

	rows, err := dao.Db.Table("product_version_services").Select("product_version_services.id, " +
		" product_version_services.service_name, " +
		" product_version_services.docker_image_tag, notes.text").Joins("LEFT JOIN notes on product_version_services.service_name = notes.service_name").Where(&model2.ProductVersionService{ProductVersionID: id}).Rows()

	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return make([]model2.ProductVersionService, 0), nil
		}
		return nil, err
	}

	for rows.Next() {
		notes = ""
		ID = 0
		serviceName = ""
		dockerImageTag = ""
		notes = ""
		rows.Scan(&ID, &serviceName, &dockerImageTag, &notes)
		p := &model2.ProductVersionService{ServiceName: serviceName, DockerImageTag: dockerImageTag, Notes: notes}
		p.ID = ID
		list = append(list, *p)
	}
	return list, nil
}

//ListProductVersionsByID - ListProductVersionsByID
func (dao ProductDAOImpl) ListProductVersionsByID(id int) (*model2.ProductVersion, error) {
	var result model2.ProductVersion
	if err := dao.Db.First(&result, id).Error; err != nil {
		return nil, err
	}
	return &result, nil
}
