package backups

import (
	"context"
	"errors"
)

// remoteStore abstracts listing and fetching backup archives from whichever
// destination is currently configured in backup.env (s3/webdav/ssh/azure/
// dropbox). doReindex, handleRestoreFromBackup and handleDownloadBackup all
// go through this interface instead of talking to a specific destination
// directly, so adding a new destination type only means adding one more
// implementation here.
type remoteStore interface {
	// List returns the backup filenames present at the configured remote
	// path (or bucket/container/folder, depending on the backend).
	List(ctx context.Context) ([]string, error)
	// Fetch downloads one backup archive to a local temp file. The caller
	// must call cleanup() once done with the local copy.
	Fetch(ctx context.Context, filename string) (localPath string, cleanup func(), err error)
	// Classify reports what a single archive contains (types/databases/
	// files/crons), without necessarily keeping a full local copy around
	// afterward.
	Classify(ctx context.Context, filename string) (BackupInfo, error)
}

// detectActiveTarget reports which backup.env section (s3/webdav/ssh/azure/
// dropbox) the given config map belongs to, the same way
// handleBackupTarget/handleBackupsPage's grouped.MatchedSections check
// does, just against the flat map readBackupEnv already returns instead of
// re-parsing the file.
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

// newRemoteStore builds the remoteStore implementation matching whichever
// destination is currently active in config.
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

// classifyViaDownload is the shared fallback for every backend that has no
// way to list an archive's contents without a full download first (i.e.
// everything except SSH, which can run `tar -tzf` remotely over the same
// connection instead): fetch the archive to a local temp file, classify it
// exactly like the SSH path does, then clean up.
func classifyViaDownload(ctx context.Context, store remoteStore, filename string) (BackupInfo, error) {
	localPath, cleanup, err := store.Fetch(ctx, filename)
	if err != nil {
		return BackupInfo{}, err
	}
	defer cleanup()

	return classifyLocalArchive(filename, localPath)
}
