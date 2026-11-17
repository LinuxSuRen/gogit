package cmd

import (
	"github.com/linuxsuren/gogit/pkg/version"
	"github.com/spf13/cobra"
)

func newVersionCmd() (c *cobra.Command) {
	c = &cobra.Command{
		Use:     "version",
		Aliases: []string{"v"},
		Short:   "Print the version of gogit",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			cmd.Println(version.Version)
			return
		},
	}
	return
}
