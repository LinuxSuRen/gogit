package cmd

import (
	"github.com/linuxsuren/gogit/pkg"
	"github.com/spf13/cobra"
	"strings"
)

func NewBuildCmd() (cmd *cobra.Command) {
	opt := &option{}
	cmd = &cobra.Command{
		Use:    "gogit",
		Short:  "Send the build status to a PR of Gitlab/GitHub",
		PreRun: opt.preRun,
		RunE:   opt.runE,
	}

	flags := cmd.Flags()
	flags.StringVarP(&opt.provider, "provider", "p", "github", "The provider of git, such as: gitlab, github")
	flags.StringVarP(&opt.server, "server", "s", "", "The server address of target git provider, only need when it's a private provider")
	flags.StringVarP(&opt.owner, "owner", "o", "", "Owner of a git repository")
	flags.StringVarP(&opt.repo, "repo", "r", "", "Name of target git repository")
	flags.IntVarP(&opt.pr, "pr", "", 1, "The pull request number")
	flags.StringVarP(&opt.username, "username", "u", "", "Username of the git repository")
	flags.StringVarP(&opt.token, "token", "t", "", "The access token of the git repository")
	flags.StringVarP(&opt.status, "status", "", "",
		"Build status, such as: pending, success, cancelled, error")
	flags.StringVarP(&opt.target, "target", "", "https://github.com/LinuxSuRen/gogit", "Address of the build server")
	flags.StringVarP(&opt.label, "label", "", "",
		"Identity of a build status")
	flags.StringVarP(&opt.description, "description", "", "",
		"The description of a build status")

	_ = cmd.MarkFlagRequired("repo")
	_ = cmd.MarkFlagRequired("pr")
	_ = cmd.MarkFlagRequired("username")
	_ = cmd.MarkFlagRequired("token")
	return
}

func (o *option) preRun(cmd *cobra.Command, args []string) {
	if o.owner == "" {
		o.owner = o.username
	}
	if o.label == "" {
		o.label = "gogit"
	}
	if o.description == "" {
		o.description = ""
	}

	// keep the status be compatible with different system
	switch o.status {
	case "Succeeded":
		// from Argo Workflow
		o.status = "success"
	}
	o.status = strings.ToLower(o.status)
}

func (o *option) runE(cmd *cobra.Command, args []string) (err error) {
	err = pkg.Reconcile(cmd.Context(), pkg.RepoInformation{
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

type option struct {
	provider    string
	server      string
	username    string
	token       string
	owner       string
	repo        string
	pr          int
	status      string
	target      string
	label       string
	description string
}
