package websites

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"os"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"syscall"

	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
)

// file size caps for the manager pages' read-only views
const (
	maxManifestBytes = 1 << 20
	maxLockBytes     = 8 << 20
	maxSourceBytes   = 256 << 10
)

// readSiteFile reads a regular file under /var/www/html/ from the account's html_data volume, false when it's missing, not a plain file or over max
func readSiteFile(userContext, containerPath string, max int64) ([]byte, bool) {
	clean := path.Clean(containerPath)
	if !strings.HasPrefix(clean, "/var/www/html/") {
		return nil, false
	}
	hostPath, ok := cmsMappedDir(userContext, clean)
	if !ok {
		return nil, false
	}
	// O_NONBLOCK so a fifo planted in the docroot can't stall the page
	f, err := tenantfs.OpenFile(hostPath, os.O_RDONLY|syscall.O_NONBLOCK, 0)
	if err != nil {
		return nil, false
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil || !fi.Mode().IsRegular() || fi.Size() > max {
		return nil, false
	}
	data, err := io.ReadAll(io.LimitReader(f, max))
	if err != nil {
		return nil, false
	}
	return data, true
}

// composeServiceFor returns a site container's service block from docker-compose.yml, keyed by the lowercased container name before its first underscore
func composeServiceFor(userContext, containerName string) map[string]any {
	name := containerName
	if idx := strings.Index(name, "_"); idx != -1 {
		name = name[:idx]
	}
	composeData, err := docker.LoadCompose(userContext)
	if err != nil {
		return nil
	}
	services, _ := composeData["services"].(map[string]any)
	svc, _ := services[strings.ToLower(name)].(map[string]any)
	return svc
}

// envVarsOf is a service's `environment:` list, one KEY=VALUE per line
func envVarsOf(svc map[string]any) string {
	rawEnv, _ := svc["environment"].([]any)
	lines := make([]string, 0, len(rawEnv))
	for _, v := range rawEnv {
		if s, ok := v.(string); ok {
			lines = append(lines, s)
		}
	}
	return strings.Join(lines, "\n")
}

// composeArgv turns a compose command/entrypoint value (string or list) into argv-ish words
func composeArgv(v any) []string {
	switch t := v.(type) {
	case string:
		if s := strings.TrimSpace(t); s != "" {
			return []string{s}
		}
	case []any:
		out := make([]string, 0, len(t))
		for _, x := range t {
			if s, ok := x.(string); ok {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}

var shellWrapRE = regexp.MustCompile(`^(?:/bin/|/usr/bin/)?(?:sh|bash)\s+-c\s+(.+)$`)

// unwrapShell strips the `sh -c "..."` wrapper pm2.go writes around every runtime's start command
func unwrapShell(argv []string) string {
	if len(argv) >= 3 && (path.Base(argv[0]) == "sh" || path.Base(argv[0]) == "bash") && argv[1] == "-c" {
		return strings.TrimSpace(strings.Join(argv[2:], " "))
	}
	s := strings.TrimSpace(strings.Join(argv, " "))
	if m := shellWrapRE.FindStringSubmatch(s); m != nil {
		inner := strings.TrimSpace(m[1])
		if len(inner) >= 2 && (inner[0] == '"' || inner[0] == '\'') && inner[len(inner)-1] == inner[0] {
			inner = inner[1 : len(inner)-1]
		}
		return strings.TrimSpace(inner)
	}
	return s
}

// composeStartCommand is the full command a service starts with, entrypoint first
func composeStartCommand(svc map[string]any) string {
	argv := append(composeArgv(svc["entrypoint"]), composeArgv(svc["command"])...)
	if len(argv) == 0 {
		return ""
	}
	return unwrapShell(argv)
}

// ComposeMount is one named volume or bind mount of a service
type ComposeMount struct {
	Source, Target string
}

// composeMounts lists a service's volumes in short ("src:dst[:ro]") or long syntax
func composeMounts(svc map[string]any) []ComposeMount {
	raw, _ := svc["volumes"].([]any)
	var out []ComposeMount
	for _, v := range raw {
		switch t := v.(type) {
		case string:
			parts := strings.Split(t, ":")
			if len(parts) >= 2 {
				out = append(out, ComposeMount{Source: parts[0], Target: parts[1]})
			} else if len(parts) == 1 && parts[0] != "" {
				out = append(out, ComposeMount{Target: parts[0]})
			}
		case map[string]any:
			src, _ := t["source"].(string)
			dst, _ := t["target"].(string)
			if dst != "" {
				out = append(out, ComposeMount{Source: src, Target: dst})
			}
		}
	}
	return out
}

// composeNetworks lists a service's networks, list or map form
func composeNetworks(svc map[string]any) []string {
	var out []string
	switch t := svc["networks"].(type) {
	case []any:
		for _, v := range t {
			if s, ok := v.(string); ok {
				out = append(out, s)
			}
		}
	case map[string]any:
		for k := range t {
			out = append(out, k)
		}
		sort.Strings(out)
	}
	return out
}

// ManifestRow is one package or script in the Packages tab, Installed is filled from a lock file when there is one
type ManifestRow struct {
	Name, Value, Installed string
}

// ManifestGroup is a titled table in the Packages tab
type ManifestGroup struct {
	Title        string
	ValueLabel   string
	HasInstalled bool
	Rows         []ManifestRow
}

// Manifest is a runtime app's dependency file, its raw text and the parsed tables
type Manifest struct {
	File   string
	Path   string
	Text   string
	Found  bool
	Groups []ManifestGroup
}

// manifestFile is the dependency file each runtime reads
func manifestFile(appType string) string {
	switch appType {
	case "nodejs":
		return "package.json"
	case "ruby":
		return "Gemfile"
	case "java":
		return "pom.xml"
	default:
		return "requirements.txt"
	}
}

// loadManifest reads and parses the app's manifest (plus its lock file) from dir, an empty Manifest with Found=false when it's missing
func loadManifest(userContext, appType, dir string) Manifest {
	m := Manifest{File: manifestFile(appType)}
	m.Path = path.Join(dir, m.File)
	data, ok := readSiteFile(userContext, m.Path, maxManifestBytes)
	if !ok {
		return m
	}
	m.Found, m.Text = true, string(data)
	lock := func(name string) []byte {
		b, _ := readSiteFile(userContext, path.Join(dir, name), maxLockBytes)
		return b
	}
	switch appType {
	case "nodejs":
		m.Groups = parsePackageJSON(data, lock("package-lock.json"))
	case "ruby":
		m.Groups = parseGemfile(data, lock("Gemfile.lock"))
	case "java":
		m.Groups = parsePom(data)
	default:
		m.Groups = parseRequirements(data)
	}
	return m
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// parsePackageJSON lists scripts and dependencies, installed versions come from package-lock.json (v2/v3 "packages" or v1 "dependencies")
func parsePackageJSON(data, lockData []byte) []ManifestGroup {
	var pkg struct {
		Scripts         map[string]string `json:"scripts"`
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if json.Unmarshal(data, &pkg) != nil {
		return nil
	}
	installed := map[string]string{}
	var lock struct {
		Packages     map[string]struct{ Version string } `json:"packages"`
		Dependencies map[string]struct{ Version string } `json:"dependencies"`
	}
	if len(lockData) > 0 && json.Unmarshal(lockData, &lock) == nil {
		for k, v := range lock.Dependencies {
			installed[k] = v.Version
		}
		for k, v := range lock.Packages {
			if name, ok := strings.CutPrefix(k, "node_modules/"); ok && !strings.Contains(name, "/node_modules/") {
				installed[name] = v.Version
			}
		}
	}
	var groups []ManifestGroup
	if len(pkg.Scripts) > 0 {
		g := ManifestGroup{Title: "Scripts", ValueLabel: "Command"}
		for _, k := range sortedKeys(pkg.Scripts) {
			g.Rows = append(g.Rows, ManifestRow{Name: k, Value: pkg.Scripts[k]})
		}
		groups = append(groups, g)
	}
	for _, d := range []struct {
		title string
		deps  map[string]string
	}{{"Dependencies", pkg.Dependencies}, {"Dev dependencies", pkg.DevDependencies}} {
		if len(d.deps) == 0 {
			continue
		}
		g := ManifestGroup{Title: d.title, ValueLabel: "Constraint", HasInstalled: len(installed) > 0}
		for _, k := range sortedKeys(d.deps) {
			g.Rows = append(g.Rows, ManifestRow{Name: k, Value: d.deps[k], Installed: installed[k]})
		}
		groups = append(groups, g)
	}
	return groups
}

var reqNameRE = regexp.MustCompile(`^([A-Za-z0-9][A-Za-z0-9._-]*)(\[[^\]]*\])?\s*(.*)$`)

// parseRequirements lists requirements.txt packages with their version spec, skipping comments, options and URLs
func parseRequirements(data []byte) []ManifestGroup {
	g := ManifestGroup{Title: "Pinned in requirements.txt", ValueLabel: "Version"}
	for _, line := range strings.Split(string(data), "\n") {
		if i := strings.Index(line, " #"); i != -1 {
			line = line[:i]
		}
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "-") || strings.Contains(line, "://") {
			continue
		}
		if i := strings.Index(line, ";"); i != -1 {
			line = strings.TrimSpace(line[:i])
		}
		m := reqNameRE.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		spec := strings.TrimSpace(m[3])
		if after, ok := strings.CutPrefix(spec, "=="); ok && !strings.ContainsAny(after, ",<>=!~*") {
			spec = strings.TrimSpace(after)
		}
		if spec == "" {
			spec = "any"
		}
		g.Rows = append(g.Rows, ManifestRow{Name: m[1], Value: spec})
	}
	if len(g.Rows) == 0 {
		return nil
	}
	return []ManifestGroup{g}
}

var (
	gemLineRE = regexp.MustCompile(`^\s*gem\s+["']([^"']+)["']((?:\s*,\s*["'][^"']*["'])*)`)
	gemArgRE  = regexp.MustCompile(`["']([^"']*)["']`)
	gemLockRE = regexp.MustCompile(`^    ([A-Za-z0-9._-]+) \(([^)]+)\)\s*$`)
)

// parseGemfile lists the Gemfile's gems with their constraints and Gemfile.lock versions, then the other locked gems
func parseGemfile(data, lockData []byte) []ManifestGroup {
	locked := map[string]string{}
	var lockOrder []string
	for _, line := range strings.Split(string(lockData), "\n") {
		if m := gemLockRE.FindStringSubmatch(strings.TrimRight(line, "\r")); m != nil {
			if _, seen := locked[m[1]]; !seen {
				lockOrder = append(lockOrder, m[1])
			}
			locked[m[1]] = m[2]
		}
	}
	g := ManifestGroup{Title: "Gems in the Gemfile", ValueLabel: "Constraint", HasInstalled: len(locked) > 0}
	direct := map[string]bool{}
	for _, line := range strings.Split(string(data), "\n") {
		m := gemLineRE.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		var cons []string
		for _, a := range gemArgRE.FindAllStringSubmatch(m[2], -1) {
			cons = append(cons, a[1])
		}
		value := strings.Join(cons, ", ")
		if value == "" {
			value = "any"
		}
		direct[m[1]] = true
		g.Rows = append(g.Rows, ManifestRow{Name: m[1], Value: value, Installed: locked[m[1]]})
	}
	var groups []ManifestGroup
	if len(g.Rows) > 0 {
		groups = append(groups, g)
	}
	rest := ManifestGroup{Title: "Locked in Gemfile.lock", ValueLabel: "Version"}
	for _, name := range lockOrder {
		if !direct[name] {
			rest.Rows = append(rest.Rows, ManifestRow{Name: name, Value: locked[name]})
		}
	}
	if len(rest.Rows) > 0 {
		groups = append(groups, rest)
	}
	return groups
}

// parsePom lists a pom.xml's coordinates, Java release and dependencies
func parsePom(data []byte) []ManifestGroup {
	var pom struct {
		GroupID    string `xml:"groupId"`
		ArtifactID string `xml:"artifactId"`
		Version    string `xml:"version"`
		Packaging  string `xml:"packaging"`
		Parent     struct {
			GroupID string `xml:"groupId"`
			Version string `xml:"version"`
		} `xml:"parent"`
		Properties struct {
			Release string `xml:"maven.compiler.release"`
			Source  string `xml:"maven.compiler.source"`
			Java    string `xml:"java.version"`
		} `xml:"properties"`
		Dependencies []struct {
			GroupID    string `xml:"groupId"`
			ArtifactID string `xml:"artifactId"`
			Version    string `xml:"version"`
			Scope      string `xml:"scope"`
		} `xml:"dependencies>dependency"`
	}
	if xml.Unmarshal(data, &pom) != nil {
		return nil
	}
	if pom.GroupID == "" {
		pom.GroupID = pom.Parent.GroupID
	}
	if pom.Version == "" {
		pom.Version = pom.Parent.Version
	}
	project := ManifestGroup{Title: "Project", ValueLabel: "Value"}
	add := func(name, value string) {
		if value = strings.TrimSpace(value); value != "" {
			project.Rows = append(project.Rows, ManifestRow{Name: name, Value: value})
		}
	}
	add("groupId", pom.GroupID)
	add("artifactId", pom.ArtifactID)
	add("version", pom.Version)
	add("packaging", pom.Packaging)
	release := pom.Properties.Release
	if release == "" {
		release = pom.Properties.Source
	}
	if release == "" {
		release = pom.Properties.Java
	}
	add("Java release", release)
	var groups []ManifestGroup
	if len(project.Rows) > 0 {
		groups = append(groups, project)
	}
	deps := ManifestGroup{Title: "Dependencies", ValueLabel: "Version"}
	for _, d := range pom.Dependencies {
		value := strings.TrimSpace(d.Version)
		if value == "" {
			value = "managed"
		}
		if d.Scope != "" {
			value += " (" + strings.TrimSpace(d.Scope) + ")"
		}
		deps.Rows = append(deps.Rows, ManifestRow{Name: strings.TrimSpace(d.GroupID) + ":" + strings.TrimSpace(d.ArtifactID), Value: value})
	}
	if len(deps.Rows) > 0 {
		groups = append(groups, deps)
	}
	return groups
}

// ComposerRow is one composer.json requirement next to the version composer.lock installed
type ComposerRow struct {
	Name, Constraint, Installed string
	Dev                         bool
}

// loadComposerRequirements reads composer.json and composer.lock from a PHP app's docroot, nil when composer.json is missing
func loadComposerRequirements(userContext, docroot string) []ComposerRow {
	data, ok := readSiteFile(userContext, path.Join(docroot, "composer.json"), maxManifestBytes)
	if !ok {
		return nil
	}
	lockData, _ := readSiteFile(userContext, path.Join(docroot, "composer.lock"), maxLockBytes)
	return parseComposer(data, lockData)
}

// parseComposer pairs require/require-dev constraints with composer.lock's installed versions
func parseComposer(data, lockData []byte) []ComposerRow {
	var c struct {
		Require    map[string]string `json:"require"`
		RequireDev map[string]string `json:"require-dev"`
	}
	if json.Unmarshal(data, &c) != nil {
		return nil
	}
	installed := map[string]string{}
	var lock struct {
		Packages    []struct{ Name, Version string } `json:"packages"`
		PackagesDev []struct{ Name, Version string } `json:"packages-dev"`
	}
	if len(lockData) > 0 && json.Unmarshal(lockData, &lock) == nil {
		for _, p := range append(lock.Packages, lock.PackagesDev...) {
			installed[strings.ToLower(p.Name)] = p.Version
		}
	}
	var rows []ComposerRow
	for _, set := range []struct {
		req map[string]string
		dev bool
	}{{c.Require, false}, {c.RequireDev, true}} {
		for _, name := range sortedKeys(set.req) {
			rows = append(rows, ComposerRow{Name: name, Constraint: set.req[name], Installed: installed[strings.ToLower(name)], Dev: set.dev})
		}
	}
	return rows
}

// loadSiteSource reads a site builder page's index.html for the read-only source view
func loadSiteSource(userContext, docroot string) (string, bool) {
	data, ok := readSiteFile(userContext, path.Join(docroot, "index.html"), maxSourceBytes)
	if !ok {
		return "", false
	}
	return string(data), true
}

// SiteEntry is one top-level file or folder in a manager's Files card
type SiteEntry struct {
	Name     string
	Dir      bool
	Size     string // "" for folders
	Modified string // 2006-01-02
}

// listSiteDir lists a docroot's top level, folders first, up to max entries plus the number of dotfiles left out
func listSiteDir(userContext, dir string, max int) ([]SiteEntry, int) {
	clean := path.Clean(dir)
	if !strings.HasPrefix(clean+"/", "/var/www/html/") {
		return nil, 0
	}
	hostPath, ok := cmsMappedDir(userContext, clean)
	if !ok {
		return nil, 0
	}
	entries, err := tenantfs.ReadDir(hostPath)
	if err != nil {
		return nil, 0
	}
	var dirs, files []SiteEntry
	hidden := 0
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".") {
			hidden++
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		se := SiteEntry{Name: e.Name(), Dir: e.IsDir(), Modified: info.ModTime().UTC().Format("2006-01-02")}
		if se.Dir {
			dirs = append(dirs, se)
			continue
		}
		se.Size = humanSize(info.Size())
		files = append(files, se)
	}
	out := append(dirs, files...)
	if len(out) > max {
		out = out[:max]
	}
	return out, hidden
}

// humanSize formats a byte count like the file manager does
func humanSize(n int64) string {
	const unit = 1024
	if n < unit {
		return strconv.FormatInt(n, 10) + " B"
	}
	div, exp := int64(unit), 0
	for m := n / unit; m >= unit; m /= unit {
		div *= unit
		exp++
	}
	v := float64(n) / float64(div)
	s := strconv.FormatFloat(v, 'f', 1, 64)
	if v >= 10 {
		s = strconv.FormatFloat(v, 'f', 0, 64)
	}
	return s + " " + string("KMGTPE"[exp]) + "B"
}

var composeVarRE = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)(?::-([^}]*))?\}`)

// expandCompose fills ${VAR} and ${VAR:-default} the way compose does, lookup returns "" for an unset var
func expandCompose(s string, lookup func(string) string) string {
	return composeVarRE.ReplaceAllStringFunc(s, func(m string) string {
		sm := composeVarRE.FindStringSubmatch(m)
		if v := lookup(sm[1]); v != "" {
			return v
		}
		return sm[2]
	})
}

// svcImage is a compose service's image with its .env tag filled in and without the registry host
func svcImage(userContext string, svc map[string]any) string {
	img, _ := svc["image"].(string)
	img = expandCompose(img, func(k string) string {
		v, _ := docker.GetEnvValue(userContext, k)
		return strings.Trim(v, `"'`)
	})
	if i := strings.Index(img, "/"); i > 0 && strings.ContainsAny(img[:i], ".:") {
		img = img[i+1:]
	}
	return strings.TrimPrefix(img, "library/")
}

