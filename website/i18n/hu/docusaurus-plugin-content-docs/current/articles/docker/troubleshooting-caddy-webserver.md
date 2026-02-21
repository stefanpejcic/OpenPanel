# Troubleshooting Caddy Webserver

OpenPanel uses Caddy webserver to handle SSL and reverse proxy to per-user webservers.

If Caddy is disabled, all websites hosted on the server are down.

To troubleshoot Caddy-related issues:

## 1. Check if Caddy is running

Run:

```bash
curl -I localhost
```

You should see `Server: Caddy` in the response, for example:

```
root@demo:~# curl -I localhost
HTTP/1.1 308 Permanent Redirect
Connection: close
Location: https://localhost/
Server: Caddy
Date: Thu, 23 Oct 2025 13:11:03 GMT
```

If it appears, **Caddy is running properly**. If a specific website or domain isn’t working, troubleshoot that domain individually.

---

## 2. Check Caddy status

Check if Caddy is 'active' or 'running':

```bash
docker ps -a
```

If the status is `exited`, `restarting`, or `starting`, it indicates an issue with the container.

---

## 3. Validate the configuration

If Caddy has a syntax error in the main configuration file or any domain configuration, it may exit on startup and restart repeatedly.

To validate the configuration:

```bash
docker exec caddy caddy validate --config /etc/caddy/Caddyfile
```

* `warn` or `info` messages are fine.
* Any `error` messages must be resolved before Caddy will start correctly.

Main configuration file for Caddy is: `/etc/openpanel/caddy/Caddyfile` and per-domain files are stored in: `/etc/openpanel/caddy/domains/*.conf`

---

## 4. Increase resource limits

By default, Caddy is limited to 1 CPU core and 1GB RAM. To increase the limits:

1. Edit the `.env` file:

```bash
nano /root/.env
```

2. Adjust the values:

```
CADDY_CPU="1.0"
CADDY_RAM="1.0G"
```

3. Save the file and rebuild the container:

```bash
cd /root
docker compose down caddy
docker compose up -d caddy
```

---

## 5. View logs

If the above steps don’t resolve the issue, check the logs:

```bash
docker logs -f caddy
```

Copy any errors and search online or post them on [our Discord channel](https://discord.openpanel.com/) or [OpenPanel forums](https://community.openpanel.org/t/openadmin) for help.
