---
sidebar_label: "Configuring OpenPanel Backups"
---

# How to Configure Automatic Server and Account Backups

OpenPanel has a unique feature where end-users can configure their remote backups. This provides users with more freedom and control over the schedule, what to backup and finally more privacy as Admin does not have access to their destination.

Backups can be configured either by the system administrator (admin-configured) or by end users (user-configured). Each mode has distinct setup and restore procedures.

---

## Configuration Options

| Feature                    | Admin-Configured Backups                              | User-Configured Backups               |
| -------------------------- | ------------------------------------------------------ | ------------------------------------- |
| Backup configuration       | Admin edits `backup.env`                               | Users configure via Backups page      |
| Backup module status       | Enabled - Destinations/Settings pages locked for users | Enabled, fully self-service           |
| Who sets backup schedule   | Admin                                                   | User                                  |
| Backup destination control | Admin                                                   | User                                  |
| Who lists/restores backups | User (from their own panel)                             | User                                  |
| Admin access to backups    | Full                                                    | None                                  |


### 1. Admin-Configured

In this mode, the **admin has full control** over backup scheduling, retention, and destination settings. End users can still browse and restore their own backups from the panel, but they **can't view or change** the destination or its credentials.

---

#### 1: Set the Central Schedule

**Path:**
`OpenAdmin > Backups > Users > Settings`
**Action:** Set the schedule dropdown to **Daily**, **Weekly**, or **Monthly** and save.

This does two things: it schedules `opencli docker-backup` to run centrally for every user, and it locks the Backups **Destinations** and **Settings** pages in the panel for accounts created from then on (via an `admin.backups` marker - see [User Backups](/docs/admin/backups/user/) for details, including how to apply it to an existing account). Their **List Backups** and **Restore Logs** pages stay available, so users can still restore from the backups the admin is taking.

---

#### 2: Edit Template

Modify the backup configuration template that applies to new accounts. This file defines default backup settings, such as remote storage destinations.

**File path:**
`/etc/openpanel/backups/backups.env`

To enable and configure a remote SSH destination, uncomment and update the following variables:

```env
########### SSH/SFTP STORAGE
# SSH_HOST_NAME=""
# SSH_PORT="22"
# SSH_REMOTE_PATH=""
# SSH_USER=""
# SSH_PASSWORD=""
# SSH_IDENTITY_FILE="/var/www/html/id_rsa"
# SSH_IDENTITY_PASSPHRASE=""
```

**Example:**

```env
########### SSH/SFTP STORAGE
SSH_HOST_NAME="185.119.22.54"
SSH_PORT="22"
SSH_REMOTE_PATH="/backups/"
SSH_USER="root"
SSH_PASSWORD="NotSoStrongP@ssword"
# SSH_IDENTITY_FILE="/var/www/html/id_rsa"
# SSH_IDENTITY_PASSPHRASE=""
```

> 🔗 For more destination types and examples, see [Backups Documentation](/docs/panel/backups/#destinations)

---

#### 3: Edit Schedule

To set the backup frequency, go to:

**Path:**
`OpenAdmin > Server > Scheduled Actions`

Locate the cron job for the command:

```bash
opencli docker-backup
```

Adjust the schedule as needed. This command will trigger backups according to your defined schedule for **all active users** on the server.

---

### 2. User-Configured

In this mode, the **Backups module is enabled** to allow users to configure their own backups based on their needs.

**Setup:**

* The admin must **enable the Backups module** in OpenPanel.
* Backups feature must be enabled on all relevant feature sets tied to hosting plans to allow user access.
* End users can set:

  * Backup destination (e.g., remote storage, custom paths)
  * Backup schedule (when backups run)
  * What data to back up (files, databases, or both)
  * Resource limits (e.g., bandwidth or CPU used during backup)

> 🔗 For end-user configuration, see [Backups Documentation](/docs/panel/backups/)

**Notes:**

* Users are responsible for managing their backups.
* Users can manually trigger backup process at any time if *Docker* feature is enabled.
* Admins do **not** have access to users' backup destinations or configurations.

---

## Restore Procedures

### Restore in Admin-Configured Backup Mode

* Users restore their own backups the same way as in User-Configured mode: **Backups > Restore** in the panel, which stays available even though Destinations/Settings are locked - see [Restore & Download](/docs/panel/backups/#restore--download).
* The admin can also restore on a user's behalf manually, either via terminal commands or through the OpenPanel UI terminal:

  * For databases: dropping the relevant tables and importing the database dump from backup files.
  * For files: using FileManager or command line to delete corrupted files and re-upload backup copies.

---

### Restore in User-Configured Backup Mode

* End users are responsible for restoring their own backups, as backups are stored in user-defined destinations inaccessible to admins.
* Users follow similar restore steps as in the admin mode but must perform actions themselves using provided tools or instructions.
* Admins cannot restore or access backups on behalf of users in this mode.

---

