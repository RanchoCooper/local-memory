package commands

import (
	"fmt"
	"os"

	"localmemory/config"
	"localmemory/core"

	"github.com/spf13/cobra"
)

var (
	forgetHard   bool
	forgetForce bool
)

var ForgetCmd = &cobra.Command{
	Use:   "forget <id>",
	Short: "Delete a memory",
	Long: `Delete a memory from LocalMemory.

Default is soft delete, memory can be recovered by ID.
Use --hard for permanent deletion (cannot be recovered).

Examples:
  localmemory forget <memory-id>
  localmemory forget <memory-id> --hard
  localmemory forget <memory-id> --force`,

	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		memoryID := args[0]

		// Initialize storage
		sqliteStore, err := initSQLiteStore()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to init storage: %v\n", err)
			os.Exit(1)
		}
		defer sqliteStore.Close()

		// Verify memory exists
		memory, err := sqliteStore.GetByID(memoryID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get memory: %v\n", err)
			os.Exit(1)
		}
		if memory == nil {
			fmt.Fprintf(os.Stderr, "Memory not found: %s\n", memoryID)
			os.Exit(1)
		}

		cfg := config.Get()
		if cfg == nil {
			cfg = config.Default()
		}

		forget := core.NewForget(sqliteStore, nil)

		if forgetHard {
			// Permanent deletion
			if !forgetForce {
				fmt.Printf("Warning: Permanently deleting memory '%s', this cannot be recovered!\n", memory.Key)
				fmt.Printf("Confirm deletion? Enter 'yes' to continue: ")
				var confirm string
				fmt.Scanln(&confirm)
				if confirm != "yes" {
					fmt.Println("Cancelled")
					os.Exit(0)
				}
			}

			if err := forget.HardDelete(memoryID); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to delete memory: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Memory permanently deleted: %s\n", memoryID)
		} else {
			// Soft delete
			if err := forget.Delete(memoryID); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to delete memory: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Memory deleted: %s\n", memory.Key)
			fmt.Printf("  ID: %s\n", memoryID)
			fmt.Printf("  Can be recovered by ID\n")
		}
	},
}
