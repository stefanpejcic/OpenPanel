package appinstall

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
)

// one call from repo to a running site, built on the per-language installers, so the compose entry, proxy and Sites row are the same as a panel install

const (
	deployDefaultWait = 180
	deployMaxWait     = 600
	deployProbeEvery  = 5 * time.Second
)

// deployStepsKept is how many installer progress lines go back to the caller
const deployStepsKept = 12

var deployServiceRE = regexp.MustCompile(`^[a-z][a-z0-9]{1,30}$`)

// RegisterDeployAPI wires the deploy routes, mux is kept so the installer runs behind its own feature gate
func RegisterDeployAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "pm2", "POST /api/deploy/detect", func(w http.ResponseWriter, r *http.Request) { apiDeployDetect(w, r) })
	apiregistry.Handle(mux, a, "pm2", "POST /api/deploy", func(w http.ResponseWriter, r *http.Request) { apiDeploy(a, mux, w, r) })
	apiregistry.Handle(mux, a, "pm2", "GET /api/deploy/status", func(w http.ResponseWriter, r *http.Request) { apiDeployStatus(a, w, r) })
}

func writeDeployJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func detectError(w http.ResponseWriter, err error) {
	status := http.StatusBadGateway
	if errors.Is(err, ErrUnsupportedStack) {
		status = http.StatusUnprocessableEntity
	}
	writeDeployJSON(w, status, map[string]string{"error": "Could not detect how to run the repository: " + err.Error()})
}

func apiDeployDetect(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RepoURL string `json:"repo_url"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	repo := strings.TrimSpace(body.RepoURL)
	if repo == "" || !isValidGitURL(repo) {
		writeDeployJSON(w, http.StatusBadRequest, map[string]string{"error": "repo_url must be a public https:// git repository URL."})
		return
	}
	plan, err := DetectStack(r.Context(), repo)
	if err != nil {
		detectError(w, err)
		return
	}
	writeDeployJSON(w, http.StatusOK, plan)
}

type deployRequest struct {
	RepoURL      string `json:"repo_url"`
	Domain       string `json:"domain"`
	Subdirectory string `json:"subdirectory"`
	ServiceName  string `json:"service_name"`
	AppType      string `json:"app_type"`
	StartCmd     string `json:"start_cmd"`
	Port         int    `json:"port"`
	Version      string `json:"version"`
	Install      *bool  `json:"install_dependencies"`
	CPULimit     string `json:"cpu_limit"`
	MemLimit     string `json:"mem_limit"`
	WaitSeconds  *int   `json:"wait_seconds"`
}

// defaultServiceName turns example.com/blog into exampleblog, short enough for env keys
func defaultServiceName(domain, subdirectory string) string {
	var b strings.Builder
	for _, c := range strings.ToLower(strings.SplitN(domain, ".", 2)[0] + subdirectory) {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
			b.WriteRune(c)
		}
	}
	name := b.String()
	if name == "" || name[0] < 'a' {
		name = "app" + name
	}
	if len(name) > 24 {
		name = name[:24]
	}
	return name
}

// applyOverrides lets the caller fix anything detection got wrong, and returns a reason when an override is unsafe
func applyOverrides(plan *StackPlan, req deployRequest) string {
	if req.AppType != "" {
		if _, ok := kindsByAppType[req.AppType]; !ok {
			return "app_type must be nodejs, python, ruby or java."
		}
		plan.AppType = req.AppType
	}
	if req.StartCmd != "" {
		// it ends up in .env as KEY="..." and inside sh -c "...", so no quotes, backslashes or new lines
		if !isValidCustomCommand(req.StartCmd) || strings.ContainsAny(req.StartCmd, "\"\\`") || len(req.StartCmd) > 500 {
			return "start_cmd can't contain quotes, backslashes, backticks, .. or new lines."
		}
		plan.StartCmd = req.StartCmd
	}
	if req.Port != 0 {
		if req.Port < 1 || req.Port > 65535 {
			return "port must be 1 to 65535."
		}
		plan.Port = req.Port
	}
	if req.Version != "" {
		if req.Version != "latest" && !isValidVersion(req.Version) {
			return "version must be latest or a numeric image tag like 20 or 3.12."
		}
		plan.Version = req.Version
	}
	if req.Install != nil {
		plan.Install = *req.Install
	}
	return ""
}

