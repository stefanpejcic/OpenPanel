package backups

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"golang.org/x/crypto/ssh"
)

// readBackupEnv returns the full (commented lines excluded, quotes
// stripped) backup.env key/value map, with SSH_IDENTITY_FILE rewritten
// from its /var/www/html/ docroot-relative form to the real host path.
//
// SSH_IDENTITY_FILE is absent entirely for any non-SSH destination; a Go
// map simply returns "" for a missing key, so the "ok" check below handles
// that case without any special-casing.
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

// processBackup lists one remote .tar.gz's members (without downloading
// it) and classifies what's inside.
func processBackup(client *ssh.Client, backup, remotePath string) BackupInfo {
	if !strings.HasSuffix(backup, ".tar.gz") && !strings.HasSuffix(backup, ".tgz") {
		return BackupInfo{BackupFile: backup, Types: []string{}, Databases: []string{}}
	}

	remoteFilePath := remotePath + "/" + backup
	out, _ := runSSHCommand(client, "tar -tzf "+shellQuoteArg(remoteFilePath)+" 2>/dev/null")
	return classifyTarListing(backup, out)
}

// classifyTarListing works out which sections (html/vhosts/mail/mysql/
// postgres/crons) and database dumps are present, given a `tar -tzf`
// listing's stdout. Split out from processBackup so it's testable
// without a live SSH session.
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

// doReindex runs in a background goroutine, connects over SSH, lists the
// remote backup directory, classifies each archive (3 at a time), writes
// jsonFile, and always removes lockFile.
func doReindex(userHome string, config map[string]string, jsonFile, lockFile string) {
	defer os.Remove(lockFile)

	writeError := func(msg string) {
		b, _ := json.Marshal(map[string]string{"error": msg})
		_ = os.WriteFile(jsonFile, b, 0o644)
	}

	client, err := dialSSH(config)
	if err != nil {
		writeError(err.Error())
		return
	}
	defer client.Close()

	remotePath := config["SSH_REMOTE_PATH"]
	out, err := runSSHCommand(client, "ls -1 "+shellQuoteArg(remotePath))
	if err != nil && out == "" {
		writeError(err.Error())
		return
	}

	var backupNames []string
	for _, line := range strings.Split(out, "\n") {
		if line != "" {
			backupNames = append(backupNames, line)
		}
	}

	results := make([]BackupInfo, len(backupNames))
	sem := make(chan struct{}, 3)
	var wg sync.WaitGroup
	for i, name := range backupNames {
		wg.Add(1)
		sem <- struct{}{}
		go func(i int, name string) {
			defer wg.Done()
			defer func() { <-sem }()
			results[i] = processBackup(client, name, remotePath)
		}(i, name)
	}
	wg.Wait()

	// results is indexed by submission order, not completion order, so
	// the output list is deterministic even though the workers finish in
	// whatever order the goroutines happen to complete.
	b, err := json.MarshalIndent(results, "", "    ")
	if err != nil {
		writeError(err.Error())
		return
	}
	_ = os.WriteFile(jsonFile, b, 0o644)
}

func injected(a *appctx.App, r *http.Request) (username, userContext string, err error) {
	userID, _ := auth.UserID(r)
	data, err := a.InjectData(r.Context(), userID)
	if err != nil {
		return "", "", err
	}
	username, _ = data["current_username"].(string)
	userContext, _ = data["context"].(string)
	return username, userContext, nil
}
