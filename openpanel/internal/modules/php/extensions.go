package php

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
)

// ----------------------------
// Extensions catalog (from docker-php-extension-installer), cached 24h.
// ----------------------------

const (
	extensionsTableFile = "/etc/openpanel/php/extensions_table.json"
	extensionsTableURL  = "https://raw.githubusercontent.com/mlocati/docker-php-extension-installer/master/README.md"
	extensionsTableTTL  = 24 * time.Hour
)

var extensionsTableHeaderVersionRE = regexp.MustCompile(`PHP\s+([\d.]+)`)
var extensionsTableNameFooterRE = regexp.MustCompile(`\[.*$`)

func splitMarkdownTableRow(line string) []string {
	trimmed := strings.Trim(strings.TrimSpace(line), "|")
	parts := strings.Split(trimmed, "|")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

// parseExtensionsTable parses the extension-installer README's markdown
// support table into {extension_name: {"8.3": true, ...}}.
func parseExtensionsTable(body string) map[string]map[string]bool {
	lines := strings.Split(body, "\n")
	headerIdx := -1
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "| Extension |") && strings.Contains(line, "PHP") {
			headerIdx = i
			break
		}
	}
	if headerIdx == -1 {
		return nil
	}

	headerCells := splitMarkdownTableRow(lines[headerIdx])
	versions := make([]string, len(headerCells))
	for i := 1; i < len(headerCells); i++ {
		if m := extensionsTableHeaderVersionRE.FindStringSubmatch(headerCells[i]); m != nil {
			versions[i] = m[1]
		}
	}

	table := map[string]map[string]bool{}
	for i := headerIdx + 2; i < len(lines); i++ {
		if !strings.HasPrefix(strings.TrimSpace(lines[i]), "|") {
			break
		}
		cells := splitMarkdownTableRow(lines[i])
		if len(cells) != len(headerCells) {
			continue
		}
		name := strings.TrimSpace(extensionsTableNameFooterRE.ReplaceAllString(cells[0], ""))
		if name == "" {
			continue
		}
		perVersion := map[string]bool{}
		for idx := 1; idx < len(cells); idx++ {
			if versions[idx] == "" {
				continue
			}
			perVersion[versions[idx]] = strings.Contains(cells[idx], "&check;")
		}
		table[name] = perVersion
	}
	return table
}

func readExtensionsTableCache() (map[string]map[string]bool, bool) {
	content, err := os.ReadFile(extensionsTableFile)
	if err != nil {
		return nil, false
	}
	var data map[string]map[string]bool
	if json.Unmarshal(content, &data) != nil {
		return nil, false
	}
	return data, true
}

func fetchAvailableExtensionsTable(ctx context.Context) map[string]map[string]bool {
	if info, err := os.Stat(extensionsTableFile); err == nil && time.Since(info.ModTime()) < extensionsTableTTL {
		if data, ok := readExtensionsTableCache(); ok {
			return data
		}
	}

	table := func() map[string]map[string]bool {
		client := &http.Client{Timeout: 5 * time.Second}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, extensionsTableURL, nil)
		if err != nil {
			return nil
		}
		resp, err := client.Do(req)
		if err != nil {
			return nil
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil
		}
		return parseExtensionsTable(string(body))
	}()

	if table == nil {
		if data, ok := readExtensionsTableCache(); ok {
			return data
		}
		return map[string]map[string]bool{}
	}

	_ = os.MkdirAll(filepath.Dir(extensionsTableFile), 0o755)
	if b, err := json.Marshal(table); err == nil {
		_ = os.WriteFile(extensionsTableFile, b, 0o644)
	}
	return table
}

