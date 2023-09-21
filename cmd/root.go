package cmd

import "github.com/spf13/cobra"

// NewRootCommand returns the root command
func NewRootCommand() (c *cobra.Command) {
	c = &cobra.Command{
		Use:   "gogit",
		Short: "Git client across Gitlab/GitHub",
	}

	c.AddCommand(newCheckoutCommand(),
		newStatusCmd(), newCommentCommand(),
		newPullRequestCmd())
	return
}
