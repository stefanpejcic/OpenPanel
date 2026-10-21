package backups

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"
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
	// Size is the archive size reported by the destination, DatabaseSizes the uncompressed dump size per database
	Size          int64            `json:"size,omitempty"`
	DatabaseSizes map[string]int64 `json:"database_sizes,omitempty"`
	// DatabaseEngines is mysql or postgres per database
	DatabaseEngines map[string]string `json:"database_engines,omitempty"`
	// Lists are the 00-meta/<kind>.tsv lists (files, mail, vhosts, dns) saved for picking single items to restore
	Lists []string `json:"lists,omitempty"`
	// IndexVersion lets a reindex redo entries indexed before a field was added
	IndexVersion int `json:"index_version,omitempty"`
}

const indexVersion = 3

var dbTimestampSuffixRE = regexp.MustCompile(`_\d{4}-\d{2}-\d{2}_\d{2}-\d{2}-\d{2}$`)

// classifyTarListing works out which sections (html/vhosts/mail/mysql/postgres/crons) and database dumps are present, given a `tar -tzf` style listing, for archives made before the 00-meta/ manifest existed.
func classifyTarListing(backup, listing string) BackupInfo {
	info := BackupInfo{BackupFile: backup, Types: []string{}, Databases: []string{}}

	seenTypes := map[string]bool{}
	seenDBs := map[string]bool{}
	sawEntries := false

	for _, entry := range strings.Split(listing, "\n") {
		entry = strings.Trim(strings.TrimSpace(entry), "/")
		if entry == "" {
			continue
		}
		entry, _ = canonicalName(entry)
		sawEntries = true
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

		// emails/ holds the real mail dirs, mail/ is the legacy volume of symlinks
		if section == "emails" {
			section = "mail"
		}
		switch section {
		case "html", "vhosts", "mail", "mysql", "postgres", "dns":
			info.HasFiles = true
			if !seenTypes[section] {
				seenTypes[section] = true
				info.Types = append(info.Types, section)
			}
		}

		// only dumps count, a site's own .sql files under html/ aren't databases
		filename := parts[len(parts)-1]
		if (section == "mysql" || section == "postgres") && (strings.HasSuffix(filename, ".sql") || strings.HasSuffix(filename, ".sql.gz")) {
			dbName := strings.TrimSuffix(strings.TrimSuffix(filename, ".sql.gz"), ".sql")
			dbName = dbTimestampSuffixRE.ReplaceAllString(dbName, "")
			if dbName != "" && !seenDBs[dbName] {
				seenDBs[dbName] = true
				info.Databases = append(info.Databases, dbName)
				if info.DatabaseEngines == nil {
					info.DatabaseEngines = map[string]string{}
				}
				info.DatabaseEngines[dbName] = section
			}
		}
	}

	if len(info.Types) == 0 && sawEntries {
		info.Error = "Archive is not in the backup/ layout this page can restore (for example a full account backup made by the administrator)."
	}
	return info
}

func shellQuoteArg(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

const metaDir = "backup/00-meta"

// listKinds are the 00-meta/<kind>.tsv lists the backup container's files-list.sh hook writes
var listKinds = map[string]bool{"files": true, "mail": true, "vhosts": true, "dns": true}

// listKind returns the kind of a 00-meta/<kind>.tsv member, "" for anything else
func listKind(name string) string {
	file, ok := strings.CutPrefix(name, metaDir+"/")
	if !ok {
		return ""
	}
	kind, ok := strings.CutSuffix(file, ".tsv")
	if !ok || !listKinds[kind] {
		return ""
	}
	return kind
}

// backupMeta is one engine's file in backup/00-meta/, written by the db dump hooks right before the archive is made
type backupMeta struct {
	Format    int              `json:"format"`
	Engine    string           `json:"engine"`
	Databases []string         `json:"databases"`
	Sizes     map[string]int64 `json:"sizes"`
}

// classifyArchiveStream reads the archive's 00-meta/ files (always the first members) and stops there, without a usable manifest it either lists every member or, with listAll false, reports found=false so the caller can list it some other way
// saveList gets each 00-meta/<kind>.tsv the archive has, nil skips them
func classifyArchiveStream(backup string, r io.Reader, listAll bool, saveList func(kind string, r io.Reader) error) (info BackupInfo, found bool, err error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return BackupInfo{}, false, err
	}
	defer gz.Close()

	var names []string
	var metas []backupMeta
	dbSizes := map[string]int64{}
	pastMeta := false
	var saved []string
	tr := tar.NewReader(gz)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return BackupInfo{}, false, err
		}
		name := strings.Trim(header.Name, "/")
		canonical, _ := canonicalName(name)
		names = append(names, canonical)
		if _, isDump := dumpEngine(canonical); isDump {
			if db, ok := dumpDBName(canonical); ok {
				dbSizes[db] += header.Size
			}
		}
		if pastMeta {
			continue
		}

		if name == "backup" || name == metaDir {
			continue
		}
		if kind := listKind(name); kind != "" && header.Typeflag == tar.TypeReg && saveList != nil {
			if err := saveList(kind, tr); err != nil {
				return BackupInfo{}, false, err
			}
			saved = append(saved, kind)
			continue
		}
		if strings.HasPrefix(name, metaDir+"/") {
			if header.Typeflag == tar.TypeReg && strings.HasSuffix(name, ".json") {
				var m backupMeta
				if json.NewDecoder(io.LimitReader(tr, 1<<20)).Decode(&m) == nil && m.Format >= 1 {
					metas = append(metas, m)
				}
			}
			continue
		}

		if info, ok := infoFromMeta(backup, metas); ok {
			return withLists(info, saved), true, nil
		}
		if !listAll {
			return BackupInfo{}, false, nil
		}
		pastMeta = true
	}

	if info, ok := infoFromMeta(backup, metas); ok {
		return withLists(info, saved), true, nil
	}
	if !listAll {
		return BackupInfo{}, false, nil
	}
	info = classifyTarListing(backup, strings.Join(names, "\n"))
	info = withLists(info, saved)
	for _, db := range info.Databases {
		if size, ok := dbSizes[db]; ok {
			if info.DatabaseSizes == nil {
				info.DatabaseSizes = map[string]int64{}
			}
			info.DatabaseSizes[db] = size
		}
	}
	return info, false, nil
}

