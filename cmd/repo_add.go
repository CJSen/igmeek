package cmd

import (
	"bufio"
	"context"
	"fmt"
	"strings"

	"github.com/CJSen/igmeek/cli/internal/api"
	"github.com/CJSen/igmeek/cli/internal/config"
	"github.com/spf13/cobra"
)

var verifyRepoAccessFunc = func(ctx context.Context, token, owner, repo string) error {
	client := api.NewClient(token)
	return client.VerifyRepo(ctx, owner, repo)
}

var repoAddCmd = &cobra.Command{
	Use:   "add [owner/repo]",
	Short: "Add a repository configuration",
	Long:  "Add a GitHub repository to the configuration. Accepts an 'owner/repo' argument, or prompts interactively if omitted. Verifies repository access before saving. Automatically sets as current repo if none is selected.",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runRepoAdd,
}

func init() {
	repoCmd.AddCommand(repoAddCmd)
}

func runRepoAdd(cmd *cobra.Command, args []string) error {
	var fullName string

	if len(args) > 0 {
		fullName = strings.TrimSpace(args[0])
	} else {
		fmt.Fprint(cmd.OutOrStdout(), "Enter repository (owner/repo or GitHub URL): ")
		reader := bufio.NewReader(cmd.InOrStdin())
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}
		fullName = strings.TrimSpace(input)
	}

	fullName, err := config.NormalizeRepoInput(fullName)
	if err != nil {
		return err
	}
	parts := strings.SplitN(fullName, "/", 2)
	owner, repo := parts[0], parts[1]

	globalDir := globalDataDirFunc()
	cfg, err := config.LoadConfig(config.ConfigPath(globalDir))
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := verifyRepoAccessFunc(context.Background(), GetToken(), owner, repo); err != nil {
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

	cfg.AddRepo(fullName)

	if cfg.CurrentRepo == "" {
		cfg.CurrentRepo = fullName
	}

	if err := cfg.Save(config.ConfigPath(globalDir)); err != nil {
		return fmt.Errorf("failed to save global config: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Added repository: %s\n", fullName)
	result, _, err := runSyncForRepoFunc(cmd, fullName, "")
	if err != nil {
		return fmt.Errorf("repository was added, but sync failed. You can retry with 'igmeek sync': %w", err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Synced %d issues, %d labels from %s\n", result.IssuesCount, result.LabelsCount, fullName)
	return nil
}
