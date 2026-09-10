# Containers

Manage containers: open a shell into them, check for and apply image updates, prefetch autostart images, back them up, view log sizes, and collect resource usage stats.

### docker

Command opens interactive shell into user containers.

To view list of all users:
```bash
opencli docker
```

To view list of active containers for a user:
```bash
opencli docker <USERNAME>
```

To open terminal directly in a container for a user:
```bash
opencli docker <USERNAME> <container>
```


### Update

Checks for image updates across root's stack and every non-suspended user's stack (via [skopeo](https://github.com/containers/skopeo)), prints what's available, and optionally pulls + recreates the affected services. Each image is pulled at most once into the shared image store, so multiple stacks referencing the same image only download it once.

```bash
opencli docker-update [options]
```

Options:

- `-y`, `--yes` - apply updates without prompting for confirmation.
- `--dry-run` - only check and print available updates, never prompt or apply.

With no options, available updates are printed and confirmation is asked for 10 seconds — if there's no answer in that time, the script exits without changing anything.

### Autostart

Prefetches (pulls) container images for services listed as autostart, skipping if free disk space is low.

```bash
opencli docker-autostart [-f|--force]
```

### Backup

Generates a backup for all users. This is the bulk/system-wide counterpart to `opencli user-backup`, see [Users](/docs/articles/opencli/user/#backup-user).

```bash
opencli docker-backup
```

### logs

View logs sizes for user and system containers.


```bash
# opencli docker-logs

Usage: opencli docker-logs [options]

Options:
  <USERNAME>                                    Display log sizes for specified user.
  --system                                      Display log sizes just for system containers.
  --users                                       Display log sizes just for user containers.
  --all                                         Display log sizes for all user and system containers.

Examples:
  opencli docker-logs stefan
  opencli docker-logs --users
  opencli docker-logs --system
  opencli docker-logs --all

```


### Collect Stats

To collect container resource usage information (cpu, ram, i/o) for all users:
```bash
opencli docker-collect_stats
```

`collect_stats` script will also rotate data according to [`resource_usage_retention` setting](/cli/config.html#resource-usage-retention)
