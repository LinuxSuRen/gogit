package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/spf13/cobra"
)

func newCheckoutCommand() (c *cobra.Command) {
	opt := &checkoutOption{}

	c = &cobra.Command{
		Use:     "checkout",
		Aliases: []string{"co"},
		Short:   "Clone and checkout the git repository with branch, tag, or pull request",
		Example: "gogit checkout https://github.com/linuxsuren/gogit",
		PreRunE: opt.preRunE,
		RunE:    opt.runE,
	}

	flags := c.Flags()
	flags.StringVarP(&opt.url, "url", "", "", "The git repository URL")
	flags.StringVarP(&opt.remote, "remote", "", "origin", "The remote name")
	flags.StringVarP(&opt.branch, "branch", "", "master", "The branch want to checkout")
	flags.StringVarP(&opt.tag, "tag", "", "", "The tag want to checkout")
	flags.IntVarP(&opt.pr, "pr", "", -1, "The pr number want to checkout, -1 means do nothing")
	flags.StringVarP(&opt.target, "target", "", ".", "Clone git repository to the target path")
	flags.StringVarP(&opt.versionOutput, "version-output", "", "", "Write the version to target file")
	return
}

func (o *checkoutOption) preRunE(c *cobra.Command, args []string) (err error) {
	if o.url == "" && len(args) > 0 {
		o.url = args[0]
	}
	return
}

func (o *checkoutOption) runE(c *cobra.Command, args []string) (err error) {
	var repoDir string
	if repoDir, err = filepath.Abs(o.target); err != nil {
		return
	}

	var gitAuth transport.AuthMethod
	if gitAuth, err = getAuth(o.url); err != nil {
		return
	}

	if _, serr := os.Stat(filepath.Join(repoDir, ".git")); serr != nil {
		if _, err = git.PlainClone(repoDir, false, &git.CloneOptions{
			RemoteName:    o.remote,
			Auth:          gitAuth,
			URL:           o.url,
			ReferenceName: plumbing.NewBranchReferenceName(o.branch),
			Progress:      c.OutOrStdout(),
		}); err != nil {
			err = fmt.Errorf("failed to clone git repository '%s' into '%s', error: %v", o.url, repoDir, err)
			return
		}
	}

	var repo *git.Repository
	if repo, err = git.PlainOpen(repoDir); err == nil {
		var wd *git.Worktree
		var remotes []*git.Remote

		if remotes, err = repo.Remotes(); err != nil {
			return
		}

		remoteURL := remotes[0].Config().URLs[0]
		kind := detectGitKind(remoteURL)
		// need to get auth again if the repo was exist
		if gitAuth, err = getAuth(remoteURL); err != nil {
			return
		}

		if wd, err = repo.Worktree(); err == nil {
			if o.tag != "" {
				if err = wd.Checkout(&git.CheckoutOptions{
					Branch: plumbing.NewTagReferenceName(o.tag),
				}); err != nil {
					err = fmt.Errorf("unable to checkout git branch: %s, error: %v", o.tag, err)
					return
				}
			}

			if o.pr > 0 {
				if err = repo.Fetch(&git.FetchOptions{
					RemoteName: o.remote,
					Auth:       gitAuth,
					Progress:   c.OutOrStdout(),
					RefSpecs:   []config.RefSpec{config.RefSpec(prRef(o.pr, kind))},
				}); err != nil && err != git.NoErrAlreadyUpToDate {
					err = fmt.Errorf("failed to fetch '%s', error: %v", o.remote, err)
					return
				}

				if err = wd.Checkout(&git.CheckoutOptions{
					Create: true,
					Branch: plumbing.NewBranchReferenceName(fmt.Sprintf("pr-%d", o.pr)),
				}); err != nil && !strings.Contains(err.Error(), "already exists") {
					err = fmt.Errorf("unable to checkout git branch: %s, error: %v", o.tag, err)
					return
				}
			}

			var head *plumbing.Reference
			if head, err = repo.Head(); err == nil {
				if o.versionOutput != "" {
					err = os.WriteFile(o.versionOutput, []byte(head.Name().Short()), 0444)
				}
			}
		}
	}
	return
}

func getAuth(remote string) (auth transport.AuthMethod, err error) {
	if strings.HasPrefix(remote, "git@") {
		rsa := os.ExpandEnv("$HOME/.ssh/id_rsa")
		auth, err = ssh.NewPublicKeysFromFile("git", rsa, "")
	}
	return
}

func detectGitKind(gitURL string) (kind string) {
	kind = "gitlab"
	if strings.Contains(gitURL, "github.com") {
		kind = "github"
	}
	return
}

// see also https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/reviewing-changes-in-pull-requests/checking-out-pull-requests-locally?gt
func prRef(pr int, kind string) (ref string) {
	switch kind {
	case "gitlab":
		ref = fmt.Sprintf("refs/merge-requests/%d/head:pr-%d", pr, pr)
	case "github":
		ref = fmt.Sprintf("refs/pull/%d/head:pr-%d", pr, pr)
	}
	return
}

type checkoutOption struct {
	url           string
	remote        string
	branch        string
	tag           string
	pr            int
	target        string
	versionOutput string
}
