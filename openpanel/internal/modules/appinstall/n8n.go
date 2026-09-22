package appinstall

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"html/template"
	"strings"
	"time"

	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
)

// n8n-specific install/display config, kept apart from nodejs.go/python.go/ruby.go/java.go so each type is easy to find and change independently. Unlike those, n8n is a fixed workflow-automation app with no startup file, package manager, or git deploy - see Kind.Simple/Kind.DataVolume.

const n8nIconSVG = `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="#EA4B71" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="w-6 h-6 icon icon-tabler icons-tabler-outline icon-tabler-hierarchy"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><circle cx="12" cy="5" r="2" /><circle cx="5" cy="19" r="2" /><circle cx="19" cy="19" r="2" /><path d="M12 7v3.5a3.5 3.5 0 0 1 -3.5 3.5h-1.5M12 10.5a3.5 3.5 0 0 0 3.5 3.5h1.5" /><path d="M5 17v-1.5M19 17v-1.5" /></svg>`

var n8nDisplay = kindDisplay{
	Icon:  template.HTML(n8nIconSVG), //nolint:gosec // static, server-defined markup, not user input
	Label: "n8n",
}

var N8N = Kind{
	AppType: "n8n", DisplayAppType: "n8n", PyOrNode: "N8N", Title: "Install n8n",
	Simple: true, DataVolume: true, NeedsWebSocket: true,
}

// n8nOwnerSetupScript POSTs the JSON it reads from stdin to n8n's own POST /rest/owner/setup REST endpoint - the same call its browser-based /setup wizard makes - creating the first ("owner") account non-interactively, then verifies the credentials actually work with a POST /rest/login call. The verification step (and the delay before it) matter: during n8n's cold start, /rest/owner/setup has been observed to return 200 and even pass an *immediate* login check, while a login moments later with the exact same credentials still fails with 401 - manual testing traced this to the owner's password not being durably persisted until slightly after the 200 response, so a bare "200 means done" check, or even a same-tick verification, isn't reliable. A short pause before verifying is what actually catches it. Run via `podman exec` (see setupN8NOwner) rather than from this Go process directly: the process has no network route into the per-user podman network the "www" bridge lives on, but a command executed inside the n8n container itself can always reach its own app on localhost. Reads the payload from stdin instead of interpolating it into the script text, so the owner's email/name/password (user-supplied form input) never has to be escaped into a shell/JS literal. Stderr is prefixed "retryable:" for failures worth another full attempt (see setupN8NOwner) versus a real, non-transient error (e.g. n8n rejecting the password as too weak).
const n8nOwnerSetupScript = `let d='';process.stdin.on('data',c=>d+=c);process.stdin.on('end',async()=>{let payload;try{payload=JSON.parse(d)}catch(e){process.stderr.write('bad-payload: '+String(e));process.exit(1)}try{const setupRes=await fetch('http://localhost:5678/rest/owner/setup',{method:'POST',headers:{'Content-Type':'application/json'},body:d});if(!setupRes.ok){process.stderr.write(await setupRes.text());process.exit(1)}await new Promise(r=>setTimeout(r,2500));const loginRes=await fetch('http://localhost:5678/rest/login',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({email:payload.email,password:payload.password})});if(!loginRes.ok){process.stderr.write('retryable: login verification failed: '+await loginRes.text());process.exit(1)}process.exit(0)}catch(e){process.stderr.write('retryable: '+String(e));process.exit(1)}});`

// n8nOwnerSetupPollAttempts/n8nOwnerSetupPollInterval bound how long setupN8NOwner waits for n8n's HTTP server to actually start accepting requests - State.Running=true (already confirmed by the caller) only means the container process started, not that n8n has finished its startup migrations and bound port 5678, which reliably takes a few seconds longer
const (
	n8nOwnerSetupPollAttempts = 20
	n8nOwnerSetupPollInterval = 2 * time.Second
)

// setupN8NOwner creates n8n's first ("owner") account right after install, so the user lands on n8n's login screen instead of its own /setup wizard.
func setupN8NOwner(ctx context.Context, userContext, serviceName, email, firstName, lastName, password string) error {
	payload, marshalErr := json.Marshal(map[string]string{
		"email": email, "firstName": firstName, "lastName": lastName, "password": password,
	})
	if marshalErr != nil {
		return marshalErr
	}

	var lastErr error
	for attempt := 0; attempt < n8nOwnerSetupPollAttempts; attempt++ {
		if attempt > 0 {
			time.Sleep(n8nOwnerSetupPollInterval)
		}
		argv := podmanmanager.PodmanArgv(userContext, "exec", "-i", serviceName, "node", "-e", n8nOwnerSetupScript)
		cmd := podmanmanager.Command(ctx, userContext, argv)
		cmd.Stdin = bytes.NewReader(payload)
		out, runErr := cmd.CombinedOutput()
		if runErr == nil {
			return nil
		}
		if msg := strings.TrimSpace(string(out)); msg != "" {
			lastErr = errors.New(msg)
		} else {
			lastErr = runErr
		}
		// every failure mode here has turned out to be a startup race rather than a real validation error: "fetch failed" before n8n's HTTP server is listening, a default Express "Cannot POST" while it's listening but hasn't finished mounting routes yet, and the login-verification race (see n8nOwnerSetupScript). The password itself is already validated by the caller before this is ever called, so retrying unconditionally up to the attempt budget is safe - /rest/owner/setup is idempotent once the owner exists.
	}
	return lastErr
}
