package commands

import (
	"fmt"
	"os"

	"localmemory/config"
	"localmemory/core"
	"localmemory/storage"
	"localmemory/storage/vector"

	"github.com/spf13/cobra"
)

var (
	saveType      string
	saveScope     string
	saveMediaType string
	saveTags      []string
)

var SaveCmd = &cobra.Command{
	Use:   "save <text>",
	Short: "Save a memory",
	Long: `Save text as a memory to the LocalMemory system.

Examples:
  localmemory save "User prefers Go language"
  localmemory save "User preference" --type preference --scope global
  localmemory save "Image description" --type fact --media-type image`,

	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize storage
		sqliteStore, err := initSQLiteStore()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to init storage: %v\n", err)
			os.Exit(1)
		}
		defer sqliteStore.Close()

		// Set defaults
		if saveScope == "" {
			saveScope = "global"
		}
		if saveType == "" {
			saveType = "fact"
		}

		// Create memory
		memory := &core.Memory{
			Type:      core.MemoryType(saveType),
			Scope:     core.Scope(saveScope),
			MediaType: core.MediaType(saveMediaType),
			Key:       args[0][:min(50, len(args[0]))],
			Value:     args[0],
			Tags:      saveTags,
			Metadata: core.Metadata{
				Source: "cli",
			},
		}

		// Initialize Store (without vector store for MVP simplification)
		store := core.NewStore(sqliteStore, nil, nil)

		// Save memory
		if err := store.Save(memory); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to save memory: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Memory saved:\n")
		fmt.Printf("  ID: %s\n", memory.ID)
		fmt.Printf("  Key: %s\n", memory.Key)
		fmt.Printf("  Type: %s\n", memory.Type)
		fmt.Printf("  Scope: %s\n", memory.Scope)
		fmt.Printf("  Created: %d\n", memory.CreatedAt)
	},
}

func init() {
	SaveCmd.Flags().StringVar(&saveType, "type", "", "Memory type (preference, fact, event, skill, goal)")
	SaveCmd.Flags().StringVar(&saveScope, "scope", "", "Memory scope (global, session, agent)")
	SaveCmd.Flags().StringVar(&saveMediaType, "media-type", "text", "Media type (text, image)")
	SaveCmd.Flags().StringArrayVar(&saveTags, "tag", []string{}, "Tags for the memory")
}

func initSQLiteStore() (*storage.SQLiteStore, error) {
	cfg := config.Get()
	return storage.NewSQLiteStore(cfg.Database.Path)
}

func initVectorStore() (vector.VectorStore, error) {
	cfg := config.Get()
	return vector.NewVectorStore(cfg.VectorDB.Type, vector.QdrantConfig{
		URL:        cfg.VectorDB.URL,
		Collection: cfg.VectorDB.Collection,
		VectorSize: 384, // MiniLM embedding size
	})
}

func initConfig() {
	if config.Get() == nil {
		config.Default()
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
