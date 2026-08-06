// Package backups implements the backup destination/settings wizard
// (env-file-based credentials for S3/WebDAV/SSH/Azure/Dropbox targets,
// backed by the "backup" service container) plus SSH-based remote backup
// reindexing, restore, and download.
package backups

import (
	"context"
	"os/exec"
	"strings"

	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
)

// sectionOrder/sectionKeys define the backup target sections in a fixed
// order (s3, webdav, ssh, azure, dropbox), used everywhere the route
// handlers build matched_sections/targets lists or JSON output - order
// matters for API responses, so a plain Go map (unordered) isn't a
// substitute.
var sectionOrder = []string{"s3", "webdav", "ssh", "azure", "dropbox"}

var sectionKeys = map[string][]string{
	"s3": {
		"AWS_S3_BUCKET_NAME", "AWS_S3_PATH", "AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY",
		"AWS_IAM_ROLE_ENDPOINT", "AWS_ENDPOINT", "AWS_ENDPOINT_PROTO", "AWS_ENDPOINT_INSECURE",
		"AWS_ENDPOINT_CA_CERT", "AWS_STORAGE_CLASS", "AWS_PART_SIZE",
	},
	"webdav": {"WEBDAV_URL", "WEBDAV_USERNAME", "WEBDAV_PATH", "WEBDAV_PASSWORD", "WEBDAV_URL_INSECURE"},
	"ssh": {
		"SSH_HOST_NAME", "SSH_PORT", "SSH_REMOTE_PATH", "SSH_USER", "SSH_PASSWORD",
		"SSH_IDENTITY_FILE", "SSH_IDENTITY_PASSPHRASE",
	},
	"azure": {
		"AZURE_STORAGE_ACCOUNT_NAME", "AZURE_STORAGE_PRIMARY_ACCOUNT_KEY", "AZURE_STORAGE_CONNECTION_STRING",
		"AZURE_STORAGE_CONTAINER_NAME", "AZURE_STORAGE_ENDPOINT", "AZURE_STORAGE_ACCESS_TIER",
	},
	"dropbox": {"DROPBOX_REMOTE_PATH", "DROPBOX_APP_KEY", "DROPBOX_APP_SECRET", "DROPBOX_CONCURRENCY_LEVEL", "DROPBOX_REFRESH_TOKEN"},
}

func isSectionKey(target, key string) bool {
	for _, k := range sectionKeys[target] {
		if k == key {
			return true
		}
	}
	return false
}

// isBackupInProgress checks the backup container's lock file via `podman
// exec backup test -f /var/run/lock/dockervolumebackup.lock`.
func isBackupInProgress(ctx context.Context, userContext string) bool {
	argv := podmanmanager.PodmanArgv(userContext, "exec", "backup", "test", "-f", "/var/run/lock/dockervolumebackup.lock")
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Env = podmanmanager.PodmanEnv(userContext)
	err := cmd.Run()
	return err == nil
}

// isBackupContainerRunning reports whether the user's "backup" container
// is currently up.
func isBackupContainerRunning(ctx context.Context, userContext string) bool {
	argv := podmanmanager.PodmanArgv(userContext, "ps", "--filter", "name=backup", "--format", "{{.Names}}")
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Env = podmanmanager.PodmanEnv(userContext)
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), "backup")
}
