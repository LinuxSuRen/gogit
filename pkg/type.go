package pkg

import "fmt"

type RepoInformation struct {
	Provider string
	Server   string
	Owner    string
	Repo     string
	PrNumber int

	Username, Token string

	Status      string
	Target      string
	Label       string
	Description string
}

func (r RepoInformation) getRepoPath() string {
	return fmt.Sprintf("%s/%s", r.Owner, r.Repo)
}
