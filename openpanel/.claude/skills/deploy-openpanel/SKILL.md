---
name: deploy-openpanel
description: Build the OpenPanel Go binary and deploy it into the test server's 'openpanel' podman container, then smoke-test it live.
---

# Deploy OpenPanel to the test server

Deploys the current working tree of this repo (`cmd/openpanel`) to a test
server by replacing the binary inside the running `openpanel` podman
container, then verifies it actually works by logging into the live app.

**Never write the server address, SSH password, or app login credentials
into any file in this repo.** Ask for them in the conversation each time
this skill runs, use them only inline in the Bash commands that need them,
and don't echo them back or persist them anywhere.

## Step 0 — Get connection details for this run

Ask the user (if not already given earlier in this conversation) for:
1. The server's IP address or hostname (call it `$SERVER` below).
2. The root SSH password for that server.
3. OpenPanel *user* login credentials to use for the post-deploy smoke
   test (a username/password that can log into `https://$SERVER:2083` —
   confirm the port too, in case it's not 2083 on this server; this is the
   end-user panel, a different app/port than OpenAdmin's 2087).

## Step 1 — Build and test locally first

From the repo root:

```bash
go build ./... && go test ./internal/modules/<relevant-module>/... -v 2>&1 | tail -60
```

Fix or flag any failures before deploying.

Then cross-compile for the server. **This container is Alpine (musl libc),
not glibc** — a normal `go build` produces a dynamically-linked glibc
binary that crashes on exec with `missing dynamic library`. Always build
static with `CGO_ENABLED=0`:

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /tmp/claude-*/scratchpad/openpanel-amd64 ./cmd/openpanel
file /tmp/claude-*/scratchpad/openpanel-amd64   # should say "statically linked", not "dynamically linked"
```

(Use this session's actual scratchpad path, not a literal glob.)

## Step 2 — Copy the binary to the server

```bash
export SSHPASS='<password from step 0>'
sshpass -e scp -o StrictHostKeyChecking=accept-new \
  <scratchpad>/openpanel-amd64 \
  root@<server>:/root/openpanel-amd64.new
```

## Step 3 — Swap the binary into the running container and restart it

The binary lives at `/openpanel` inside the container (per its
Dockerfile's `CMD ["/openpanel"]`), not on the host filesystem — use
`podman cp` / `podman exec`, not `scp` directly into a host path.

```bash
export SSHPASS='<password>'
sshpass -e ssh root@<server> '
set -e
podman exec openpanel cp /openpanel /openpanel.bak-$(date +%Y%m%d%H%M%S)
podman cp /root/openpanel-amd64.new openpanel:/openpanel
podman exec openpanel chmod +x /openpanel
podman restart openpanel
sleep 4
podman ps -a --filter "name=openpanel" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
podman logs --tail 15 openpanel 2>&1
'
```

Confirm the `openpanel` row shows `Up ...` (not `Exited`), and the log
tail ends with a clean startup (`BOOTSTRAP - listening on :2083 (tls=true)`
and no `missing dynamic library` or panic output). If it's `Exited`,
`podman logs openpanel` will show why — the previous binary is at
`/openpanel.bak-<timestamp>` inside the container and can be copied back
over `/openpanel` followed by another `podman restart openpanel` to roll
back.

## Step 4 — Smoke-test it live (log in and check it actually works)

Prefer an actual browser check if a browser-automation tool/skill (e.g.
`claude-in-chrome`) is available in this environment: load that skill
first, open `https://<server>:2083/login` (or whatever port was confirmed
in Step 0), log in with the user credentials from Step 0, navigate to
whatever page is relevant to the change you just deployed, and take a
screenshot to visually confirm it renders correctly and the feature works.

If no browser tool is available, fall back to this curl-based login flow
(same gorilla/csrf setup as OpenAdmin — login form fields are `username`
and `password`, field name `csrf_token`, and the app requires a matching
Referer header on HTTPS POSTs):

```bash
cd <scratchpad>
BASE="https://<server>:<port from step 0>"

login_page=$(curl -sk -c cookies.txt -e "$BASE/login" "$BASE/login")
csrf=$(echo "$login_page" | grep -oP 'name="csrf_token" value="\K[^"]+' | head -1 \
  | python3 -c "import sys,html; print(html.unescape(sys.stdin.read().strip()))")

curl -sk -b cookies.txt -c cookies.txt -e "$BASE/login" -o login_resp.html \
  -w "login POST status: %{http_code}\n" \
  -X POST "$BASE/login" \
  --data-urlencode "username=<username>" \
  --data-urlencode "password=<password>" \
  --data-urlencode "csrf_token=$csrf"
# Success redirects to /dashboard (302/303).

# Fetch a page relevant to what changed and check it looks right, e.g.
# for cron-related changes:
curl -sk -b cookies.txt "$BASE/cronjobs" | grep -o "No overlap"
curl -sk -b cookies.txt "$BASE/cronjobs/new" | grep -o "no_overlap"
```

Pick the page/grep target based on whatever was actually changed this
session.

## Step 5 — Report back

Summarize concisely: build/test result, deploy result (container up or
not, clean logs or not), and the smoke-test result (what you actually saw
on the live page). If anything failed, say exactly what and at which step
— don't declare success unless the live page genuinely showed the expected
content.
