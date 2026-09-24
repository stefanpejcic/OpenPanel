---
sidebar_label: "OpenPanel Installation on Hetzner"
description: "Step-by-step guide to installing the OpenPanel hosting control panel on a Hetzner Cloud or dedicated server, including firewall, reverse DNS and port 25 settings."
---

# How to Install a Hosting Control Panel on Hetzner

This guide walks you through installing **OpenPanel** on a [Hetzner](https://www.hetzner.com/) Cloud server. The same steps work for Hetzner dedicated (Robot) servers.

Hetzner is a popular choice for OpenPanel: servers are inexpensive, have fast NVMe storage, and both x86 and ARM64 (Ampere) servers are supported.

---

## Prerequisites

- A [Hetzner Cloud](https://console.hetzner.cloud/) account.
- An SSH key (recommended) or root password.
- Optional: a domain name for the panel, e.g. `server.example.com`.

---

## Step 1: Create a Server

1. Log in to the **Hetzner Cloud Console** and open (or create) a project.
2. Click **Add Server**.
3. Choose a **location** close to your visitors.
4. Select the **image**: **Ubuntu 24.04** (recommended). Debian 12/13, AlmaLinux and Rocky Linux also work.
5. Choose a **type**:
   - **x86 (Intel/AMD)** or **Arm64 (Ampere)** - OpenPanel runs on both.
   - At least **2 GB RAM**; **4 GB RAM or more** is recommended for production. Use the [resource calculator](/calculator) to size the server for the number of websites you plan to host.
6. Make sure **Public IPv4** is enabled - OpenPanel requires an IPv4 address.
7. Add your **SSH key**.
8. Click **Create & Buy now**.

:::tip
Leave **Cloud config** empty, or paste the install command there to install OpenPanel automatically on first boot - see [Install OpenPanel with Cloud-Init](/docs/articles/install-update/install-using-cloudinit/).
:::

---

## Step 2: Connect via SSH

```bash
ssh root@your-server-ip
```

Replace `your-server-ip` with the IPv4 address shown in the Cloud Console.

![Terminal showing the first SSH connection as root: confirming the server's host key fingerprint, the Ubuntu 24.04 welcome message, and the root prompt](/img/openpanel-screenshots/install/ssh-hetzner.png#screenshot)

---

## Step 3: Run the OpenPanel Installer

```bash
bash <(curl -sSL https://openpanel.org)
```

The installer takes about 5 minutes. To set a domain, admin username, or other options up front, use the [install command generator](/install).

When it finishes, the installer prints the **OpenAdmin URL**, **username**, and **password**.

![OpenPanel installer output in a terminal: the OpenPanel banner with version, OS, IP and Podman engine, each installation step marked OK including the Hetzner DNS resolvers fix, and the OpenAdmin URL with the generated username and password at the end](/img/openpanel-screenshots/install/installer-output.png#screenshot)

---

## Step 4: Log in to OpenAdmin

Open the URL printed by the installer in your browser:

```
https://your-server-ip:2087
```

Log in with the credentials from the previous step. Then follow the [post-install steps](/docs/admin/intro/#post-install-steps).

![OpenAdmin login form with the username and password fields, Remember me, passkey sign-in and the Switch to OpenPanel button](/img/openadmin-screenshots/000_intro-login.png#gh-light-mode-only)
![OpenAdmin login form with the username and password fields, Remember me, passkey sign-in and the Switch to OpenPanel button](/img/openadmin-screenshots/000_intro-login_dark.png#gh-dark-mode-only)

---

## Hetzner-Specific Settings

### DNS Resolvers

Hetzner servers use Hetzner's own DNS resolvers by default, which can cause problems pulling container images (for example [issue #471](https://github.com/stefanpejcic/OpenPanel/issues/471)). We recommend using Cloudflare's public resolvers instead.

The OpenPanel installer does this automatically when it detects a Hetzner image (`/etc/hetzner-build`) - you'll see `Hetzner detected — adding Cloudflare DNS resolvers...` in the install output.

If your server wasn't detected (for example a custom image or a dedicated Robot server), or image downloads fail during installation, set the resolvers manually:

```bash
mv /etc/resolv.conf /etc/resolv.conf.bak
printf "nameserver 1.1.1.1\nnameserver 1.0.0.1\n" > /etc/resolv.conf
```

Google's resolvers (`8.8.8.8` and `8.8.4.4`) work too. To undo the change, restore the backup: `mv /etc/resolv.conf.bak /etc/resolv.conf`.

### Hetzner Cloud Firewall

OpenPanel installs its own firewall ([Sentinel Firewall / CSF](/docs/admin/security/firewall/)), so a Hetzner Cloud Firewall is optional. If you attach one, allow these **inbound** ports:

| Service | Ports |
|---|---|
| Websites | `80`, `443` (TCP and UDP for HTTP/3) |
| OpenPanel / OpenAdmin | `2083`, `2087` |
| DNS | `53` (TCP and UDP) |
| Email (Enterprise) | `25`, `143`, `465`, `587`, `993` |
| FTP (Enterprise) | `21`, `21000-21010` |
| phpMyAdmin | `2053`, `8888` |
| SSH | `22` |
| User services | `32768-60999` |

### Outgoing Email (Port 25)

Hetzner blocks outgoing traffic on ports **25** and **465** for new Cloud accounts to prevent spam. Websites and the panel work normally, but the server cannot deliver email to other mail servers until the block is lifted.

- Once your account has been active for a while and has a paid invoice, you can request unblocking from the Cloud Console (**Limits** / support request).
- Or send email through an external relay - see [Port 25 Blocked: How to Use an SMTP Relay](/docs/articles/email/port-25-blocked-smtp-relay/).

### Reverse DNS (PTR)

If the server will send email, set the reverse DNS of the IPv4 address to the server hostname:

1. In the Cloud Console, open the server and go to **Networking**.
2. Click the **...** menu next to the IPv4 address and choose **Edit Reverse DNS**.
3. Enter the server hostname, e.g. `server.example.com`.

More details: [How to Set Up Reverse DNS (PTR)](/docs/articles/email/reverse-dns-ptr-record/).

---

## Next Steps

- [Set a domain and SSL for the panel](/docs/admin/settings/general/#domain)
- [Configure custom nameservers](/docs/articles/domains/how-to-configure-nameservers-in-openpanel/)
- [Create hosting plans](/docs/admin/plans/hosting_plans/)
- [Secure the server for production](/docs/articles/security/securing-openpanel/)
