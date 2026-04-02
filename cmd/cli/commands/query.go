package commands

import (
	"fmt"
	"os"

	"localmemory/config"
	"localmemory/core"

	"github.com/spf13/cobra"
)

var (
	queryTopK  int
	queryScope string
)

var QueryCmd = &cobra.Command{
	Use:   "query <text>",
	Short: "语义检索记忆",
	Long: `通过自然语言查询检索相关记忆。

示例：
  localmemory query "用户偏好什么语言"
  localmemory query "项目技术栈" --topk 10 --scope global`,

	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// 初始化配置
		cfg := config.Get()
		if cfg == nil {
			cfg = config.Default()
		}

		// 初始化存储
		sqliteStore, err := initSQLiteStore()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to init storage: %v\n", err)
			os.Exit(1)
		}
		defer sqliteStore.Close()

		// MVP 阶段：简单关键词搜索
		// 完整的语义搜索需要 embedding 服务
		searchQuery := args[0]
		if queryTopK <= 0 {
			queryTopK = cfg.CLI.DefaultTopK
		}
		if queryScope == "" {
			queryScope = cfg.CLI.DefaultScope
		}

		// 列出记忆并简单过滤
		listReq := &core.ListRequest{
			Scope:  core.Scope(queryScope),
			Limit:  100,
			Offset: 0,
		}

		recall := core.NewRecall(sqliteStore, nil, nil, core.NewRanker(cfg.Decay.Lambda))
		resp, err := recall.List(listReq)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to list memories: %v\n", err)
			os.Exit(1)
		}

		// 简单关键词匹配（MVP）
		var matched []*core.Memory
		for _, m := range resp.Memories {
			if containsIgnoreCase(m.Value, searchQuery) || containsIgnoreCase(m.Key, searchQuery) {
				matched = append(matched, m)
				if len(matched) >= queryTopK {
					break
				}
			}
		}

		fmt.Printf("找到 %d 条相关记忆（共 %d 条）:\n\n", len(matched), resp.Total)
		for i, m := range matched {
			fmt.Printf("%d. [%s] %s\n", i+1, m.Type, m.Key)
			fmt.Printf("   Value: %s\n", truncate(m.Value, 100))
			fmt.Printf("   Scope: %s | Confidence: %.2f\n", m.Scope, m.Confidence)
			fmt.Printf("   ID: %s\n\n", m.ID)
		}

		if len(matched) == 0 {
			fmt.Println("未找到相关记忆")
		}
	},
}

func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && containsLower(s, substr)
}

func containsLower(s, substr string) bool {
	s = toLower(s)
	substr = toLower(substr)
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func toLower(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
