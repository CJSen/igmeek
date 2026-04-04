package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/CJSen/igmeek/internal/config"
	"github.com/spf13/cobra"
)

var repoDelCmd = &cobra.Command{
	Use:   "del",
	Short: "Remove a repository configuration",
	Long:  "Remove a repository from the configuration and delete its local data directory (index and tag cache). Prompts for selection if multiple repositories are configured.",
	RunE:  runRepoDel,
}

func init() {
	repoCmd.AddCommand(repoDelCmd)
}

func runRepoDel(cmd *cobra.Command, args []string) error {
	globalDir := config.GetGlobalDataDir()
	cfg, err := config.LoadConfig(config.ConfigPath(globalDir))
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if len(cfg.Repos) == 0 {
		return fmt.Errorf("no repositories configured")
	}

	var target string
	if len(cfg.Repos) == 1 {
		target = cfg.Repos[0]
	} else {
		fmt.Println("Select a repository to remove:")
		for i, r := range cfg.Repos {
			fmt.Printf("  %d. %s\n", i+1, r)
		}

		fmt.Print("Enter number: ")
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		num, err := strconv.Atoi(strings.TrimSpace(input))
		if err != nil || num < 1 || num > len(cfg.Repos) {
			return fmt.Errorf("invalid selection")
		}

		target = cfg.Repos[num-1]
	}

	var newRepos []string
	for _, r := range cfg.Repos {
		if r != target {
			newRepos = append(newRepos, r)
		}
	}
	cfg.Repos = newRepos

	if cfg.CurrentRepo == target {
		if len(newRepos) > 0 {
			cfg.CurrentRepo = newRepos[0]
		} else {
			cfg.CurrentRepo = ""
		}
	}

	if err := cfg.Save(config.ConfigPath(globalDir)); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	repoDir := config.GetRepoDir(globalDir, target)
	os.RemoveAll(repoDir)

	fmt.Printf("Removed repository: %s\n", target)
	return nil
}
