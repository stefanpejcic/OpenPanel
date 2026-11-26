---
sidebar_position: 1
---

# User backups

*OpenAdmin > Backups > Users* controls how hosting users' own account backups (websites, databases, etc., via each user's `docker-volume-backup` service) run. It's only available to Administrators and Users, not Resellers. It has three tabs: **Settings**, **Configuration**, and **Runs**.

There are two mutually exclusive modes, controlled entirely by the schedule of the `opencli docker-backup` [System Cron Job](/docs/admin/advanced/crons/):

- **User Configured** (default), each user manages their own backup from their own account, when the Backups module is enabled for them. Nothing runs centrally; the cron entry is set to a disabled placeholder schedule.
- **Admin Configured**: the Administrator runs backups centrally, on one schedule, via `opencli docker-backup`, for every account it manages: the accounts with an `admin.backups` marker (see below). Users of those accounts can no longer view or change their own backup destination or settings in the panel, they can still list existing backups and restore from them (their **Backups** page opens the Restore & Download page directly), but where and how backups are stored is entirely up to the Administrator.

See the [Configuring OpenPanel Backups](/docs/articles/backups/configuring-backups/) article for the fuller comparison between the two approaches.

### Settings

A single dropdown decides the mode:

- **Disabled**: user configured (the default).
- **Daily** / **Weekly** / **Monthly**, admin configured, all running at 03:00 server time. Fine-tune the exact time afterward from [Scheduled Actions](/docs/admin/advanced/crons/).

Saving updates the `opencli docker-backup` cron entry's schedule accordingly.

![User Backups Settings tab with the backup schedule dropdown](/img/openadmin-screenshots/backups/user-settings.png#gh-light-mode-only)
![User Backups Settings tab with the backup schedule dropdown](/img/openadmin-screenshots/backups/user-settings_dark.png#gh-dark-mode-only)

Saving to Daily/Weekly/Monthly also drops an empty `admin.backups` marker file into `/etc/openpanel/skeleton/`, so every account created *from then on* is provisioned admin-configured (its Backups > Destinations and Backups > Settings pages are locked, per-account, from the moment it's created). Switching back to Disabled removes the marker from the skeleton, so accounts created afterward go back to fully user-configured.

This only affects **new** accounts, it doesn't touch any account that already exists. To hand an existing account's backups to the Administrator after the fact, create an empty `admin.backups` file yourself in that account's own config directory:

```bash
touch /etc/openpanel/openpanel/core/users/<context>/admin.backups
```

`<context>` is the account's context: the name of its home directory (`/home/<context>`) and system user. It's the same as the username, unless the account was renamed with `opencli user-rename`, which only changes the username. The panel, `opencli docker-backup` and the Restore & Download page all look for the marker under the context.

`opencli docker-backup` only backs up accounts that have this file, accounts without it manage their own backups and are skipped (*Skipping user: … (manages its own backups)* in the [Runs](#runs) log). Removing the file returns the account to user-configured (assuming the Backups module itself is still enabled for it).

:::info Renamed accounts
Before 2.0.25, `opencli user-rename` moved the marker to a folder named after the new username, so a renamed admin-configured account was skipped by `opencli docker-backup` and its **Restore & Download** page failed with *openat …: no such file or directory* (the panel looked for the SSH key inside the account's website files). Since 2.0.25 the marker stays under the context. For accounts renamed earlier, check that the marker exists and create it if it doesn't:

```bash
ls /etc/openpanel/openpanel/core/users/<context>/admin.backups
```
:::

### Restore speed on SSH/SFTP

When users restore from an SSH/SFTP destination, the panel streams the archive from the destination. For admin-configured accounts it downloads with OpenSSH (`ssh … cat <archive>`) when all of these are true:

- the account has the `admin.backups` marker,
- the destination logs in with a key (`SSH_IDENTITY_FILE`), without `SSH_PASSWORD` and without `SSH_IDENTITY_PASSPHRASE`,
- the `ssh` client is available in the OpenPanel container.

The key is copied to a temporary file readable only by the panel (OpenSSH refuses keys other users can read) and removed after the download. When `SSH_HOST_PUBLIC_KEY` is set, OpenSSH only connects to a server with that key, without it the host key isn't checked, the same as the built-in client.

In every other case, and automatically when OpenSSH fails before the download starts (for example a storage that allows SFTP only and refuses commands, or a login error), the panel uses its built-in SFTP client. Accounts that configure their own destination always use the built-in client, so their destination can't point the panel at private or internal addresses.

On the same link OpenSSH is usually 2-3x faster: in our tests 8-15 MB/s, the built-in client about 4-5 MB/s. A restore that stops halfway because the connection dropped is logged as failed, never as done.

### Configuration

Edits `/etc/openpanel/backups/backup.env`, the default `docker-volume-backup` settings (destination, retention, notifications, etc.) every **new** user account is provisioned with. This does not change any existing user's own backup settings, only what future accounts start with.

A **Restore default** button fetches OpenPanel's shipped default `backup.env` from GitHub into the field (client-side, no server round trip), click **Save** afterward to actually apply it.

![User Backups Configuration tab with the default backup.env editor and the Restore Default button](/img/openadmin-screenshots/backups/user-configuration.png#gh-light-mode-only)
![User Backups Configuration tab with the default backup.env editor and the Restore Default button](/img/openadmin-screenshots/backups/user-configuration_dark.png#gh-dark-mode-only)

### Runs

Raw log of every `opencli docker-backup` run (`/var/log/openpanel/admin/docker-backup.log`).

![User Backups Runs tab with the log of past backup runs](/img/openadmin-screenshots/backups/user-runs.png#gh-light-mode-only)
![User Backups Runs tab with the log of past backup runs](/img/openadmin-screenshots/backups/user-runs_dark.png#gh-dark-mode-only)

**Run backup now** triggers `opencli docker-backup` immediately for every account the Administrator manages. It's only available in **Admin Configured** mode, in Disabled mode there's no central schedule for it to act on, since each user's own settings apply instead.
