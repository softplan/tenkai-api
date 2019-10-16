package helmapi

import (
	"fmt"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/softplan/tenkai-api/pkg/global"
	"io"
	"os"

	"k8s.io/helm/pkg/getter"
	"k8s.io/helm/pkg/helm/helmpath"
	"k8s.io/helm/pkg/repo"
)

type repoAddCmd struct {
	name     string
	url      string
	username string
	password string
	home     helmpath.Home
	noupdate bool
	certFile string
	keyFile  string
	caFile   string
	out      io.Writer
}

//AddRepository - Add new repository
func (svc HelmServiceImpl) AddRepository(repo model.Repository) error {
	settings.Home = global.HelmDir
	add := &repoAddCmd{out: os.Stdout}
	add.name = repo.Name
	add.url = repo.URL
	add.username = repo.Username
	add.password = repo.Password
	add.home = settings.Home
	add.caFile = ""
	return add.run()
}

func (a *repoAddCmd) run() error {

	if a.username != "" && a.password == "" {
		return fmt.Errorf("password must be te for user %q", a.username)
	}

	if err := addRepository(a.name, a.url, a.username, a.password, a.home, a.certFile, a.keyFile, a.caFile, a.noupdate); err != nil {
		return err
	}
	fmt.Fprintf(a.out, "%q has been added to your repositories\n", a.name)
	return nil
}

func addRepository(name, url, username, password string, home helmpath.Home, certFile, keyFile, caFile string, noUpdate bool) error {
	f, err := repo.LoadRepositoriesFile(home.RepositoryFile())
	if err != nil {
		return err
	}

	if noUpdate && f.Has(name) {
		return fmt.Errorf("repository name (%s) already exists, please specify a different name", name)
	}

	cif := home.CacheIndex(name)
	c := repo.Entry{
		Name:     name,
		Cache:    cif,
		URL:      url,
		Username: username,
		Password: password,
		CertFile: certFile,
		KeyFile:  keyFile,
		CAFile:   caFile,
	}

	r, err := repo.NewChartRepository(&c, getter.All(settings))
	if err != nil {
		return err
	}

	if err := r.DownloadIndexFile(""); err != nil {
		return fmt.Errorf("Looks like %q is not a valid chart repository or cannot be reached: %s", url, err.Error())
	}

	c.Cache = "./" + c.Name + "-index.yaml"
	f.Update(&c)

	return f.WriteFile(home.RepositoryFile(), 0644)
}
