# Configuring OpenPanel Backups

OpenPanel has a unique feature where end-users can configure their remote backups. This provides users with more freedom and control over the schedule, what to backup and finally mire privacy as Admin does not have access to their destination.

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


### 1. Admin-Configured Backups

In this mode, the **admin fully controls** backup scheduling, retention, and destination settings. End users do **not** have access to modify backup configurations.

**Setup:**

* The admin must **disable the Backups module** in the user interface to prevent users from changing backup settings.
* All backup configuration is managed via the `backups.env` file. This file defines:

  * Backup schedule (e.g., daily, weekly)
  * Retention policy (how many backups to keep)
  * Destination (local path, remote storage, etc.)

**Notes:**

* Users cannot configure or trigger backups in this mode.
* The admin is responsible for monitoring and maintaining backup integrity.

---

### 2. User-Configured Backups

In this mode, the **Backups module is enabled** to allow users to configure their own backups based on their needs.

**Setup:**

* The admin must **enable the Backups module** in OpenPanel.
* Backups feature must be enabled on all relevant feature sets tied to hosting plans to allow user access.
* End users can set:

  * Backup destination (e.g., remote storage, custom paths)
  * Backup schedule (when backups run)
  * What data to back up (files, databases, or both)
  * Resource limits (e.g., bandwidth or CPU used during backup)

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

