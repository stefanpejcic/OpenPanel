package filemanager

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
)

// moveItemToTrash moves a file or folder into the user's trash directory,
// picking a unique name if one already exists there. This is the piece
// handleDeleteFile depends on directly; the /trash list/restore/empty page
// is a separate later phase.
func moveItemToTrash(ctx context.Context, a *appctx.App, itemPath, itemName, userContext string) (string, error) {
	trashDir := "/home/" + userContext + "/.local/share/Trash"
	trashInfoPath := filepath.Join(trashDir, ".trash_restore")

	if err := os.MkdirAll(trashDir, 0o755); err != nil {
		return "", err
	}

	trashedName := uniqueTrashName(trashDir, itemName)
	trashedPath := filepath.Join(trashDir, trashedName)
	if err := os.Rename(itemPath, trashedPath); err != nil {
		return "", err
	}

	deletionDate := time.Now().Format("2006-01-02T15:04:05")
	line := fmt.Sprintf("%s=%s|deletion_date=%s\n", trashedName, itemPath, deletionDate)
	f, err := os.OpenFile(trashInfoPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err == nil {
		_, _ = f.WriteString(line)
		_ = f.Close()
	}

	if uid, uidErr := a.GetUID(ctx, userContext); uidErr == nil && uid > 0 {
		_ = os.Chown(trashInfoPath, uid, uid)
	}

	return trashedPath, nil
}

func uniqueTrashName(trashDir, originalName string) string {
	candidate := originalName
	counter := 1
	for {
		if _, err := os.Stat(filepath.Join(trashDir, candidate)); os.IsNotExist(err) {
			return candidate
		}
		candidate = fmt.Sprintf("%s.%d", originalName, counter)
		counter++
	}
}
