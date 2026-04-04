package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/CJSen/igmeek/internal/api"
	"github.com/CJSen/igmeek/internal/config"
	"github.com/CJSen/igmeek/internal/sync"
	"github.com/spf13/cobra"
)

var repoAddCmd = &cobra.Command{
	Use:   "add [owner/repo]",
	Short: "Add a repository configuration",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runRepoAdd,
}

func init() {
	repoCmd.AddCommand(repoAddCmd)
}

func runRepoAdd(cmd *cobra.Command, args []string) error {
	var fullName string

	if len(args) > 0 {
		fullName = args[0]
	} else {
		fmt.Print("Enter repository (owner/repo): ")
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}
		fullName = strings.TrimSpace(input)
	}

	owner, repo, err := sync.ParseOwnerRepo(fullName)
	if err != nil {
		return err
	}

	globalDir := config.GetGlobalDataDir()
	cfg, err := config.LoadConfig(config.ConfigPath(globalDir))
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client := api.NewClient(GetToken())
	if err := client.VerifyRepo(context.Background(), owner, repo); err != nil {
		return fmt.Errorf("cannot access repository: %w", err)
	}

	repoDir := config.GetRepoDir(globalDir, fullName)
	repoConfig := &config.RepoConfig{
		Owner:    owner,
		Repo:     repo,
		FullName: fullName,
	}

	if err := repoConfig.Save(repoDir); err != nil {
		return fmt.Errorf("failed to save repo config: %w", err)
	}

	found := false
	for _, r := range cfg.Repos {
		if r == fullName {
			found = true
			break
		}
	}
	if !found {
		cfg.Repos = append(cfg.Repos, fullName)
	}

	if cfg.CurrentRepo == "" {
		cfg.CurrentRepo = fullName
	}

	if err := cfg.Save(config.ConfigPath(globalDir)); err != nil {
		return fmt.Errorf("failed to save global config: %w", err)
	}

	fmt.Printf("Added repository: %s\n", fullName)
	return nil
}
