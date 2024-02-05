---
sidebar_position: 8
---

# Backups

Manage backup jobs, destiantions, exclude lists, restore, etc.

## Backup Jobs

To list all current backup jobs run command:
```bash
opencli backup-job list
```
### Create a job
```bash
opencli backup-job create
```
### Edit backup job
```bash
opencli backup-job edit
```
### Run backup job
```bash
opencli backup-run <ID> --force-run
```
### Delete backup job
```bash
opencli backup-job delete <ID>
```

## Destination

To list all current destinations run command:

```bash
opencli backup-destination list
```

Example output:
```bash
Available destination IDs:
2143
2145
2142
2144
2141
```

### Add Destination

To create a new destination run the following command:

```bash
opencli backup-destination create <HOSTNAME> <PASSWORD_FOR_KEY_FILE> <PORT> <USER> <PATH_TO_SSH_KEY_FILE> <STORAGE_PERCENTAGE>
```
Example:
```bash
opencli backup-destination create 15.19.90.29 strongpass 22 root /root/.ssh/id_rsa 85
```

### Edit Destination

To edit destination configuration run the following command:

```bash
opencli backup-destination edit <ID> <HOSTNAME> <PASSWORD_FOR_KEY_FILE> <PORT> <USER> <PATH_TO_SSH_KEY_FILE> <STORAGE_PERCENTAGE>
```
Example:
```bash
opencli backup-destination edit 241 15.19.90.29 strongpass 22 root /root/.ssh/id_rsa 85

Destination  ID: '241' edited successfully!
Previous destination configuration:
{
  "hostname": "18.19.90.29",
  "password": "pass222",
  "ssh_port": 22,
  "ssh_user": "new_ssh_user",
  "ssh_key_path": "/root/.ssh/id_rsa",
  "storage_limit": "100"
}
New destination configuration:
{
  "hostname": "15.19.90.29",
  "password": "strongpass",
  "ssh_port": 22,
  "ssh_user": "root",
  "ssh_key_path": "/root/.ssh/id_rsa",
  "storage_limit": "85"
}
```

### Delete Destination

To delete a destination run the following command:

```bash
opencli backup-destination delete <ID>
```

:::tip
Destination can not be deleted if it is used by any backup job.
:::


### Validate Destination

To validate a destination run the following command:

```bash
opencli backup-destination validate <ID>
```

Command will test;

- if ssh key exists and its permissions
- ssh connection to the destination server
- check if destination folder exists
- check permissions and owner for destination folder
- check available disk space on destination and compare it with storage limit 


Examples:

```bash
opencli backup-destination validate 2144
Validating SSH connection with the destination, running command: 'ssh -i /root/.ssh/id_rsa root@18.19.90.29 -p 22'
SSH connection successful.
Destination directory /root/backup on the remote server is owned by root user.
Disk usage on remote destination is over the storage limit (70%) for /root/backup. Backups jobs will not run!

```

```bash
opencli backup-destination validate 2145
Validating SSH connection with the destination, running command: 'ssh -i /root/.ssh/id_rsa root@18.19.90.29 -p 22'
SSH connection successful.
Destination directory /root/backup on the remote server is owned by root user.
Disk space on remote destination is below the storage limit (85%) for /root/backup. Backups can run without a problem.
```


```bash
opencli backup-destination validate 1
Validating SSH connection with the destination, running command: 'ssh -i /root/.ssh/id_rsa root@18.19.90.29 -p 202'
ssh: connect to host 18.19.90.29 port 202: Connection refused
Validation failed! SSH connection failed or timed out.
SSH Connection Status: 255
```


## Restore


### Full Account restore
To restore user files from a backup run command:

```bash
opencli backup-restore <DATE> <USER> --all
```

### Partial restore
To restore only specific files from a backup, specify what to restore instead of the `--all` flag:

```bash
opencli backup-restore <DATE> <USER> [-files --entrypoint --nginx-conf --mysql-conf --mysql-data --php-versions --crontab --user-data --core-users --stats-users --apache-ssl-conf --domain-access-reports --ssh --timezone]
```

