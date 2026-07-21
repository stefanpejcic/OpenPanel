# OpenPanel Installation on Microsoft Azure

This guide will walk you through deploying **OpenPanel** on an Azure Virtual Machine with Ubuntu.

---

## Prerequisites

- An active [Microsoft Azure account](https://portal.azure.com/).
- Basic familiarity with SSH, Azure Virtual Machines, and network/firewall settings.

---

## Step 1: Create a Virtual Machine

1. Log in to the **[Azure Portal](https://portal.azure.com/)**.
2. Navigate to **Virtual Machines → Create → Virtual Machine**.
3. Select an **image**:
   - Ubuntu 24.04
4. Choose a **size** meeting the minimum requirements:
   - **1 CPU, 2 GB RAM** (recommended: `Standard_B1s` or equivalent)
5. Choose **authentication type**:
   - **SSH public key** (recommended)
   - Enter a username and upload or paste your SSH public key
6. Configure networking and firewall settings:
   - Allow all TCP traffic **or** specific ports:
     - `22` (SSH)
     - `80` (HTTP)
     - `443` (HTTPS)
     - `2083` (OpenPanel)
     - `2087` (OpenAdmin)
     - `443 UDP` (for HTTP/3 support)

---

## Step 2: Connect to Your Virtual Machine

Use SSH to connect:

```bash
ssh -i your_private_key azure@yourIpAddress
````

> Replace `your_private_key` with your private key file and `yourIpAddress` with the public IP of your VM.

---

## Step 3: Run the OpenPanel Installer

Once logged in, run the installer script:

```bash
bash <(curl -sSL https://openpanel.org)
```

> Follow the prompts to select your preferred database engine and complete installation.

---

## Step 4: Access OpenPanel

1. Open your browser and navigate to:

```
https://yourIpAddress:2087
```

2. Log in using the credentials created during installation.

---

## Notes

* Ensure your VM meets [OpenPanel minimum requirements](https://openpanel.com/docs/admin/intro/#requirements).
* After installation, consider restricting firewall access for better security.
* Use the [OpenPanel Install Command Generator](https://openpanel.com/install) for advanced configuration options.

---

**Congratulations!** You have successfully installed OpenPanel on Microsoft Azure.
