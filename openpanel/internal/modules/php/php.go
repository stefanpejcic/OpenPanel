// Package php ports modules/php.py plus its phpd/* siblings
// (php_ini.py, php_options.py, php_extensions.py) and modules/phpmyadmin.py:
// default/per-domain PHP version selection, php.ini editing, PHP option
// tuning, extension management, and the phpMyAdmin redirect.
package php

import (
	"context"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cache"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
)

// phpVersionFromSegment extracts the version from a "php<version>" URL path
// segment (e.g. "php8.2" -> "8.2"). Go's net/http.ServeMux only allows a
// wildcard to span an entire path segment (see its Patterns doc), unlike
// Flask's <version> converter which could sit inside a literal "php<version>"
// segment - so routes register the whole segment as a wildcard and every
// handler unwraps it with this helper instead.
func phpVersionFromSegment(seg string) string {
	return strings.TrimPrefix(seg, "php")
}

// phpVersionFromIniSegment extracts the version from a "php<version>.ini"
// segment (e.g. "php8.2.ini" -> "8.2"), used by the php.ini editor route
// (/php/php<version>.ini/editor in the Python source).
func phpVersionFromIniSegment(seg string) string {
	return strings.TrimSuffix(phpVersionFromSegment(seg), ".ini")
}

func injected(a *appctx.App, r *http.Request) (username, userContext string, err error) {
	userID, _ := auth.UserID(r)
	data, err := a.InjectData(r.Context(), userID)
	if err != nil {
		return "", "", err
	}
	username, _ = data["current_username"].(string)
	userContext, _ = data["context"].(string)
	return username, userContext, nil
}

func flashAndRedirect(a *appctx.App, w http.ResponseWriter, r *http.Request, category, message, path string) {
	sess, _ := a.Sessions.Get(r, session.CookieName)
	flash.Add(sess, category, message)
	_ = a.Sessions.Save(r, w, sess)
	http.Redirect(w, r, path, http.StatusFound)
}

func flashSess(a *appctx.App, w http.ResponseWriter, r *http.Request, category, message string) {
	sess, _ := a.Sessions.Get(r, session.CookieName)
	flash.Add(sess, category, message)
	_ = a.Sessions.Save(r, w, sess)
}

// checkPHPIniSyntax mirrors check_php_ini_syntax(): have the PHP CLI inside
// the relevant container parse php.ini, returning the parse error string if
// invalid, or "" if valid or unable to check (e.g. service not running).
func checkPHPIniSyntax(ctx context.Context, userContext, version string) string {
	webServer := webserver.GetEnvFileValue(userContext, "WEB_SERVER")
	var container, iniPath string
	if strings.Contains(strings.ToLower(webServer), "litespeed") {
		container = webServer
		iniPath = "/usr/local/lsws/lsphp" + strings.ReplaceAll(version, ".", "") + "/etc/php/" + version + "/litespeed/php.ini"
	} else {
		container = "php-fpm-" + version
		iniPath = "/etc/php/" + version + "/fpm/php.ini"
	}

	if !docker.ComposeContainer(ctx, userContext, container, "status") {
		return ""
	}

	argv := podmanmanager.PodmanArgv(userContext, "exec", container, "php", "-c", iniPath, "-d", "display_errors=0", "-r", "exit(0);")
	cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	cmd := podmanmanager.Command(cctx, userContext, argv)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	_ = cmd.Run()

	return strings.TrimSpace(stderr.String())
}

// domainConfVersionRE mirrors the `php-fpm-(\d+\.\d+)` search used by
// get_php_v_for_domain() and php_version() to read the version straight out
// of a domain's vhost config.
var domainConfVersionRE = regexp.MustCompile(`php-fpm-(\d+\.\d+)`)

var validDomainRE = regexp.MustCompile(`^[a-zA-Z0-9.-]+$`)
var validContextRE = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// litespeedTagVersionRE mirrors the `(\d)(\d)$` fallback used to derive a
// PHP version from an OPENLITESPEED_VERSION/LITESPEED_VERSION tag (e.g.
// "1.8.5-lsphp83" -> "8.3") when the container isn't running to ask directly.
var litespeedTagVersionRE = regexp.MustCompile(`(\d)(\d)$`)

// GetPHPVForDomain mirrors get_php_v_for_domain(): the PHP version
// currently configured for domainURL, read from its vhost file (PHP-FPM) or
// queried live from the running container (LiteSpeed). Exported for later
// app-installer phases (wordpress.py, drupal.py, mautic.py, websites.py)
// that call the Python equivalent.
func GetPHPVForDomain(ctx context.Context, a *appctx.App, userContext, domainURL string) string {
	webServer := webserver.GetEnvFileValue(userContext, "WEB_SERVER")

	if !validDomainRE.MatchString(domainURL) || !validContextRE.MatchString(userContext) {
		return "/"
	}

	if strings.Contains(strings.ToLower(webServer), "litespeed") {
		command := `php -d memory_limit=-1 -d open_basedir=none -d disable_functions= -d display_errors=0 -d error_log=/dev/null -r "echo PHP_VERSION, PHP_EOL;"`
		argv := podmanmanager.PodmanArgv(userContext, "exec", webServer, "sh", "-c", command)
		out, err := podmanmanager.Command(ctx, userContext, argv).Output()
		if err == nil {
			return strings.TrimSpace(string(out))
		}

		var tagKey string
		if webServer == "openlitespeed" {
			tagKey = "OPENLITESPEED_VERSION"
		} else if webServer == "litespeed" {
			tagKey = "LITESPEED_VERSION"
		}
		tag := webserver.GetEnvFileValue(userContext, tagKey)
		if tag == "latest" {
			return "8.3"
		}
		if m := litespeedTagVersionRE.FindStringSubmatch(tag); m != nil {
			return m[1] + "." + m[2]
		}
		return "8.3"
	}

	confPath := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_webserver_data/_data/" + domainURL + ".conf"
	if content, err := os.ReadFile(confPath); err == nil {
		lines := strings.SplitN(string(content), "\n", 2)
		if len(lines) > 0 {
			if m := domainConfVersionRE.FindStringSubmatch(lines[0]); m != nil {
				return m[1]
			}
		}
	}

	var version string
	err := a.DB.QueryRowContext(ctx, "SELECT php_version FROM domains WHERE domain_url = ?", domainURL).Scan(&version)
	if err != nil || version == "" {
		return "/"
	}
	return version
}

