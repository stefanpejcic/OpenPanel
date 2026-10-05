package dokuwiki

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/appkit"
	"gist.github.com/stefanpejcic/openpanel/internal/core/installmanifest"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/cmsapp"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/waf"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/websites"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// dokuwikiStableTarball is DokuWiki's version-agnostic "always current stable" download, wrapped in a single top-level dokuwiki-<version>/ dir inside the tarball
const dokuwikiStableTarball = "https://download.dokuwiki.org/src/dokuwiki/dokuwiki-stable.tgz"

// phpVersionBelow reports whether version (e.g. "7.2") is older than wantMajor.wantMinor, opposite of sofawiki's ceiling check - an unparseable version fails safe as "too old"
func phpVersionBelow(version string, wantMajor, wantMinor int) bool {
	major, minor := 0, 0
	if _, err := fmt.Sscanf(version, "%d.%d", &major, &minor); err != nil {
		return true
	}
	if major != wantMajor {
		return major < wantMajor
	}
	return minor < wantMinor
}

// hashDokuwikiPassword shells out to php's password_hash() in the site's own php-fpm container instead of reimplementing bcrypt in Go, so the hash format always matches what install.php would produce - password is passed as a separate argv element, not interpolated into the shell string
func hashDokuwikiPassword(ctx context.Context, userContext, phpContainer, password string) (string, error) {
	argv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "php", "-r",
		"echo password_hash($argv[1], PASSWORD_BCRYPT);", "--", password)
	out, err := podmanmanager.Command(ctx, userContext, argv).Output()
	if err != nil {
		return "", err
	}
	hash := strings.TrimSpace(string(out))
	if hash == "" {
		return "", fmt.Errorf("empty hash returned")
	}
	return hash, nil
}

