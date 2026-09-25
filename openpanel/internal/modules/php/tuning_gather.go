package php

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/dbtuning"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
)

// errorLogTail is how many container log lines are scanned for known problems
const errorLogTail = 1000

var maxChildrenRE = regexp.MustCompile(`pm\.max_children\s*=\s*(\d+)`)

// phpContainer is the service that runs PHP for this version, the web server itself on LiteSpeed
func phpContainer(userContext, version string) string {
	if ws := webserver.GetEnvFileValue(userContext, "WEB_SERVER"); strings.Contains(strings.ToLower(ws), "litespeed") {
		return ws
	}
	return "php-fpm-" + version
}

// effectiveIni asks PHP inside the container for the live values of keys, the CLI always reports max_execution_time as 0 so that one is left out
func effectiveIni(ctx context.Context, userContext, container string, keys []string) map[string]string {
	out := map[string]string{}
	var ask []string
	for _, k := range keys {
		if k != "max_execution_time" {
			ask = append(ask, k)
		}
	}
	b, _ := json.Marshal(ask)
	script := `$o=[];foreach(json_decode($argv[1]) as $k){$v=ini_get($k);if($v!==false)$o[$k]=(string)$v;}echo json_encode($o);`
	cctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	argv := podmanmanager.PodmanArgv(userContext, "exec", container, "php", "-r", script, string(b))
	if res, err := podmanmanager.Command(cctx, userContext, argv).Output(); err == nil {
		_ = json.Unmarshal(res, &out)
	}
	return out
}

// fpmMaxChildren sums pm.max_children over the pools php-fpm actually loaded
func fpmMaxChildren(ctx context.Context, userContext, container string) int {
	cctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	argv := podmanmanager.PodmanArgv(userContext, "exec", container, "php-fpm", "-tt")
	res, _ := podmanmanager.Command(cctx, userContext, argv).CombinedOutput()
	total := 0
	for _, m := range maxChildrenRE.FindAllStringSubmatch(string(res), -1) {
		n, _ := strconv.Atoi(m[1])
		total += n
	}
	return total
}

// envSize reads compose style limits like 1.0G from .env
func envSize(v string) int64 {
	v = strings.ToUpper(strings.TrimSpace(v))
	mult := int64(1)
	switch {
	case strings.HasSuffix(v, "G"):
		mult, v = gb, strings.TrimSuffix(v, "G")
	case strings.HasSuffix(v, "M"):
		mult, v = mb, strings.TrimSuffix(v, "M")
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil || f <= 0 {
		return 0
	}
	return int64(f * float64(mult))
}

// gatherPHPTuningStats collects everything BuildPHPRecommendations needs for one PHP version
func gatherPHPTuningStats(ctx context.Context, a *appctx.App, userID int, userContext, version string) PHPTuningStats {
	s := PHPTuningStats{Version: version, Settings: map[string]string{}, AvailableKeys: loadKeysFromFile(userContext)}

	if content, err := os.ReadFile("/home/" + userContext + "/php.ini/" + version + ".ini"); err == nil {
		file := configEntriesToMap(parseConfigContent(string(content)))
		for _, k := range s.AvailableKeys {
			if v, ok := file[k]; ok {
				s.Settings[k] = strings.Trim(v, `"`)
			}
		}
	}

	container := phpContainer(userContext, version)
	s.Running = docker.IsServiceRunning(ctx, userContext, container)
	if s.Running {
		for k, v := range effectiveIni(ctx, userContext, container, s.AvailableKeys) {
			if _, set := s.Settings[k]; !set {
				s.Settings[k] = v
			}
		}
		if strings.HasPrefix(container, "php-fpm-") {
			s.MaxChildren = fpmMaxChildren(ctx, userContext, container)
		}
		c := dbtuning.InspectContainer(ctx, userContext, container)
		s.RAMLimit, s.MemUsage, s.CPUs, s.OOMKilled = c.MemLimit, c.MemUsage, c.CPUs, c.OOMKilled
		if body, status := docker.FetchContainerLog(ctx, a, userContext, container, errorLogTail); status == http.StatusOK {
			for _, line := range strings.Split(body, "\n") {
				if strings.Contains(line, "PHP ") || strings.Contains(line, "WARNING") || strings.Contains(line, "exceeds the limit") {
					s.LogLines = append(s.LogLines, line)
				}
			}
		}
	}
	if _, set := s.Settings["max_execution_time"]; !set {
		s.Settings["max_execution_time"] = "30"
	}
	if s.RAMLimit <= 0 {
		s.RAMLimit = envSize(webserver.GetEnvFileValue(userContext, "PHP_FPM_"+strings.ReplaceAll(version, ".", "_")+"_RAM"))
	}

	domainsList, _ := a.AllDomainsForUser(ctx, userID)
	rows, _, _ := buildPHPDomainRows(userContext, domainsList, nil)
	docroots := map[string]string{}
	for _, d := range domainsList {
		docroots[d.DomainURL] = d.Docroot
	}
	htmlRoot := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data"
	for _, row := range rows {
		if row.PHPVersion != version {
			continue
		}
		s.Domains = append(s.Domains, row.DomainURL)
		if rel, err := filepath.Rel("/var/www/html", filepath.Clean(docroots[row.DomainURL])); err == nil && !strings.HasPrefix(rel, "..") {
			if _, err := os.Stat(filepath.Join(htmlRoot, rel, "wp-config.php")); err == nil {
				s.WordPressSites++
			}
		}
	}
	return s
}

// handlePHPOptionsRecommendations returns tuning suggestions for one PHP version's options page, loaded with ajax
func handlePHPOptionsRecommendations(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	_, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	version := phpVersionFromSegment(r.PathValue("phpversion"))
	if !phpVersionFormRE.MatchString(version) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid PHP version."})
		return
	}
	if _, err := os.Stat("/home/" + userContext + "/php.ini/" + version + ".ini"); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "PHP " + version + " is not installed."})
		return
	}
	userID, _ := auth.UserID(r)
	writeJSON(w, http.StatusOK, BuildPHPRecommendations(gatherPHPTuningStats(r.Context(), a, userID, userContext, version)))
}
