package dbms

import (
	"github.com/jinzhu/gorm"
	"github.com/softplan/tenkai-api/dbms/model"
)

//CreateProduct - Create a new product
func (database *Database) CreateProduct(e model.Product) (int, error) {
	if err := database.Db.Create(&e).Error; err != nil {
		return -1, err
	}
	return int(e.ID), nil
}

//CreateProductVersion - Create a new product version
func (database *Database) CreateProductVersion(e model.ProductVersion) (int, error) {
	if err := database.Db.Create(&e).Error; err != nil {
		return -1, err
	}
	return int(e.ID), nil
}

//CreateProductVersionService - Create a new product version
func (database *Database) CreateProductVersionService(e model.ProductVersionService) (int, error) {
	if err := database.Db.Create(&e).Error; err != nil {
		return -1, err
	}
	return int(e.ID), nil
}

//EditProduct - Updates an existing product
func (database *Database) EditProduct(e model.Product) error {
	if err := database.Db.Save(&e).Error; err != nil {
		return err
	}
	return nil
}

//EditProductVersion - Updates an existing product version
func (database *Database) EditProductVersion(e model.ProductVersion) error {
	if err := database.Db.Save(&e).Error; err != nil {
		return err
	}
	return nil
}

//EditProductVersionService - Updates an existing product version
func (database *Database) EditProductVersionService(e model.ProductVersionService) error {
	if err := database.Db.Save(&e).Error; err != nil {
		return err
	}
	return nil
}

//DeleteProduct - Deletes a product
func (database *Database) DeleteProduct(id int) error {
	if err := database.Db.Unscoped().Delete(model.Product{}, id).Error; err != nil {
		return err
	}
	return nil
}

//DeleteProductVersion - Deletes a productVersion
func (database *Database) DeleteProductVersion(id int) error {
	if err := database.Db.Unscoped().Delete(model.ProductVersion{}, id).Error; err != nil {
		return err
	}
	return nil
}

//DeleteProductVersionService - Deletes a productVersionService
func (database *Database) DeleteProductVersionService(id int) error {
	if err := database.Db.Unscoped().Delete(model.ProductVersionService{}, id).Error; err != nil {
		return err
	}
	return nil
}

//ListProducts - List products
func (database *Database) ListProducts() ([]model.Product, error) {
	list := make([]model.Product, 0)
	if err := database.Db.Find(&list).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return make([]model.Product, 0), nil
		}
		return nil, err
	}
	return list, nil
}

//ListProductsVersions - List products versions
func (database *Database) ListProductsVersions(id int) ([]model.ProductVersion, error) {
	list := make([]model.ProductVersion, 0)
	if err := database.Db.Where(&model.ProductVersion{ProductID: id}).Find(&list).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return make([]model.ProductVersion, 0), nil
		}
		return nil, err
	}
	return list, nil
}

//ListProductsVersionServices - List products versions
func (database *Database) ListProductsVersionServices(id int) ([]model.ProductVersionService, error) {
	list := make([]model.ProductVersionService, 0)
	if err := database.Db.Where(&model.ProductVersionService{ProductVersionID: id}).Find(&list).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return make([]model.ProductVersionService, 0), nil
		}
		return nil, err
	}
	return list, nil
}

//ListProductVersionServicesLatest - List from the latest Product Version
func (database *Database) ListProductVersionServicesLatest(productID, productVersionID int) ([]model.ProductVersionService, error) {
	item := model.ProductVersion{}
	list := make([]model.ProductVersionService, 0)

	if err := database.Db.Where(&model.ProductVersion{ProductID: productID}).Not("id", productVersionID).Order("created_at desc").Limit(1).Find(&item).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return make([]model.ProductVersionService, 0), nil
		}
		return list, err
	}

	list, err := database.ListProductsVersionServices(int(item.ID))
	if err != nil {
		return list, err
	}

	return list, nil
}
