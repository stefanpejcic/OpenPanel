---
sidebar_label: "OpenPanel Installation on Contabo"
description: "How to install the OpenPanel hosting control panel on a Contabo VPS or VDS: choosing a plan and OS, running the installer, reverse DNS and email tips."
---

# How to Install a Hosting Control Panel on Contabo

This guide shows how to install **OpenPanel** on a [Contabo](https://contabo.com/) Cloud VPS or VDS.

Contabo servers come with a lot of RAM and disk for the price, which makes them a good fit for hosting many small websites with OpenPanel.

---

## Prerequisites

- A Contabo **Cloud VPS** or **Cloud VDS** (KVM virtualization - supported by OpenPanel).
- The root password or SSH key you set when ordering.
- Optional: a domain name for the panel, e.g. `server.example.com`.

---

## Step 1: Order or Reinstall the Server

When ordering a new server:

1. Pick a **region** close to your visitors.
2. Under **Image**, choose **Ubuntu 24.04** (recommended). Debian, AlmaLinux and Rocky Linux also work.
3. Set a strong **root password** and, if you have one, add your **SSH key**.

If you already have a Contabo server with another OS or panel on it, reinstall it first - OpenPanel requires a clean server:

1. Log in to the [Contabo Customer Control Panel](https://my.contabo.com/).
2. Go to **Your services**, click **Manage** next to the server, and choose **Reinstall**.
3. Select **Ubuntu 24.04** under **Standard** images (not an image with a panel or apps preinstalled) and confirm.

:::info
Any Contabo plan meets the [minimum requirements](/docs/admin/intro/#requirements). Use the [resource calculator](/calculator) to check how many websites the plan can comfortably host.
:::

---

## Step 2: Connect via SSH

The IP address and login details are in the email Contabo sends once the server is ready.

```bash
ssh root@your-server-ip
```

![Terminal showing the first SSH connection as root: confirming the server's host key fingerprint, the Ubuntu 24.04 welcome message, and the root prompt](/img/openpanel-screenshots/install/ssh-contabo.png#screenshot)

---

## Step 3: Run the OpenPanel Installer

```bash
bash <(curl -sSL https://openpanel.org)
```

The installation takes about 5 minutes. For optional flags (domain, admin username, ports, swap size), use the [install command generator](/install).

When it finishes, the installer prints the **OpenAdmin URL**, **username** and **password**.

![OpenPanel installer output in a terminal: the OpenPanel banner with version, OS, IP and Podman engine, each installation step marked OK, and the OpenAdmin URL with the generated username and password at the end](/img/openpanel-screenshots/install/installer-output-generic.png#screenshot)

---

## Step 4: Log in to OpenAdmin

Open the URL from the installer output in your browser:

```
https://your-server-ip:2087
```

Log in with the printed credentials and continue with the [post-install steps](/docs/admin/intro/#post-install-steps).

![OpenAdmin login form with the username and password fields, Remember me, passkey sign-in and the Switch to OpenPanel button](/img/openadmin-screenshots/000_intro-login.png#gh-light-mode-only)
![OpenAdmin login form with the username and password fields, Remember me, passkey sign-in and the Switch to OpenPanel button](/img/openadmin-screenshots/000_intro-login_dark.png#gh-dark-mode-only)

---

## Contabo-Specific Settings

### Reverse DNS (PTR)

For email to be delivered reliably, the server IP needs a reverse DNS record that matches the server hostname:

1. In the Contabo Customer Control Panel, open **Reverse DNS management**.
2. Find your server IP, click the edit icon, and enter the hostname, e.g. `server.example.com`.

See [How to Set Up Reverse DNS (PTR)](/docs/articles/email/reverse-dns-ptr-record/) for how to check it.

### Sending Email

Contabo IP ranges are often listed on email blocklists, so mail sent directly from the server can land in spam even with SPF, DKIM and DMARC set up. If that happens:

- Check the IP on a blocklist checker and request delisting.
- Or route outgoing email through an SMTP relay - see [Port 25 Blocked: How to Use an SMTP Relay](/docs/articles/email/port-25-blocked-smtp-relay/).

### Firewall

OpenPanel installs [Sentinel Firewall (CSF)](/docs/admin/security/firewall/) automatically. If you also use a firewall in the Contabo panel, allow at least ports `22`, `53`, `80`, `443`, `2083` and `2087` inbound - the full list is in the [requirements](/docs/admin/intro/#requirements).

---

## Next Steps

- [Set a domain and SSL for the panel](/docs/admin/settings/general/#domain)
- [Configure custom nameservers](/docs/articles/domains/how-to-configure-nameservers-in-openpanel/)
- [Create hosting plans](/docs/admin/plans/hosting_plans/)
- [Secure the server for production](/docs/articles/security/securing-openpanel/)
