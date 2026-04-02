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
	Short: "保存记忆",
	Long:  `将文本保存为记忆到 LocalMemory 系统。

示例：
  localmemory save "用户喜欢 Go 语言"
  localmemory save "用户偏好" --type preference --scope global
  localmemory save "图片描述" --type fact --media-type image`,

	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// 初始化存储
		sqliteStore, err := initSQLiteStore()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to init storage: %v\n", err)
			os.Exit(1)
		}
		defer sqliteStore.Close()

		// 创建记忆
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

		// 初始化 Store（不使用向量存储，MVP 简化）
		store := core.NewStore(sqliteStore, nil, nil)

		// 保存记忆
		if err := store.Save(memory); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to save memory: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("记忆已保存:\n")
		fmt.Printf("  ID: %s\n", memory.ID)
		fmt.Printf("  Key: %s\n", memory.Key)
		fmt.Printf("  Type: %s\n", memory.Type)
		fmt.Printf("  Scope: %s\n", memory.Scope)
		fmt.Printf("  Created: %d\n", memory.CreatedAt)
	},
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