// withLists records the saved lists, a dns list also means the archive has dns zones
func withLists(info BackupInfo, saved []string) BackupInfo {
	info.Lists = saved
	for _, kind := range saved {
		if kind == "dns" && !slices.Contains(info.Types, "dns") {
			info.Types = append(info.Types, "dns")
		}
	}
	return info
}

// infoFromMeta needs mysql.json, it's missing when the db container still runs without the 00-meta mount and then its dumps would go unlisted
func infoFromMeta(backup string, metas []backupMeta) (BackupInfo, bool) {
	byEngine := map[string]backupMeta{}
	for _, m := range metas {
		byEngine[m.Engine] = m
	}
	if _, ok := byEngine["mysql"]; !ok {
		return BackupInfo{}, false
	}

	// html, mail, vhosts and crons.ini are always mounted into the backup container, same order classifyTarListing finds them in
	info := BackupInfo{BackupFile: backup, Types: []string{"crons", "html", "mail"}, Databases: []string{}, HasFiles: true, HasCrons: true}
	seen := map[string]bool{}
	for _, engine := range []string{"mysql", "postgres"} {
		m, ok := byEngine[engine]
		if !ok {
			continue
		}
		info.Types = append(info.Types, engine)
		for _, db := range m.Databases {
			if db != "" && !seen[db] {
				seen[db] = true
				info.Databases = append(info.Databases, db)
				if info.DatabaseEngines == nil {
					info.DatabaseEngines = map[string]string{}
				}
				info.DatabaseEngines[db] = engine
				if size, ok := m.Sizes[db]; ok {
					if info.DatabaseSizes == nil {
						info.DatabaseSizes = map[string]int64{}
					}
					info.DatabaseSizes[db] = size
				}
			}
		}
	}
	info.Types = append(info.Types, "vhosts")
	return info, true
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
// ReindexUser rebuilds a user's available_backups.json in the foreground, for opencli docker-backup after it uploads a backup
func ReindexUser(userContext string) error {
	if !validReindexContext.MatchString(userContext) {
		return fmt.Errorf("invalid user context %q", userContext)
	}
	config, userHome := readBackupEnv(userContext)
	if _, err := os.Stat(userHome); err != nil {
		return err
	}
	jsonFile := filepath.Join(userHome, "available_backups.json")
	lockFile := filepath.Join(userHome, "reindex.lock")
	if _, err := os.Stat(lockFile); err == nil {
		return errors.New("reindex already in progress")
	}
	if err := os.WriteFile(lockFile, []byte("running"), 0o644); err != nil {
		return err
	}
	doReindex(userHome, config, backupNamePrefix(config, userContext), jsonFile, lockFile)

	// doReindex reports failures inside the index file, pass them on so docker-backup can log them
	var failed struct {
		Error string `json:"error"`
	}
	if content, err := os.ReadFile(jsonFile); err == nil && json.Unmarshal(content, &failed) == nil && failed.Error != "" {
		return errors.New(failed.Error)
	}
	return nil
}

var validReindexContext = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

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
	files, err := store.List(ctx)
	if err != nil {
		writeError(err.Error())
		return
	}
	sizes := map[string]int64{}
	var allNames []string
	for _, f := range files {
		sizes[f.Name] = f.Size
		allNames = append(allNames, f.Name)
	}
	backupNames := filterBackupNames(allNames, prefix)
	pruneFileLists(userHome, backupNames)

	// archive names carry a timestamp and never change, so anything indexed fine last time is reused while its size still matches
	known := map[string]BackupInfo{}
	if content, err := os.ReadFile(jsonFile); err == nil {
		var previous []BackupInfo
		if json.Unmarshal(content, &previous) == nil {
			for _, info := range previous {
				if info.Error == "" && info.BackupFile != "" {
					known[info.BackupFile] = info
				}
			}
		}
	}

	results := make([]BackupInfo, len(backupNames))
	sem := make(chan struct{}, 3)
	var wg sync.WaitGroup
	for i, name := range backupNames {
		if info, ok := known[name]; ok && info.IndexVersion == indexVersion && info.Size == sizes[name] && info.Size > 0 && listsExist(userHome, name, info.Lists) {
			results[i] = info
			continue
		}
		wg.Add(1)
		sem <- struct{}{}
		go func(i int, name string) {
			defer wg.Done()
			defer func() { <-sem }()
			info, err := classifyArchive(ctx, store, name, func(kind string, r io.Reader) error { return saveList(userHome, name, kind, r) })
			if err != nil {
				info = BackupInfo{BackupFile: name, Types: []string{}, Databases: []string{}, Error: err.Error()}
			}
			info.Size = sizes[name]
			info.IndexVersion = indexVersion
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
