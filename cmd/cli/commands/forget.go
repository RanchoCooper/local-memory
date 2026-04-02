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
	Short: "删除记忆",
	Long: `删除 LocalMemory 中的记忆。

默认执行软删除，记忆可通过 ID 恢复。
使用 --hard 执行永久删除（不可恢复）。

示例：
  localmemory forget <memory-id>
  localmemory forget <memory-id> --hard
  localmemory forget <memory-id> --force`,

	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		memoryID := args[0]

		// 初始化存储
		sqliteStore, err := initSQLiteStore()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to init storage: %v\n", err)
			os.Exit(1)
		}
		defer sqliteStore.Close()

		// 验证记忆存在
		memory, err := sqliteStore.GetByID(memoryID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get memory: %v\n", err)
			os.Exit(1)
		}
		if memory == nil {
			fmt.Fprintf(os.Stderr, "记忆不存在: %s\n", memoryID)
			os.Exit(1)
		}

		cfg := config.Get()
		if cfg == nil {
			cfg = config.Default()
		}

		forget := core.NewForget(sqliteStore, nil)

		if forgetHard {
			// 永久删除
			if !forgetForce {
				fmt.Printf("警告：永久删除记忆 '%s'，此操作不可恢复！\n", memory.Key)
				fmt.Printf("确认删除？输入 'yes' 继续: ")
				var confirm string
				fmt.Scanln(&confirm)
				if confirm != "yes" {
					fmt.Println("已取消")
					os.Exit(0)
				}
			}

			if err := forget.HardDelete(memoryID); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to delete memory: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("记忆已永久删除: %s\n", memoryID)
		} else {
			// 软删除
			if err := forget.Delete(memoryID); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to delete memory: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("记忆已删除: %s\n", memory.Key)
			fmt.Printf("  ID: %s\n", memoryID)
			fmt.Printf("  可通过 ID 恢复\n")
		}
	},
}
