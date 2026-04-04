package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/CJSen/igmeek/internal/config"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize igmeek with your GitHub token",
	RunE:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	globalDir := config.GetGlobalDataDir()
	if err := config.EnsureGlobalDir(globalDir); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	fmt.Print("Enter your GitHub Personal Access Token (needs 'repo' scope): ")
	reader := bufio.NewReader(os.Stdin)
	token, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}
	token = strings.TrimSpace(token)

	if token == "" {
		return fmt.Errorf("token cannot be empty")
	}

	cfgPath := config.ConfigPath(globalDir)
	var cfg *config.GlobalConfig
	if existing, err := config.LoadConfig(cfgPath); err == nil {
		cfg = existing
	} else {
		cfg = &config.GlobalConfig{
			Repos: []string{},
		}
	}

	cfg.Token = token
	if err := cfg.Save(cfgPath); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println("Token saved to configuration file.")
	return nil
}