// lookupOwnedDomain returns the domain_id of an exact domain the caller owns
func lookupOwnedDomain(ctx context.Context, a *appctx.App, userID int, domain string) (string, bool) {
	if domain == "" || !a.CheckExactDomainBelongsToUser(ctx, userID, domain) {
		return "", false
	}
	var id int
	if err := a.DB.QueryRowContext(ctx, "SELECT domain_id FROM domains WHERE domain_url = ? AND user_id = ?", domain, userID).Scan(&id); err != nil {
		return "", false
	}
	return strconv.Itoa(id), true
}

// installOutcome is the end of the installer's NDJSON stream
type installOutcome struct {
	OK    bool     `json:"ok"`
	Error string   `json:"error,omitempty"`
	Steps []string `json:"steps"`
}

func parseInstallStream(raw []byte, status int) installOutcome {
	out := installOutcome{Steps: []string{}}
	if status >= 400 {
		var e map[string]any
		_ = json.Unmarshal(raw, &e)
		msg, _ := e["error"].(string)
		if msg == "" {
			msg = strings.TrimSpace(string(raw))
		}
		out.Error = msg
		return out
	}
	for _, line := range bytes.Split(raw, []byte("\n")) {
		var ev map[string]any
		if json.Unmarshal(bytes.TrimSpace(line), &ev) != nil {
			continue
		}
		if msg, _ := ev["error"].(string); msg != "" {
			out.Error = msg
		}
		if msg, _ := ev["status"].(string); msg != "" {
			out.Steps = append(out.Steps, msg)
		}
	}
	if len(out.Steps) > deployStepsKept {
		out.Steps = out.Steps[len(out.Steps)-deployStepsKept:]
	}
	out.OK = out.Error == "" && len(out.Steps) > 0
	return out
}