// ComposerPackage is one package composer.lock installed
type ComposerPackage struct {
	Name, Version string
	Dev           bool
}

// loadComposerPackages lists what composer.lock installed, production first, nil without a lock file
func loadComposerPackages(userContext, docroot string) []ComposerPackage {
	data, ok := readSiteFile(userContext, path.Join(docroot, "composer.lock"), maxLockBytes)
	if !ok {
		return nil
	}
	return parseComposerLock(data)
}

func parseComposerLock(data []byte) []ComposerPackage {
	var lock struct {
		Packages    []struct{ Name, Version string } `json:"packages"`
		PackagesDev []struct{ Name, Version string } `json:"packages-dev"`
	}
	if json.Unmarshal(data, &lock) != nil {
		return nil
	}
	var out []ComposerPackage
	for _, p := range lock.Packages {
		out = append(out, ComposerPackage{Name: p.Name, Version: p.Version})
	}
	for _, p := range lock.PackagesDev {
		out = append(out, ComposerPackage{Name: p.Name, Version: p.Version, Dev: true})
	}
	return out
}

var (
	blankStripRE = regexp.MustCompile(`(?is)<(style|script|head)\b.*?</(style|script|head)>|<!--.*?-->`)
	blankTagRE   = regexp.MustCompile(`(?s)<[^>]*>`)
	blankMediaRE = regexp.MustCompile(`(?i)<(img|video|iframe|svg|canvas|picture|object|embed)\b`)
)

// blankPage is true when a builder page has nothing visible, no text and no media
func blankPage(src string) bool {
	body := blankStripRE.ReplaceAllString(src, "")
	if blankMediaRE.MatchString(body) {
		return false
	}
	return strings.TrimSpace(blankTagRE.ReplaceAllString(body, "")) == ""
}

// svcString is a plain string key of a compose service
func svcString(svc map[string]any, key string) string {
	v, _ := svc[key].(string)
	return v
}