// handleInstallStream drives a DokuWiki install end to end over NDJSON: download+extract the stable tarball, write the conf files directly, fix ownership, record the site, and remove install.php
func handleInstallStream(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	in, ok := cmsapp.StartInstall(a, w, r)
	if !ok {
		return
	}
	defer appkit.RemoveLockFile(in.Username)
	ctx := in.Ctx
	currentUsername := in.Username
	userContext := in.UserContext
	emit := in.Emit
	ipAddress := in.IP
	domainID := in.DomainID
	dom := in.Domain
	subdirectory := in.Subdirectory

	adminUser := strings.ToLower(strings.TrimSpace(web.FormOr(r, "admin_user", "admin")))
	if adminUser == "" || strings.ContainsAny(adminUser, " :\t\n") {
		emit(map[string]any{"error": "Invalid admin username."})
		return
	}
	adminPassword := r.FormValue("admin_password")
	if adminPassword == "" {
		adminPassword = appkit.RandomString(16)
	}
	adminFullName := web.FormOr(r, "admin_full_name", "Administrator")
	siteTitle := web.FormOr(r, "site_title", dom.DomainURL)

	docroot := dom.Docroot.String
	selectedDomain := dom.DomainURL
	installPath := docroot
	if subdirectory != "" {
		installPath = strings.TrimSuffix(docroot, "/") + "/" + subdirectory
		selectedDomain = selectedDomain + "/" + subdirectory
	}

	webServer := webserver.GetEnvFileValue(userContext, "WEB_SERVER")
	isLitespeed := strings.Contains(strings.ToLower(webServer), "litespeed")
	phpVersion := dom.PHPVersion.String
	phpContainer := webServer
	if !isLitespeed {
		phpContainer = "php-fpm-" + phpVersion
	}

	// DokuWiki needs PHP 7.4+, opposite of sofawiki's PHP 8-fatals ceiling
	if !isLitespeed && phpVersionBelow(phpVersion, 7, 4) {
		emit(map[string]any{"error": "DokuWiki requires PHP 7.4 or newer, but this domain is set to PHP " + phpVersion + ". Change the domain's PHP version and try again."})
		return
	}

	emit(map[string]any{"status": "Starting PHP container: " + phpContainer})
	if !cmsapp.EnsureContainerRunning(ctx, userContext, phpContainer) {
		emit(map[string]any{"error": "PHP container failed to start. Please check it from Services."})
		return
	}

	htmlVolume := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/"
	docrootWithoutWWW := strings.TrimPrefix(strings.TrimPrefix(installPath, "/var/www/html/"), "/")
	hostOSPath := filepath.Join(htmlVolume, docrootWithoutWWW)

	if _, statErr := os.Stat(hostOSPath); statErr == nil {
		if entries, readErr := os.ReadDir(hostOSPath); readErr == nil && len(entries) > 0 {
			emit(map[string]any{"error": "Directory " + installPath + " already exists and is not empty."})
			return
		}
	}

	emit(map[string]any{"status": "Downloading and extracting DokuWiki"})
	installedVersion, out, runErr := runDokuwikiExtract(ctx, userContext, phpContainer, installPath)
	if runErr != nil {
		emit(map[string]any{"error": "Download/extract failed: " + strings.TrimSpace(string(out))})
		cmsapp.EmitCleanupFiles(ctx, userContext, phpContainer, installPath, emit)
		return
	}

	emit(map[string]any{"status": "Hashing admin password"})
	passwordHash, hashErr := hashDokuwikiPassword(ctx, userContext, phpContainer, adminPassword)
	if hashErr != nil {
		emit(map[string]any{"error": "Failed to hash admin password: " + hashErr.Error()})
		cmsapp.EmitCleanupFiles(ctx, userContext, phpContainer, installPath, emit)
		return
	}

	emit(map[string]any{"status": "Writing DokuWiki configuration"})
	if confErr := writeDokuwikiConfig(ctx, userContext, phpContainer, installPath, dokuwikiConfigParams{
		Title:      siteTitle,
		AdminUser:  adminUser,
		AdminHash:  passwordHash,
		AdminName:  adminFullName,
		AdminEmail: web.FormOr(r, "admin_email", "admin@"+dom.DomainURL),
	}); confErr != nil {
		emit(map[string]any{"error": "Failed to write configuration: " + confErr.Error()})
		cmsapp.EmitCleanupFiles(ctx, userContext, phpContainer, installPath, emit)
		return
	}

	emit(map[string]any{"status": "Setting files permissions and owner to '" + userContext + "'"})
	if uid, uidErr := podmanmanager.GetUID(userContext); uidErr == nil {
		uidStr := strconv.Itoa(uid)
		_ = podmanmanager.Command(ctx, userContext, podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "chown", "-R", uidStr+":"+uidStr, installPath)).Run()
	}

	// record exactly what this install created so a later uninstall can remove precisely that, not the whole docroot
	_ = installmanifest.Record(hostOSPath)

	emit(map[string]any{"status": "Saving website information to Site Manager"})
	adminEmail := web.FormOr(r, "admin_email", "admin@"+dom.DomainURL)
	if _, insertErr := a.DB.ExecContext(ctx,
		"INSERT INTO sites (site_name, domain_id, admin_email, version, type) VALUES (?, ?, ?, ?, ?)",
		selectedDomain, domainID, adminEmail, installedVersion, "dokuwiki"); insertErr != nil {
		emit(map[string]any{"error": "DokuWiki installed, but an error occurred while saving to Site Manager: " + insertErr.Error()})
		return
	}
	websites.TriggerScreenshotGeneration(a, selectedDomain)
	waf.EnableProfileForNewSite(a, selectedDomain, "dokuwiki")

	_ = logger.RecordUserAction(a.Config, currentUsername, "installed DokuWiki on domain "+selectedDomain, ipAddress)
	web.Flash(a, w, r, "success", web.Tr(a, r, "DokuWiki installed successfully on %(selected_domain)s", "selected_domain", selectedDomain))
	emit(map[string]any{"status": "DokuWiki installation completed!", "admin_user": adminUser, "admin_password": adminPassword})
}