// runInstaller replays the form the panel's own install page posts, through mux so the per-language feature, token scope and demo mode still apply
func runInstaller(mux http.Handler, r *http.Request, appType string, form url.Values) installOutcome {
	req := httptest.NewRequest(http.MethodPost, "/api/"+appType+"/install", strings.NewReader(form.Encode()))
	req = req.WithContext(r.Context())
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", r.Header.Get("Authorization"))
	req.RemoteAddr = r.RemoteAddr
	for _, h := range []string{"X-Forwarded-For", "CF-Connecting-IP"} {
		for _, v := range r.Header.Values(h) {
			req.Header.Add(h, v)
		}
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return parseInstallStream(rec.Body.Bytes(), rec.Code)
}

// appHealth is what deploy and deploy status report about a running app
type appHealth struct {
	Container  string   `json:"container_state"`
	Listening  bool     `json:"listening"`
	HTTPStatus int      `json:"http_status,omitempty"`
	ProbeError string   `json:"probe_error,omitempty"`
	Logs       []string `json:"logs,omitempty"`
}

// probeCommand asks the app's own runtime to fetch it, so no extra tools are needed in the image
func probeCommand(appType string, port int) []string {
	target := "http://127.0.0.1:" + strconv.Itoa(port) + "/"
	switch appType {
	case "nodejs":
		return []string{"node", "-e", "require('http').get('" + target + "',r=>{console.log(r.statusCode);process.exit(0)}).on('error',e=>{console.log('ERR '+e.code);process.exit(1)})"}
	case "python":
		return []string{"python", "-c", "import urllib.request as u,urllib.error as e\ntry:\n print(u.urlopen('" + target + "',timeout=5).status)\nexcept e.HTTPError as x:\n print(x.code)\nexcept Exception as x:\n print('ERR',x)"}
	case "ruby":
		return []string{"ruby", "-rnet/http", "-e", "begin; puts Net::HTTP.get_response(URI('" + target + "')).code; rescue => e; puts 'ERR ' + e.class.name; end"}
	}
	return nil
}

var probeRunner = func(ctx context.Context, userContext, container string, argv []string) (string, error) {
	cctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	out, err := podmanmanager.Command(cctx, userContext, podmanmanager.PodmanArgv(userContext, append([]string{"exec", container}, argv...)...)).CombinedOutput()
	return string(out), err
}

var containerStateFn = func(ctx context.Context, userContext, container string) string {
	return docker.GetContainerStatus(ctx, userContext, container).State
}

func probeApp(ctx context.Context, userContext, container, appType string, port int) appHealth {
	h := appHealth{Container: containerStateFn(ctx, userContext, container)}
	if h.Container != "running" {
		return h
	}
	argv := probeCommand(appType, port)
	if argv == nil {
		h.ProbeError = "no HTTP probe for " + appType + " apps, check the logs"
		return h
	}
	out, _ := probeRunner(ctx, userContext, container, argv)
	line := strings.TrimSpace(out)
	if i := strings.LastIndexByte(line, '\n'); i >= 0 {
		line = strings.TrimSpace(line[i+1:])
	}
	if code, err := strconv.Atoi(line); err == nil {
		h.Listening, h.HTTPStatus = true, code
	} else {
		h.ProbeError = "nothing answering on port " + strconv.Itoa(port) + " yet (" + line + ")"
	}
	return h
}

func recentLogs(ctx context.Context, userContext, container string, lines int) []string {
	body, status := docker.FetchContainerLogSince(ctx, userContext, container, lines, 24*time.Hour)
	if status != http.StatusOK {
		return nil
	}
	out := strings.Split(strings.TrimRight(body, "\n"), "\n")
	if len(out) == 1 && out[0] == "" {
		return nil
	}
	return out
}

// waitForApp probes until the app answers, the container stops, or wait runs out
func waitForApp(ctx context.Context, userContext, container, appType string, port int, wait time.Duration) appHealth {
	deadline := time.Now().Add(wait)
	for {
		h := probeApp(ctx, userContext, container, appType, port)
		if h.Listening || h.Container == "exited" || h.Container == "not_found" || time.Now().After(deadline) || ctx.Err() != nil {
			return h
		}
		select {
		case <-ctx.Done():
			return h
		case <-time.After(deployProbeEvery):
		}
	}
}

func apiDeploy(a *appctx.App, mux http.Handler, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	var req deployRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeDeployJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON body."})
		return
	}
	req.RepoURL = strings.TrimSpace(req.RepoURL)
	req.Domain = strings.ToLower(strings.TrimSpace(req.Domain))
	req.Subdirectory = strings.Trim(strings.ToLower(strings.TrimSpace(req.Subdirectory)), "/")
	if req.RepoURL == "" || !isValidGitURL(req.RepoURL) {
		writeDeployJSON(w, http.StatusBadRequest, map[string]string{"error": "repo_url must be a public https:// git repository URL."})
		return
	}
	domainID, ok := lookupOwnedDomain(ctx, a, userID, req.Domain)
	if !ok {
		writeDeployJSON(w, http.StatusForbidden, map[string]string{"error": "domain must be a domain on this account, like example.com."})
		return
	}
	service := req.ServiceName
	if service == "" {
		service = defaultServiceName(req.Domain, req.Subdirectory)
	}
	if !deployServiceRE.MatchString(service) {
		writeDeployJSON(w, http.StatusBadRequest, map[string]string{"error": "service_name must be 2 to 31 lowercase letters and digits, starting with a letter."})
		return
	}
	wait := deployDefaultWait
	if req.WaitSeconds != nil {
		wait = *req.WaitSeconds
	}
	if wait < 0 || wait > deployMaxWait {
		writeDeployJSON(w, http.StatusBadRequest, map[string]string{"error": "wait_seconds must be 0 to 600."})
		return
	}

	plan := &StackPlan{}
	if req.AppType == "" || req.StartCmd == "" {
		detected, err := DetectStack(ctx, req.RepoURL)
		if err != nil {
			detectError(w, err)
			return
		}
		plan = detected
	}
	if msg := applyOverrides(plan, req); msg != "" {
		writeDeployJSON(w, http.StatusBadRequest, map[string]string{"error": msg})
		return
	}
	if plan.Port == 0 {
		writeDeployJSON(w, http.StatusBadRequest, map[string]string{"error": "port is required when app_type and start_cmd are both given."})
		return
	}

	form := url.Values{
		"domain_id": {domainID}, "service_name": {service}, "subdirectory": {req.Subdirectory},
		"git_repo_url": {req.RepoURL}, "port": {strconv.Itoa(plan.Port)}, "custom_cmd": {plan.StartCmd},
		"startup_file": {plan.StartupFile}, "version": {plan.Version},
		"cpu_limit": {req.CPULimit}, "mem_limit": {req.MemLimit},
	}
	if plan.Install {
		form.Set("requirements", "1")
	}
	install := runInstaller(mux, r, plan.AppType, form)

	username, userContext, _ := injectedContext(a, r)
	site := req.Domain
	if req.Subdirectory != "" {
		site += "/" + req.Subdirectory
	}
	resp := map[string]any{"site": site, "service": service, "plan": plan, "install": install}
	if !install.OK {
		resp["logs"] = recentLogs(ctx, userContext, service, 30)
		writeDeployJSON(w, http.StatusUnprocessableEntity, resp)
		return
	}
	_ = logger.RecordUserAction(a.Config, username, "deployed "+req.RepoURL+" to "+site+" as "+plan.AppType+" app "+service, reqip.ClientIP(r))

	health := waitForApp(ctx, userContext, service, plan.AppType, plan.Port, time.Duration(wait)*time.Second)
	if !health.Listening {
		health.Logs = recentLogs(ctx, userContext, service, 30)
	}
	resp["health"] = health
	status := http.StatusOK
	switch {
	case health.Listening && health.HTTPStatus < 500:
		resp["next"] = "The app is answering. Open https://" + site + "/ to check it."
	case health.Container == "running":
		status = http.StatusAccepted
		resp["next"] = "Installed, but the app isn't answering yet (it may still be installing dependencies or building). Poll GET /api/deploy/status?site=" + site + " and read the logs."
	default:
		status = http.StatusAccepted
		resp["next"] = "Installed, but the container is " + health.Container + ". Read the logs. A fix in the repo needs a push and POST /api/pm2/" + site + "/restart, a different start_cmd or port needs DELETE /api/pm2/" + site + " and a new deploy."
	}
	writeDeployJSON(w, status, resp)
}

