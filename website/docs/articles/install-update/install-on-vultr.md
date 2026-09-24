---
sidebar_label: "OpenPanel Installation on Vultr"
---

# How to Install a Hosting Control Panel on Vultr

This guide will walk you through deploying **OpenPanel** on a Vultr instance with Ubuntu.

---

## Prerequisites

- An active [Vultr account](https://www.vultr.com/).
- Basic familiarity with SSH and terminal commands.

---

## Step 1: Create a Vultr Instance

1. Log in to your **Vultr account**.
2. Click **Products → Deploy New Instance**.
3. Choose a **server location** near your users.
4. Select an **OS image**:
   - Ubuntu 24.04
5. Choose a **server size** meeting minimum requirements:
   - **1 CPU, 2 GB RAM** (recommended: `Cloud Compute` or equivalent)
6. Configure **SSH key authentication** (recommended) and add your public key.
7. Complete the instance creation.

---

## Step 2: Connect to Your Instance

Use SSH to connect:

```bash
ssh root@yourInstanceIp
```

> Replace `yourInstanceIp` with the public IP of your Vultr instance.

![Terminal showing the first SSH connection as root: confirming the server's host key fingerprint, the Ubuntu 24.04 welcome message, and the root prompt](/img/openpanel-screenshots/install/ssh-vultr.png#screenshot)

---

## Step 3: Run the OpenPanel Installer

Run the installation script:

```bash
bash <(curl -sSL https://openpanel.org)
```

The installer runs without any prompts and takes about 5 minutes. When it finishes, it prints the **OpenAdmin URL**, **username** and **password**.

![OpenPanel installer output in a terminal: the OpenPanel banner with version, OS, IP and Podman engine, each installation step marked OK, and the OpenAdmin URL with the generated username and password at the end](/img/openpanel-screenshots/install/installer-output-generic.png#screenshot)

---

## Step 4: Access OpenPanel

1. Open your browser and navigate to:

```
https://yourInstanceIp:2087
```

2. Log in using the credentials created during installation.

---

## Notes

* Ensure your Vultr instance meets [OpenPanel minimum requirements](https://openpanel.com/docs/admin/intro/#requirements).
* For production environments, consider restricting firewall access after installation.
* Use the [OpenPanel Install Command Generator](https://openpanel.com/install) for advanced configuration options.

---

**Congratulations!** You have successfully installed OpenPanel on Vultr.
