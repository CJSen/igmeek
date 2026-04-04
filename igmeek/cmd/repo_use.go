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

var repoUseCmd = &cobra.Command{
	Use:   "use",
	Short: "Select the current working repository",
	Long:  "Switch the active repository for subsequent commands. If only one repository is configured, it is automatically selected. Otherwise, prompts for selection.",
	RunE:  runRepoUse,
}

func init() {
	repoCmd.AddCommand(repoUseCmd)
}

func runRepoUse(cmd *cobra.Command, args []string) error {
	globalDir := config.GetGlobalDataDir()
	cfg, err := config.LoadConfig(config.ConfigPath(globalDir))
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if len(cfg.Repos) == 0 {
		return fmt.Errorf("no repositories configured. Run 'igmeek repo add' first")
	}

	if len(cfg.Repos) == 1 {
		cfg.CurrentRepo = cfg.Repos[0]
		if err := cfg.Save(config.ConfigPath(globalDir)); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		fmt.Printf("Switched to: %s\n", cfg.CurrentRepo)
		return nil
	}

	fmt.Println("Select a repository:")
	for i, r := range cfg.Repos {
		marker := " "
		if r == cfg.CurrentRepo {
			marker = "*"
		}
		fmt.Printf("  %d. %s%s\n", i+1, r, marker)
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

	cfg.CurrentRepo = cfg.Repos[num-1]
	if err := cfg.Save(config.ConfigPath(globalDir)); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Switched to: %s\n", cfg.CurrentRepo)
	return nil
}
