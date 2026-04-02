package core

import "errors"

var (
	// ErrMemoryNotFound indicates the memory does not exist
	ErrMemoryNotFound = errors.New("memory not found")

	// ErrMemoryNotDeleted indicates the memory is not deleted, cannot restore
	ErrMemoryNotDeleted = errors.New("memory not deleted")

	// ErrCannotLinkToSelf indicates cannot link a memory to itself
	ErrCannotLinkToSelf = errors.New("cannot link memory to itself")

	// ErrEmbeddingServiceRequired indicates embedding service is required
	ErrEmbeddingServiceRequired = errors.New("embedding service is required for this operation")

	// ErrVectorStoreRequired indicates vector store is required
	ErrVectorStoreRequired = errors.New("vector store is required for this operation")

	// ErrInvalidVectorSize indicates vector dimension mismatch
	ErrInvalidVectorSize = errors.New("invalid vector size")
)
