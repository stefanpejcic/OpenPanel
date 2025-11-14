---
title: Using OpenPanel for File Storage Hosting
description: A step-by-step guide to setting up file storage hosting plans with OpenPanel Enterprise
slug: storage-file-hosting-with-openpanel
authors: stefanpejcic
tags: [OpenPanel, FTP]
image: https://openpanel.com/img/blog/file-storage-hosting-with-openpanel.png
hide_table_of_contents: true
---

OpenPanel makes it easy to create hosting plans dedicated purely to file storage. With these plans, users have access only to **File Manager** and **FTP**, allowing them to manage files remotely—for example, for backups—without full website hosting access.

<!--truncate-->

![preview](https://i.postimg.cc/63dZ7RJt/image.png)

---

## Setting Up Storage-Only Hosting Plans

OpenPanel allows you to define feature sets per plan, so you can quickly create **files-only hosting plans**. The setup involves three main steps: enabling modules, creating a feature set, and assigning it to hosting plans.

---

### Step 1: Enable Required Modules

Navigate to **OpenAdmin > Settings > Modules** and enable the modules necessary for file storage hosting:

* **File Manager** – Manage files and backups
* **FTP** – Create and manage FTP accounts

![screenshot](https://i.postimg.cc/Tf0BnRL5/enable-modules.png)

---

### Step 2: Create a Feature Set

Next, define a feature set for your files-only plans:

1. Go to **OpenAdmin > Hosting Plans > Feature Manager**
2. Under *Add a new feature list*, enter `files_only` and click **Create**
![newPlan](https://i.postimg.cc/1tNkDmJm/image.png)
3. Select your newly created feature set under *Manage feature set*
![ManageFeature](https://i.postimg.cc/fLwd9ywX/image.png)
4. Enable the relevant features: **File Manager**, **FTP**, and any optional features you want, Save the changes
![EnablingFeatures](https://i.postimg.cc/8CCJp3wM/image.png)

---

### Step 3: Create Hosting Plans

Once the feature set is ready, you can create your storage-only hosting plans:

1. Go to **OpenAdmin > Hosting Plans > User Packages** and click **Create New**
2. Configure plan limits and select `files_only` under *Feature Set*
3. Click **Create Plan**

Example plans:

* **File Hosting**
* **FTP Hosting**

![NewPlan](https://i.postimg.cc/KYdM0fr5/New-Plan.png)

---

### Step 4: Create User Accounts

After creating the plans, you can add users manually via OpenAdmin, through the terminal, or using billing systems such as **WHMCS** or **FOSSBilling**.

![screenshot](/img/admin/2025-06-09_08-20.png)

---

## Key Features of File Storage Plans

Users on storage-only plans gain access to the following:

### File Management

A feature-rich **File Manager** allows users to upload, edit, and organize files with ease.

![FileManager](https://i.postimg.cc/Y2mhXQhf/file-Manager.png)

---

### Remote FTP Access

Users can create FTP sub-accounts, restrict access to specific directories, and manage files remotely. This is ideal for backups or for accessing files via popular FTP clients like **FileZilla** or **CyberDuck**.

![FTPAccount](https://i.postimg.cc/QN5cF2gy/image.png)

Check out our [FTP setup guide for admins](https://openpanel.com/docs/articles/user-experience/how-to-setup-ftp-in-openpanel/) to find more detailed info  .

---

## Why Choose OpenPanel for File Hosting?

OpenPanel’s feature sets make it simple to offer storage-only plans, giving users access only to **File Manager** and **FTP** without unnecessary tools. It’s an efficient, secure, and scalable solution for file hosting needs.

Ready to get started? [Start Here](https://openpanel.com/enterprise/)
