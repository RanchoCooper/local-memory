package main

import (
	"fmt"
	"os"

	"localmemory/cmd/cli/commands"
	"localmemory/config"

	"github.com/spf13/cobra"
)

var (
	cfgFile string
	cfg    *config.Config
)

func main() {
	// 加载配置
	var err error
	cfg, err = config.Load(cfgFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 创建根命令
	rootCmd := &cobra.Command{
		Use:   "localmemory",
		Short: "LocalMemory - AI Agent 的本地记忆系统",
		Long: `LocalMemory 为 AI Agent 提供本地优先的持久化、可检索、可进化的长期记忆系统。

支持的记忆类型：
  - preference: 用户偏好
  - fact: 客观事实
  - event: 事件记录
  - skill: 技能/能力
  - goal: 目标/意图

示例：
  localmemory save "用户喜欢 Go 语言"
  localmemory query "用户偏好什么语言"
  localmemory list --scope global
  localmemory forget <id>`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// 重新加载配置（可能已被修改）
			cfg, _ = config.Load(cfgFile)
		},
	}

	// 添加全局 flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "配置文件路径 (默认: ~/.localmemory/config.json)")

	// 添加子命令
	rootCmd.AddCommand(commands.SaveCmd)
	rootCmd.AddCommand(commands.QueryCmd)
	rootCmd.AddCommand(commands.ListCmd)
	rootCmd.AddCommand(commands.ForgetCmd)
	rootCmd.AddCommand(commands.StatsCmd)

	// 执行命令
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
