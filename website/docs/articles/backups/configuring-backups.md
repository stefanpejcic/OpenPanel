# Configuring OpenPanel Backups

OpenPanel has a unique feature where end-users can configure their remote backups. This provides users with more freedom and control over the schedule, what to backup and finally more privacy as Admin does not have access to their destination.

Backups can be configured either by the system administrator (admin-configured) or by end users (user-configured). Each mode has distinct setup and restore procedures.

---

## Configuration Options

| Feature                    | Admin-Configured Backups       | User-Configured Backups               |
| -------------------------- | ------------------------------ | ------------------------------------- |
| Backup configuration       | Admin edits `backups.env`      | Users configure via Backups page      |
| Backup module status       | Must be disabled for users     | Must be enabled for users             |
| Who sets backup schedule   | Admin                          | User                                  |
| Backup destination control | Admin                          | User                                  |
| Restore performed by       | Admin                          | User                                  |
| Admin access to backups    | Full                           | None                                  |


### 1. Admin-Configured

In this mode, the **admin has full control** over backup scheduling, retention, and destination settings. End users are **not allowed** to modify any backup configurations.

---

#### 1: Disable Backups Module

To prevent users from changing backup settings, disable the **Backups** module from the admin interface.

**Path:**
`OpenAdmin > Settings > Modules`
**Action:** Deactivate the **Backups** module.

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

> 🔗 For more destination types and examples, see [Backups Documentation](/docs/panel/files/backups/#destinations)

---

#### 3: Edit Schedule

To set the backup frequency, go to:

**Path:**
`OpenAdmin > Advanced > System Cron Jobs`

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

> 🔗 For end-user configuration, see [Backups Documentation](/docs/panel/files/backups/)

**Notes:**

* Users are responsible for managing their backups.
* Users can manually trigger backup process at any time if *Docker* feature is enabled.
* Admins do **not** have access to users' backup destinations or configurations.

---

## Restore Procedures

### Restore in Admin-Configured Backup Mode

* The admin performs restores manually, either via terminal commands or through the OpenPanel UI terminal.
* Common restore steps include:

  * For databases: dropping the relevant tables and importing the database dump from backup files.
  * For files: using FileManager or command line to delete corrupted files and re-upload backup copies.

---

### Restore in User-Configured Backup Mode

* End users are responsible for restoring their own backups, as backups are stored in user-defined destinations inaccessible to admins.
* Users follow similar restore steps as in the admin mode but must perform actions themselves using provided tools or instructions.
* Admins cannot restore or access backups on behalf of users in this mode.

---

