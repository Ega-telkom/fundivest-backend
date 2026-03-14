package storage

import "context"

// FileStorage interface untuk storage abstraction
type FileStorage interface {
    Save(ctx context.Context, filename string, data []byte) error
    Get(ctx context.Context, filename string) ([]byte, error)
    Delete(ctx context.Context, filename string) error
    Exists(ctx context.Context, filename string) (bool, error)
}