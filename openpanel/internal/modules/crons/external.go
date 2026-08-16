package crons

import (
	"context"
	"os"
	"strings"
)

// AddJob appends a new [job-exec "<comment>"] block to userContext's
// crons.ini and (re)starts the cron container - the same effect as
// handleSaveCronjob's non-interactive path (see handlers.go), exported for
// other modules (e.g. an app installer that needs its own periodic task,
// such as Moodle's admin/cli/cron.php) to register a job directly at
// install time instead of going through the HTTP form. comment must be
// unique within the user's crons.ini - callers that might install more
// than once per user should pick a comment that encodes the site (e.g.
// "moodle-" + domain) and pair this with RemoveJobByComment at uninstall.
func AddJob(ctx context.Context, userContext, comment, schedule, container, command string, noOverlap bool) error {
	path := cronFilePath(userContext)
	block := "[job-exec \"" + comment + "\"]\n" +
		"schedule = " + schedule + "\n" +
		"container = " + container + "\n" +
		"command = " + command
	if noOverlap {
		block += "\nno-overlap"
	}
	if err := writeCronFile(path, block, false); err != nil {
		return err
	}
	restartOrActivateCron(ctx, userContext)
	return nil
}

// RemoveJobByComment deletes the [job-exec "<comment>"] block (if any) from
// userContext's crons.ini and restarts the cron container - the uninstall-
// time counterpart to AddJob. A missing crons.ini or a comment that isn't
// present are both treated as success (nothing to remove).
func RemoveJobByComment(ctx context.Context, userContext, comment string) error {
	path := cronFilePath(userContext)
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	blocks := splitCronBlocks(string(content))
	kept := make([]string, 0, len(blocks))
	found := false
	for _, block := range blocks {
		headerMatch := cronJobHeaderRE.FindStringSubmatch(block)
		if headerMatch != nil && strings.TrimSpace(headerMatch[1]) == comment {
			found = true
			continue
		}
		kept = append(kept, block)
	}
	if !found {
		return nil
	}

	newContent := strings.Join(kept, "\n\n")
	if newContent != "" {
		newContent += "\n\n"
	}
	if writeErr := os.WriteFile(path, []byte(newContent), 0o644); writeErr != nil {
		return writeErr
	}
	restartOrActivateCron(ctx, userContext)
	return nil
}
