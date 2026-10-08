package backups

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"golang.org/x/crypto/ssh"
)

// readBackupEnv returns the full backup.env key/value map (commented lines excluded, quotes stripped), with SSH_IDENTITY_FILE rewritten from its /var/www/html/ docroot-relative form to the real host path - it's absent entirely for any non-SSH destination, and a Go map just returns "" for a missing key so the "ok" check below handles that without special-casing
func readBackupEnv(username string) (map[string]string, string) {
	userHome := "/home/" + username
	envFile := filepath.Join(userHome, "backup.env")

	config := map[string]string{}
	if content, err := os.ReadFile(envFile); err == nil {
		for _, line := range strings.Split(string(content), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			key, value, ok := strings.Cut(line, "=")
			if !ok {
				continue
			}
			value = strings.Trim(strings.TrimSpace(value), `"'`)
			config[strings.TrimSpace(key)] = value
		}
	}

	if sshKeyPath, ok := config["SSH_IDENTITY_FILE"]; ok && strings.HasPrefix(sshKeyPath, "/var/www/html/") {
		config["SSH_IDENTITY_FILE"] = userHome + "/docker-data/volumes/" + username + "_html_data/_data/" + filepath.Base(sshKeyPath)
	}

	return config, userHome
}

// BackupInfo summarizes one remote backup archive's contents.
type BackupInfo struct {
	BackupFile string   `json:"backup_file"`
	Types      []string `json:"types"`
	Databases  []string `json:"databases"`
	HasFiles   bool     `json:"has_files"`
	HasCrons   bool     `json:"has_crons"`
	Error      string   `json:"error,omitempty"`
}

var dbTimestampSuffixRE = regexp.MustCompile(`_\d{4}-\d{2}-\d{2}_\d{2}-\d{2}-\d{2}$`)

// processBackup lists one remote .tar.gz's members (without downloading it) and classifies what's inside
func processBackup(client *ssh.Client, backup, remotePath string) BackupInfo {
	if !strings.HasSuffix(backup, ".tar.gz") && !strings.HasSuffix(backup, ".tgz") {
		return BackupInfo{BackupFile: backup, Types: []string{}, Databases: []string{}}
	}

	remoteFilePath := remotePath + "/" + backup
	out, _ := runSSHCommand(client, "tar -tzf "+shellQuoteArg(remoteFilePath)+" 2>/dev/null")
	return classifyTarListing(backup, out)
}

// classifyTarListing works out which sections (html/vhosts/mail/mysql/postgres/crons) and database dumps are present, given a `tar -tzf` listing's stdout. Split out from processBackup so it's testable without a live SSH session.
func classifyTarListing(backup, listing string) BackupInfo {
	info := BackupInfo{BackupFile: backup, Types: []string{}, Databases: []string{}}

	seenTypes := map[string]bool{}
	seenDBs := map[string]bool{}

	for _, entry := range strings.Split(listing, "\n") {
		entry = strings.Trim(strings.TrimSpace(entry), "/")
		if entry == "" {
			continue
		}
		parts := strings.Split(entry, "/")
		if len(parts) < 2 || parts[0] != "backup" {
			continue
		}
		section := parts[1]

		if section == "crons.ini" {
			info.HasCrons = true
			if !seenTypes["crons"] {
				seenTypes["crons"] = true
				info.Types = append(info.Types, "crons")
			}
			continue
		}

		switch section {
		case "html", "vhosts", "mail", "mysql", "postgres":
			info.HasFiles = true
			if !seenTypes[section] {
				seenTypes[section] = true
				info.Types = append(info.Types, section)
			}
		}

		filename := parts[len(parts)-1]
		if strings.HasSuffix(filename, ".sql") || strings.HasSuffix(filename, ".sql.gz") {
			dbName := strings.TrimSuffix(strings.TrimSuffix(filename, ".sql.gz"), ".sql")
			dbName = dbTimestampSuffixRE.ReplaceAllString(dbName, "")
			if dbName != "" && !seenDBs[dbName] {
				seenDBs[dbName] = true
				info.Databases = append(info.Databases, dbName)
			}
		}
	}

	return info
}

