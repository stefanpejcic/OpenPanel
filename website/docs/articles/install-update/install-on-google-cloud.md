---
sidebar_label: "OpenPanel Installation on Google Cloud"
---

# How to Install a Hosting Control Panel on Google Cloud

This guide will walk you through deploying **OpenPanel** on a Google Cloud VM instance with Ubuntu.

---

## Prerequisites

- An active [Google Cloud account](https://console.cloud.google.com/).
- Basic familiarity with SSH, Compute Engine, and firewall settings.

---

## Step 1: Create a VM Instance

1. Log in to the **Google Cloud Console**.
2. Navigate to **Compute Engine → VM instances → Create Instance**.
3. Choose a **boot disk image**:
   - Ubuntu 24.04
4. Select a **machine type** meeting the minimum requirements:
   - **1 CPU, 2 GB RAM** (recommended: `e2-small` or equivalent)
5. Add a **network tag** (e.g., `OPENPANEL`) for firewall rules.
6. In **Network interfaces**, assign a **static external IP**.
7. Click **Create** to launch the VM.

---

## Step 2: Configure Firewall Rules

1. Navigate to **VPC Network → Firewall rules → Create Firewall Rule**.
2. Configure the rule:
   - **Targets:** Specified target tags → `OPENPANEL`
   - **Source IP ranges:** `0.0.0.0/0`
   - **Protocols and ports:** `tcp:2087` (admin panel port)
3. Save the firewall rule.

> OpenPanel manages its own internal firewall; the above rule allows external access to the admin panel.

---

## Step 3: Connect to Your VM

Use SSH to connect:

```bash
ssh username@yourStaticIpAddress
```

> Replace `username` with your VM username (e.g., `ubuntu` or `debian`) and `yourStaticIpAddress` with your static external IP.

The installer must run as **root**, so switch to the root user:

```bash
sudo -i
```

![Terminal showing the first SSH connection as username and switching to root with sudo -i: confirming the server's host key fingerprint, the Ubuntu 24.04 welcome message, and the root prompt](/img/openpanel-screenshots/install/ssh-google-cloud.png#screenshot)

---

## Step 4: Run the OpenPanel Installer

Once connected, run the installer script:

```bash
bash <(curl -sSL https://openpanel.org)
```

The installer runs without any prompts and takes about 5 minutes. When it finishes, it prints the **OpenAdmin URL**, **username** and **password**.

![OpenPanel installer output in a terminal: the OpenPanel banner with version, OS, IP and Podman engine, each installation step marked OK, and the OpenAdmin URL with the generated username and password at the end](/img/openpanel-screenshots/install/installer-output-generic.png#screenshot)

---

## Step 5: Access OpenPanel

1. Open your browser and navigate to:

```
https://yourStaticIpAddress:2087
```

2. Log in using the credentials created during installation.

---

## Notes

* Ensure your VM meets [OpenPanel minimum requirements](https://openpanel.com/docs/admin/intro/#requirements).
* After installation, consider restricting firewall access for better security.
* Use the [OpenPanel Install Command Generator](https://openpanel.com/install) for advanced configuration options.

---

**Congratulations!** You have successfully installed OpenPanel on Google Cloud.
