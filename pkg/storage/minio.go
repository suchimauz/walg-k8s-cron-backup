package storage

import (
	"context"
	"fmt"

	"github.com/minio/minio-go/v7"
)

// Minio storage struct implemets Provider interface methods
type FileStorage struct {
	client   *minio.Client
	bucket   string
	endpoint string
}

// Constructor
func NewFileStorage(client *minio.Client, bucket, endpoint string) *FileStorage {
	return &FileStorage{
		client:   client,
		bucket:   bucket,
		endpoint: endpoint,
	}
}

// Required method for Provider interface
func (fs *FileStorage) Upload(ctx context.Context, input UploadInput) (string, error) {
	opts := minio.PutObjectOptions{
		ContentType: input.ContentType,
	}

	_, err := fs.client.PutObject(ctx, fs.bucket, input.Name, input.File, input.Size, opts)
	if err != nil {
		return "", err
	}

	return fs.generateFileURL(input.Name), nil
}

// Get path with bucket <bucket>://<path_to_file_in_bucket>
func (fs *FileStorage) generateFileURL(filename string) string {
	return fmt.Sprintf("%s://%s", fs.bucket, filename)
}
