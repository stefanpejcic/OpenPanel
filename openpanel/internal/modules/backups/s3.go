package backups

import (
	"context"
	"errors"
	"io"
	"path"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// s3Store is the remoteStore implementation for the "s3" backup.env section. Talks to whatever S3-compatible endpoint the "backup" container is configured against (AWS S3 by default, or self-hosted via AWS_ENDPOINT) using the same credentials, so nothing extra needs configuring.
type s3Store struct {
	config map[string]string
}

// client builds a fresh minio client from config - cheap enough (no network round-trip on construction) that there's no need to cache one across calls, matching sshStore's per-call-connection approach
func (s *s3Store) client() (*minio.Client, string, string, error) {
	bucket := s.config["AWS_S3_BUCKET_NAME"]
	if bucket == "" {
		return nil, "", "", errors.New("AWS_S3_BUCKET_NAME is not configured")
	}
	accessKey := s.config["AWS_ACCESS_KEY_ID"]
	secretKey := s.config["AWS_SECRET_ACCESS_KEY"]
	if accessKey == "" || secretKey == "" {
		return nil, "", "", errors.New("AWS_ACCESS_KEY_ID/AWS_SECRET_ACCESS_KEY are not configured")
	}

	endpoint := s.config["AWS_ENDPOINT"]
	if endpoint == "" {
		endpoint = "s3.amazonaws.com"
	}
	secure := s.config["AWS_ENDPOINT_PROTO"] != "http"

	c, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: secure,
	})
	if err != nil {
		return nil, "", "", err
	}

	prefix := strings.Trim(s.config["AWS_S3_PATH"], "/")
	return c, bucket, prefix, nil
}

func (s *s3Store) List(ctx context.Context) ([]remoteFile, error) {
	client, bucket, prefix, err := s.client()
	if err != nil {
		return nil, err
	}

	listPrefix := prefix
	if listPrefix != "" {
		listPrefix += "/"
	}

	var files []remoteFile
	for object := range client.ListObjects(ctx, bucket, minio.ListObjectsOptions{Prefix: listPrefix, Recursive: false}) {
		if object.Err != nil {
			return nil, object.Err
		}
		if strings.HasSuffix(object.Key, "/") {
			continue // a "folder" placeholder, not a backup file
		}
		files = append(files, remoteFile{Name: path.Base(object.Key), Size: object.Size})
	}
	return files, nil
}

// GetObject is lazy, so this only downloads as much as the caller actually reads
func (s *s3Store) Open(ctx context.Context, filename string) (io.ReadCloser, error) {
	client, bucket, prefix, err := s.client()
	if err != nil {
		return nil, err
	}

	objectKey := filename
	if prefix != "" {
		objectKey = prefix + "/" + filename
	}
	return client.GetObject(ctx, bucket, objectKey, minio.GetObjectOptions{})
}
