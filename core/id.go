package core

import (
	"github.com/google/uuid"
)

// GenerateID 生成一个新的 UUID v4 字符串
// 用作 Memory 的唯一标识符
// UUID v4 使用随机数生成，具有足够低的冲突概率
func GenerateID() string {
	return uuid.New().String()
}
