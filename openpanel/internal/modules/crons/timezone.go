package crons

import (
	"bytes"
	"context"
	"encoding/json"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"

	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/php"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// hostLocaltimePath is the host's /etc/localtime, bind-mounted into the openpanel container
var (
	hostLocaltimePath = "/etc/localtime"
	zoneinfoDir       = "/usr/share/zoneinfo"
)

// cronTimeZone is the zone the user's cron container runs schedules in: its TZ, else the host's zone when /etc/localtime is mounted into it, else UTC like ofelia defaults to
func cronTimeZone(userContext string) *time.Location {
	raw, err := os.ReadFile(filepath.Join("/home", userContext, "docker-compose.yml"))
	if err != nil {
		return time.UTC
	}
	tz, mountsLocaltime := cronServiceZone(raw)
	if tz != "" {
		if ref := envRefRE.FindStringSubmatch(tz); ref != nil {
			tz = ref[2]
			if v, ok := docker.GetEnvValue(userContext, ref[1]); ok && v != "" {
				tz = v
			}
		}
		if loc, err := time.LoadLocation(strings.Trim(tz, `"' `)); err == nil {
			return loc
		}
	}
	if mountsLocaltime {
		return hostLocation()
	}
	return time.UTC
}

// envRefRE matches ${NAME} and ${NAME:-default}
var envRefRE = regexp.MustCompile(`^\$\{(\w+)(?::-([^}]*))?\}$`)

// cronServiceZone reads the cron service's TZ and whether it mounts /etc/localtime from a compose file
func cronServiceZone(compose []byte) (tz string, mountsLocaltime bool) {
	var doc struct {
		Services map[string]struct {
			Environment yaml.Node `yaml:"environment"`
			Volumes     []any     `yaml:"volumes"`
		} `yaml:"services"`
	}
	if yaml.Unmarshal(compose, &doc) != nil {
		return "", false
	}
	svc, ok := doc.Services["cron"]
	if !ok {
		return "", false
	}
	// environment is either a map or a list of NAME=value
	switch svc.Environment.Kind {
	case yaml.MappingNode:
		for i := 0; i+1 < len(svc.Environment.Content); i += 2 {
			if svc.Environment.Content[i].Value == "TZ" {
				tz = svc.Environment.Content[i+1].Value
			}
		}
	case yaml.SequenceNode:
		for _, item := range svc.Environment.Content {
			if v, ok := strings.CutPrefix(item.Value, "TZ="); ok {
				tz = v
			}
		}
	}
	for _, v := range svc.Volumes {
		switch vol := v.(type) {
		case string:
			parts := strings.Split(vol, ":")
			mountsLocaltime = mountsLocaltime || (len(parts) > 1 && parts[1] == "/etc/localtime")
		case map[string]any:
			mountsLocaltime = mountsLocaltime || vol["target"] == "/etc/localtime"
		}
	}
	return tz, mountsLocaltime
}

var (
	hostZoneOnce sync.Once
	hostZone     *time.Location
)

// hostLocation is the host's zone under its IANA name, found by matching /etc/localtime against the zoneinfo files, since the browser needs a name to format times in it
func hostLocation() *time.Location {
	hostZoneOnce.Do(func() {
		hostZone = time.UTC
		data, err := os.ReadFile(hostLocaltimePath)
		if err != nil {
			return
		}
		name := ""
		_ = filepath.WalkDir(zoneinfoDir, func(path string, d fs.DirEntry, err error) error {
			if name != "" {
				return fs.SkipAll
			}
			if err != nil {
				return nil
			}
			rel, _ := filepath.Rel(zoneinfoDir, path)
			// posix/ and right/ are copies of the same zones
			if d.IsDir() && (rel == "posix" || rel == "right") {
				return fs.SkipDir
			}
			if d.IsDir() || !strings.Contains(rel, "/") && rel != "UTC" {
				return nil
			}
			if b, err := os.ReadFile(path); err == nil && bytes.Equal(b, data) {
				name = rel
			}
			return nil
		})
		if name == "" {
			return
		}
		if loc, err := time.LoadLocationFromTZData(name, data); err == nil {
			hostZone = loc
		}
	})
	return hostZone
}

// setComposeCronTZ sets TZ on the cron service of a compose file, editing only those lines so the rest of the file keeps its formatting
func setComposeCronTZ(compose, tz string) (string, error) {
	lines := strings.Split(compose, "\n")
	indent := func(l string) int { return len(l) - len(strings.TrimLeft(l, " ")) }
	blank := func(l string) bool { t := strings.TrimSpace(l); return t == "" || strings.HasPrefix(t, "#") }

	// the cron: key directly under services:
	svc, svcIndent := -1, 0
	inServices, servicesIndent, childIndent := false, 0, -1
	for i, l := range lines {
		if blank(l) {
			continue
		}
		if strings.TrimSpace(l) == "services:" {
			inServices, servicesIndent, childIndent = true, indent(l), -1
			continue
		}
		if !inServices {
			continue
		}
		if indent(l) <= servicesIndent {
			inServices = false
			continue
		}
		if childIndent == -1 {
			childIndent = indent(l)
		}
		if indent(l) == childIndent && strings.TrimSpace(l) == "cron:" {
			svc, svcIndent = i, indent(l)
			break
		}
	}
	if svc == -1 {
		return "", errString("The cron service isn't in docker-compose.yml.")
	}
	end := len(lines)
	for i := svc + 1; i < len(lines); i++ {
		if !blank(lines[i]) && indent(lines[i]) <= svcIndent {
			end = i
			break
		}
	}
	keyIndent := -1
	env := -1
	for i := svc + 1; i < end; i++ {
		if blank(lines[i]) {
			continue
		}
		if keyIndent == -1 {
			keyIndent = indent(lines[i])
		}
		if indent(lines[i]) == keyIndent && strings.TrimSpace(lines[i]) == "environment:" {
			env = i
		}
	}
	if keyIndent == -1 {
		keyIndent = svcIndent + 4
	}
	step := keyIndent - svcIndent
	pad := func(n int) string { return strings.Repeat(" ", n) }

	if env == -1 {
		// no environment yet, add it right under cron:
		add := []string{pad(keyIndent) + "environment:", pad(keyIndent+step) + "TZ: " + tz}
		lines = append(lines[:svc+1], append(add, lines[svc+1:]...)...)
		return strings.Join(lines, "\n"), nil
	}
	entryIndent := keyIndent + step
	for i := env + 1; i < end; i++ {
		if blank(lines[i]) {
			continue
		}
		if indent(lines[i]) <= keyIndent {
			break
		}
		entryIndent = indent(lines[i])
		t := strings.TrimSpace(lines[i])
		switch {
		case strings.HasPrefix(t, "TZ:"):
			lines[i] = pad(entryIndent) + "TZ: " + tz
			return strings.Join(lines, "\n"), nil
		case strings.HasPrefix(t, "- TZ="):
			lines[i] = pad(entryIndent) + "- TZ=" + tz
			return strings.Join(lines, "\n"), nil
		}
	}
	entry := "TZ: " + tz
	if next := strings.TrimSpace(lines[env+1]); strings.HasPrefix(next, "- ") {
		entry = "- TZ=" + tz
	}
	lines = append(lines[:env+1], append([]string{pad(entryIndent) + entry}, lines[env+1:]...)...)
	return strings.Join(lines, "\n"), nil
}

// validCronTimeZone accepts the zones the PHP options page offers
func validCronTimeZone(tz string) bool {
	if tz == "UTC" {
		return true
	}
	if _, err := time.LoadLocation(tz); err != nil {
		return false
	}
	for _, z := range php.AvailableTimezones() {
		if z == tz {
			return true
		}
	}
	return false
}

// setCronTimeZone writes TZ for the user's cron service, then recreates it so the new zone applies, a stopped cron with no enabled jobs stays stopped
func setCronTimeZone(ctx context.Context, userContext, tz string) error {
	if !validCronTimeZone(tz) {
		return errString("Unknown time zone.")
	}
	path := filepath.Join("/home", userContext, "docker-compose.yml")
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	updated, err := setComposeCronTZ(string(raw), tz)
	if err != nil {
		return err
	}
	if got, _ := cronServiceZone([]byte(updated)); got != tz {
		return errString("Could not set TZ in docker-compose.yml.")
	}
	info, _ := os.Stat(path)
	if err := os.WriteFile(path, []byte(updated), info.Mode().Perm()); err != nil {
		return err
	}
	jobs, _ := os.ReadFile(cronFilePath(userContext))
	if hasCronJobs(strings.Split(string(jobs), "\n")) {
		// up -d recreates the container, a restart would keep the old environment
		docker.StartOrStopContainer(ctx, userContext, "cron", "activate", "detached")
	}
	return nil
}

// handleCronTimeZone saves the zone picked on the Cron Jobs summary
func handleCronTimeZone(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	username, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	_ = r.ParseForm()
	tz := strings.TrimSpace(r.Form.Get("timezone"))
	if err := setCronTimeZone(r.Context(), userContext, tz); err != nil {
		flashAndRedirect(a, w, r, "error", web.Tr(a, r, err.Error()), "/cronjobs")
		return
	}
	_ = logger.RecordUserAction(a.Config, username, "changed the cron time zone to "+tz, reqip.ClientIP(r))
	flashAndRedirect(a, w, r, "success", web.Tr(a, r, "Cron time zone set to %(tz)s.", "tz", tz), "/cronjobs")
}

// apiCronTimeZoneGet returns the zone the cron service runs in
func apiCronTimeZoneGet(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	_, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeAPICronsJSON(w, http.StatusOK, map[string]string{"timezone": cronTimeZone(userContext).String()})
}

// apiCronTimeZonePut sets the zone, body {"timezone": "Europe/Belgrade"}
func apiCronTimeZonePut(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	username, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	var body struct {
		Timezone string `json:"timezone"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	tz := strings.TrimSpace(body.Timezone)
	if err := setCronTimeZone(r.Context(), userContext, tz); err != nil {
		writeAPICronsJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	_ = logger.RecordUserAction(a.Config, username, "changed the cron time zone to "+tz, reqip.ClientIP(r))
	writeAPICronsJSON(w, http.StatusOK, map[string]string{"message": "Cron time zone set to " + tz, "timezone": tz})
}
