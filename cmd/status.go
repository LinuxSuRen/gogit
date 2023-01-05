package cmd

import (
	"github.com/linuxsuren/gogit/pkg"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

func newStatusCmd() (cmd *cobra.Command) {
	opt := &statusOption{}
	cmd = &cobra.Command{
		Use:     "status",
		Short:   "Send the build token to a PR of Gitlab/GitHub",
		PreRunE: opt.preRunE,
		RunE:    opt.runE,
	}

	flags := cmd.Flags()
	flags.StringVarP(&opt.provider, "provider", "p", "github", "The provider of git, such as: gitlab, github")
	flags.StringVarP(&opt.server, "server", "s", "", "The server address of target git provider, only need when it's a private provider")
	flags.StringVarP(&opt.owner, "owner", "o", "", "Owner of a git repository")
	flags.StringVarP(&opt.repo, "repo", "r", "", "Name of target git repository")
	flags.IntVarP(&opt.pr, "pr", "", 1, "The pull request number")
	flags.StringVarP(&opt.username, "username", "u", "", "Username of the git repository")
	flags.StringVarP(&opt.token, "token", "t", "",
		"The access token of the git repository. Or you could provide a file path, such as: file:///var/token")
	flags.StringVarP(&opt.status, "status", "", "",
		"Build token, such as: pending, success, cancelled, error")
	flags.StringVarP(&opt.target, "target", "", "https://github.com/LinuxSuRen/gogit", "Address of the build server")
	flags.StringVarP(&opt.label, "label", "", "",
		"Identity of a build token")
	flags.StringVarP(&opt.description, "description", "", "",
		"The description of a build token")

	_ = cmd.MarkFlagRequired("repo")
	_ = cmd.MarkFlagRequired("pr")
	_ = cmd.MarkFlagRequired("username")
	_ = cmd.MarkFlagRequired("token")
	return
}

func (o *statusOption) preRunE(cmd *cobra.Command, args []string) (err error) {
	if o.owner == "" {
		o.owner = o.username
	}
	if o.label == "" {
		o.label = "gogit"
	}
	if o.description == "" {
		o.description = ""
	}

	// keep the token be compatible with different system
	switch o.status {
	case "Succeeded":
		// from Argo Workflow
		o.status = "success"
	}
	o.status = strings.ToLower(o.status)

	if strings.HasPrefix(o.token, "file://") {
		tokenFile := strings.TrimPrefix(o.token, "file://")
		var data []byte
		if data, err = os.ReadFile(tokenFile); err == nil {
			o.token = string(data)
		}
	}
	return
}

func (o *statusOption) runE(cmd *cobra.Command, args []string) (err error) {
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

type statusOption struct {
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