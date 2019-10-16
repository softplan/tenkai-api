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
	ListProductsVersions(id int) ([]model2.ProductVersion, error)
	ListProductsVersionServices(id int) ([]model2.ProductVersionService, error)
	ListProductVersionServicesLatest(productID, productVersionID int) ([]model2.ProductVersionService, error)
	CreateProductVersionCopying(payload model2.ProductVersion) (int, error)
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

	if payload.CopyLatestRelease {
		list, err := dao.ListProductVersionServicesLatest(payload.ProductID, id)
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
	if err := dao.Db.Save(&e).Error; err != nil {
		return err
	}
	return nil
}

//EditProductVersion - Updates an existing product version
func (dao ProductDAOImpl) EditProductVersion(e model2.ProductVersion) error {
	if err := dao.Db.Save(&e).Error; err != nil {
		return err
	}
	return nil
}

//EditProductVersionService - Updates an existing product version
func (dao ProductDAOImpl) EditProductVersionService(e model2.ProductVersionService) error {
	if err := dao.Db.Save(&e).Error; err != nil {
		return err
	}
	return nil
}

//DeleteProduct - Deletes a product
func (dao ProductDAOImpl) DeleteProduct(id int) error {
	if err := dao.Db.Unscoped().Delete(model2.Product{}, id).Error; err != nil {
		return err
	}
	return nil
}

//DeleteProductVersion - Deletes a productVersion
func (dao ProductDAOImpl) DeleteProductVersion(id int) error {
	if err := dao.Db.Unscoped().Delete(model2.ProductVersion{}, id).Error; err != nil {
		return err
	}
	return nil
}

//DeleteProductVersionService - Deletes a productVersionService
func (dao ProductDAOImpl) DeleteProductVersionService(id int) error {
	if err := dao.Db.Unscoped().Delete(model2.ProductVersionService{}, id).Error; err != nil {
		return err
	}
	return nil
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
	if err := dao.Db.Where(&model2.ProductVersionService{ProductVersionID: id}).Find(&list).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return make([]model2.ProductVersionService, 0), nil
		}
		return nil, err
	}
	return list, nil
}

//ListProductVersionServicesLatest - List from the latest Product Version
func (dao ProductDAOImpl) ListProductVersionServicesLatest(productID, productVersionID int) ([]model2.ProductVersionService, error) {
	item := model2.ProductVersion{}
	list := make([]model2.ProductVersionService, 0)

	if err := dao.Db.Where(&model2.ProductVersion{ProductID: productID}).Not("id", productVersionID).Order("created_at desc").Limit(1).Find(&item).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return make([]model2.ProductVersionService, 0), nil
		}
		return list, err
	}

	list, err := dao.ListProductsVersionServices(int(item.ID))
	if err != nil {
		return list, err
	}

	return list, nil
}
