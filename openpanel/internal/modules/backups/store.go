package backups

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// remoteStore abstracts listing and fetching backup archives from whichever destination is currently configured in backup.env. doReindex, handleRestoreFromBackup and handleDownloadBackup all go through this interface, so adding a new destination type only means adding one more implementation here.
type remoteStore interface {
	// List returns the backup files present at the configured remote path (or bucket/container/folder, depending on the backend)
	List(ctx context.Context) ([]remoteFile, error)
	// Open streams one backup archive, the caller closes it, closing early stops the download
	Open(ctx context.Context, filename string) (io.ReadCloser, error)
}

// remoteFile is one entry of a destination listing, Size is 0 when the backend doesn't report it
type remoteFile struct {
	Name string
	Size int64
}

// memberLister is for backends that can list an archive remotely (ssh runs tar -tzf there), used for old archives with no manifest
type memberLister interface {
	ListMembers(ctx context.Context, filename string) (string, error)
}

// detectActiveTarget reports which backup.env section the given config map belongs to, the same way handleBackupTarget/handleBackupsPage's grouped.MatchedSections check does, just against the flat map readBackupEnv returns instead of re-parsing the file
func detectActiveTarget(config map[string]string) string {
	for _, section := range sectionOrder {
		for _, key := range sectionKeys[section] {
			if config[key] != "" {
				return section
			}
		}
	}
	return ""
}

// newRemoteStore builds the remoteStore implementation matching whichever destination is currently active in config
func newRemoteStore(config map[string]string) (remoteStore, error) {
	switch detectActiveTarget(config) {
	case "ssh":
		return &sshStore{config: config}, nil
	case "s3":
		return &s3Store{config: config}, nil
	case "webdav":
		return &webdavStore{config: config}, nil
	case "azure":
		return &azureStore{config: config}, nil
	case "dropbox":
		return &dropboxStore{config: config}, nil
	default:
		return nil, errors.New("no backup destination configured")
	}
}

// fetchArchive downloads one backup archive to a local temp file, the caller must call cleanup() once done with it
func fetchArchive(ctx context.Context, store remoteStore, filename string) (localPath string, cleanup func(), err error) {
	rc, err := store.Open(ctx, filename)
	if err != nil {
		return "", nil, err
	}
	defer rc.Close()

	tmpDir, err := os.MkdirTemp("", "backup_restore_")
	if err != nil {
		return "", nil, err
	}
	cleanup = func() { _ = os.RemoveAll(tmpDir) }

	localPath = filepath.Join(tmpDir, filepath.Base(filename))
	localFile, err := os.Create(localPath)
	if err != nil {
		cleanup()
		return "", nil, err
	}
	defer localFile.Close()

	if _, err := io.Copy(localFile, rc); err != nil {
		cleanup()
		return "", nil, err
	}
	return localPath, cleanup, nil
}

// classifyArchive reads only the 00-meta/ manifest at the front of the archive when there is one, older archives get listed in full
func classifyArchive(ctx context.Context, store remoteStore, filename string, saveList func(kind string, r io.Reader) error) (BackupInfo, error) {
	if !strings.HasSuffix(filename, ".tar.gz") && !strings.HasSuffix(filename, ".tgz") {
		return BackupInfo{BackupFile: filename, Types: []string{}, Databases: []string{}}, nil
	}

	rc, err := store.Open(ctx, filename)
	if err != nil {
		return BackupInfo{}, err
	}
	var saved []string
	if saveList != nil {
		inner := saveList
		saveList = func(kind string, r io.Reader) error {
			saved = append(saved, kind)
			return inner(kind, r)
		}
	}
	lister, canList := store.(memberLister)
	// a remote lister is cheaper than pulling the whole archive through here, so only peek at the start in that case
	info, found, err := classifyArchiveStream(filename, rc, !canList, saveList)
	rc.Close()
	if err != nil || found || !canList {
		return info, err
	}

	listing, err := lister.ListMembers(ctx, filename)
	if err != nil && strings.TrimSpace(listing) == "" {
		return BackupInfo{BackupFile: filename, Types: []string{}, Databases: []string{}, Error: "Could not read archive contents: " + err.Error()}, nil
	}
	return withLists(classifyTarListing(filename, listing), saved), nil
}
