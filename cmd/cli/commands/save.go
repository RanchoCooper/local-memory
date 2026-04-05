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
	saveProfile   string
	saveExtract   bool
)

var SaveCmd = &cobra.Command{
	Use:   "save <text>",
	Short: "Save a memory",
	Long: `Save text as a memory to the LocalMemory system.

Examples:
  localmemory save "User prefers Go language"
  localmemory save "user_preference" "User prefers Go language"
  localmemory save "Image description" --type fact --media-type image
  localmemory save "key" "value" --type preference --scope global`,

	Args: cobra.RangeArgs(1, 2),
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
		if saveProfile == "" {
			saveProfile = config.Get().Profile.ID
		}
		if saveProfile == "" {
			saveProfile = "default"
		}

		// Parse key-value: save <key> <value> or save <value> (key defaults to first 50 chars of value)
		var key, value string
		if len(args) == 2 {
			key = args[0][:min(50, len(args[0]))]
			value = args[1]
		} else {
			value = args[0]
			key = args[0][:min(50, len(args[0]))]
		}

		// Create memory
		memory := &core.Memory{
			ProfileID: saveProfile,
			Type:      core.MemoryType(saveType),
			Scope:     core.Scope(saveScope),
			MediaType: core.MediaType(saveMediaType),
			Key:       key,
			Value:     value,
			Tags:      saveTags,
			Metadata: core.Metadata{
				Source: "cli",
			},
		}

		// Initialize Store (without vector store for MVP simplification)
		store := core.NewStore(sqliteStore, nil, nil)

		// Save memory
		if saveExtract {
			// Extract and save atomic facts
			parent, facts, err := store.ExtractAndSave(memory)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to extract facts: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Memory saved with %d atomic facts:\n", len(facts))
			fmt.Printf("  Parent ID: %s\n", parent.ID)
			fmt.Printf("  Parent Key: %s\n", parent.Key)
			for i, fact := range facts {
				factType := "UNKNOWN"
				if len(fact.Tags) > 0 {
					factType = fact.Tags[0]
				}
				fmt.Printf("  Fact %d: [%s] %s\n", i+1, factType, truncate(fact.Value, 60))
			}
		} else {
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
		}
	},
}

func init() {
	SaveCmd.Flags().StringVar(&saveType, "type", "", "Memory type (preference, fact, event, skill, goal)")
	SaveCmd.Flags().StringVar(&saveScope, "scope", "", "Memory scope (global, session, agent)")
	SaveCmd.Flags().StringVar(&saveMediaType, "media-type", "text", "Media type (text, image)")
	SaveCmd.Flags().StringArrayVar(&saveTags, "tag", []string{}, "Tags for the memory")
	SaveCmd.Flags().StringVar(&saveProfile, "profile", "", "Profile ID (default: from config)")
	SaveCmd.Flags().BoolVar(&saveExtract, "extract", false, "Extract atomic facts from text")
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