func shellQuoteArg(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// classifyLocalArchive is classifyTarListing's counterpart for an archive already downloaded to a local file, used by any remoteStore backend that can't inspect contents without downloading first (see classifyViaDownload in store.go). Walks the same tar+gzip structure restoreFilesFromTar/scanSQLMembers read, then reuses classifyTarListing so both paths agree on what "html"/"mysql"/"crons" mean.
func classifyLocalArchive(backup, localPath string) (BackupInfo, error) {
	f, err := os.Open(localPath)
	if err != nil {
		return BackupInfo{}, err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return BackupInfo{}, err
	}
	defer gz.Close()

	var names []string
	tr := tar.NewReader(gz)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return BackupInfo{}, err
		}
		names = append(names, header.Name)
	}

	return classifyTarListing(backup, strings.Join(names, "\n")), nil
}

// backupNamePrefix returns the fixed start of this account's archive names (BACKUP_PRUNING_PREFIX, else BACKUP_FILENAME up to its first format verb), or "" if it can't be worked out
func backupNamePrefix(config map[string]string, userContext string) string {
	name := config["BACKUP_PRUNING_PREFIX"]
	if name == "" {
		name = config["BACKUP_FILENAME"]
	}
	// compose sets USERNAME to the account's context for the backup container
	name = strings.NewReplacer("${USERNAME}", userContext, "$USERNAME", userContext).Replace(name)
	if i := strings.Index(name, "%"); i >= 0 {
		name = name[:i]
	}
	if i := strings.Index(name, "{{"); i >= 0 {
		name = name[:i]
	}
	if strings.Contains(name, "$") {
		return ""
	}
	return name
}

// filterBackupNames keeps only names starting with prefix, so a destination shared by several accounts only gets this one's archives indexed
func filterBackupNames(names []string, prefix string) []string {
	if prefix == "" {
		return names
	}
	var own []string
	for _, name := range names {
		if strings.HasPrefix(name, prefix) {
			own = append(own, name)
		}
	}
	return own
}

// doReindex runs in a background goroutine, connects to whichever destination is currently configured, lists the remote backups, classifies each archive (3 at a time), writes jsonFile, and always removes lockFile
func doReindex(userHome string, config map[string]string, prefix, jsonFile, lockFile string) {
	defer os.Remove(lockFile)

	writeError := func(msg string) {
		b, _ := json.Marshal(map[string]string{"error": msg})
		_ = os.WriteFile(jsonFile, b, 0o644)
	}

	store, err := newRemoteStore(config)
	if err != nil {
		writeError(err.Error())
		return
	}

	ctx := context.Background()
	backupNames, err := store.List(ctx)
	if err != nil {
		writeError(err.Error())
		return
	}
	backupNames = filterBackupNames(backupNames, prefix)

	results := make([]BackupInfo, len(backupNames))
	sem := make(chan struct{}, 3)
	var wg sync.WaitGroup
	for i, name := range backupNames {
		wg.Add(1)
		sem <- struct{}{}
		go func(i int, name string) {
			defer wg.Done()
			defer func() { <-sem }()
			info, err := store.Classify(ctx, name)
			if err != nil {
				info = BackupInfo{BackupFile: name, Types: []string{}, Databases: []string{}, Error: err.Error()}
			}
			results[i] = info
		}(i, name)
	}
	wg.Wait()

	// results is indexed by submission order, not completion order, so the output stays deterministic regardless of which goroutine finishes first
	b, err := json.MarshalIndent(results, "", "    ")
	if err != nil {
		writeError(err.Error())
		return
	}
	_ = os.WriteFile(jsonFile, b, 0o644)
}
