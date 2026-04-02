package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var StatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "显示统计信息",
	Long: `显示 LocalMemory 记忆系统的统计信息。

示例：
  localmemory stats`,

	Run: func(cmd *cobra.Command, args []string) {
		// 初始化存储
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

		fmt.Println("LocalMemory 统计信息")
		fmt.Println("==================")
		fmt.Printf("\n记忆总数: %d\n", stats.Total)
		fmt.Printf("已删除: %d\n", stats.Deleted)

		fmt.Println("\n按类型分布:")
		for t, count := range stats.ByType {
			fmt.Printf("  %s: %d\n", t, count)
		}

		fmt.Println("\n按作用域分布:")
		for s, count := range stats.ByScope {
			fmt.Printf("  %s: %d\n", s, count)
		}

		fmt.Println("\n按媒体类型分布:")
		for m, count := range stats.ByMedia {
			fmt.Printf("  %s: %d\n", m, count)
		}
	},
}
