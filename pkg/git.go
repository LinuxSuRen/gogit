package pkg

import (
	"context"
	"fmt"
	"strings"

	"github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/go-scm/scm/factory"
)

// CreateStatus is the main entry of this reconciler
func CreateStatus(ctx context.Context, repoInfo RepoInformation) (err error) {
	if maker := NewMaker(ctx, repoInfo); maker != nil {
		err = maker.CreateStatus(ctx, scm.ToState(repoInfo.Status), repoInfo.Label, repoInfo.Description)
	}
	return
}

// CreateComment creates a comment against the pull request
//
// It will update the comment there is a comment has the same ender
func CreateComment(ctx context.Context, repoInfo RepoInformation, message, identity string) (err error) {
	if maker := NewMaker(ctx, repoInfo); maker != nil {
		err = maker.CreateComment(ctx, message, identity)
	}
	return
}

// NewMaker creates a maker
func NewMaker(ctx context.Context, repoInfo RepoInformation) (maker *StatusMaker) {
	if repoInfo.PrNumber == -1 {
		fmt.Println("skip due to pr number is -1")
		return
	}

	repo := repoInfo.GetRepoPath()
	maker = NewStatusMaker(repo, repoInfo.Token)
	maker.WithTarget(repoInfo.Target).WithPR(repoInfo.PrNumber).
		WithServer(repoInfo.Server).
		WithProvider(repoInfo.Provider).
		WithUsername(repoInfo.Username).
		WithToken(repoInfo.Token)
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

// CommentEndMarker is the identify for matching existing comment
const CommentEndMarker = "Comment from [gogit](https://github.com/linuxsuren/gogit)."

// CreateComment creates a comment
func (s *StatusMaker) CreateComment(ctx context.Context, message, endMarker string) (err error) {
	var scmClient *scm.Client
	if scmClient, err = factory.NewClient(s.provider, s.server, s.token, func(c *scm.Client) {
		c.Username = s.username
	}); err != nil {
		return
	}

	var comments []*scm.Comment
	if comments, _, err = scmClient.PullRequests.ListComments(ctx, s.repo, s.pr, &scm.ListOptions{
		Page: 1,
		Size: 100,
	}); err != nil {
		if err = IgnoreError(err, "Not Found"); err != nil {
			err = fmt.Errorf("cannot any comments %v", err)
			return
		}
	}

	commentIDs := getCommentIDs(comments, endMarker)
	commentInput := &scm.CommentInput{
		Body: fmt.Sprintf("%s\n\n%s", message, endMarker),
	}

	if len(commentIDs) == 0 {
		// not found existing comment, create a new one
		_, _, err = scmClient.PullRequests.CreateComment(ctx, s.repo, s.pr, commentInput)
		err = WrapError(err, "failed to create comment, repo is %q, pr is %d: %v", s.repo, s.pr)
	} else {
		_, _, err = scmClient.PullRequests.EditComment(ctx, s.repo, s.pr, commentIDs[0], commentInput)
		err = WrapError(err, "failed to edit comment: %v")

		// remove the duplicated comments
		for i := 1; i < len(commentIDs); i++ {
			_, _ = scmClient.PullRequests.DeleteComment(ctx, s.repo, s.pr, commentIDs[i])
		}
	}
	return
}

func getCommentIDs(comments []*scm.Comment, endMarker string) (commentIDs []int) {
	for i := range comments {
		comment := comments[i]
		if strings.HasSuffix(comment.Body, endMarker) {
			commentIDs = append(commentIDs, comment.ID)
		}
	}
	return
}

// CreateStatus creates a generic status
func (s *StatusMaker) CreateStatus(ctx context.Context, status scm.State, label, desc string) (err error) {
	var scmClient *scm.Client
	if scmClient, err = factory.NewClient(s.provider, s.server, s.token, func(c *scm.Client) {
		c.Username = s.username
	}); err != nil {
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
	} else {
		err = fmt.Errorf("failed to find pull requests %v", err)
	}
	return
}

// ListStatus list the status
func (s *StatusMaker) ListStatus(ctx context.Context, label, desc string) (err error) {
	var scmClient *scm.Client
	if scmClient, err = factory.NewClient(s.provider, s.server, s.token, func(c *scm.Client) {
		c.Username = s.username
	}); err != nil {
		return
	}

	var pullRequest *scm.PullRequest
	if pullRequest, _, err = scmClient.PullRequests.Find(ctx, s.repo, s.pr); err == nil {
		var exists []*scm.Status
		if exists, _, err = scmClient.Repositories.ListStatus(ctx, s.repo, pullRequest.Sha, &scm.ListOptions{
			Page: 1,
			Size: 100, // assume this list has not too many items
		}); err != nil {
			err = fmt.Errorf("failed to list the existing status, error: %v", err)
			return
		}

		for _, item := range exists {
			if item.Label == label {
				fmt.Println(item.State)
			}
		}
	} else {
		err = fmt.Errorf("failed to find pull requests [%d] from [%s] %v", s.pr, s.repo, err)
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
