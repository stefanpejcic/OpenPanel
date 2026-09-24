---
sidebar_position: 26
---

# n8n

Install and manage [n8n](https://n8n.io/) - open-source workflow automation - in its own container on one of your domains. OpenPanel sets up the container, the reverse proxy with WebSocket support, the webhook URL and the n8n owner account, so you land directly on n8n's login screen.

Step-by-step guide: [How to Self-Host n8n on Your Own Server](/docs/articles/apps/install-n8n/)

---

## Install n8n

Navigate to **OpenPanel > Websites > Install App** and click **Install n8n**.

![Install n8n form with the application name, fixed port 5678, version, domain, owner account and resource limits](/img/openpanel-screenshots/applications/n8n_install-form.png#gh-light-mode-only)
![Install n8n form with the application name, fixed port 5678, version, domain, owner account and resource limits](/img/openpanel-screenshots/applications/n8n_install-form_dark.png#gh-dark-mode-only)

On the install page, configure:

* **Name** – The name of the application and its container as displayed in OpenPanel.
* **Port** – Fixed to `5678`, the port n8n listens on inside the container.
* **Version** – Any `n8nio/n8n` release from Docker Hub. The newest one is selected by default.
* **Domain / Subfolder** – The domain (and optional subfolder) where n8n will be available. Leave the subfolder blank to use the domain root.
* **Owner account** – First name, last name, email and password (8-64 characters) of n8n's first user. Click **Generate** for a random password.
* **Advanced options** – CPU, memory and PIDs limits for the container. The default of 1 CPU and 1 GB of memory is the recommended minimum.

Unlike the Node.js, Python, Ruby and Java installers, there is no startup file, custom command, package install or Git repository - n8n runs its own fixed entrypoint.

After you click **Start Installation**, OpenPanel:

1. Adds an `n8nio/n8n` service to your account's `docker-compose.yml`, with its own data volume for `/home/node/.n8n` (workflows, credentials and settings).
2. Sets `N8N_HOST`, `WEBHOOK_URL` and `N8N_EDITOR_BASE_URL` to the chosen domain, so webhooks and editor links use your real `https://` address.
3. Starts the container and configures the web server as a reverse proxy to it, including WebSockets for the live editor.
4. Waits for n8n to finish starting, then creates the owner account through n8n's own setup API.

Progress is shown below the form. When it finishes, you're redirected to the Site Manager.

---

## Manage n8n

n8n shows up on the **Site Manager** page (`/sites`) alongside your other websites. Click **Manage** next to it to open its page, which has four tabs.

### Overview

![Overview tab of an n8n application with its status, n8n version, CPU and memory limits, Stop and Restart actions and files](/img/openpanel-screenshots/applications/n8n-site.png#gh-light-mode-only)
![Overview tab of an n8n application with its status, n8n version, CPU and memory limits, Stop and Restart actions and files](/img/openpanel-screenshots/applications/n8n-site_dark.png#gh-dark-mode-only)

* **Status** – Whether the container is running.
* **n8n Version** – The version currently in use.
* **CPU limit** / **Memory limit** – The configured resource limits.
* **Actions** – **Start**, **Stop** and **Restart** the container. The icon in the corner opens the Containers page.
* **Files** – The domain folder and its size, with shortcuts to the Disk Usage and Inodes explorers and FTP accounts.

Further down, the same tab lets you change the application's settings:

![Container name and port, n8n version picker and CPU, memory and PIDs limits on the Overview tab](/img/openpanel-screenshots/applications/n8n-settings.png#gh-light-mode-only)
![Container name and port, n8n version picker and CPU, memory and PIDs limits on the Overview tab](/img/openpanel-screenshots/applications/n8n-settings_dark.png#gh-dark-mode-only)

* **Container** – The service name and port (`N8N:5678`) other services in your account can use to reach n8n.
* **Version** – Switch to another n8n release to update (or roll back) n8n. Restart the application afterwards. Workflows and credentials are kept on the data volume.
* **Resource limits** – CPU cores, memory (GB) and PIDs for the container. Click **Save changes**, then restart.

### Env Vars

![Env Vars tab with custom KEY=VALUE environment variables for the n8n container and the Save Environment Variables button](/img/openpanel-screenshots/applications/n8n-envvars.png#gh-light-mode-only)
![Env Vars tab with custom KEY=VALUE environment variables for the n8n container and the Save Environment Variables button](/img/openpanel-screenshots/applications/n8n-envvars_dark.png#gh-dark-mode-only)

Add custom environment variables for the n8n container, one `KEY=VALUE` per line - for example `GENERIC_TIMEZONE`, execution data pruning, or SMTP settings for n8n's user emails. See the [n8n environment variable reference](https://docs.n8n.io/hosting/configuration/environment-variables/).

Click **Save Environment Variables**, then restart the application from the **Overview** tab for the changes to take effect.

### Logs

![Logs tab showing the n8n container log](/img/openpanel-screenshots/applications/n8n-logs.png#gh-light-mode-only)
![Logs tab showing the n8n container log](/img/openpanel-screenshots/applications/n8n-logs_dark.png#gh-dark-mode-only)

Shows the n8n container log - startup, database migrations and errors from workflows and nodes. Use **Raw view** to switch the formatting, and the buttons next to it to copy the log or open it in a new tab.

### Remove

![Remove tab with the Confirm delete and Cancel buttons shown after clicking Delete Application](/img/openpanel-screenshots/applications/n8n-remove.png#gh-light-mode-only)
![Remove tab with the Confirm delete and Cancel buttons shown after clicking Delete Application](/img/openpanel-screenshots/applications/n8n-remove_dark.png#gh-dark-mode-only)

**Delete Application** stops n8n, deletes its container **and its data volume**, and removes it from Site Manager. All workflows, credentials and executions are lost, so export your workflows from n8n first. Click **Confirm delete** to proceed.