func apiDeployStatus(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	site := strings.ToLower(strings.Trim(strings.TrimSpace(r.URL.Query().Get("site")), "/"))
	domainRoot, _, _ := strings.Cut(site, "/")
	if site == "" || !a.CheckDomainBelongsToUser(ctx, userID, domainRoot) {
		writeDeployJSON(w, http.StatusForbidden, map[string]string{"error": "site must be a site on this account, like example.com or example.com/blog."})
		return
	}
	var container, siteType string
	var ports sql.NullInt64
	err := a.DB.QueryRowContext(ctx, "SELECT container, type, ports FROM sites WHERE site_name = ?", site).Scan(&container, &siteType, &ports)
	if err != nil {
		writeDeployJSON(w, http.StatusNotFound, map[string]string{"error": "No app is installed on " + site + "."})
		return
	}
	kind, ok := kindsByAppType[strings.ToLower(siteType)]
	if !ok {
		writeDeployJSON(w, http.StatusNotFound, map[string]string{"error": site + " is a " + siteType + " site, not a deployed app."})
		return
	}
	_, userContext, _ := injectedContext(a, r)
	container = strings.ToLower(container)
	port := int(ports.Int64)
	h := appHealth{Container: containerStateFn(ctx, userContext, container)}
	if port > 0 {
		h = probeApp(ctx, userContext, container, kind.AppType, port)
	} else {
		h.ProbeError = "this app has no proxy port recorded, so it can't be probed"
	}
	lines := 30
	if v, err := strconv.Atoi(r.URL.Query().Get("lines")); err == nil && v > 0 && v <= 500 {
		lines = v
	}
	h.Logs = recentLogs(ctx, userContext, container, lines)
	writeDeployJSON(w, http.StatusOK, map[string]any{"site": site, "service": container, "app_type": kind.AppType, "port": port, "health": h})
}
