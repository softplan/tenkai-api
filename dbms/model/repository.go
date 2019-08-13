package model

type Repository struct {
	Name     string `json:"name"`
	Url      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type RepositoryResult struct {
	Repositories []Repository `json:"repositories"`
}

type DefaultRepoRequest struct {
	Reponame string `json:"reponame"`
}
