package backups

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

var dumpNameRE = regexp.MustCompile(`^[A-Za-z0-9_$-]+$`)

// downloadTargets turns picked items into archive paths, unlike restores these don't need the domain on the account since nothing is written
func downloadTargets(kind string, values []string) ([]string, error) {
	var targets []string
	for _, v := range values {
		v = strings.TrimSpace(v)
		switch kind {
		case "databases":
			engine, name, ok := strings.Cut(v, "/")
			if !ok || (engine != "mysql" && engine != "postgres") || !dumpNameRE.MatchString(name) {
				return nil, fmt.Errorf("invalid database: %s", v)
			}
			// no extension, the dump can be .sql, .sql.gz or timestamped like restore accepts
			targets = append(targets, "backup/"+engine+"/"+name)
		case "mailboxes":
			domain, user, ok := strings.Cut(strings.ToLower(v), "/")
			if !ok || !domainNameRE.MatchString(domain) || user == "" || strings.ContainsAny(user, "/\x00") || user == "." || user == ".." {
				return nil, fmt.Errorf("invalid mailbox: %s", v)
			}
			targets = append(targets, "backup/emails/"+domain+"/"+user)
		case "domains":
			domain := strings.ToLower(v)
			if !domainNameRE.MatchString(domain) {
				return nil, fmt.Errorf("invalid domain: %s", v)
			}
			targets = append(targets, "backup/dns/"+domain+".zone", "backup/vhosts/"+domain+".conf")
		case "paths":
			clean, ok := cleanBackupPath(v)
			if !ok || clean == "" {
				return nil, fmt.Errorf("invalid path: %s", v)
			}
			targets = append(targets, "backup/html/"+clean)
		default:
			return nil, fmt.Errorf("unknown kind: %s", kind)
		}
	}
	if len(targets) == 0 {
		return nil, errors.New("nothing selected")
	}
	return targets, nil
}

// handleDownloadItems streams the picked items of one backup to the browser, a single database dump as is and anything else as a .tar.gz, without reading past the last one
func handleDownloadItems(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	_, username, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	_ = r.ParseMultipartForm(1 << 20)
	backup := strings.TrimSpace(r.Form.Get("backup_file"))
	if backup == "" || filepath.Base(backup) != backup {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid backup filename."})
		return
	}
	kind := r.Form.Get("kind")
	targets, err := downloadTargets(kind, r.Form["items"])
	if err != nil {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	config, _ := readBackupEnv(userContext)
	if !mayAccessBackup(userContext, config, backup) {
		web.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "This backup does not belong to your account."})
		return
	}
	store, err := newRemoteStore(config)
	if err != nil {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	rc, err := store.Open(r.Context(), backup)
	if err != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	defer rc.Close()
	sent, err := streamItems(w, rc, kind, backup, targets)
	if err != nil && sent == 0 {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if err != nil {
		// headers are already out, so drop the connection to make the browser mark the download as failed instead of saving a truncated file
		panic(http.ErrAbortHandler)
	}
	if sent == 0 {
		web.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "The selected items were not found in this backup."})
		return
	}
	_ = logger.RecordUserAction(a.Config, username, "downloaded "+strings.Join(r.Form["items"], ", ")+" from backup "+backup, reqip.ClientIP(r))
}

// streamItems copies the entries under targets out of the archive to w, returns how many it sent, an item cut off by an error counts so the caller knows output started
func streamItems(w http.ResponseWriter, archive io.Reader, kind, backup string, targets []string) (int, error) {
	gz, err := gzip.NewReader(archive)
	if err != nil {
		return 0, err
	}
	defer gz.Close()
	// one database is most useful as the plain .sql, everything else gets packed
	single := kind == "databases" && len(targets) == 1
	archiveName := kind + "-" + strings.TrimSuffix(strings.TrimSuffix(backup, ".tar.gz"), ".tgz") + ".tar.gz"
	var tw *tar.Writer
	var outGz *gzip.Writer
	sent := 0
	matches, stopAt := insidePath, targets
	if kind == "databases" {
		matches, stopAt = isDumpOf, nil
		for _, t := range targets {
			stopAt = append(stopAt, path.Dir(t))
		}
	}
	layout := newArchiveLayout()
	tr := tar.NewReader(gz)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return sent, err
		}
		raw := strings.Trim(header.Name, "/")
		name := layout.observe(raw)
		if layout.pastAll(raw, stopAt) {
			break
		}
		if header.Typeflag != tar.TypeReg && header.Typeflag != tar.TypeDir {
			continue
		}
		wanted := false
		for _, t := range targets {
			if matches(name, t) {
				wanted = true
				break
			}
		}
		if !wanted {
			continue
		}

		if single {
			if header.Typeflag != tar.TypeReg {
				continue
			}
			contentType := "application/sql"
			if strings.HasSuffix(name, ".gz") {
				contentType = "application/gzip"
			}
			w.Header().Set("Content-Type", contentType)
			w.Header().Set("Content-Disposition", "attachment; filename=\""+path.Base(name)+"\"")
			w.Header().Set("Content-Length", strconv.FormatInt(header.Size, 10))
			if _, err := io.Copy(w, tr); err != nil {
				return 1, err
			}
			sent++
			break
		}
		if tw == nil {
			w.Header().Set("Content-Type", "application/gzip")
			w.Header().Set("Content-Disposition", "attachment; filename=\""+archiveName+"\"")
			outGz = gzip.NewWriter(w)
			tw = tar.NewWriter(outGz)
		}
		// inside the download the paths start at what was picked, without the backup/ prefix
		out := *header
		out.Name = strings.TrimPrefix(name, "backup/")
		if header.Typeflag == tar.TypeDir {
			out.Name += "/"
		}
		if err := tw.WriteHeader(&out); err != nil {
			return sent + 1, err
		}
		if header.Typeflag == tar.TypeReg {
			if _, err := io.Copy(tw, tr); err != nil {
				return sent + 1, err
			}
		}
		sent++
	}
	if tw != nil {
		_ = tw.Close()
		_ = outGz.Close()
	}
	return sent, nil
}

// isDumpOf reports whether member is a dump of target, backup/<engine>/<db> from downloadTargets
func isDumpOf(member, target string) bool {
	engine, isEngine := dumpEngine(member)
	db, isDump := dumpDBName(member)
	return isEngine && isDump && target == "backup/"+engine+"/"+db
}
