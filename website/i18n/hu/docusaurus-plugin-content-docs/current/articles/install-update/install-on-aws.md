# Installation on AWS EC2

This guide will walk you through deploying **OpenPanel** on an AWS EC2 instance with Ubuntu or Debian.

## Prerequisites

- An active [AWS account](https://aws.amazon.com/console/).
- Basic familiarity with EC2, SSH, and terminal commands.

---

## Step 1: Launch an EC2 Instance

1. Log in to the **[AWS Management Console](https://aws.amazon.com/console/)**.
2. Navigate to **EC2** → **Instances** → **Launch instances**.
3. Select an operating system (choose one of the following):
   - Ubuntu 24.04
   - Debian 11  
   *(x86 or ARM architecture is supported)*.
4. Choose an instance type meeting the minimum requirements:
   - **1 CPU**
   - **1 GB RAM**  
   *(Recommended: `t3.small` or equivalent)*.

---

## Step 2: Configure SSH Access

1. Create or select an existing **key pair** for SSH access.
2. Download the private key (`.pem`) and store it securely.

---

## Step 3: Configure Network Settings

1. Select the desired **VPC** and **subnet**.
2. Create a **security group** allowing all TCP traffic:  
   > OpenPanel manages its own firewall, so allowing all TCP is acceptable.
3. Allocate and associate an **Elastic IP** to your instance for a fixed public IP.

---

## Step 4: Connect to Your Instance

Use SSH to connect to your EC2 instance:

```bash
ssh -i your_private_key.pem ubuntu@yourElasticIpAddress
```

> Replace `your_private_key.pem` with your key file and `yourElasticIpAddress` with the allocated Elastic IP.


---



## Step 5: Run the OpenPanel Installer

Once logged in, run the installer script:

```bash
bash <(curl -sSL https://openpanel.org)
```

---

## Step 6: Access OpenPanel

After installation:

1. Open your browser and navigate to your Elastic IP and port `2087`.

---

## Notes

* Ensure your server meets [the minimum requirements](https://openpanel.com/docs/admin/intro/#requirements) for OpenPanel.
* For production environments, consider restricting access to your security group after installation for security.
* Refer to the [OpenPanel Install Command Generator](https://openpanel.com/install) for advanced configuration.

---

**Congratulations!** You have successfully installed OpenPanel on AWS EC2.

