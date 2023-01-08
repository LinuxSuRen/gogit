package cmd

import (
	"github.com/linuxsuren/gogit/pkg"
	"github.com/spf13/cobra"
)

func newCommentCommand() (c *cobra.Command) {
	opt := &commentOption{}
	c = &cobra.Command{
		Use:     "comment",
		Short:   "Create a comment against the pull request",
		Example: `gogit comment --provider github --username linuxsuren --repo test --pr 45 --token $GITHUB_TOKEN -m LGTM`,
		Aliases: []string{"c"},
		PreRunE: opt.preRunE,
		RunE:    opt.runE,
	}

	flags := c.Flags()
	opt.addFlags(flags)
	flags.StringVarP(&opt.message, "message", "m", "", "The comment body")
	flags.StringVarP(&opt.identity, "identity", "", pkg.CommentEndMarker, "The identity for matching exiting comment")
	return
}

func (o *commentOption) runE(c *cobra.Command, args []string) (err error) {
	err = pkg.CreateComment(c.Context(), pkg.RepoInformation{
		Provider: o.provider,
		Server:   o.server,
		Owner:    o.owner,
		Repo:     o.repo,
		PrNumber: o.pr,
		Username: o.username,
		Token:    o.token,
	}, o.message, o.identity)
	return
}

func (o *commentOption) preRunE(c *cobra.Command, args []string) (err error) {
	o.preHandle()
	return
}

type commentOption struct {
	gitProviderOption
	message  string
	identity string
}
