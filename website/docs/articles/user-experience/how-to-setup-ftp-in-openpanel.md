# How to setup FTP in OpenPanel

[OpenPanel Enterprise](https://openpanel.com/enterprise/) supports FTP access. Once enabled, a shared FTP server becomes available for all users.

Follow these steps to enable FTP in OpenPanel:

---

## 1. Activate the Enterprise License

Ensure that you're running the **Enterprise Edition** of OpenPanel. FTP support is only available in this version.

[Upgrading to OpenPanel Enterprise and activating License](/docs/articles/license/upgrade_to_openpanel_enterprise_and-activate_license/)

---

## 2. Enable the FTP Module

OpenPanel uses a modular system where features can be individually enabled or disabled.

To enable the FTP module:

* Go to **OpenAdmin > Settings > Modules**
* Find the **FTP** module and click **Activate**

![FTP Module Activation](https://i.postimg.cc/zGBwhJVz/2025-07-09-11-34.png)

---

## 3. Start the FTP Service

The FTP feature uses the `vsftpd` server.

To start it:

* Navigate to **OpenAdmin > Services > FTP Accounts**
* Click **Start FTP Server**

![Start FTP Server](https://i.postimg.cc/pdMTHCNw/2025-07-09-11-35.png)

---

## 4. Enable the FTP Feature in Hosting Plans

FTP access must be explicitly enabled in the hosting plan's feature set.

To do this:

* Go to **OpenAdmin > Hosting Plans > Feature Manager**
* Select the feature set you want to modify
* Enable the **ftp** feature

![Enable FTP Feature](https://i.postimg.cc/mrSbGGy9/2025-07-09-11-38.png)

---

## 5. Restart OpenPanel

Restarting OpenPanel ensures that the newly enabled module and feature sets take effect immediately for all users.

To restart:

* Go to **OpenAdmin > Services > Status**
* Click **Restart** next to **OpenPanel**

![Restart OpenPanel](https://i.postimg.cc/pd1PdJ3V/2025-07-09-11-40.png)

---

## ✅ FTP Setup Complete

FTP is now enabled for all hosting plans that include the **ftp** feature. Affected users can access it under **OpenPanel > Files > FTP Accounts**.

[![2025-07-09-11-42.png](https://i.postimg.cc/rsjn8QYQ/2025-07-09-11-42.png)](https://postimg.cc/y3JXjXKZ)

---

## ⚠️ Troubleshooting

If a user sees the following error:

> `FTP service is not yet started, please contact Administrator to enable it.`

It means the FTP service is not running. Revisit **Step 3: Start the FTP Service** to resolve this.