Available options:
- `--all` - restores all user files.
- `--files` - restores home directory
- `--entrypoint` - restore services and their status
- `--apache-conf` - restore Apache configuration
- `--nginx-conf` - restore Nginx configuration
- `--mysql-conf` - restore MySQL configuration
- `--mysql-data` - restore MySQL databases and users
- `--php-versions` - restore php versions and their php.ini files
- `--crontab` - restore cronjobs
- `--user-data` - restore user password, email address, 2fa settings, preferences..
- `--core-users` - restore OpenPanel data for the user account
- `--stats-users` - restore resource usage statistics for the account
- `--apache-ssl-conf` - restore VirtualHosts and SSL certificates
- `--domain-access-reports` - restore html reports for domain visitors
- `--ssh` - restore SSH users and passwords
- `--timezone` - restore timezone setting


## List Backups for user

`opencli backup-list` allows you to view available backups for a single user.

```bash
opencli backup-list <USERNAME> [--json]
```


## View backup details

`opencli backup-details` allows you to view information about a specific user backup.

```bash
opencli backup-list <USERNAME> <BACKUP_JOB_ID> <DATE> [--json]
```

Example:
```bash
root@server:/# openpanel backup-details nesto 2 20240131170700

backup_job_id=2
destination_id=1
destination_directory=20240131170700
start_time=Wed Jan 31 17:07:52 UTC 2024
end_time=Wed Jan 31 17:09:35 UTC 2024
total_exec_time=103
contains=Full Backup
status=Completed
```



## Index

`opencli backup-index` allows you to re-index backups from a remote destination and make them available for users.

```bash
opencli backup-index <ID>
```

Example:
```bash
opencli backup-index 21

Indexing backups for 5 users from destination: 18.19.90.29 and directory: /root/backup
Processing user stefan (1/5)
Indexed 6 backups for user stefan.
Processing user demo (2/5)
Indexed 6 backups for user demo.
Processing user demo123 (3/5)
Indexed 6 backups for user demo123.
Processing user pera (4/5)
Indexed 5 backups for user pera.
Processing user nesto (5/5)
Indexed 5 backups for user nesto.
Index complete, found a total of 28 backups for all 5 users.
```


## Check

`opencli backup-check` checks process id for any running jobs and terminates the backupjob if process is not running.

```bash
opencli backup-check
```


## Logs

### List Logs

### View Log File

### Delete Log File


## Scheduler

`opencli backup-scheduler` command is executed daily to schedule new backup jobs and ensure they are carried out on time.

Example:
```bash
opencli backup-scheduler --debug

0 1 * * SUN opencli backup-run 2  --entrypoint --files --mysql_conf
0 1 * * * opencli backup-run 3 --conf
```


## Config

`opencli backup-config` allows you to change general backup settings that affect all backup jobs and destinations.



### Get

The `get` parameter allows you to view current backup settings.

```bash
opencli backup-config get <OPTION>
```

Example:
```bash
# opencli backup-config get time_format
24
```

### Update

The `update` parameter allows you to change backuo settings.


```bash
opencli backup-config update <OPTION> <NEW-VALUE>
```

Example:
```bash
opencli backup-config update time_format 12
Updated time_format to 12
```

### Options

Currently available backup configuration options:

- debug (yes|no)
- error_report (yes|no)
- workplace_dir (path)
- downloads_dir (path)
- delete_orphan_backups (days)
- days_to_keep_logs (days)
- time_format (12|24)
- avg_load_limit (number)
- concurent_jobs (number)
- backup_restore_ttl (number)
- cpu_limit (number)
- io_read_limit (number)
- io_write_limit (number)
- enable_notification (yes|no)
- email_notification (yes|no)
- send_emails_to (email)
- notify_on_every_job (yes|no)
- notify_on_failed_backups (yes|no)
- notify_on_no_backups (yes|no)
- notify_if_no_backups_after (days)
