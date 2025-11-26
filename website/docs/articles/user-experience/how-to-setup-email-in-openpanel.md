# Setup Email

[OpenPanel Enterprise](https://openpanel.com/enterprise/) supports Emails. Once enabled, a shared Email server becomes available for all users.

Follow these steps to enable Emails in OpenPanel:

---

## 1. Activate the Enterprise License

Ensure that you're running the **Enterprise Edition** of OpenPanel. Email support is only available in this version.

[Upgrading to OpenPanel Enterprise and activating License](/docs/articles/license/upgrade_to_openpanel_enterprise_and-activate_license/)

---

## 2. Enable the Email Module

OpenPanel uses a modular system where features can be individually enabled or disabled.

To enable the Email module:

* Go to **OpenAdmin > Settings > Modules**
* Find the **Emails** module and click **Activate**

![Email Module Activation](https://i.postimg.cc/9FvhmLSX/2025-07-09-17-12.png)

---

## 3. Install the Email Service

The Email feature uses the [`docker-mailserver`](https://docker-mailserver.github.io/docker-mailserver/latest/) stack.

To start it:

* Navigate to **OpenAdmin > Emails > Email Accounts**
* Copy the command displayed on the page `opencli email-server install`
* Paste it on Terminal and wait for the process to finish.

[![2025-07-09-17-14.png](https://i.postimg.cc/YCzMd1np/2025-07-09-17-14.png)](https://postimg.cc/G4tW2snN)

---

## 4. Enable the Emails Feature in Hosting Plans

Emails access must be explicitly enabled in the hosting plan's feature set.

To do this:

* Go to **OpenAdmin > Hosting Plans > Feature Manager**
* Select the feature set you want to modify
* Enable the **emails** feature

[![2025-07-09-17-17.png](https://i.postimg.cc/NjfzwmGG/2025-07-09-17-17.png)](https://postimg.cc/zV6jCLR4)

---

## 5. Restart OpenPanel

Restarting OpenPanel ensures that the newly enabled module and feature sets take effect immediately for all users.

To restart:

* Go to **OpenAdmin > Services > Status**
* Click **Restart** next to **OpenPanel**

![Restart OpenPanel](https://i.postimg.cc/pd1PdJ3V/2025-07-09-11-40.png)

---

## Done

Email is now enabled for all hosting plans that include the **emails** feature. Affected users can access Webmail an Email accounts form access it under **OpenPanel > Emails**.

[![2025-07-09-17-19.png](https://i.postimg.cc/44QPNySY/2025-07-09-17-19.png)](https://postimg.cc/144wvmPS)

---

## Troubleshooting

If the mail server fails to start, try launching it manually from the terminal:

```bash
cd /usr/local/mail/openmail && docker compose up mailserver
```

Review the startup logs for any errors, resolve them, and once the service is running correctly, stop it and restart it in detached mode using the `-d` flag.

The most common reasons the mail server fails to start include:

* **Another service already using port 25**—for example, Exim.
  **Solution:** Disable or stop the conflicting service.

* **Firewall or iptables error** that block docker starting.
  **Solution:** Restart the Docker service to reset network rules.

* **Invalid Configuration** for mailserver.
  **Solution:** Revert the default mailserver .env file.
