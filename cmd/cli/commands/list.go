package commands

import (
	"fmt"
	"os"
	"strconv"

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
	Short: "列出记忆",
	Long: `列出 LocalMemory 中的记忆。

示例：
  localmemory list
  localmemory list --scope global
  localmemory list --limit 20 --offset 0
  localmemory list --all  # 包含已删除的记忆`,

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

		// 设置默认值
		if listLimit <= 0 {
			listLimit = 20
		}
		if listScope == "" {
			listScope = cfg.CLI.DefaultScope
		}

		// 构建查询请求
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

		fmt.Printf("记忆列表（共 %d 条）:\n\n", resp.Total)

		if len(resp.Memories) == 0 {
			fmt.Println("暂无记忆")
			return
		}

		for i, m := range resp.Memories {
			deletedMark := ""
			if m.Deleted {
				deletedMark = " [已删除]"
			}
			fmt.Printf("%d. [%s%s] %s\n", listOffset+i+1, m.Type, deletedMark, m.Key)
			fmt.Printf("   Value: %s\n", truncate(m.Value, 80))
			fmt.Printf("   Scope: %s | Confidence: %.2f | Created: %d\n",
				m.Scope, m.Confidence, m.CreatedAt)
			fmt.Printf("   ID: %s\n\n", m.ID)
		}

		// 分页提示
		if resp.Total > listLimit {
			fmt.Printf("显示 %d-%d 条，共 %d 条\n",
				listOffset+1, listOffset+len(resp.Memories), resp.Total)
			fmt.Printf("使用 --offset %d 查看更多\n", listOffset+listLimit)
		}
	},
}

var offsetArg string

var listCmdOld = &cobra.Command{
	Use:   "list [offset]",
	Short: "列出记忆",
	Long: `列出 LocalMemory 中的记忆。

示例：
  localmemory list
  localmemory list 10`,

	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			offset, err := strconv.Atoi(args[0])
			if err == nil {
				listOffset = offset
			}
		}
		ListCmd.Run(cmd, args)
	},
}
