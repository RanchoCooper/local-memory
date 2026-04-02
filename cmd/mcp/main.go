package main

import (
	"fmt"
	"os"

	"localmemory/agent/mcp"
	"localmemory/config"
)

func main() {
	// 加载配置
	cfg, err := config.Load("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 创建 MCP Server
	server, err := mcp.NewMCPServer(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create MCP server: %v\n", err)
		os.Exit(1)
	}

	// 启动 MCP Server
	if err := server.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "MCP server error: %v\n", err)
		os.Exit(1)
	}
}
