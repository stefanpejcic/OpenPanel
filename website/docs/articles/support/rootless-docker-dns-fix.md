---
sidebar_label: "Fix Outbound DNS/Internet in User Containers"
---

# How to Fix No Internet / DNS Access in User Containers

## Symptom

Containers of an OpenPanel user have no outbound internet access, and DNS resolution fails for external hostnames, even though:

- The host server has full internet access.
- System containers (for example Caddy) resolve DNS correctly.
- The issue only affects the per-user containers.

**Example errors** (run as root on the server):

```bash
source /usr/local/opencli/lib/podman.sh

podman_user <user> exec apache wget -q --spider --timeout=5 https://example.com/feed.xml
# wget: bad address 'example.com'

podman_user <user> exec apache nslookup example.com 8.8.8.8
# ;; connection timed out; no servers could be reached
```

**Real-world impact:** In applications that fetch remote content (e.g. RSS/feed modules in a CMS, plugin updates, payment APIs), each unresolved host adds a DNS timeout to the page load. A normal page load of a couple of seconds can turn into more than a minute.

---

## Fix (Podman)

OpenPanel runs each user's containers with rootless **Podman**. The DNS servers for all of a user's containers are set in the user's Podman configuration, not in `docker-compose.yml`.

### 1. Test from the right container

Test from the container that actually makes the outbound request - usually the PHP-FPM container, not the web server:

```bash
source /usr/local/opencli/lib/podman.sh
podman_user <user> exec <phpfpm-container> curl -sI https://example.com/
```

### 2. Check the user's DNS configuration

```bash
cat /home/<user>/.config/containers/containers.conf
```

It should contain:

```ini
[containers]
no_new_privileges = true
dns_servers = ["1.1.1.1", "1.0.0.1"]
```

If the file or the `dns_servers` line is missing (for example for an account created on an older version, or edited by hand), add it. If your network blocks public resolvers, use your provider's DNS servers instead.

### 3. Recreate the user's containers

DNS settings apply when a container is created, so recreate the running services:

```bash
source /usr/local/opencli/lib/podman.sh
cd /home/<user>
podman_compose_user <user> down
podman_compose_user <user> up -d <service> [<service> ...]
```

Only start the services the user actually uses (for example `apache php-fpm-8.3 mysql`).

### 4. Verify

```bash
podman_user <user> exec <phpfpm-container> curl -sI https://example.com/
podman_user <user> exec apache wget -q -O- https://example.com/feed.xml
```

Both should now return output instead of timing out.

### 5. Diagnostics (if the issue persists)

Collect this output before contacting support:

```bash
opencli report --public
podman_user <user> info
podman_user <user> network ls
cat /etc/containers/containers.conf.d/*.conf
```

Also check that the server itself can resolve names (`dig example.com`) and that the firewall allows outbound DNS (port `53` UDP/TCP) - see [Sentinel Firewall](/docs/admin/security/firewall/).

---

## Older Installations (Rootless Docker)

Servers installed before OpenPanel 2.0 may still run user containers with rootless **Docker**. On those servers the cause and fix are different:

### Root Cause

Rootless Docker uses `vpnkit` for networking. `vpnkit` intercepts DNS resolution before it reaches any nameservers configured inside the container or in `docker-compose.yml`. As a result:

- Adding `dns: [8.8.8.8, 1.1.1.1]` under the service in `docker-compose.yml` has **no effect**.
- Manually editing `/etc/resolv.conf` inside the container has **no effect**.

The actual fix must happen at the rootless Docker **daemon** level, not the container or compose level.

### Steps to Fix

#### 1. Test from the correct container

Make sure you're testing from the container that actually serves outbound requests for the application (e.g. a PHP-FPM container), not just the web server container, since application-level outbound requests often run through a different process.

```bash
docker --context <user> exec <phpfpm-container> curl https://example.com/
```

#### 2. Check the per-user daemon configuration

Locate the rootless daemon config for the affected user:

```bash
/home/<user>/.config/docker/daemon.json
```

Compare it against the platform's official reference configuration for rootless Docker daemons.

If fields are missing or incomplete, replace `daemon.json` with the full expected format, [view example](https://github.com/stefanpejcic/openpanel-configuration/blob/main/docker/daemon/rootless.json)


Don't forget to restart the Docker daemon for that user from step 4.

#### 3. Verify the per-user network driver

Confirm that the internal network in the user's `docker-compose.yml` (`/home/<USER>/docker-compose.yml)` uses the bridge driver:

```yaml
networks:
  internal:
    driver: bridge
```

#### 4. Restart the rootless Docker daemon for that user

```bash
machinectl shell <user>@ /bin/bash -c "systemctl --user start docker"
```

(Use `restart` instead of `start` if the daemon is already running.)

#### 5. Verify connectivity after restart

```bash
docker --context <user> exec <phpfpm-container> curl https://example.com/
docker --context <user> exec apache wget -q -O- https://example.com/feed.xml
```

Both commands should now return valid output instead of timing out.

#### 6. Diagnostic commands (if the issue persists)

Collect and review the following before escalating to platform support:

```bash
opencli report --public
docker info
docker --context=<user> network inspect <network-name>
```

### Notes

- Newly created users may already have a correct `daemon.json` by default — this issue can occur when an existing user's config predates the current reference format or was manually edited.
- If the fix doesn't resolve DNS, double-check whether `slirp4netns` is required for your specific OS/network driver combination, and share diagnostic output with platform support to help reproduce the environment.

### Summary Checklist

- [ ] Test DNS from the application/backend container, not just the web server container
- [ ] Compare the per-user `daemon.json` to the official rootless reference config
- [ ] Ensure the internal network driver is set to `bridge`
- [ ] Restart the user's rootless Docker daemon
- [ ] Re-test outbound connectivity from both the web server and backend containers
