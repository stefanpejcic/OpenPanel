package backups

import (
	"context"
	"errors"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// s3Store is the remoteStore implementation for the "s3" backup.env
// section. It talks to whatever S3-compatible endpoint the "backup"
// container itself is configured against (AWS S3 by default, or any
// self-hosted/compatible service via AWS_ENDPOINT) using the same
// AWS_ACCESS_KEY_ID/AWS_SECRET_ACCESS_KEY credentials, so there's nothing
// destination-specific to configure beyond what backup.env already has.
type s3Store struct {
	config map[string]string
}

// client builds a fresh minio client from config - cheap enough (no
// network round-trip on construction) that there's no need to cache one
// across List/Fetch/Classify calls, matching sshStore's per-call-connection
// approach.
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

func (s *s3Store) List(ctx context.Context) ([]string, error) {
	client, bucket, prefix, err := s.client()
	if err != nil {
		return nil, err
	}

	listPrefix := prefix
	if listPrefix != "" {
		listPrefix += "/"
	}

	var names []string
	for object := range client.ListObjects(ctx, bucket, minio.ListObjectsOptions{Prefix: listPrefix, Recursive: false}) {
		if object.Err != nil {
			return nil, object.Err
		}
		if strings.HasSuffix(object.Key, "/") {
			continue // a "folder" placeholder, not a backup file
		}
		names = append(names, path.Base(object.Key))
	}
	return names, nil
}

func (s *s3Store) Fetch(ctx context.Context, filename string) (localPath string, cleanup func(), err error) {
	client, bucket, prefix, err := s.client()
	if err != nil {
		return "", nil, err
	}

	objectKey := filename
	if prefix != "" {
		objectKey = prefix + "/" + filename
	}

	tmpDir, err := os.MkdirTemp("", "backup_restore_")
	if err != nil {
		return "", nil, err
	}
	cleanup = func() { _ = os.RemoveAll(tmpDir) }

	localPath = filepath.Join(tmpDir, filename)
	if err := client.FGetObject(ctx, bucket, objectKey, localPath, minio.GetObjectOptions{}); err != nil {
		cleanup()
		return "", nil, err
	}

	return localPath, cleanup, nil
}

func (s *s3Store) Classify(ctx context.Context, filename string) (BackupInfo, error) {
	return classifyViaDownload(ctx, s, filename)
}
