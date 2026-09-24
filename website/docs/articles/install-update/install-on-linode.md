---
sidebar_label: "OpenPanel Installation on Linode (Akamai)"
description: "How to install the OpenPanel hosting control panel on a Linode (Akamai Cloud) instance: creating the Linode, running the installer, Cloud Firewall, SMTP restrictions and reverse DNS."
---

# How to Install a Hosting Control Panel on Linode (Akamai Cloud)

This guide shows how to install **OpenPanel** on a [Linode](https://www.linode.com/) instance, now part of **Akamai Cloud**.

---

## Prerequisites

- A [Linode / Akamai Cloud](https://cloud.linode.com/) account.
- An SSH key (recommended) or a strong root password.
- Optional: a domain name for the panel, e.g. `server.example.com`.

---

## Step 1: Create a Linode

1. Log in to the **Cloud Manager** and click **Create** → **Linode**.
2. Under **Choose a Distribution**, select **Ubuntu 24.04 LTS** (recommended). Debian, AlmaLinux and Rocky Linux also work.
3. Choose a **Region** close to your visitors.
4. Choose a **Linode Plan**:
   - **Shared CPU** plans are fine for most hosting servers.
   - At least **2 GB RAM**; **4 GB or more** is recommended for production. Use the [resource calculator](/calculator) to size the plan.
5. Set a **Linode Label**, **Root Password**, and add your **SSH key**.
6. Click **Create Linode** and wait for the status to change to **Running**.

:::tip
You can paste the install command into **Add User Data** (Metadata / cloud-init) to install OpenPanel automatically on first boot - see [Install OpenPanel with Cloud-Init](/docs/articles/install-update/install-using-cloudinit/).
:::

---

## Step 2: Connect via SSH

```bash
ssh root@your-linode-ip
```

The IPv4 address is shown on the Linode's summary page.

![Terminal showing the first SSH connection as root: confirming the server's host key fingerprint, the Ubuntu 24.04 welcome message, and the root prompt](/img/openpanel-screenshots/install/ssh-linode.png#screenshot)

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

```
https://your-linode-ip:2087
```

Log in with the printed credentials and continue with the [post-install steps](/docs/admin/intro/#post-install-steps).

![OpenAdmin login form with the username and password fields, Remember me, passkey sign-in and the Switch to OpenPanel button](/img/openadmin-screenshots/000_intro-login.png#gh-light-mode-only)
![OpenAdmin login form with the username and password fields, Remember me, passkey sign-in and the Switch to OpenPanel button](/img/openadmin-screenshots/000_intro-login_dark.png#gh-dark-mode-only)

---

## Linode-Specific Settings

### Cloud Firewall

OpenPanel installs [Sentinel Firewall (CSF)](/docs/admin/security/firewall/) automatically, so a Linode **Cloud Firewall** is optional. If you attach one, set the default inbound policy to **Drop** and allow at least ports `22`, `53` (TCP and UDP), `80`, `443`, `2083` and `2087`. The full list, including email and FTP ports, is in the [requirements](/docs/admin/intro/#requirements).

### SMTP Restrictions (Port 25)

New Linode accounts have outgoing **SMTP ports 25, 465 and 587 restricted** to prevent abuse. Websites and the panel work, but email sent from the server will not be delivered.

To send email:

- Open a **support ticket** asking Linode to lift the SMTP restriction, and describe your use case (set up reverse DNS first - they usually ask for it).
- Or send email through an external relay - see [Port 25 Blocked: How to Use an SMTP Relay](/docs/articles/email/port-25-blocked-smtp-relay/).

### Reverse DNS (PTR)

1. Open the Linode and go to the **Network** tab.
2. In the **IP Addresses** section, click **...** next to the IPv4 address and choose **Edit RDNS**.
3. Enter the server hostname, e.g. `server.example.com`, and save.

The hostname must already resolve to the Linode's IP. See [How to Set Up Reverse DNS (PTR)](/docs/articles/email/reverse-dns-ptr-record/).

---

## Next Steps

- [Set a domain and SSL for the panel](/docs/admin/settings/general/#domain)
- [Configure custom nameservers](/docs/articles/domains/how-to-configure-nameservers-in-openpanel/)
- [Create hosting plans](/docs/admin/plans/hosting_plans/)
- [Secure the server for production](/docs/articles/security/securing-openpanel/)