func extensionsSupportedForVersion(ctx context.Context, version string) []string {
	table := fetchAvailableExtensionsTable(ctx)
	var names []string
	for name, versions := range table {
		if versions[version] {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

// ----------------------------
// Inspecting/toggling extensions inside the container
// ----------------------------

const confDPath = "/usr/local/etc/php/conf.d"

var extConfNameRE = regexp.MustCompile(`(?i)docker-php-ext-([a-z0-9_]+)`)

// ensurePHPServiceRunning starts the PHP service if it isn't already
// running, checking success via the command's exit code plus a live
// IsServiceRunning recheck rather than sniffing command output - real
// podman-compose stdout for a successful `up -d` doesn't reliably contain
// any fixed marker word (it just prints the container name), so a text
// sniff would reject genuinely successful starts.
func ensurePHPServiceRunning(ctx context.Context, userContext, service string) (ok bool, errMessage string) {
	if docker.IsServiceRunning(ctx, userContext, service) {
		return true, ""
	}
	result := docker.StartOrStopContainer(ctx, userContext, service, "activate", "run")
	if !result.Success || !docker.IsServiceRunning(ctx, userContext, service) {
		return false, result.Message
	}
	return true, ""
}

// getActiveAndDisabledExtensions inspects the running PHP container to
// determine which extensions are active (loaded by `php -m`) versus
// disabled (their conf.d file renamed to *.disabled).
func getActiveAndDisabledExtensions(ctx context.Context, userContext, service string) (active, disabled map[string]bool) {
	active = map[string]bool{}
	disabled = map[string]bool{}

	modArgv := podmanmanager.PodmanArgv(userContext, "exec", service, "php", "-m")
	if out, err := runShort(ctx, userContext, modArgv); err == nil {
		for _, line := range strings.Split(out, "\n") {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "[") {
				active[strings.ToLower(line)] = true
			}
		}
	}

	lsArgv := podmanmanager.PodmanArgv(userContext, "exec", service, "sh", "-c", "ls -1 "+confDPath+"/ 2>/dev/null")
	out, _ := runShort(ctx, userContext, lsArgv)
	for _, fname := range strings.Split(out, "\n") {
		if !strings.HasSuffix(fname, ".disabled") {
			continue
		}
		name := strings.ToLower(fname)
		if m := extConfNameRE.FindStringSubmatch(fname); m != nil {
			name = strings.ToLower(m[1])
		}
		disabled[name] = true
	}
	return active, disabled
}

func runShort(ctx context.Context, userContext string, argv []string) (string, error) {
	cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	out, err := podmanmanager.Command(cctx, userContext, argv).Output()
	return string(out), err
}

func findExtensionConfFile(ctx context.Context, userContext, service, extension string, wantDisabled bool) string {
	suffix := ".ini"
	if wantDisabled {
		suffix = ".ini.disabled"
	}
	argv := podmanmanager.PodmanArgv(userContext, "exec", service, "sh", "-c", "ls -1 "+confDPath+"/ 2>/dev/null")
	out, _ := runShort(ctx, userContext, argv)
	needle := strings.ToLower(extension)
	for _, fname := range strings.Split(out, "\n") {
		if strings.HasSuffix(fname, suffix) && strings.Contains(strings.ToLower(fname), needle) {
			return fname
		}
	}
	return ""
}

// toggleExtension enables or disables an extension by renaming its conf.d
// file to (or from) a .disabled suffix inside the PHP container.
func toggleExtension(ctx context.Context, userContext, service, extension string, enable bool) (ok bool, errMessage string) {
	var src, dst string
	if enable {
		src = findExtensionConfFile(ctx, userContext, service, extension, true)
		if src == "" {
			return false, "Extension is not installed, or is already enabled."
		}
		dst = strings.TrimSuffix(src, ".disabled")
	} else {
		src = findExtensionConfFile(ctx, userContext, service, extension, false)
		if src == "" {
			return false, "Extension is not installed, or is already disabled."
		}
		dst = src + ".disabled"
	}

	cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	argv := podmanmanager.PodmanArgv(userContext, "exec", service, "mv", confDPath+"/"+src, confDPath+"/"+dst)
	cmd := podmanmanager.Command(cctx, userContext, argv)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return false, strings.TrimSpace(stderr.String())
	}
	return true, ""
}

// EnsureExtensionInstalled makes sure a PHP extension is present and
// enabled in the given user's PHP-FPM service before some other module's
// installer runs (e.g. OJS requires "ftp" - see internal/modules/ojs's
// install.go). Reuses this file's own install machinery (`phpaddmod`, the
// same command the Extensions page's "Install" button runs) rather than
// duplicating it, so results and failure modes are identical to the
// browser-driven Extensions flow.
//
// Three cases, cheapest first:
//  1. Already active (`php -m` lists it) - no-op.
//  2. Installed but disabled (a "<ext>.ini.disabled" file exists) - just
//     re-enable it (rename off ".disabled"), then restart PHP-FPM.
//  3. Not installed at all - run `phpaddmod <extension>` inside the
//     container (this can take a while - a real package install/compile,
//     not a config toggle), then restart PHP-FPM.
//
// In all restart cases this blocks until the service reports running again
// (mirrors ensureContainerRunning's identical poll loop used throughout
// this codebase's CMS installers), so the caller can safely proceed
// straight to using the extension afterward.
func EnsureExtensionInstalled(ctx context.Context, userContext, service, extension string) error {
	active, _ := getActiveAndDisabledExtensions(ctx, userContext, service)
	if active[strings.ToLower(extension)] {
		return nil
	}

	if disabledConf := findExtensionConfFile(ctx, userContext, service, extension, true); disabledConf != "" {
		if ok, errMsg := toggleExtension(ctx, userContext, service, extension, true); !ok {
			return fmt.Errorf("enabling PHP extension %q: %s", extension, errMsg)
		}
	} else {
		installCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
		defer cancel()
		argv := podmanmanager.PodmanArgv(userContext, "exec", service, "phpaddmod", extension)
		cmd := podmanmanager.Command(installCtx, userContext, argv)
		var stderr strings.Builder
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			msg := strings.TrimSpace(stderr.String())
			if installCtx.Err() == context.DeadlineExceeded {
				msg = "installation timed out"
			} else if msg == "" {
				msg = err.Error()
			}
			return fmt.Errorf("installing PHP extension %q: %s", extension, msg)
		}
	}

	docker.ComposeContainer(ctx, userContext, service, "restart")
	const attempts = 15
	for i := 0; i < attempts; i++ {
		time.Sleep(2 * time.Second)
		if docker.IsServiceRunning(ctx, userContext, service) {
			return nil
		}
	}
	return fmt.Errorf("PHP-FPM service %q did not come back up after installing %q", service, extension)
}

