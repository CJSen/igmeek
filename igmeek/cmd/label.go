package cmd

import (
	"github.com/spf13/cobra"
)

var labelCmd = &cobra.Command{
	Use:   "label",
	Short: "Manage repository labels",
	Long:  "Create and list GitHub labels for the configured repository. Labels are used by Gmeek to control which issues are published.",
}

func init() {
	rootCmd.AddCommand(labelCmd)
}
