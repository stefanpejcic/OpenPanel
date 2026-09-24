---
sidebar_label: "OpenPanel Installation on a Home Server"
description: "How to self-host websites from home with OpenPanel: hardware, port forwarding, dynamic IP, CGNAT, Cloudflare Tunnel, email limitations and accessing sites from your own network."
---

# How to Self-Host Websites on a Home Server with OpenPanel

You can run **OpenPanel** on a spare PC, mini PC, NAS or small ARM board at home and host your own websites, apps and services. Installation is the same as on a VPS - the extra work is making the server reachable from the internet.

For a Raspberry Pi, see [Install OpenPanel on a Raspberry Pi](/docs/articles/install-update/install-on-raspberry-pi/). For a VM on Proxmox, see [Install OpenPanel on Proxmox VE](/docs/articles/install-update/install-on-proxmox/).

---

## What You Need

- **A dedicated machine or VM** - OpenPanel needs a clean OS with nothing else installed (no other web server, Docker or panel).
- **CPU**: any 64-bit x86 (AMD64) or ARM64 (AArch64) processor.
- **RAM**: at least **2 GB**, **4 GB or more** recommended.
- **Disk**: at least **20 GB**, an SSD is strongly recommended.
- **OS**: Ubuntu 24.04 LTS (recommended), Debian, AlmaLinux or Rocky Linux - see the [supported systems](/docs/admin/intro/#requirements).
- **Wired network connection** - more reliable than Wi-Fi for a server.

Use the [resource calculator](/calculator) to see how many websites your hardware can handle.

:::info Check your ISP's terms
Some home internet plans don't allow running servers, and some ISPs block incoming ports `80` and `443`. If yours does, use a [Cloudflare Tunnel](#option-b-cloudflare-tunnel-no-port-forwarding) instead of port forwarding.
:::

---

## Step 1: Install the Operating System

1. Install **Ubuntu Server 24.04 LTS** on the machine and enable **OpenSSH server** during setup.
2. Give the server a **fixed local IP** - either set a static IP during installation, or create a **DHCP reservation** for it in your router. Port forwarding breaks if the local IP changes.

---

## Step 2: Install OpenPanel

Connect from another computer and run the installer as root:

```bash
ssh your-user@192.168.1.50
sudo -i
bash <(curl -sSL https://openpanel.org)
```

When the installation finishes, open `https://192.168.1.50:2087` from your local network and log in with the printed credentials.

![OpenPanel installer output in a terminal: the OpenPanel banner with version, OS, IP and Podman engine, each installation step marked OK, and the OpenAdmin URL with the generated username and password at the end](/img/openpanel-screenshots/install/installer-output-generic.png#screenshot)

---

## Step 3: Make Websites Reachable from the Internet

### Option A: Port Forwarding

In your router's admin page (often under **NAT**, **Port Forwarding** or **Virtual Servers**), forward these ports to the server's local IP:

| Port | Needed for |
|---|---|
| `80`, `443` TCP (+ `443` UDP) | Websites and SSL certificates - **required** |
| `2083` TCP | OpenPanel user panel |
| `2087` TCP | OpenAdmin - better to keep it local-only or restrict it by IP |
| `53` TCP/UDP | Only if you use OpenPanel as the DNS server for your domains |

Then point your domain's `A` record to your **public** IP (the one shown at [ip.openpanel.com](https://ip.openpanel.com)). OpenPanel detects the public IP automatically and uses it in new DNS zones.

:::tip Use an external DNS provider at home
At home it is usually easier to keep DNS at your registrar or Cloudflare instead of running OpenPanel as your nameserver - you don't need to forward port 53, and DNS keeps working when your server is offline.
:::

#### CGNAT check

If your router's WAN IP is different from the IP shown on [ip.openpanel.com](https://ip.openpanel.com), or starts with `100.64`-`100.127`, your ISP uses **CGNAT** and port forwarding will not work. Ask your ISP for a public IP, or use option B.

### Option B: Cloudflare Tunnel (No Port Forwarding)

A Cloudflare Tunnel makes an outgoing connection from your server to Cloudflare, so you don't need a public IP or open ports. It works behind CGNAT and on networks where the ISP blocks incoming traffic.

Follow [Host Websites Behind a Cloudflare Tunnel](/docs/articles/server/cloudflared-tunnel-openpanel/).

---

## Step 4: Handle a Changing Public IP

Most home connections get a **dynamic** public IP that changes from time to time. When it changes, your domains point to the old IP and the websites go offline.

Options:

- **Ask your ISP for a static IP** - the simplest fix, often available for a small fee.
- **Use a Cloudflare Tunnel** - the IP doesn't matter at all.
- **Keep DNS at Cloudflare or another provider with an API**, and run a dynamic DNS client (such as `ddclient`) on the server to update the `A` records when the IP changes.

If you host DNS on another OpenPanel server, its [Dynamic DNS](/docs/panel/domains/dynamic-dns/) feature gives you an update URL you can call from a cron job at home to keep a record in sync.

---

## Step 5: Accessing Your Sites from Inside Your Network

Many home routers don't support **NAT loopback (hairpin NAT)**, so opening `https://yourdomain.com` from a device on the same network may time out even though it works from outside.

Fixes:

- Enable **NAT loopback / NAT reflection** in the router, if available.
- Or add a local DNS override (in your router, Pi-hole, or the `hosts` file of your computer) that points your domains to the server's local IP, e.g. `192.168.1.50 yourdomain.com`.

---

## Email from a Home Server

Sending email directly from a home connection almost never works:

- Most ISPs **block outgoing port 25**.
- Home IP ranges are listed on spam blocklists, and you can't set **reverse DNS (PTR)** for them.

Websites can still send email through an external SMTP relay - see [Port 25 Blocked: How to Use an SMTP Relay](/docs/articles/email/port-25-blocked-smtp-relay/).

---

## Keeping It Running

- **Power**: a small **UPS** protects the server and disks from power cuts.
- **Backups**: configure [OpenPanel backups](/docs/articles/backups/configuring-backups/) to an external drive or remote storage - a home server has no provider snapshots.
- **Security**: keep OpenAdmin (`2087`) off the internet or [restrict it by IP](/docs/articles/dev-experience/limit_access_to_openadmin/), and follow the [server hardening checklist](/docs/articles/security/securing-openpanel/).
