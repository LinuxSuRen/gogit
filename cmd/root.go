package cmd

import (
	"github.com/linuxsuren/gogit/pkg/version"
	"github.com/spf13/cobra"
)

// NewRootCommand returns the root command
func NewRootCommand() (c *cobra.Command) {
	c = &cobra.Command{
		Use:     "gogit",
		Short:   "Git client across Gitlab/GitHub",
		Version: version.Version,
	}

	c.AddCommand(newCheckoutCommand(),
		newStatusCmd(), newCommentCommand(),
		newPullRequestCmd(), newCommitCmd(),
		newVersionCmd())
	return
}