// updatePHPVersionPreference mirrors update_php_version_preference().
func updatePHPVersionPreference(userContext, newPHPVersion string) bool {
	webServer := webserver.GetEnvFileValue(userContext, "WEB_SERVER")
	newPHPVersion = strings.TrimPrefix(newPHPVersion, "php")

	var placeholder string
	switch webServer {
	case "openlitespeed":
		placeholder = "OPENLITESPEED_VERSION"
		newPHPVersion = "1.8.5-lsphp" + strings.ReplaceAll(newPHPVersion, ".", "")
	case "litespeed":
		placeholder = "LITESPEED_VERSION"
		newPHPVersion = "6.3.5-lsphp" + strings.ReplaceAll(newPHPVersion, ".", "")
	default:
		placeholder = "DEFAULT_PHP_VERSION"
	}

	configFilePath := "/home/" + userContext + "/.env"
	content, err := os.ReadFile(configFilePath)
	if err != nil {
		return false
	}

	lines := strings.Split(string(content), "\n")
	prefix := placeholder + "="
	for i, line := range lines {
		if strings.HasPrefix(line, prefix) {
			lines[i] = placeholder + "=" + newPHPVersion
		}
	}
	return os.WriteFile(configFilePath, []byte(strings.Join(lines, "\n")), 0o644) == nil
}

// fetchPHPVersions mirrors fetch_php_versions(), memoized 1h like the
// Python @cache.memoize(timeout=3600).
func FetchPHPVersions(ctx context.Context, a *appctx.App, userContext string) []string {
	versions, _ := cache.Memoize(ctx, a.Cache, "fetch_php_versions:"+userContext, time.Hour, func() ([]string, error) {
		return computeFetchPHPVersions(ctx, userContext), nil
	})
	return versions
}

func computeFetchPHPVersions(ctx context.Context, userContext string) []string {
	webServer := webserver.GetEnvFileValue(userContext, "WEB_SERVER")

	if strings.Contains(strings.ToLower(webServer), "litespeed") {
		return fetchLitespeedTags(ctx, webServer)
	}

	argv := podmanmanager.PodmanComposeArgv("-f", "/home/"+userContext+"/docker-compose.yml", "config", "--services")
	out, err := podmanmanager.Command(ctx, userContext, argv).Output()
	if err != nil {
		return nil
	}

	var versions []string
	for _, service := range strings.Split(string(out), "\n") {
		service = strings.TrimSpace(service)
		if v, ok := strings.CutPrefix(service, "php-fpm-"); ok {
			versions = append(versions, v)
		}
	}
	sortVersionsDesc(versions)
	return versions
}

// sortVersionsDesc mirrors sorted(..., key=lambda x: tuple(map(int,
// x.split('.'))), reverse=True): numeric major.minor comparison, not
// lexical.
func sortVersionsDesc(versions []string) {
	less := func(a, b string) bool {
		amaj, amin := parseMajorMinor(a)
		bmaj, bmin := parseMajorMinor(b)
		if amaj != bmaj {
			return amaj > bmaj
		}
		return amin > bmin
	}
	for i := 1; i < len(versions); i++ {
		for j := i; j > 0 && less(versions[j], versions[j-1]); j-- {
			versions[j], versions[j-1] = versions[j-1], versions[j]
		}
	}
}

func parseMajorMinor(v string) (int, int) {
	parts := strings.SplitN(v, ".", 2)
	maj, _ := strconv.Atoi(parts[0])
	min := 0
	if len(parts) > 1 {
		min, _ = strconv.Atoi(parts[1])
	}
	return maj, min
}

// stopPHPServiceIfRunningAndUnused mirrors
// stop_php_service_if_running_and_unused().
func stopPHPServiceIfRunningAndUnused(ctx context.Context, userContext, oldPHPVersion string) {
	phpDefaultVersion := webserver.GetEnvFileValue(userContext, "DEFAULT_PHP_VERSION")
	if oldPHPVersion == phpDefaultVersion {
		return
	}

	phpFPMVersion := "php-fpm-" + oldPHPVersion
	basePath := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_webserver_data/_data/"

	entries, err := os.ReadDir(basePath)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".conf") {
			continue
		}
		content, err := os.ReadFile(basePath + entry.Name())
		if err != nil {
			continue
		}
		if strings.Contains(string(content), phpFPMVersion) {
			return
		}
	}

	docker.StartOrStopContainer(ctx, userContext, phpFPMVersion, "deactivate", "")
}
