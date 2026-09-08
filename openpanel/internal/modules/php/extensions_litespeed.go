package php

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
)

// ----------------------------
// LiteSpeed/OpenLiteSpeed PHP extensions.
// Unlike php-fpm's one-container-per-version setup, a single "openlitespeed"/"litespeed" container serves every PHP version, and extensions are Debian packages from litespeedtech's apt repo ("lsphp<major><minor>-<extension>") whose .ini lands directly in the interpreter's scan directory - so disabling mirrors the php-fpm ".ini.disabled" rename convention to keep the same toggle UI/state model.
// ----------------------------

// litespeedVersionDir is e.g. "8.3" -> "lsphp83".
func litespeedVersionDir(version string) string {
	return "lsphp" + strings.ReplaceAll(version, ".", "")
}

// litespeedPackagePrefix is e.g. "8.3" -> "lsphp83-".
func litespeedPackagePrefix(version string) string {
	return litespeedVersionDir(version) + "-"
}

// litespeedConfDir is the ini scan directory for one PHP version inside the
// LiteSpeed container.
func litespeedConfDir(version string) string {
	return "/usr/local/lsws/" + litespeedVersionDir(version) + "/etc/php/" + version + "/mods-available"
}

// litespeedExcludedPackages are lsphp<ver>-* packages that aren't standalone extensions (build/debug/meta, or needing manual license acceptance) and shouldn't be offered for install
var litespeedExcludedPackages = map[string]bool{
	"common":         true,
	"dev":            true,
	"dbg":            true,
	"pear":           true,
	"modules-source": true,
	"ioncube":        true,
}

// litespeedExtensionAliases maps a package/extension name to the ini basenames it may install, for packages whose ini files don't share the package's name (e.g. "mysql" ships "mysqli.ini" and "pdo_mysql.ini")
var litespeedExtensionAliases = map[string][]string{
	"mysql": {"mysql", "mysqli", "pdo_mysql"},
}

// litespeedExtensionMatchNames returns the ini basenames that belong to extension/package "name" - its own name, a "pdo_"-prefixed variant, and any explicit alias
func litespeedExtensionMatchNames(name string) []string {
	lname := strings.ToLower(name)
	if aliases, ok := litespeedExtensionAliases[lname]; ok {
		return aliases
	}
	return []string{lname, "pdo_" + lname}
}

var litespeedIniPriorityPrefixRE = regexp.MustCompile(`^\d+-`)

// litespeedIniBaseName strips a package's numeric load-priority prefix (e.g. "50-redis.ini" -> "redis") and the .ini/.ini.disabled suffix from a mods-available filename
func litespeedIniBaseName(fname string) string {
	base := strings.TrimSuffix(strings.TrimSuffix(fname, ".disabled"), ".ini")
	base = litespeedIniPriorityPrefixRE.ReplaceAllString(base, "")
	return strings.ToLower(base)
}

