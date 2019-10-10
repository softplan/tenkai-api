package productsrv

import (
	"github.com/softplan/tenkai-api/dbms"
	"github.com/softplan/tenkai-api/dbms/model"
)

// CreateProductVersion create a new version product version
func CreateProductVersion(dbms dbms.Database, payload model.ProductVersion) (int, error) {
	id, err := dbms.CreateProductVersion(payload)
	if err != nil {
		return -1, err
	}

	if payload.CopyLatestRelease {
		list, err := dbms.ListProductVersionServicesLatest(payload.ProductID, id)
		if err != nil {
			return -1, err
		}
		var pvs *model.ProductVersionService
		for _, l := range list {
			pvs = &model.ProductVersionService{}
			pvs.ProductVersionID = id
			pvs.ServiceName = l.ServiceName
			pvs.DockerImageTag = l.DockerImageTag

			if _, err := dbms.CreateProductVersionService(*pvs); err != nil {
				return -1, err
			}
		}
	}
	return id, nil
}
