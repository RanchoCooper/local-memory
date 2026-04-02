package main

import (
	"fmt"
	"os"

	"localmemory/cmd/cli/commands"
	"localmemory/config"

	"github.com/spf13/cobra"
)

var (
	cfgFile string
	cfg    *config.Config
)

func main() {
	// Load configuration
	var err error
	cfg, err = config.Load(cfgFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Create root command
	rootCmd := &cobra.Command{
		Use:   "localmemory",
		Short: "LocalMemory - Local-first memory system for AI Agents",
		Long: `LocalMemory provides a local-first, persistent, searchable, and evolvable long-term memory system for AI Agents.

Supported memory types:
  - preference: user preference
  - fact: objective fact
  - event: event record
  - skill: skill/ability
  - goal: goal/intent

Examples:
  localmemory save "User prefers Go language"
  localmemory query "What language does user prefer"
  localmemory list --scope global
  localmemory forget <id>`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			cfg, _ = config.Load(cfgFile)
		},
	}

	// Add global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Config file path (default: ~/.localmemory/config.json)")

	// Add subcommands
	rootCmd.AddCommand(commands.SaveCmd)
	rootCmd.AddCommand(commands.QueryCmd)
	rootCmd.AddCommand(commands.ListCmd)
	rootCmd.AddCommand(commands.ForgetCmd)
	rootCmd.AddCommand(commands.StatsCmd)

	// Execute command
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
