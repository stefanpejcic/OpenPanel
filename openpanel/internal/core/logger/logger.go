// Package logger ports modules/core/logger.py: per-user activity logs used
// by modules and plugins to record actions like "logged in", "created
// database x", etc.
package logger

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gist.github.com/stefanpejcic/openpanel/internal/core/config"
)

const activityLogDir = "/etc/openpanel/openpanel/core/users"

// RecordUserAction appends a line to the user's activity.log, then trims it
// if it has grown past the configured retention, matching
// record_user_action().
func RecordUserAction(cfg config.Config, username, action, ipAddress string) error {
	username = strings.TrimPrefix(strings.TrimPrefix(username, "SUSPENDED_"), "suspended_")
	username = strings.ReplaceAll(username, " ", "")

	logFile := filepath.Join(activityLogDir, username, "activity.log")
	if err := os.MkdirAll(filepath.Dir(logFile), 0o755); err != nil {
		return err
	}

	if ipAddress == "" {
		ipAddress = "0.0.0.0"
	}
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	entry := fmt.Sprintf("%s  %s User %s %s\n", timestamp, ipAddress, username, action)

	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	if _, err := f.WriteString(entry); err != nil {
		f.Close()
		return err
	}
	f.Close()

	trimIfOversized(cfg, logFile)
	return nil
}

func trimIfOversized(cfg config.Config, logFile string) {
	maxLines := atoiDefault(cfg.Get("activity_lines_retention", ""), 1000)
	maxSize := int64(atoiDefault(cfg.Get("activity_max_size_bytes", ""), 2_000_000))

	info, err := os.Stat(logFile)
	if err != nil || info.Size() <= maxSize {
		return
	}

	toRead := info.Size()
	if toRead > 200_000 {
		toRead = 200_000
	}

	f, err := os.Open(logFile)
	if err != nil {
		return
	}
	defer f.Close()

	if _, err := f.Seek(-toRead, io.SeekEnd); err != nil {
		return
	}
	chunk := make([]byte, toRead)
	if _, err := f.Read(chunk); err != nil {
		return
	}

	lines := splitLines(chunk)
	if len(lines) > maxLines {
		lines = lines[len(lines)-maxLines:]
	}
	_ = os.WriteFile(logFile, []byte(strings.Join(lines, "\n")), 0o644)
}

func splitLines(chunk []byte) []string {
	var lines []string
	scanner := bufio.NewScanner(bytes.NewReader(chunk))
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}

func atoiDefault(s string, def int) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}
