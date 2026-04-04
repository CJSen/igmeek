package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

const (
	ExitSuccess      = 0
	ExitGeneralError = 1
	ExitParamError   = 2
	ExitAuthError    = 3
	ExitNetworkError = 4
)

var rootCmd = &cobra.Command{
	Use:   "igmeek",
	Short: "Local-first GitHub Issue/Tag management CLI for Gmeek blogs",
	Long: `igmeek is a CLI tool for managing GitHub Issues and Tags
for blogs built with the Gmeek framework.

It allows you to create, update, close, and reopen Issues
from your local terminal, with label management tailored
for Gmeek's label-driven publishing workflow.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(ExitGeneralError)
	}
}
