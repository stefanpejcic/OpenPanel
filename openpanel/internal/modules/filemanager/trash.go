package filemanager

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/paths"
)

const trashRel = ".local/share/Trash"

// moveItemToTrash moves a file or folder into the user's trash directory, picking a unique name if one already exists there - this is the piece handleDeleteFile depends on directly, the /trash list/restore/empty page is a separate later phase
func moveItemToTrash(ctx context.Context, a *appctx.App, root *paths.UserRoot, rel, itemName, userContext string) (string, error) {
	home, err := os.OpenRoot("/home/" + userContext)
	if err != nil {
		return "", err
	}
	defer home.Close()

	if err := home.MkdirAll(trashRel, 0o755); err != nil {
		return "", err
	}

	trashedName := uniqueTrashName(home, path.Base(itemName))
	trashedRel := path.Join(trashRel, trashedName)
	if err := paths.RenameBetween(root.Root, rel, home, trashedRel); err != nil {
		return "", err
	}

	deletionDate := time.Now().Format("2006-01-02T15:04:05")
	line := fmt.Sprintf("%s=%s|deletion_date=%s\n", trashedName, root.Abs(rel), deletionDate)
	infoRel := path.Join(trashRel, ".trash_restore")
	f, err := home.OpenFile(infoRel, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err == nil {
		_, _ = f.WriteString(line)
		_ = f.Close()
	}

	if uid, uidErr := a.GetUID(ctx, userContext); uidErr == nil && uid > 0 {
		_ = home.Lchown(infoRel, uid, uid)
	}

	return filepath.Join("/home", userContext, trashedRel), nil
}

func uniqueTrashName(home *os.Root, originalName string) string {
	candidate := originalName
	counter := 1
	for {
		if _, err := home.Lstat(path.Join(trashRel, candidate)); os.IsNotExist(err) {
			return candidate
		}
		candidate = fmt.Sprintf("%s.%d", originalName, counter)
		counter++
	}
}
