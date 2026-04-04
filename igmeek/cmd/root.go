package cmd

import (
	"errors"
	"os"

	"github.com/CJSen/igmeek/internal/config"
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
	PersistentPreRunE: preRun,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		var tokenErr *TokenError
		if errors.As(err, &tokenErr) {
			os.Exit(ExitAuthError)
		}
		os.Exit(ExitGeneralError)
	}
}

func preRun(cmd *cobra.Command, args []string) error {
	if cmd.Name() == "init" || cmd.Name() == "help" || cmd.Name() == "completion" {
		return nil
	}

	token := GetToken()
	if token == "" {
		return &TokenError{Msg: "no GitHub token found. Please set IMGEEK_GITHUB_TOKEN environment variable or run 'igmeek init'"}
	}
	return nil
}

func GetToken() string {
	if token := os.Getenv("IMGEEK_GITHUB_TOKEN"); token != "" {
		return token
	}

	globalDir := config.GetGlobalDataDir()
	cfgPath := config.ConfigPath(globalDir)
	if cfg, err := config.LoadConfig(cfgPath); err == nil {
		return cfg.Token
	}

	return ""
}

type TokenError struct {
	Msg string
}

func (e *TokenError) Error() string {
	return e.Msg
}
