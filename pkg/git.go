package pkg

import (
	"context"
	"fmt"
	"github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/go-scm/scm/factory"
)

// Reconcile is the main entry of this reconciler
func Reconcile(ctx context.Context, repoInfo RepoInformation) (err error) {
	if repoInfo.PrNumber == -1 {
		fmt.Println("skip due to pr number is -1")
		return
	}

	repo := repoInfo.getRepoPath()
	maker := NewStatusMaker(repo, repoInfo.Token)
	maker.WithTarget(repoInfo.Target).WithPR(repoInfo.PrNumber).
		WithServer(repoInfo.Server).
		WithProvider(repoInfo.Provider).
		WithUsername(repoInfo.Username).
		WithToken(repoInfo.Token)

	err = maker.Create(ctx, scm.ToState(repoInfo.Status), repoInfo.Label, repoInfo.Description)
	return
}

// StatusMaker responsible for Pull Requests status creating
type StatusMaker struct {
	provider string
	server   string
	repo     string
	pr       int
	token    string
	username string
	target   string

	// expirationCheck checks if the current status is expiration that compared to the previous one
	expirationCheck expirationCheckFunc
}

// NewStatusMaker creates an instance of statusMaker
func NewStatusMaker(repo, token string) *StatusMaker {
	return &StatusMaker{
		repo:  repo,
		token: token,
		expirationCheck: func(previousStatus *scm.Status, currentStatus *scm.StatusInput) bool {
			return previousStatus != nil && previousStatus.State == currentStatus.State
		},
	}
}

type expirationCheckFunc func(previousStatus *scm.Status, currentStatus *scm.StatusInput) bool

// WithExpirationCheck set the expiration check function
func (s *StatusMaker) WithExpirationCheck(check expirationCheckFunc) *StatusMaker {
	s.expirationCheck = check
	return s
}

// WithUsername sets the username
func (s *StatusMaker) WithUsername(username string) *StatusMaker {
	s.username = username
	return s
}

// WithToken sets the token
func (s *StatusMaker) WithToken(token string) *StatusMaker {
	s.token = token
	return s
}

// WithProvider sets the Provider
func (s *StatusMaker) WithProvider(provider string) *StatusMaker {
	s.provider = provider
	return s
}

// WithServer sets the server
func (s *StatusMaker) WithServer(server string) *StatusMaker {
	s.server = server
	return s
}

// WithTarget sets the Target URL
func (s *StatusMaker) WithTarget(target string) *StatusMaker {
	s.target = target
	return s
}

// WithPR sets the pr number
func (s *StatusMaker) WithPR(pr int) *StatusMaker {
	s.pr = pr
	return s
}

// Create creates a generic status
func (s *StatusMaker) Create(ctx context.Context, status scm.State, label, desc string) (err error) {
	var scmClient *scm.Client
	scmClient, err = factory.NewClient(s.provider, s.server, s.token, func(c *scm.Client) {
		c.Username = s.username
	})
	if err != nil {
		return
	}

	var pullRequest *scm.PullRequest
	if pullRequest, _, err = scmClient.PullRequests.Find(ctx, s.repo, s.pr); err == nil {
		var previousStatus *scm.Status
		if previousStatus, err = s.FindPreviousStatus(ctx, scmClient, pullRequest.Sha, label); err != nil {
			return
		}

		currentStatus := &scm.StatusInput{
			Desc:   desc,
			Label:  label,
			State:  status,
			Target: s.target,
		}
		// avoid the previous building status override newer one
		if !s.expirationCheck(previousStatus, currentStatus) {
			_, _, err = scmClient.Repositories.CreateStatus(ctx, s.repo, pullRequest.Sha, currentStatus)
		}
	}
	return
}

// FindPreviousStatus finds the existing status by sha and label
func (s *StatusMaker) FindPreviousStatus(ctx context.Context, scmClient *scm.Client, sha, label string) (target *scm.Status, err error) {
	var exists []*scm.Status
	if exists, _, err = scmClient.Repositories.ListStatus(ctx, s.repo, sha, &scm.ListOptions{
		Page: 1,
		Size: 100, // assume this list has not too many items
	}); err != nil {
		err = fmt.Errorf("failed to list the existing status, error: %v", err)
		return
	}

	for _, item := range exists {
		if item.Label == label {
			target = item
			break
		}
	}
	return
}
