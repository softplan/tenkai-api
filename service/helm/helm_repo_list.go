package helmapi

import (
	"errors"
	"github.com/softplan/tenkai-api/dbms/model"
	"github.com/softplan/tenkai-api/global"
	"k8s.io/helm/pkg/repo"
)

//GetRepositories - Returns a repository list
func GetRepositories() ([]model.Repository, error) {

	settings.Home = global.HELM_DIR

	logFields := global.AppFields{global.FUNCTION: "GetRepositories"}

	global.Logger.Info(logFields, "Starting GetRepositories: "+settings.Home.RepositoryFile())

	var repositories []model.Repository

	f, err := repo.LoadRepositoriesFile(settings.Home.RepositoryFile())
	if err != nil {
		global.Logger.Error(logFields, "Error loading repositories "+err.Error())
		return nil, err
	}
	if len(f.Repositories) == 0 {
		global.Logger.Error(logFields, "no repositories to show")
		return nil, errors.New("no repositories to show")
	}

	for _, re := range f.Repositories {
		rep := &model.Repository{Name: re.Name, URL: re.URL, Username: re.Username, Password: re.Password}
		repositories = append(repositories, *rep)
	}

	global.Logger.Info(logFields, "Returning repositories")

	return repositories, nil
}
