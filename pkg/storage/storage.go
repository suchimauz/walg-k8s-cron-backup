package storage

import (
	"context"
	"io"
)

// Object struct for upload file for some storage provider
type UploadInput struct {
	File        io.Reader
	Name        string
	Size        int64
	ContentType string
}

type Provider interface {
	Upload(ctx context.Context, input UploadInput) (string, error)
}
