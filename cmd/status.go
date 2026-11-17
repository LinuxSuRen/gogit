package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/go-scm/scm/factory"
	"github.com/linuxsuren/gogit/pkg"
	"github.com/spf13/cobra"
)

func newStatusCmd() (cmd *cobra.Command) {
	opt := &statusOption{}
	cmd = &cobra.Command{
		Use:     "status",
		Short:   "Send the build status to a PR of Gitlab/GitHub",
		PreRunE: opt.preRunE,
		RunE:    opt.runE,
	}

	opt.addFlags(cmd)
	flags := cmd.Flags()
	flags.StringVarP(&opt.status, "status", "", "",
		"Build status, such as: pending, success, cancelled, error")
	flags.StringVarP(&opt.target, "target", "", "https://github.com/LinuxSuRen/gogit", "Address of the build server")
	flags.StringVarP(&opt.label, "label", "", "",
		"Identity of a build status")
	flags.StringVarP(&opt.description, "description", "", "",
		"The description of a build status")
	flags.BoolVarP(&opt.print, "print", "", false, "Print the status list then exit")
	return
}

// statusAliases keeps the status be compatible with different systems
var statusAliases = map[string]string{
	"succeeded": "success", // from Argo Workflows
	"failed":    "failure",
	"canceled":  "cancelled",
}

func (o *statusOption) preRunE(cmd *cobra.Command, args []string) (err error) {
	o.preHandle()
	if o.label == "" {
		o.label = "gogit"
	}

	// keep the status be compatible with different system
	o.status = strings.ToLower(o.status)
	o.status = strings.TrimSpace(o.status)
	if alias, ok := statusAliases[o.status]; ok {
		o.status = alias
	}

	if o.status != "" {
		if scm.ToState(o.status) == scm.StateUnknown {
			err = fmt.Errorf("invalid build status %q, please choose from: pending, running, success, failure, cancelled, expected, error", o.status)
			return
		}
	}

	if strings.HasPrefix(o.token, "file://") {
		tokenFile := strings.TrimPrefix(o.token, "file://")
		var data []byte
		if data, err = os.ReadFile(tokenFile); err != nil {
			err = fmt.Errorf("failed to read the token file %s: %v", tokenFile, err)
			return
		}
		// token files usually carry a trailing newline which breaks the authentication
		o.token = strings.TrimSpace(string(data))
	}
	return
}

func (o *statusOption) runE(cmd *cobra.Command, args []string) (err error) {
	if o.print {
		maker := pkg.NewMaker(cmd.Context(), pkg.RepoInformation{
			Provider:    o.provider,
			Server:      o.server,
			Owner:       o.owner,
			Repo:        o.repo,
			PrNumber:    o.pr,
			Target:      o.target,
			Username:    o.username,
			Token:       o.token,
			Status:      o.status,
			Label:       o.label,
			Description: o.description,
		})
		if maker != nil {
			err = maker.ListStatus(cmd.Context(), o.label, o.description)
		}
		return
	}

	err = pkg.CreateStatus(cmd.Context(), pkg.RepoInformation{
		Provider:    o.provider,
		Server:      o.server,
		Owner:       o.owner,
		Repo:        o.repo,
		PrNumber:    o.pr,
		Target:      o.target,
		Username:    o.username,
		Token:       o.token,
		Status:      o.status,
		Label:       o.label,
		Description: o.description,
	})
	return
}

type gitProviderOption struct {
	provider string
	server   string
	username string
	token    string
	owner    string
	repo     string
	pr       int
}

func (o *gitProviderOption) addFlags(c *cobra.Command) {
	flags := c.Flags()
	flags.StringVarP(&o.provider, "provider", "p", "github", "The provider of git, such as: gitlab, github")
	flags.StringVarP(&o.server, "server", "s", "", "The server address of target git provider, only need when it's a private provider")
	flags.StringVarP(&o.owner, "owner", "o", "", "Owner of a git repository")
	flags.StringVarP(&o.repo, "repo", "r", "", "Name of target git repository")
	flags.IntVarP(&o.pr, "pr", "", 1, "The pull request number")
	flags.StringVarP(&o.username, "username", "u", "", "Username of the git repository")
	flags.StringVarP(&o.token, "token", "t", "",
		"The access token of the git repository. Or you could provide a file path, such as: file:///var/token")

	_ = c.MarkFlagRequired("repo")
	_ = c.MarkFlagRequired("pr")
	_ = c.MarkFlagRequired("username")
	_ = c.MarkFlagRequired("token")
}

func (o *gitProviderOption) preHandle() {
	if o.owner == "" {
		o.owner = o.username
	}
}

func (o *gitProviderOption) getClient() (scmClient *scm.Client, err error) {
	scmClient, err = factory.NewClient(o.provider, o.server, o.token, func(c *scm.Client) {
		c.Username = o.username
	})
	return
}

type statusOption struct {
	gitProviderOption
	status      string
	target      string
	label       string
	description string
	print       bool
}
