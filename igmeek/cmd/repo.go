package cmd

import (
	"github.com/spf13/cobra"
)

var repoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Manage repository configurations",
	Long:  "Add, remove, list, and switch between repository configurations. Repositories are stored in the global data directory and isolated by owner/repo.",
}

func init() {
	rootCmd.AddCommand(repoCmd)
}
