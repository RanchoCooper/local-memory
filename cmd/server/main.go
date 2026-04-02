package main

import (
	"fmt"
	"log"
	"os"

	"localmemory/config"
	"localmemory/server"
)

var (
	cfgFile string
)

func main() {
	// 加载配置
	cfg, err := config.Load(cfgFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 创建并启动服务器
	srv := server.NewServer(cfg)

	port := cfg.Server.Port
	if port <= 0 {
		port = 8080
	}

	log.Printf("LocalMemory 服务器启动中...")
	log.Printf("监听端口: %d", port)
	log.Printf("API 文档: http://localhost:%d/docs", port)

	if err := srv.Run(fmt.Sprintf(":%d", port)); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