// runDokuwikiExtract downloads and extracts the stable tarball to a scratch dir then copies its top-level dokuwiki-<version>/ contents into installPath with cp -a (installPath may already exist as the docroot, even if empty, so mv won't work - same as sofawiki's runSofawikiExtract), returning the version read from VERSION
func runDokuwikiExtract(ctx context.Context, userContext, phpContainer, installPath string) (version string, out []byte, err error) {
	scratch := "/tmp/openpanel-dokuwiki-" + strconv.FormatInt(time.Now().UnixNano(), 36)
	script := `set -e
rm -rf ` + scratch + ` ` + scratch + `.tgz
mkdir -p ` + scratch + `
curl -sL -o ` + scratch + `.tgz ` + dokuwikiStableTarball + `
tar -xzf ` + scratch + `.tgz -C ` + scratch + `
SRC=$(find ` + scratch + ` -mindepth 1 -maxdepth 1 -type d | head -1)
mkdir -p ` + installPath + `
cp -a "$SRC"/. ` + installPath + `/
cat "$SRC"/VERSION
rm -rf ` + scratch + ` ` + scratch + `.tgz
rm -f ` + installPath + `/install.php`

	argv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "sh", "-c", script)
	out, err = podmanmanager.Command(ctx, userContext, argv).CombinedOutput()
	if err != nil {
		return "", out, err
	}
	version = strings.TrimSpace(string(out))
	if version == "" {
		version = "unknown"
	}
	return version, out, nil
}

type dokuwikiConfigParams struct {
	Title      string
	AdminUser  string
	AdminHash  string
	AdminName  string
	AdminEmail string
}

// writeDokuwikiConfig writes conf/local.php, users.auth.php and acl.auth.php directly, matching what install.php's wizard writes for a single-admin ACL-enabled wiki - AdminHash must already be hashed (see hashDokuwikiPassword), this never touches a clear-text password
func writeDokuwikiConfig(ctx context.Context, userContext, phpContainer, installPath string, p dokuwikiConfigParams) error {
	localPHP := "<?php\n" +
		"$conf['title'] = '" + cmsapp.EscapePHPSingleQuoted(p.Title) + "';\n" +
		"$conf['lang'] = 'en';\n" +
		"$conf['license'] = 'cc-by-sa';\n" +
		"$conf['useacl'] = 1;\n" +
		"$conf['superuser'] = '@admin';\n"

	usersAuth := "# users.auth.php\n" +
		"# <?php exit()?>\n" +
		"# Don't modify the lines above\n" +
		"#\n" +
		"# Auto-generated by OpenPanel install\n" +
		"#\n" +
		"# Format:\n" +
		"# login:passwordhash:Real Name:email:groups,comma,separated\n\n" +
		p.AdminUser + ":" + p.AdminHash + ":" + escapeAuthField(p.AdminName) + ":" + escapeAuthField(p.AdminEmail) + ":admin,user\n"

	aclAuth := "# acl.auth.php\n" +
		"# <?php exit()?>\n" +
		"# Don't modify the lines above\n" +
		"#\n" +
		"# Auto-generated by OpenPanel install\n\n" +
		"*               @ALL          8\n"

	script := `set -e
mkdir -p ` + installPath + `/conf
cat > ` + installPath + `/conf/local.php << 'OPENPANEL_EOF'
` + localPHP + `OPENPANEL_EOF
cat > ` + installPath + `/conf/users.auth.php << 'OPENPANEL_EOF'
` + usersAuth + `OPENPANEL_EOF
cat > ` + installPath + `/conf/acl.auth.php << 'OPENPANEL_EOF'
` + aclAuth + `OPENPANEL_EOF`

	argv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "sh", "-c", script)
	out, err := podmanmanager.Command(ctx, userContext, argv).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// escapeAuthField strips the colon delimiter users.auth.php's line format depends on, since a stray colon in a display name/email would silently shift every field after it
func escapeAuthField(value string) string {
	return strings.ReplaceAll(value, ":", "")
}