// litespeedExtensionsSupportedForVersion lists the extension packages litespeedtech's apt repo offers for this PHP version, queried live from the container's local apt cache so unlike the php-fpm catalog this isn't file-cached
func litespeedExtensionsSupportedForVersion(ctx context.Context, userContext, container, version string) []string {
	prefix := litespeedPackagePrefix(version)
	argv := podmanmanager.PodmanArgv(userContext, "exec", container, "apt-cache", "pkgnames", prefix)
	out, err := runShort(ctx, userContext, argv)
	if err != nil {
		return nil
	}

	seen := map[string]bool{}
	var names []string
	for _, pkg := range strings.Split(out, "\n") {
		pkg = strings.TrimSpace(pkg)
		if pkg == "" || !strings.HasPrefix(pkg, prefix) {
			continue
		}
		name := strings.ToLower(strings.TrimPrefix(pkg, prefix))
		if name == "" || litespeedExcludedPackages[name] || seen[name] {
			continue
		}
		seen[name] = true
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// getLitespeedInstalledPackages returns the lowercased, prefix-stripped set of lsphp<ver>-* packages currently installed via dpkg
func getLitespeedInstalledPackages(ctx context.Context, userContext, container, version string) map[string]bool {
	prefix := litespeedPackagePrefix(version)
	argv := podmanmanager.PodmanArgv(userContext, "exec", container, "sh", "-c",
		`dpkg-query -W -f='${Package} ${Status}\n' '`+prefix+`*' 2>/dev/null`)
	out, _ := runShort(ctx, userContext, argv)

	installed := map[string]bool{}
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		pkg := fields[0]
		status := strings.Join(fields[1:], " ")
		if !strings.Contains(status, "install ok installed") {
			continue
		}
		installed[strings.ToLower(strings.TrimPrefix(pkg, prefix))] = true
	}
	return installed
}

// getLitespeedExtensionsState lists the mods-available directory once and returns, by ini basename, which are enabled (a bare ".ini") and which are disabled (renamed to ".ini.disabled")
func getLitespeedExtensionsState(ctx context.Context, userContext, container, version string) (enabled, disabled map[string]bool) {
	enabled = map[string]bool{}
	disabled = map[string]bool{}

	confDir := litespeedConfDir(version)
	argv := podmanmanager.PodmanArgv(userContext, "exec", container, "sh", "-c", "ls -1 "+confDir+"/ 2>/dev/null")
	out, _ := runShort(ctx, userContext, argv)
	for _, fname := range strings.Split(out, "\n") {
		fname = strings.TrimSpace(fname)
		switch {
		case strings.HasSuffix(fname, ".ini.disabled"):
			disabled[litespeedIniBaseName(fname)] = true
		case strings.HasSuffix(fname, ".ini"):
			enabled[litespeedIniBaseName(fname)] = true
		}
	}
	return enabled, disabled
}

// litespeedExtensionState resolves one extension's row state from the installed-package set and the enabled/disabled ini basename sets built above
func litespeedExtensionState(name string, installed, enabledBase, disabledBase map[string]bool) string {
	hasEnabled, hasDisabled := false, false
	for _, m := range litespeedExtensionMatchNames(name) {
		if enabledBase[m] {
			hasEnabled = true
		}
		if disabledBase[m] {
			hasDisabled = true
		}
	}
	switch {
	case hasEnabled:
		return "active"
	case hasDisabled:
		return "disabled"
	case installed[strings.ToLower(name)]:
		// package installed but no matching ini found - assume it's loaded rather than hiding it as "not installed"
		return "active"
	default:
		return "not_installed"
	}
}

// litespeedExtensionRows builds the extensions table for one PHP version.
func litespeedExtensionRows(ctx context.Context, userContext, container, version string) []ExtensionRow {
	names := litespeedExtensionsSupportedForVersion(ctx, userContext, container, version)
	installed := getLitespeedInstalledPackages(ctx, userContext, container, version)
	enabledBase, disabledBase := getLitespeedExtensionsState(ctx, userContext, container, version)

	rows := make([]ExtensionRow, 0, len(names))
	for _, name := range names {
		rows = append(rows, ExtensionRow{Name: name, State: litespeedExtensionState(name, installed, enabledBase, disabledBase)})
	}
	return rows
}

// toggleLitespeedExtension enables or disables an already-installed extension by renaming its owned ini file(s) to/from a ".disabled" suffix
func toggleLitespeedExtension(ctx context.Context, userContext, container, version, extension string, enable bool) (ok bool, errMessage string) {
	confDir := litespeedConfDir(version)
	argv := podmanmanager.PodmanArgv(userContext, "exec", container, "sh", "-c", "ls -1 "+confDir+"/ 2>/dev/null")
	out, _ := runShort(ctx, userContext, argv)

	matchSet := map[string]bool{}
	for _, m := range litespeedExtensionMatchNames(extension) {
		matchSet[m] = true
	}

	var toRename []string
	for _, fname := range strings.Split(out, "\n") {
		fname = strings.TrimSpace(fname)
		if fname == "" || !matchSet[litespeedIniBaseName(fname)] {
			continue
		}
		isDisabled := strings.HasSuffix(fname, ".ini.disabled")
		if enable && isDisabled {
			toRename = append(toRename, fname)
		} else if !enable && !isDisabled && strings.HasSuffix(fname, ".ini") {
			toRename = append(toRename, fname)
		}
	}

	if len(toRename) == 0 {
		if enable {
			return false, "Extension is not installed, or is already enabled."
		}
		return false, "Extension is not installed, or is already disabled."
	}

	cmds := make([]string, 0, len(toRename))
	for _, fname := range toRename {
		dst := fname + ".disabled"
		if enable {
			dst = strings.TrimSuffix(fname, ".disabled")
		}
		cmds = append(cmds, "mv '"+confDir+"/"+fname+"' '"+confDir+"/"+dst+"'")
	}

	cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	argv2 := podmanmanager.PodmanArgv(userContext, "exec", container, "sh", "-c", strings.Join(cmds, " && "))
	cmd := podmanmanager.Command(cctx, userContext, argv2)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return false, strings.TrimSpace(stderr.String())
	}
	return true, ""
}

