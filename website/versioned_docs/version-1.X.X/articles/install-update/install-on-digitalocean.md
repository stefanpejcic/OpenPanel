# OpenPanel Installation on DigitalOcean

This guide will walk you through deploying **OpenPanel** on a DigitalOcean Droplet with Ubuntu.

---

## Prerequisites

- An active [DigitalOcean account](https://www.digitalocean.com/).
- Basic familiarity with SSH and terminal commands.

---

## Step 1: Create a Droplet

1. Log in to your **DigitalOcean account**.
2. Click **Create → Droplet**.
3. Choose an operating system:
   - Ubuntu 24.04
4. Select a **Droplet size**:
   - Recommended: **1 CPU, 2 GB RAM** (Premium AMD with NVMe SSD suggested for performance)
5. Choose **authentication**:
   - **SSH key** (recommended) or password authentication
6. Assign a **Reserved IP** for a static IP address.
7. Complete creation and note your Droplet’s IP address.

---

## Step 2: Connect to Your Droplet

Use SSH to connect:
```bash
ssh root@yourDropletIpAddress
````

> Replace `yourDropletIpAddress` with the static IP you assigned to your Droplet.

---

## Step 3: Run the OpenPanel Installer

Once logged in, run the installer script:
```bash
bash <(curl -sSL https://openpanel.org)
```

---

## Step 4: Access OpenPanel

1. Open your browser and navigate to:

```
https://yourDropletIpAddress:2087
```

2. Log in using the credentials created during installation.

---

## Notes

* Ensure your Droplet meets [OpenPanel minimum requirements](https://openpanel.com/docs/admin/intro/#requirements).
* For production, restrict firewall access to limit exposure after installation.
* Use the [OpenPanel Install Command Generator](https://openpanel.com/install) for advanced configuration options.

---

**Congratulations!** You have successfully installed OpenPanel on DigitalOcean.
