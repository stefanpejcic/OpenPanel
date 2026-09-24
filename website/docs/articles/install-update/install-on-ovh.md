---
sidebar_label: "OpenPanel Installation on OVHcloud"
description: "How to install the OpenPanel hosting control panel on an OVHcloud VPS or dedicated server, including the default ubuntu user, reverse DNS, the Edge Network Firewall and email."
---

# How to Install a Hosting Control Panel on OVHcloud

This guide shows how to install **OpenPanel** on an [OVHcloud](https://www.ovhcloud.com/) VPS or dedicated server (including Kimsufi and So you Start servers).

---

## Prerequisites

- An OVHcloud **VPS**, **Public Cloud instance**, or **dedicated server**.
- An SSH key (recommended) or the password OVHcloud emails you after installation.
- Optional: a domain name for the panel, e.g. `server.example.com`.

---

## Step 1: Install a Clean Operating System

OpenPanel needs a clean server with no other control panel or web server installed.

1. Log in to the [OVHcloud Control Panel](https://www.ovh.com/manager/).
2. Open **Bare Metal Cloud** and select your **VPS** or **dedicated server**.
3. Click **Reinstall my VPS** (or **Install** for dedicated servers).
4. Choose **Ubuntu 24.04** (recommended). Debian 12/13, AlmaLinux and Rocky Linux also work.
   - Pick the plain **Distribution only** image, not one with a panel or application preinstalled.
5. Add your SSH key and confirm.

The server is reinstalled in a few minutes and OVHcloud emails you the login details.

:::info
Every current OVHcloud VPS meets the [minimum requirements](/docs/admin/intro/#requirements). Use the [resource calculator](/calculator) to see how many websites a plan can host.
:::

---

## Step 2: Connect via SSH and Switch to Root

OVHcloud images log you in as a regular user (usually `ubuntu` on Ubuntu or `debian` on Debian), not `root`:

```bash
ssh ubuntu@your-server-ip
```

Then switch to the root user - the installer must run as root:

```bash
sudo -i
```

![Terminal showing the first SSH connection as ubuntu and switching to root with sudo -i: confirming the server's host key fingerprint, the Ubuntu 24.04 welcome message, and the root prompt](/img/openpanel-screenshots/install/ssh-ovh.png#screenshot)

---

## Step 3: Run the OpenPanel Installer

```bash
bash <(curl -sSL https://openpanel.org)
```

Installation takes about 5 minutes. For optional flags (domain, admin username, ports), use the [install command generator](/install).

When it finishes, the installer prints the **OpenAdmin URL**, **username** and **password**.

![OpenPanel installer output in a terminal: the OpenPanel banner with version, OS, IP and Podman engine, each installation step marked OK, and the OpenAdmin URL with the generated username and password at the end](/img/openpanel-screenshots/install/installer-output-generic.png#screenshot)

---

## Step 4: Log in to OpenAdmin

Open the printed URL in your browser:

```
https://your-server-ip:2087
```

Log in with the printed credentials and continue with the [post-install steps](/docs/admin/intro/#post-install-steps).

![OpenAdmin login form with the username and password fields, Remember me, passkey sign-in and the Switch to OpenPanel button](/img/openadmin-screenshots/000_intro-login.png#gh-light-mode-only)
![OpenAdmin login form with the username and password fields, Remember me, passkey sign-in and the Switch to OpenPanel button](/img/openadmin-screenshots/000_intro-login_dark.png#gh-dark-mode-only)

---

## OVHcloud-Specific Settings

### Edge Network Firewall

OpenPanel installs [Sentinel Firewall (CSF)](/docs/admin/security/firewall/) on the server. The OVHcloud **Edge Network Firewall** is disabled by default - if you enable it on the IP, allow at least ports `22`, `53`, `80`, `443`, `2083` and `2087` inbound. The full list is in the [requirements](/docs/admin/intro/#requirements).

### Reverse DNS (PTR)

To send email reliably, set the reverse DNS of the server IP to the server hostname:

1. In the OVHcloud Control Panel, go to **Bare Metal Cloud** → **Network** → **IP**.
2. Click the **...** button next to your IP and choose **Modify the reverse path**.
3. Enter the hostname, e.g. `server.example.com`.

OVHcloud checks that the hostname resolves to the IP, so create the `A` record first. See [How to Set Up Reverse DNS (PTR)](/docs/articles/email/reverse-dns-ptr-record/).

### Sending Email

OVHcloud does not block port 25 by default, but its anti-spam system automatically blocks it if the server sends spam - for example from a hacked website. If outgoing mail suddenly stops, check for an anti-spam notice in the Control Panel and clean up the account before unblocking.

- [Why do my emails go to spam?](/docs/articles/email/why-emails-go-to-spam/)
- [Set up DKIM](/docs/articles/email/how-to-setup-dkim-for-mailserver/)

---

## Next Steps

- [Set a domain and SSL for the panel](/docs/admin/settings/general/#domain)
- [Configure custom nameservers](/docs/articles/domains/how-to-configure-nameservers-in-openpanel/)
- [Create hosting plans](/docs/admin/plans/hosting_plans/)
- [Secure the server for production](/docs/articles/security/securing-openpanel/)
