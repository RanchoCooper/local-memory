package core

import "errors"

var (
	// ErrMemoryNotFound 记忆不存在
	ErrMemoryNotFound = errors.New("memory not found")

	// ErrMemoryNotDeleted 记忆未删除，无法恢复
	ErrMemoryNotDeleted = errors.New("memory not deleted")

	// ErrCannotLinkToSelf 无法关联到自身
	ErrCannotLinkToSelf = errors.New("cannot link memory to itself")

	// ErrEmbeddingServiceRequired 需要嵌入服务
	ErrEmbeddingServiceRequired = errors.New("embedding service is required for this operation")

	// ErrVectorStoreRequired 需要向量存储
	ErrVectorStoreRequired = errors.New("vector store is required for this operation")

	// ErrInvalidVectorSize 向量维度不匹配
	ErrInvalidVectorSize = errors.New("invalid vector size")
)
