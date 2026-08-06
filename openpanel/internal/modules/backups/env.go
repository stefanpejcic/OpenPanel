package backups

import (
	"os"
	"strings"
)

// KV is one env-file key=value pair, order-preserving (Go maps aren't -
// and the templates render form fields in the file's line order).
type KV struct {
	Key, Value string
}

// parseUncommentedEnv parses backup.env into key/value pairs: strip
// blank/comment lines, split on the first "=", strip a wrapping pair of
// quotes from the value.
//
// Both single- and double-quote wrappers are stripped, not just double
// quotes: docker-volume-backup's config templates write single-quoted
// values (e.g. AWS_ENDPOINT='s3.amazonaws.com'), and leaving those quotes
// in place would surface the literal quote characters in the settings
// form's input values. This matches readBackupEnv's (the SSH restore
// path's parser) handling of the same file.
func parseUncommentedEnv(path string) ([]KV, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var out []KV
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
		out = append(out, KV{Key: strings.TrimSpace(key), Value: value})
	}
	return out, nil
}

// EnvSections is the parsed-and-grouped shape backup_settings()/backups()
// both build: which SECTIONS targets have at least one key present in the
// file, their values (file order), and the leftover "settings" keys.
type EnvSections struct {
	MatchedSections []string
	SectionValues   map[string][]KV
	Settings        []KV
}

// groupBySections buckets parsed env entries into their matching
// destination sections (s3/webdav/ssh/azure/dropbox) plus a leftover
// "settings" group for keys that don't belong to any section - shared by
// handleBackupSettings, handleBackupTarget's GET branch, and
// handleBackupsPage.
func groupBySections(entries []KV) EnvSections {
	result := EnvSections{SectionValues: map[string][]KV{}}
	used := map[string]bool{}

	for _, section := range sectionOrder {
		var values []KV
		for _, kv := range entries {
			if isSectionKey(section, kv.Key) {
				values = append(values, kv)
			}
		}
		if len(values) > 0 {
			result.MatchedSections = append(result.MatchedSections, section)
			result.SectionValues[section] = values
			for _, kv := range values {
				used[kv.Key] = true
			}
		}
	}

	for _, kv := range entries {
		if !used[kv.Key] {
			result.Settings = append(result.Settings, kv)
		}
	}

	return result
}
