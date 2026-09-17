// Package installmanifest lets an autoinstaller (WordPress, Joomla, Drupal, ...) record exactly which
// top-level entries it created inside a site's docroot, so a later uninstall or failed-install rollback
// can delete precisely those entries instead of the whole docroot directory - the docroot itself belongs
// to the domain, not the app, and removing it forces the user to manually recreate it before reinstalling.
package installmanifest

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
)

// Name is the hidden marker file dropped in the install directory listing every top-level entry the
// installer created there.
const Name = ".op-install-manifest"

// Record snapshots dir's current top-level entries (host path, e.g. hostOSPath) and writes them to dir's
// manifest file. Call this once, right after an install finishes successfully.
func Record(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	var names []string
	for _, e := range entries {
		if e.Name() == Name {
			continue
		}
		names = append(names, e.Name())
	}
	return os.WriteFile(filepath.Join(dir, Name), []byte(strings.Join(names, "\n")), 0o644)
}

// EntriesViaContainer reads back the manifest for containerPath (as seen inside phpContainer) via podman
// exec, returning nil if none exists - e.g. an install made before this mechanism existed.
func EntriesViaContainer(ctx context.Context, userContext, phpContainer, containerPath string) []string {
	out, err := podmanmanager.Command(ctx, userContext, podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "cat", containerPath+"/"+Name)).Output()
	if err != nil {
		return nil
	}
	var names []string
	for _, line := range strings.Split(string(out), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			names = append(names, line)
		}
	}
	return names
}

// RemoveViaContainer deletes exactly the given top-level entries plus the manifest itself inside
// containerPath via podman exec, leaving containerPath itself in place.
func RemoveViaContainer(ctx context.Context, userContext, phpContainer, containerPath string, entries []string) error {
	args := []string{"exec", phpContainer, "rm", "-rf"}
	for _, e := range entries {
		args = append(args, containerPath+"/"+e)
	}
	args = append(args, containerPath+"/"+Name)
	return podmanmanager.Command(ctx, userContext, podmanmanager.PodmanArgv(userContext, args...)).Run()
}

// ClearContentsViaContainer deletes everything currently inside containerPath (but not containerPath
// itself) via podman exec. Safe to use during install-failure rollback within the same request that
// created containerPath fresh - at that point "everything in it" and "everything the install created"
// are the same set, so no manifest is needed yet.
func ClearContentsViaContainer(ctx context.Context, userContext, phpContainer, containerPath string) error {
	script := `cd "$1" && for f in .[!.]* ..?* *; do [ -e "$f" ] || continue; rm -rf -- "$f"; done`
	argv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "sh", "-c", script, "sh", containerPath)
	return podmanmanager.Command(ctx, userContext, argv).Run()
}

// ClearContents deletes everything inside dir (a host path) but leaves dir itself in place - the
// host-side equivalent of ClearContentsViaContainer, for installers that extract host-side and can't
// rely on the container's rootless "root" having permission to touch those files via podman exec.
func ClearContents(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.Name() == Name {
			continue
		}
		if rmErr := os.RemoveAll(filepath.Join(dir, e.Name())); rmErr != nil {
			return rmErr
		}
	}
	return nil
}
