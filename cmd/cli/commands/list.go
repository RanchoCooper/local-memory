package commands

import (
	"fmt"
	"os"

	"localmemory/config"
	"localmemory/core"

	"github.com/spf13/cobra"
)

var (
	listScope  string
	listLimit  int
	listOffset int
	listAll    bool
)

var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List memories",
	Long: `List memories in LocalMemory.

Examples:
  localmemory list
  localmemory list --scope global
  localmemory list --limit 20 --offset 0
  localmemory list --all  # Include deleted memories`,

	Run: func(cmd *cobra.Command, args []string) {
		// Initialize configuration
		cfg := config.Get()
		if cfg == nil {
			cfg = config.Default()
		}

		// Initialize storage
		sqliteStore, err := initSQLiteStore()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to init storage: %v\n", err)
			os.Exit(1)
		}
		defer sqliteStore.Close()

		// Set defaults
		if listLimit <= 0 {
			listLimit = 20
		}
		if listScope == "" {
			listScope = cfg.CLI.DefaultScope
		}

		// Build query request
		listReq := &core.ListRequest{
			Scope:          core.Scope(listScope),
			Limit:          listLimit,
			Offset:         listOffset,
			IncludeDeleted: listAll,
		}

		recall := core.NewRecall(sqliteStore, nil, nil, core.NewRanker(cfg.Decay.Lambda))
		resp, err := recall.List(listReq)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to list memories: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Memory list (%d total):\n\n", resp.Total)

		if len(resp.Memories) == 0 {
			fmt.Println("No memories yet")
			return
		}

		for i, m := range resp.Memories {
			deletedMark := ""
			if m.Deleted {
				deletedMark = " [deleted]"
			}
			fmt.Printf("%d. [%s%s] %s\n", listOffset+i+1, m.Type, deletedMark, m.Key)
			fmt.Printf("   Value: %s\n", truncate(m.Value, 80))
			fmt.Printf("   Scope: %s | Confidence: %.2f | Created: %d\n",
				m.Scope, m.Confidence, m.CreatedAt)
			fmt.Printf("   ID: %s\n\n", m.ID)
		}

		// Pagination hint
		if resp.Total > listLimit {
			fmt.Printf("Showing %d-%d of %d\n",
				listOffset+1, listOffset+len(resp.Memories), resp.Total)
			fmt.Printf("Use --offset %d to see more\n", listOffset+listLimit)
		}
	},
}

func init() {
	ListCmd.Flags().BoolVar(&listAll, "all", false, "Include deleted memories")
	ListCmd.Flags().IntVar(&listLimit, "limit", 20, "Number of memories to list")
	ListCmd.Flags().StringVar(&listScope, "scope", "", "Scope filter (global, session, agent)")
	ListCmd.Flags().IntVar(&listOffset, "offset", 0, "Offset for pagination")
}