// ----------------------------
// Per-version install history: which extensions this user has installed at
// least once.
// ----------------------------

const extensionsHistoryFilename = "installed.json"

func extensionsHistoryDir(userContext, version string) string {
	return "/home/" + userContext + "/php.extensions/" + version
}

func extensionsHistoryFile(userContext, version string) string {
	return filepath.Join(extensionsHistoryDir(userContext, version), extensionsHistoryFilename)
}

func loadExtensionsHistory(userContext, version string) []string {
	content, err := os.ReadFile(extensionsHistoryFile(userContext, version))
	if err != nil {
		return nil
	}
	var data struct {
		Extensions []string `json:"extensions"`
	}
	if json.Unmarshal(content, &data) != nil {
		return nil
	}
	seen := map[string]bool{}
	var names []string
	for _, n := range data.Extensions {
		l := strings.ToLower(n)
		if !seen[l] {
			seen[l] = true
			names = append(names, l)
		}
	}
	sort.Strings(names)
	return names
}

func saveExtensionsHistory(userContext, version string, names []string) []string {
	dir := extensionsHistoryDir(userContext, version)
	_ = os.MkdirAll(dir, 0o755)

	merged := map[string]bool{}
	for _, n := range loadExtensionsHistory(userContext, version) {
		merged[n] = true
	}
	for _, n := range names {
		n = strings.ToLower(strings.TrimSpace(n))
		if n != "" {
			merged[n] = true
		}
	}
	sortedNames := make([]string, 0, len(merged))
	for n := range merged {
		sortedNames = append(sortedNames, n)
	}
	sort.Strings(sortedNames)

	path := extensionsHistoryFile(userContext, version)
	tmp := path + ".tmp"
	if b, err := json.Marshal(map[string][]string{"extensions": sortedNames}); err == nil {
		if os.WriteFile(tmp, b, 0o644) == nil {
			_ = os.Rename(tmp, path)
		}
	}
	return sortedNames
}

