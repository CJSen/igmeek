package cmd

import (
	"github.com/spf13/cobra"
)

var labelCmd = &cobra.Command{
	Use:   "label",
	Short: "Manage repository labels",
}

func init() {
	rootCmd.AddCommand(labelCmd)
}
