package server

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// loggerMiddleware 日志中间件
// 记录请求方法和路径、响应状态码、处理时间
func loggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// 处理请求
		c.Next()

		// 记录日志
		latency := time.Since(start)
		status := c.Writer.Status()

		log.Printf("[%s] %s %d %v", method, path, status, latency)
	}
}

// corsMiddleware CORS 中间件
// 允许跨域请求
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// authMiddleware 认证中间件（预留）
// TODO: 实现认证逻辑
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取 Authorization header
		token := c.GetHeader("Authorization")

		if token == "" {
			// 允许无认证访问 MVP
			c.Next()
			return
		}

		// TODO: 验证 token
		c.Next()
	}
}