// litespeedRunningPHPVersion asks the container's default `php` binary for its major.minor version, needed to build mods-available paths since a single LiteSpeed container can only run one PHP version at a time
func litespeedRunningPHPVersion(ctx context.Context, userContext, container string) string {
	argv := podmanmanager.PodmanArgv(userContext, "exec", container, "php", "-r", "echo PHP_MAJOR_VERSION.'.'.PHP_MINOR_VERSION;")
	out, err := runShort(ctx, userContext, argv)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}

// installLitespeedExtensionsArgv builds the argv for a blocking apt-get install of one or more extension packages, refreshing the apt index first since the base image's index may be stale
func installLitespeedExtensionsArgv(userContext, container, version string, extensions []string) []string {
	prefix := litespeedPackagePrefix(version)
	pkgs := make([]string, len(extensions))
	for i, e := range extensions {
		pkgs[i] = prefix + strings.ToLower(e)
	}
	shCmd := "apt-get update -qq >/dev/null 2>&1; DEBIAN_FRONTEND=noninteractive apt-get install -y " + strings.Join(pkgs, " ")
	return podmanmanager.PodmanArgv(userContext, "exec", container, "sh", "-c", shCmd)
}

// ensureLitespeedExtensionInstalled is EnsureExtensionInstalled's LiteSpeed counterpart, see that function's doc comment for the three cases
func ensureLitespeedExtensionInstalled(ctx context.Context, userContext, container, extension string) error {
	version := litespeedRunningPHPVersion(ctx, userContext, container)
	if version == "" {
		return fmt.Errorf("could not determine the running PHP version for %q", container)
	}

	installed := getLitespeedInstalledPackages(ctx, userContext, container, version)
	enabledBase, disabledBase := getLitespeedExtensionsState(ctx, userContext, container, version)
	switch litespeedExtensionState(extension, installed, enabledBase, disabledBase) {
	case "active":
		return nil
	case "disabled":
		if ok, errMsg := toggleLitespeedExtension(ctx, userContext, container, version, extension, true); !ok {
			return fmt.Errorf("enabling PHP extension %q: %s", extension, errMsg)
		}
	default:
		installCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
		defer cancel()
		argv := installLitespeedExtensionsArgv(userContext, container, version, []string{extension})
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

	docker.ComposeContainer(ctx, userContext, container, "restart")
	const attempts = 15
	for i := 0; i < attempts; i++ {
		time.Sleep(2 * time.Second)
		if docker.IsServiceRunning(ctx, userContext, container) {
			return nil
		}
	}
	return fmt.Errorf("web server %q did not come back up after installing %q", container, extension)
}