func isInstallRunningInContainer(ctx context.Context, userContext, service string) bool {
	argv := podmanmanager.PodmanArgv(userContext, "exec", service, "sh", "-c",
		"ps aux | grep -E 'phpaddmod|install-php-extensions' | grep -v grep")
	out, _ := runShort(ctx, userContext, argv)
	return strings.TrimSpace(out) != ""
}

// ----------------------------
// Background installs (phpaddmod) + persistent state, following the same
// pattern as the file-manager wget download flow
// (internal/modules/filemanager/wget.go).
// ----------------------------

const installStateDir = "/tmp/php_extension_installs"

type installState struct {
	Status          string   `json:"status"`
	Message         string   `json:"message"`
	Extensions      []string `json:"extensions"`
	Version         string   `json:"version"`
	Context         string   `json:"context"`
	CurrentUsername string   `json:"current_username"`
	Service         string   `json:"service"`
	Logged          bool     `json:"logged"`
}

func installStateFilePath(installID string) (string, error) {
	if err := os.MkdirAll(installStateDir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(installStateDir, installID+".json"), nil
}

func saveInstallState(installID string, s installState) error {
	path, err := installStateFilePath(installID)
	if err != nil {
		return err
	}
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func loadInstallState(installID string) (installState, bool) {
	if installID == "" {
		return installState{}, false
	}
	path, err := installStateFilePath(installID)
	if err != nil {
		return installState{}, false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return installState{}, false
	}
	var s installState
	if json.Unmarshal(data, &s) != nil {
		return installState{}, false
	}
	return s, true
}

// runExtensionInstall runs the extension install in its own goroutine so it
// outlives the triggering request.
func runExtensionInstall(installID string) {
	info, ok := loadInstallState(installID)
	if !ok {
		return
	}

	info.Status = "installing"
	info.Message = "Installing " + strings.Join(info.Extensions, ", ") + "..."
	_ = saveInstallState(installID, info)

	argv := append(podmanmanager.PodmanArgv(info.Context, "exec", info.Service, "phpaddmod"), info.Extensions...)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	cmd := podmanmanager.Command(ctx, info.Context, argv)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	runErr := cmd.Run()

	cur, _ := loadInstallState(installID)
	if runErr == nil {
		cur.Status = "restarting"
		cur.Message = "Extensions installed, restarting PHP-FPM to apply changes..."
		_ = saveInstallState(installID, cur)

		docker.ComposeContainer(context.Background(), info.Context, info.Service, "restart")
		saveExtensionsHistory(info.Context, info.Version, info.Extensions)

		cur.Status = "done"
		cur.Message = "Extensions installed: " + strings.Join(info.Extensions, ", ")
		_ = saveInstallState(installID, cur)
		return
	}

	msg := strings.TrimSpace(stderr.String())
	if ctx.Err() == context.DeadlineExceeded {
		msg = "Installation timed out."
	} else if msg == "" {
		msg = "Installation failed."
	}
	cur.Status = "error"
	cur.Message = msg
	_ = saveInstallState(installID, cur)
}

// ----------------------------
// Routes
// ----------------------------

var extensionNameRE = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// ExtensionRow is one row of extensions.html's table.
type ExtensionRow struct {
	Name  string `json:"name"`
	State string `json:"state"` // "active" | "disabled" | "not_installed"
}

func litespeedRedirectIfNeeded(a *appctx.App, w http.ResponseWriter, r *http.Request, userContext string) bool {
	webServer := webserver.GetEnvFileValue(userContext, "WEB_SERVER")
	if strings.Contains(strings.ToLower(webServer), "litespeed") {
		flashAndRedirect(a, w, r, "warning", "PHP extension management is only available for PHP-FPM, not for Litespeed.", "/php/default")
		return true
	}
	return false
}

// handlePHPExtensionsSelect renders the PHP-version picker for the
// extensions page.
func handlePHPExtensionsSelect(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if litespeedRedirectIfNeeded(a, w, r, userContext) {
		return
	}

	installedVersions := FetchPHPVersions(ctx, a, userContext)
	renderPHPExtensionsSelectPage(a, w, r, installedVersions)
}

// handlePHPExtensions renders the extensions table for one PHP version and
// handles the enable/disable POST from it.
func handlePHPExtensions(a *appctx.App, w http.ResponseWriter, r *http.Request, versionSeg string) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if litespeedRedirectIfNeeded(a, w, r, userContext) {
		return
	}

	version := phpVersionFromSegment(versionSeg)
	service := "php-fpm-" + version
	if ok, errMsg := ensurePHPServiceRunning(ctx, userContext, service); !ok {
		flashAndRedirect(a, w, r, "error", "Failed to start PHP "+version+": "+errMsg, "/php/extensions")
		return
	}

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		extension := strings.TrimSpace(r.Form.Get("extension"))
		enable := r.Form.Get("enable") == "1"

		if !extensionNameRE.MatchString(extension) {
			flashAndRedirect(a, w, r, "error", "Invalid extension name.", "/php/php"+version+"/extensions")
			return
		}

		ok, errMsg := toggleExtension(ctx, userContext, service, extension, enable)
		if ok {
			docker.ComposeContainer(ctx, userContext, service, "restart")
			actionWord := "disabled"
			if enable {
				actionWord = "enabled"
			}
			ipAddress := reqip.ClientIP(r)
			_ = logger.RecordUserAction(a.Config, currentUsername, fmt.Sprintf("%s PHP extension %s for PHP %s", actionWord, extension, version), ipAddress)
			flashSess(a, w, r, "success", fmt.Sprintf("Extension %s %s, PHP %s restarted to apply changes.", extension, actionWord, version))
		} else {
			flashSess(a, w, r, "error", fmt.Sprintf("Could not change extension %s: %s", extension, errMsg))
		}
		http.Redirect(w, r, "/php/php"+version+"/extensions", http.StatusFound)
		return
	}

	active, disabled := getActiveAndDisabledExtensions(ctx, userContext, service)
	supportedNames := extensionsSupportedForVersion(ctx, version)

	extensions := make([]ExtensionRow, 0, len(supportedNames))
	for _, name := range supportedNames {
		lname := strings.ToLower(name)
		state := "not_installed"
		if active[lname] {
			state = "active"
		} else if disabled[lname] {
			state = "disabled"
		}
		extensions = append(extensions, ExtensionRow{Name: name, State: state})
	}

	if r.URL.Query().Get("output") == "json" {
		writeJSON(w, http.StatusOK, map[string]any{"extensions": extensions, "service": service})
		return
	}

	history := loadExtensionsHistory(userContext, version)
	currentNames := map[string]bool{}
	for _, e := range extensions {
		if e.State == "active" || e.State == "disabled" {
			currentNames[strings.ToLower(e.Name)] = true
		}
	}
	var recentlyRemoved []string
	for _, name := range history {
		if !currentNames[name] {
			recentlyRemoved = append(recentlyRemoved, name)
		}
	}
	sort.Strings(recentlyRemoved)

	renderPHPExtensionsPage(a, w, r, version, extensions, history, recentlyRemoved)
}

