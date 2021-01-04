package handlers

import (
	"bytes"
	"fmt"

	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/softplan/tenkai-api/pkg/global"
)


func (appContext *AppContext) helmInstall(chartsToDeploy []model.InstallPayload, environment model.Environment, user model.User, out *bytes.Buffer, dryRun, helmCommandOnly bool) error{
	logFields := global.AppFields{global.Function: "helmInstall"}
	global.Logger.Info(logFields, "helmInstall - begin")
	
	requestDeployment := model.RequestDeployment{}
	requestDeployment.Success = false
	requestDeployment.Processed = false
	requestDeployment.UserID = user.ID
	requestDeploymentID, err := appContext.Repositories.RequestDeploymentDAO.CreateRequestDeployment(requestDeployment)
	if err != nil {
		global.Logger.Info(logFields, "helmInstall - end")
		return err
	}

	for _, chart := range chartsToDeploy {
		
		appContext.simpleInstall(&environment, chart, out, dryRun, helmCommandOnly, fmt.Sprint(user.ID), requestDeploymentID)
	}
	global.Logger.Info(logFields, "helmInstall - end")
	return nil
}