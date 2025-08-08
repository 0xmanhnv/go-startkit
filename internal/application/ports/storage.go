package ports

import (
	"context"
	"io"
)

// ObjectStorage abstracts object storage operations.
type ObjectStorage interface {
	Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) (url string, err error)
	Get(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
}