// handlePHPAvailableExtensions returns the full extensions catalog for one
// PHP version, flagging which are already installed.
func handlePHPAvailableExtensions(a *appctx.App, w http.ResponseWriter, r *http.Request, versionSeg string) {
	ctx := r.Context()
	_, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	webServer := webserver.GetEnvFileValue(userContext, "WEB_SERVER")
	if strings.Contains(strings.ToLower(webServer), "litespeed") {
		writeJSONError(w, http.StatusBadRequest, "Not available for Litespeed.")
		return
	}

	version := phpVersionFromSegment(versionSeg)
	service := "php-fpm-" + version
	active, disabled := getActiveAndDisabledExtensions(ctx, userContext, service)
	supportedNames := extensionsSupportedForVersion(ctx, version)

	type availableExt struct {
		Name      string `json:"name"`
		Installed bool   `json:"installed"`
	}
	extensions := make([]availableExt, 0, len(supportedNames))
	for _, name := range supportedNames {
		lname := strings.ToLower(name)
		extensions = append(extensions, availableExt{Name: name, Installed: active[lname] || disabled[lname]})
	}

	var cachedUntil any
	if info, statErr := os.Stat(extensionsTableFile); statErr == nil {
		cachedUntil = info.ModTime().Add(extensionsTableTTL).Unix()
	}

	writeJSON(w, http.StatusOK, map[string]any{"extensions": extensions, "service": service, "cached_until": cachedUntil})
}

