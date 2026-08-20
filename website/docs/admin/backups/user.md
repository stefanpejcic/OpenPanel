---
sidebar_position: 2
---

# User Backups

*OpenAdmin > Backups > Users* controls how hosting users' own account backups (websites, databases, etc. — via each user's `docker-volume-backup` service) run. It's only available to Administrators and Users, not Resellers. It has three tabs: **Settings**, **Configuration**, and **Runs**.

There are two mutually exclusive modes, controlled entirely by the schedule of the `opencli docker-backup` [System Cron Job](/docs/admin/advanced/crons/):

- **User Configured** (default) — each user manages their own backup from their own account, when the Backups module is enabled for them. Nothing runs centrally; the cron entry is set to a disabled placeholder schedule.
- **Admin Configured** — the Administrator runs backups for *every* user centrally, on one schedule, via `opencli docker-backup`.

See the [Configuring OpenPanel Backups](/docs/articles/backups/configuring-backups/) article for the fuller comparison between the two approaches.

### Settings

A single dropdown decides the mode:

- **Disabled** — user configured (the default).
- **Daily** / **Weekly** / **Monthly** — admin configured, all running at 03:00 server time. Fine-tune the exact time afterward from [Scheduled Actions](/docs/admin/advanced/crons/).

Saving updates the `opencli docker-backup` cron entry's schedule accordingly.

### Configuration

Edits `/etc/openpanel/backups/backup.env` — the default `docker-volume-backup` settings (destination, retention, notifications, etc.) every **new** user account is provisioned with. This does not change any existing user's own backup settings, only what future accounts start with.

A **Restore Default** button fetches OpenPanel's shipped default `backup.env` from GitHub into the field (client-side, no server round trip) — click **Save** afterward to actually apply it.

### Runs

Raw log of every `opencli docker-backup` run (`/var/log/openpanel/admin/docker-backup.log`).

**Run Backup Now** triggers `opencli docker-backup` immediately for every user. It's only available in **Admin Configured** mode — in Disabled mode there's no central schedule for it to act on, since each user's own settings apply instead.
