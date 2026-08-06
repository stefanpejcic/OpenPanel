package trash

import "strings"

// Entry is one parsed trash item: an `ls -l` row enriched with its
// original path and deletion date from .trash_restore.
type Entry struct {
	Permissions  string
	Links        string
	Owner        string
	Group        string
	Size         string
	Date         string
	Name         string
	LinkTarget   string
	Type         string // "directory" | "file" | "symlink" | "unknown"
	DeletionDate string
	OriginalPath string
}

// trashMeta is one .trash_restore line's parsed value.
type trashMeta struct {
	originalPath string
	deletionDate string
}

// parseTrashRestoreFile parses .trash_restore's
// trashed_name=original_path|deletion_date=... lines, best-effort:
// malformed lines are simply skipped.
func parseTrashRestoreFile(content string) map[string]trashMeta {
	metadata := make(map[string]trashMeta)
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || !strings.Contains(line, "=") {
			continue
		}
		left, right, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		if idx := strings.Index(right, "|deletion_date="); idx != -1 {
			metadata[left] = trashMeta{originalPath: right[:idx], deletionDate: right[idx+len("|deletion_date="):]}
		} else {
			metadata[left] = trashMeta{originalPath: right}
		}
	}
	return metadata
}

// parseLsOutputTrash parses `ls -l`/`ls -la` output into structured
// entries, enriched with deletion date/original path looked up from
// .trash_restore's content.
func parseLsOutputTrash(output, trashInfoContent string) []Entry {
	metadata := parseTrashRestoreFile(trashInfoContent)

	var entries []Entry
	for _, line := range strings.Split(output, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}

		parts := splitFields(line, 9)
		if len(parts) < 9 {
			continue
		}

		permissions, links, owner, group, size := parts[0], parts[1], parts[2], parts[3], parts[4]
		month, day, timeOrYear, nameField := parts[5], parts[6], parts[7], parts[8]

		if nameField == "." || nameField == ".." {
			continue
		}

		var linkTarget string
		name := nameField
		if strings.HasPrefix(permissions, "l") {
			if idx := strings.Index(nameField, " -> "); idx != -1 {
				name = nameField[:idx]
				linkTarget = nameField[idx+len(" -> "):]
			}
		}
		name = strings.Trim(name, "'\"")

		var entryType string
		switch {
		case strings.HasPrefix(permissions, "d"):
			entryType = "directory"
		case strings.HasPrefix(permissions, "-"):
			entryType = "file"
		case strings.HasPrefix(permissions, "l"):
			entryType = "symlink"
		default:
			entryType = "unknown"
		}

		entry := Entry{
			Permissions: permissions, Links: links, Owner: owner, Group: group, Size: size,
			Date: month + " " + day + " " + timeOrYear, Name: name, LinkTarget: linkTarget, Type: entryType,
		}
		if meta, ok := metadata[name]; ok {
			entry.OriginalPath = meta.originalPath
			entry.DeletionDate = meta.deletionDate
		}
		entries = append(entries, entry)
	}

	return entries
}

// splitFields splits line on whitespace, stopping after maxFields-1 splits
// so the final field keeps any remaining whitespace-separated content
// intact (up to maxFields elements).
func splitFields(line string, maxFields int) []string {
	var fields []string
	rest := line
	for len(fields) < maxFields-1 {
		rest = strings.TrimLeft(rest, " \t")
		if rest == "" {
			return fields
		}
		idx := strings.IndexAny(rest, " \t")
		if idx == -1 {
			fields = append(fields, rest)
			return fields
		}
		fields = append(fields, rest[:idx])
		rest = rest[idx:]
	}
	rest = strings.TrimLeft(rest, " \t")
	if rest != "" {
		fields = append(fields, rest)
	}
	return fields
}
