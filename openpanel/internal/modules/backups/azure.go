package backups

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

// azureStore is the remoteStore implementation for the "azure" backup.env
// section (Azure Blob Storage).
type azureStore struct {
	config map[string]string
}

func (s *azureStore) client() (*azblob.Client, string, error) {
	container := s.config["AZURE_STORAGE_CONTAINER_NAME"]
	if container == "" {
		return nil, "", errors.New("AZURE_STORAGE_CONTAINER_NAME is not configured")
	}

	if connStr := s.config["AZURE_STORAGE_CONNECTION_STRING"]; connStr != "" {
		c, err := azblob.NewClientFromConnectionString(connStr, nil)
		if err != nil {
			return nil, "", err
		}
		return c, container, nil
	}

	account := s.config["AZURE_STORAGE_ACCOUNT_NAME"]
	key := s.config["AZURE_STORAGE_PRIMARY_ACCOUNT_KEY"]
	if account == "" || key == "" {
		return nil, "", errors.New("AZURE_STORAGE_ACCOUNT_NAME/AZURE_STORAGE_PRIMARY_ACCOUNT_KEY (or AZURE_STORAGE_CONNECTION_STRING) are not configured")
	}

	cred, err := azblob.NewSharedKeyCredential(account, key)
	if err != nil {
		return nil, "", err
	}

	serviceURL := s.config["AZURE_STORAGE_ENDPOINT"]
	if serviceURL == "" {
		serviceURL = "https://" + account + ".blob.core.windows.net/"
	}

	c, err := azblob.NewClientWithSharedKeyCredential(serviceURL, cred, nil)
	if err != nil {
		return nil, "", err
	}
	return c, container, nil
}

func (s *azureStore) List(ctx context.Context) ([]string, error) {
	client, container, err := s.client()
	if err != nil {
		return nil, err
	}

	var names []string
	pager := client.NewListBlobsFlatPager(container, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		if page.Segment == nil {
			continue
		}
		for _, item := range page.Segment.BlobItems {
			if item.Name != nil {
				names = append(names, *item.Name)
			}
		}
	}
	return names, nil
}

func (s *azureStore) Fetch(ctx context.Context, filename string) (localPath string, cleanup func(), err error) {
	client, container, err := s.client()
	if err != nil {
		return "", nil, err
	}

	tmpDir, err := os.MkdirTemp("", "backup_restore_")
	if err != nil {
		return "", nil, err
	}
	cleanup = func() { _ = os.RemoveAll(tmpDir) }

	localPath = filepath.Join(tmpDir, filename)
	localFile, err := os.Create(localPath)
	if err != nil {
		cleanup()
		return "", nil, err
	}
	defer localFile.Close()

	if _, err := client.DownloadFile(ctx, container, filename, localFile, nil); err != nil {
		cleanup()
		return "", nil, err
	}

	return localPath, cleanup, nil
}

func (s *azureStore) Classify(ctx context.Context, filename string) (BackupInfo, error) {
	return classifyViaDownload(ctx, s, filename)
}
