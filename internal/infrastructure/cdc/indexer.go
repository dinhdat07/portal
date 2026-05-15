package cdc

import "context"

type Indexer[T any] interface {
	Index(ctx context.Context, doc T) error
	Delete(ctx context.Context, id string) error
}
