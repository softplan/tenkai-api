package model

//Repository struct
type Repository struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
}

//RepositoryResult struct
type RepositoryResult struct {
	Repositories []Repository `json:"repositories"`
}

//DefaultRepoRequest struct
type DefaultRepoRequest struct {
	Reponame string `json:"reponame"`
}
