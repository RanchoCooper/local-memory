package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var StatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show statistics",
	Long: `Show statistics for the LocalMemory system.

Examples:
  localmemory stats`,

	Run: func(cmd *cobra.Command, args []string) {
		// Initialize storage
		sqliteStore, err := initSQLiteStore()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to init storage: %v\n", err)
			os.Exit(1)
		}
		defer sqliteStore.Close()

		stats, err := sqliteStore.GetStats()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get stats: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("LocalMemory Statistics")
		fmt.Println("==================")
		fmt.Printf("\nTotal memories: %d\n", stats.Total)
		fmt.Printf("Deleted: %d\n", stats.Deleted)

		fmt.Println("\nBy type:")
		for t, count := range stats.ByType {
			fmt.Printf("  %s: %d\n", t, count)
		}

		fmt.Println("\nBy scope:")
		for s, count := range stats.ByScope {
			fmt.Printf("  %s: %d\n", s, count)
		}

		fmt.Println("\nBy media type:")
		for m, count := range stats.ByMedia {
			fmt.Printf("  %s: %d\n", m, count)
		}
	},
}
