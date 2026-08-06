package filemanager

import "strings"

// FileEntry is one parsed row from `ls -l`/`ls -la` output.
type FileEntry struct {
	Permissions string
	Links       string
	Owner       string
	Group       string
	Size        string
	Date        string
	Name        string
	LinkTarget  string
	Type        string // "directory" | "file" | "symlink" | "unknown"
}

// parseLsOutput parses `ls -l`/`ls -la` output into structured entries,
// respecting quoted names and symlink " -> " targets.
func parseLsOutput(output string) []FileEntry {
	var entries []FileEntry

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

		entries = append(entries, FileEntry{
			Permissions: permissions, Links: links, Owner: owner, Group: group, Size: size,
			Date: month + " " + day + " " + timeOrYear, Name: name, LinkTarget: linkTarget, Type: entryType,
		})
	}

	return entries
}

// splitFields splits on runs of whitespace, but stops splitting after
// maxFields-1 splits so the final field (the filename, which may itself
// contain spaces) is left intact.
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
