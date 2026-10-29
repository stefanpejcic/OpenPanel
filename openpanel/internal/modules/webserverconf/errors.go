package webserverconf

import (
	"fmt"
	"net/http"
	"path"
	"regexp"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// testFailure is what the editor shows about a rejected save
type testFailure struct {
	Problems []string
	Hints    []string
	// line of the edited file to highlight, 0 when the problem is somewhere else
	Line int
	// domain whose VirtualHost file has the problem
	VhostDomain string
}

// containerConfPaths is where each web server's container sees the file this page edits
var containerConfPaths = map[string]string{
	"nginx":         "/etc/nginx/nginx.conf",
	"openresty":     "/usr/local/openresty/nginx/conf/nginx.conf",
	"apache":        "/usr/local/apache2/conf/httpd.conf",
	"openlitespeed": "/usr/local/lsws/conf/httpd_config.conf",
}

// vhostDirs is where each web server's container sees the domains' VirtualHost files
var vhostDirs = map[string]string{
	"nginx":         "/etc/nginx/conf.d/",
	"openresty":     "/etc/nginx/conf.d/",
	"apache":        "/etc/httpd/conf.d/",
	"openlitespeed": "/usr/local/lsws/conf/vhosts/",
}

var hostNotFound = regexp.MustCompile(`host not found in upstream "([^":]+)`)

type trFunc func(msg string, kv ...any) string

// requestTr is web.Tr with the request's translator looked up once
func requestTr(a *appctx.App, r *http.Request) trFunc {
	return translatorTr(web.RequestTranslator(a, r))
}

func translatorTr(t i18n.Translator) trFunc {
	return func(msg string, kv ...any) string {
		pairs := make([]string, len(kv))
		for i, v := range kv {
			pairs[i] = fmt.Sprint(v)
		}
		return t.Get(msg, pairs...)
	}
}

// describeTestFailure turns a failed config test's output into short problem lines named after the user's files, plus hints for the cases the output doesn't explain
func describeTestFailure(tr trFunc, entry webserverConfEntry, output string) testFailure {
	var f testFailure
	confPath := containerConfPaths[entry.ServiceName]
	vhostDir := vhostDirs[entry.ServiceName]
	hostsHinted := map[string]bool{}

	for _, p := range webserver.ConfigTestProblems(entry.ServiceName, output) {
		file := p.File
		switch {
		case file != "" && file == confPath:
			file = entry.ConfFile
			if f.Line == 0 {
				f.Line = p.Line
			}
		case file != "" && vhostDir != "" && strings.HasPrefix(file, vhostDir):
			file = path.Base(file)
			if f.VhostDomain == "" {
				f.VhostDomain = strings.TrimSuffix(file, ".conf")
			}
		}
		msg := p.Message
		if confPath != "" {
			msg = strings.ReplaceAll(msg, confPath, entry.ConfFile)
		}
		switch {
		case file != "" && p.Line > 0:
			f.Problems = append(f.Problems, tr("%(file)s, line %(line)s: %(message)s", "file", file, "line", p.Line, "message", msg))
		case file != "":
			f.Problems = append(f.Problems, tr("%(file)s: %(message)s", "file", file, "message", msg))
		default:
			f.Problems = append(f.Problems, msg)
		}

		if m := hostNotFound.FindStringSubmatch(p.Message); m != nil && !hostsHinted[m[1]] {
			hostsHinted[m[1]] = true
			f.Hints = append(f.Hints, tr("%(host)s could not be found. If it's one of your containers, start it and save again, otherwise check the name.", "host", m[1]))
		}
	}

	if f.VhostDomain != "" {
		f.Hints = append(f.Hints, tr("The problem is in the VirtualHost file of %(domain)s, not in %(conf_file)s. Fix that file first, then save again.", "domain", f.VhostDomain, "conf_file", entry.ConfFile))
	}
	switch {
	case output == webserver.TestTimedOut:
		f.Hints = append(f.Hints, tr("The test took longer than expected, the %(service_name)s image may still be downloading. Try again in a minute.", "service_name", entry.ServiceName))
	case strings.HasPrefix(output, webserver.TestNoOutput) && entry.ServiceName == "openlitespeed":
		f.Hints = append(f.Hints, tr("OpenLiteSpeed could not load this configuration at all. Check that the file isn't empty or cut off."))
	}
	return f
}
