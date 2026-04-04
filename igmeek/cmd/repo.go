package cmd

import (
	"github.com/spf13/cobra"
)

var repoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Manage repository configurations",
}

func init() {
	rootCmd.AddCommand(repoCmd)
}
