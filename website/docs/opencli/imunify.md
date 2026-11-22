# Imunify 

`opencli imunify` command enables administrators to install, start, manage, and uninstall the [ImunifyAV standalone service](https://cloudlinux.zendesk.com/hc/en-us/articles/4716287786396-Imunify360-Standalone-installation-guide-with-integration-conf-examples).

> Note: Imunify, its trademarks, and all related assets are the property of [CloudLinux Zug GmbH](https://cloudlinux.com/).

This command acts as a wrapper that installs and runs the ImunifyAV GUI as a standalone PHP server, following the official setup guide provided here: [Imunify360 Standalone installation guide with integration.conf examples
](https://cloudlinux.zendesk.com/hc/en-us/articles/4716287786396-Imunify360-Standalone-installation-guide-with-integration-conf-examples) 

## Install

To install ImunifyAV:

```bash
opencli imunify install
```

<details>
  <summary>Example output</summary>

```bash
# opencli imunify install
Installing ImunifyAV...
Creating directories...
Creating pam_deny.so file...
Creating integration.conf file...
Downloading deploy script...
Running deploy script...
...
Configuring ImunifyAV notifications to use 'OpenAdmin > Settings > Notifications'...
Setting 2CPU and 1GB Memory limits for ImunifyAV service..
Adding Containers, Images and writable filesystems to ignored files list..
Allowing users to initiate a scan..
Installing PHP if not present...
PHP already installed.
New service added successfully.
ImunifyAV UI for the user panel listens on 172.20.0.1:9000.
Install completed!
```
</details>

## User panel

`opencli imunify install` also sets up what the [ImunifyAV page in OpenPanel](/docs/admin/settings/modules/#imunifyav) needs:

* The agent's socket folder `/run/defence360agent/` is mounted into the `openpanel` container. On servers installed before the mount was in `/root/docker-compose.yml`, `install`, or `start` while the `imunifyav` module is enabled, adds the line and recreates the container once. If the compose file looks different, it prints a warning and you add the mount yourself.
* A second ImunifyAV UI runs as the `imunify-gui-panel` systemd service on the gateway of the OpenPanel network (usually `172.20.0.1:9000`), because the `openpanel` container can't reach `127.0.0.1` on the host. It only answers `/imav/` requests that carry the key in `/etc/openpanel/imunify/panel.key`, so user containers that can reach the gateway get `403`.

The service only runs while the `imunifyav` module is enabled. Turning the module on or off in *OpenAdmin > Settings > Modules* starts or stops it, and `install` and `start` leave it stopped when the module is off:

```bash
# opencli imunify start
Starting ImunifyAV...
ImunifyAV GUI started.
Enable the imunifyav module in OpenAdmin > Settings > Modules to give users the ImunifyAV page.
```

`opencli imunify start` creates the service on servers that don't have it yet and restarts it, `stop` stops it and `uninstall` removes it with `/etc/openpanel/imunify/`.

## Status

To check the status of the ImunifyAV GUI:

```bash
opencli imunify status
```

<details>
  <summary>Example output</summary>

```bash
# opencli imunify status
Imunify GUI is running.
48213 php -S 127.0.0.1:9000 -t /etc/sysconfig/imunify360/
Imunify GUI for the user panel is running.
```
</details>

## Start

To start the ImunifyAV GUI:

```bash
opencli imunify start
```

<details>
  <summary>Example output</summary>

```bash
# opencli imunify start
Starting ImunifyAV...
ImunifyAV GUI started.
ImunifyAV UI for the user panel listens on 172.20.0.1:9000.
```
</details>

## Stop

To stop the ImunifyAV GUI:

```bash
opencli imunify stop
```

<details>
  <summary>Example output</summary>

```bash
# opencli imunify stop
Stopping ImunifyAV...
ImunifyAV GUI disabled.
```
</details>

## Update

To update ImunifyAV:

```bash
opencli imunify update
```

<details>
  <summary>Example output</summary>

```bash
# opencli imunify update
Updating ImunifyAV...
...
Reading package lists... Done
imunify-antivirus is already the newest version.
```
</details>

## Uninstall

To uninstall ImunifyAV:

```bash
opencli imunify uninstall
```

<details>
  <summary>Example output</summary>

```bash
# opencli imunify uninstall
Uninstalling ImunifyAV...
ImunifyAV GUI disabled.
Removing files and directories...
Checking for _imunify user...
User '_imunify' exists. Not removing automatically to avoid breaking dependencies.
→ If you're sure, run: userdel _imunify
Removing ImunifyAV from OpenAdmin > Services Status...
Service 'ImunifyAV' removed successfully.
Uninstall complete.
```
</details>

## Scan notifications

After ImunifyAV is installed, every malware scan it starts and finishes is reported by `opencli imunify notify`, which runs on its own when ImunifyAV writes a scan event. Each scan is one notification in OpenAdmin, sent by email and webhook when it finishes, with the number of files checked and up to 5 malicious files. To turn these off, disable **ImunifyAV scan started/finished** under User Actions in **OpenAdmin > Settings > Notifications**.

When a finished scan of a user's home folder found malware, that user also gets a **Malware found** email listing up to 20 infected files, if they have [Malware found](/docs/panel/account/notifications/#malware-found) turned on.

<details>
  <summary>Example output</summary>

```bash
# opencli imunify notify
stefan notified by email about 2 malicious file(s).
```
</details>
