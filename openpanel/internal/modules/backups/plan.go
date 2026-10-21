package backups

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
)

// restorePlan knows where each archive entry goes on the host, emails and dns zones only for domains the account owns
type restorePlan struct {
	sections  map[string]string // backup/html etc. -> host path
	domains   map[string]bool
	mailStore func(domain string) string
	zonesDir  string
	// suspended reports a suspended domain, its vhost is never restored
	suspended func(domain string) bool
}

// domainSuspended reads the live caddy file, opencli domains-suspend swaps it for one serving suspended_domain.html
func domainSuspended(domain string) bool {
	content, err := os.ReadFile(filepath.Join("/etc/openpanel/caddy/domains", domain+".conf"))
	return err == nil && strings.Contains(string(content), "suspended_domain.html")
}

// mailStoreRoot mirrors opencli's backup_mail_dir: email_storage_location when it's absolute, else the account's own mail dir
func mailStoreRoot(userContext string) func(string) string {
	store := ""
	if content, err := os.ReadFile("/etc/openpanel/openadmin/config/admin.ini"); err == nil {
		for _, line := range strings.Split(string(content), "\n") {
			if v, ok := strings.CutPrefix(strings.TrimSpace(line), "email_storage_location="); ok {
				store = strings.TrimSpace(v)
			}
		}
	}
	return func(domain string) string {
		if strings.HasPrefix(store, "/") {
			return filepath.Join(store, domain)
		}
		return filepath.Join("/home", userContext, "mail", domain)
	}
}

func newRestorePlan(ctx context.Context, a *appctx.App, userID int, userContext string) (*restorePlan, error) {
	plan := &restorePlan{sections: restoreFilesPathMap(userContext), domains: map[string]bool{}, mailStore: mailStoreRoot(userContext), zonesDir: "/etc/bind/zones", suspended: domainSuspended}
	domains, err := a.AllDomainsForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, d := range domains {
		plan.domains[strings.ToLower(d.DomainURL)] = true
	}
	return plan, nil
}

// dest returns where an archive entry goes, the dir it has to stay inside and whether it gets chowned to the account
func (p *restorePlan) dest(name string) (dest, root string, chown, ok bool) {
	if rest, found := strings.CutPrefix(name, "backup/emails/"); found {
		domain, sub, nested := strings.Cut(rest, "/")
		if !nested || sub == "" || !p.domains[strings.ToLower(domain)] {
			return "", "", false, false
		}
		root = p.mailStore(domain)
		return filepath.Join(root, sub), root, true, true
	}
	if file, found := strings.CutPrefix(name, "backup/dns/"); found {
		domain, isZone := strings.CutSuffix(file, ".zone")
		if !isZone || strings.Contains(file, "/") || !p.domains[strings.ToLower(domain)] {
			return "", "", false, false
		}
		return filepath.Join(p.zonesDir, file), p.zonesDir, false, true
	}
	if file, found := strings.CutPrefix(name, "backup/vhosts/"); found {
		if domain, isConf := strings.CutSuffix(file, ".conf"); isConf && p.suspended != nil && p.suspended(strings.ToLower(domain)) {
			return "", "", false, false
		}
	}
	for prefix, hostPath := range p.sections {
		if name == prefix {
			// a single file like crons.ini, its folder is the root
			return hostPath, filepath.Dir(hostPath), true, true
		}
		if rel, found := strings.CutPrefix(name, prefix+"/"); found && rel != "" {
			return filepath.Join(hostPath, rel), hostPath, true, true
		}
	}
	return "", "", false, false
}

var errOutsideRoot = errors.New("path leaves its restore folder")

// within checks that dir, after following symlinks, is root or below it, so a symlink planted in a volume can't send a root-owned write elsewhere
func within(dir, root string) bool {
	realDir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return false
	}
	realRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return false
	}
	return realDir == realRoot || strings.HasPrefix(realDir, realRoot+string(filepath.Separator))
}

// writeTarEntry writes the current tar entry to dest, new parent dirs under root and the file itself get chowned to uid when it's above 0
func writeTarEntry(tr io.Reader, dest, root string, uid int) error {
	if dest != root && !strings.HasPrefix(dest, root+string(filepath.Separator)) {
		return errOutsideRoot
	}
	destDir := filepath.Dir(dest)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}
	if !within(destDir, root) {
		return errOutsideRoot
	}
	if info, err := os.Lstat(dest); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return errOutsideRoot
	}
	if uid > 0 {
		chownAncestors(destDir, root, uid)
	}
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, tr); err != nil {
		out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	if uid > 0 {
		_ = os.Lchown(dest, uid, uid)
	}
	return nil
}

// extractEntry writes one regular archive entry through the plan, returns the host path it landed at
func (p *restorePlan) extractEntry(tr io.Reader, name string, uid int) (string, bool) {
	dest, root, chown, ok := p.dest(name)
	if !ok {
		return "", false
	}
	owner := uid
	if !chown {
		owner = 0
	}
	if writeTarEntry(tr, dest, root, owner) != nil {
		return "", false
	}
	return dest, true
}

// restoreSelected streams the archive and restores only the entries inside targets (full archive paths), it stops reading once it's past the last one
func restoreSelected(ctx context.Context, store remoteStore, backup string, plan *restorePlan, uid int, targets []string) ([]string, error) {
	rc, err := store.Open(ctx, backup)
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	gz, err := gzip.NewReader(rc)
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	var extracted []string
	layout := newArchiveLayout()
	tr := tar.NewReader(gz)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return extracted, err
		}
		raw := strings.Trim(header.Name, "/")
		name := layout.observe(raw)
		if layout.pastAll(raw, targets) {
			break
		}
		if header.Typeflag != tar.TypeReg {
			continue
		}
		for _, t := range targets {
			if insidePath(name, t) {
				if dest, ok := plan.extractEntry(tr, name, uid); ok {
					extracted = append(extracted, dest)
				}
				break
			}
		}
	}
	return extracted, nil
}