// handlePHPExtensionsHistory gets or appends to the per-version install
// history of extensions this user has installed at least once.
func handlePHPExtensionsHistory(a *appctx.App, w http.ResponseWriter, r *http.Request, versionSeg string) {
	_, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	version := phpVersionFromSegment(versionSeg)

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		raw := r.Form["extensions[]"]
		if len(raw) == 0 {
			raw = r.Form["extensions"]
		}
		var names []string
		for _, e := range raw {
			e = strings.TrimSpace(e)
			if extensionNameRE.MatchString(e) {
				names = append(names, e)
			}
		}
		if len(names) == 0 {
			writeJSONError(w, http.StatusBadRequest, "No valid extensions given.")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"extensions": saveExtensionsHistory(userContext, version, names)})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"extensions": loadExtensionsHistory(userContext, version)})
}

// handlePHPInstallExtensions queues an asynchronous install of one or more
// PHP extensions for a version, returning an install ID for polling.
func handlePHPInstallExtensions(a *appctx.App, w http.ResponseWriter, r *http.Request, versionSeg string) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	version := phpVersionFromSegment(versionSeg)

	webServer := webserver.GetEnvFileValue(userContext, "WEB_SERVER")
	if strings.Contains(strings.ToLower(webServer), "litespeed") {
		writeJSONError(w, http.StatusBadRequest, "Not available for Litespeed.")
		return
	}

	_ = r.ParseMultipartForm(1 << 20)
	raw := r.Form["extensions[]"]
	if len(raw) == 0 {
		raw = r.Form["extensions"]
	}
	var extensions []string
	for _, e := range raw {
		e = strings.TrimSpace(e)
		if extensionNameRE.MatchString(e) {
			extensions = append(extensions, e)
		}
	}
	if len(extensions) == 0 {
		writeJSONError(w, http.StatusBadRequest, "No valid extensions selected.")
		return
	}

	service := "php-fpm-" + version
	if ok, errMsg := ensurePHPServiceRunning(ctx, userContext, service); !ok {
		writeJSONError(w, http.StatusInternalServerError, "Failed to start PHP "+version+": "+errMsg)
		return
	}

	installID := uuid.NewString()
	info := installState{
		Status: "queued", Message: "Install queued...", Extensions: extensions,
		Version: version, Context: userContext, CurrentUsername: currentUsername, Service: service,
	}
	if err := saveInstallState(installID, info); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Internal server error.")
		return
	}

	go runExtensionInstall(installID)

	writeJSON(w, http.StatusOK, map[string]any{"install_id": installID})
}

// handlePHPInstallExtensionsStatus reports the progress of a queued
// extension install, logging the user action once it completes.
func handlePHPInstallExtensionsStatus(a *appctx.App, w http.ResponseWriter, r *http.Request, versionSeg string) {
	ctx := r.Context()
	_, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	version := phpVersionFromSegment(versionSeg)
	service := "php-fpm-" + version
	installID := r.URL.Query().Get("install_id")

	containerBusy := isInstallRunningInContainer(ctx, userContext, service)
	info, ok := loadInstallState(installID)

	if ok {
		if info.Status == "done" && !info.Logged {
			ipAddress := reqip.ClientIP(r)
			_ = logger.RecordUserAction(a.Config, info.CurrentUsername, fmt.Sprintf("installed PHP extensions %s for PHP %s", strings.Join(info.Extensions, ", "), info.Version), ipAddress)
			info.Logged = true
			_ = saveInstallState(installID, info)
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"status": info.Status, "message": info.Message, "extensions": info.Extensions, "container_busy": containerBusy,
		})
		return
	}

	status := "idle"
	if containerBusy {
		status = "busy"
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": status, "container_busy": containerBusy})
}
