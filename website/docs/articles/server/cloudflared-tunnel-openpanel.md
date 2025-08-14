# Cloudflare Tunnel + OpenPanel

Cloudflare Tunnel allows you to securely expose services (websites, APIs, or internal tools) to the internet **without opening firewall ports** or exposing your server’s IP address.

This guide covers configuring Cloudflare Tunnel for an OpenPanel server.

---

## 1. Add `cloudflared` service

Edit your existing `/root/docker-compose.yml` and add:

```yaml
  cloudflared:
    image: cloudflare/cloudflared:latest
    restart: unless-stopped
    command: tunnel --config /etc/cloudflared/config.yml run
    volumes:
      - ./cloudflared:/etc/cloudflared
    network_mode: host
```

Create the Cloudflared folder:

```bash
mkdir -p /root/cloudflared
```

Create `/root/cloudflared/config.yml`:

```yaml
tunnel: my-openpanel-tunnel
credentials-file: /etc/cloudflared/<TUNNEL-ID>.json

ingress:
  - hostname: site1.example.com
    service: http://localhost
  - hostname: site2.example.com
    service: http://localhost
```

---

## 2. Install & Login (one-time)

Run the following to authenticate Cloudflare:

```bash
docker run -it --rm \
  -v /root/cloudflared:/etc/cloudflared \
  cloudflare/cloudflared:latest tunnel login
```

Open the link in your browser and log in with your Cloudflare account.

---

## 3. Create a Tunnel

```bash
docker run -it --rm \
  -v /root/cloudflared:/etc/cloudflared \
  cloudflare/cloudflared:latest tunnel create my-openpanel-tunnel
```

This generates `<TUNNEL-ID>.json` in `/root/cloudflared`.

---

## 4. Update `config.yml`

* Replace `<TUNNEL-ID>` with the actual ID.
* Update the `service` URLs to match your internal site addresses.

---

## 5. Configure DNS in Cloudflare

For each site (`site1.example.com`, `site2.example.com`):

* Type: `CNAME`
* Value: `<TUNNEL-ID>.cfargotunnel.com`
* Proxy status: **Proxied** (orange cloud)

---

## 6. Start the Tunnel

```bash
cd /root && docker compose up -d cloudflared
```

---

## 7. Block Direct Access

In your firewall, block all inbound HTTP/HTTPS traffic (and `2083` and `2087` ports) **except** Cloudflare IP ranges:

* [IPv4 ranges](https://www.cloudflare.com/ips-v4)
* [IPv6 ranges](https://www.cloudflare.com/ips-v6)
