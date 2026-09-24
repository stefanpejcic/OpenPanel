# Containers

Manage containers: open a shell into them, check for and apply image updates, prefetch autostart images, back them up, view log sizes, and collect resource usage stats.

### docker

Command opens an interactive shell into system or user containers. When the user or container is not given, an interactive `fzf` picker is shown to select it (press `ESC` to quit).

To pick a user and then one of their containers:
```bash
opencli docker
```

To pick one of the active containers for a user:
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

<details>
  <summary>Example output</summary>

```bash
# opencli docker-update
Checking 14 unique image(s)...
  ^  docker.io/library/mariadb:latest - update available
  ^  docker.io/library/nginx:latest - update available
  +  docker.io/library/redis:8.6.2-alpine - not pulled yet

3 image(s) have an update available, affecting 2 context(s):
  - root
  - user: stefan

Apply these updates now? [y/N] (auto-cancel in 10s): y
2026-09-24 10:20:01 : === Update run started ===
2026-09-24 10:20:01 : pull docker.io/library/mariadb:latest
2026-09-24 10:20:12 : pull docker.io/library/nginx:latest
2026-09-24 10:20:19 : pull docker.io/library/redis:8.6.2-alpine
2026-09-24 10:20:24 : recreating stack for root (down/up)
2026-09-24 10:20:41 :   root is back up
2026-09-24 10:20:41 : recreating stack for user: stefan (down/up)
2026-09-24 10:20:55 :   user: stefan is back up
2026-09-24 10:20:55 : === Update run finished ===
```
</details>

Options:

- `-y`, `--yes` - apply updates without prompting for confirmation.
- `--dry-run` - only check and print available updates, never prompt or apply.

With no options, available updates are printed and confirmation is asked for 10 seconds — if there's no answer in that time, the script exits without changing anything.

### Autostart

Prefetches (pulls) container images for services listed as autostart, skipping if free disk space is low.

```bash
opencli docker-autostart [-f|--force]
```

<details>
  <summary>Example output</summary>

```bash
# opencli docker-autostart
Root disk free is 142GB — proceeding with prefetch.

Autostart services: nginx php-fpm-8.3 mysql
Compose services with images: 24

pull   mysql -> docker.io/library/mysql:latest ... OK
pull   nginx -> docker.io/library/nginx:latest ... OK
skip   openlitespeed (not in autostart)
pull   php-fpm-8.3 -> docker.io/shinsenter/php:8.3-fpm ... OK
skip   redis (not in autostart)
...

Fixing permissions on /var/lib/containers/shared-storage ...
Done.

Store contents:
REPOSITORY                  TAG         IMAGE ID      CREATED      SIZE
docker.io/library/mysql     latest      a1b2c3d4e5f6  2 weeks ago  803 MB
docker.io/library/nginx     latest      b2c3d4e5f6a7  3 weeks ago  197 MB
docker.io/shinsenter/php    8.3-fpm     c3d4e5f6a7b8  5 days ago   512 MB

Shared store disk usage: 1.5G (1.47 GB) at /var/lib/containers/shared-storage
```
</details>

- `-f`, `--force` - prefetch images even when free disk space is low.

### Backup

Generates a backup for all users. This is the bulk/system-wide counterpart to `opencli user-backup`, see [Users](/docs/articles/opencli/user/#backup-user).

```bash
opencli docker-backup
```

<details>
  <summary>Example output</summary>

```bash
# opencli docker-backup
2026-09-24 02:00:01 : === Backup started at 2026-09-24 02:00:01 ===
2026-09-24 02:00:01 : Fetching list of active users...
2026-09-24 02:00:01 : Found 2 active users to process
2026-09-24 02:00:01 : Processing user: stefan (1/2)
2026-09-24 02:00:01 : Processing user: stefan
2026-09-24 02:00:03 : --- backup container output for user: stefan ---
...
2026-09-24 02:00:40 : --- end of backup container output for user: stefan ---
2026-09-24 02:00:41 : Backup completed for user: stefan (context: stefan) | Time taken: 40s
2026-09-24 02:00:41 : ------------------------------
2026-09-24 02:00:41 : Processing user: demo (2/2)
...
2026-09-24 02:01:15 : Successfully processed all 2 users
2026-09-24 02:01:15 : === Backup finished at 2026-09-24 02:01:15 | Total time: 74s ===
```
</details>

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
opencli docker-collect_stats --all
```

<details>
  <summary>Example output</summary>

```bash
# opencli docker-collect_stats --all
{"timestamp":"2026-09-24T10:30:00Z","user":"stefan","uid":1001,"memory":{"total":{"bytes":1073741824,"human":"1.0G"},"used":{"bytes":314572800,"human":"300.0M"},"free":{"bytes":759169024,"human":"724.0M"},"buff_cache":{"bytes":52428800,"human":"50.0M"},"available":{"bytes":759169024,"human":"724.0M"},"usage_pct":29},"cpu":{"usage":{"pct":100,"human":"1.0 cores"},"total":{"pct":4,"human":"0.0 cores"},"server":{"pct":400,"human":"4.0 cores"}},"tasks":{"current":37,"limit":100,"usage_pct":37},"bandwidth":{"limit":{"bits":100000000,"human":"100.0Mbit"},"total_sent":{"bytes":1048576,"human":"8.4Mbit"},"usage_pct":8},"warning":null}
{"timestamp":"2026-09-24T10:30:00Z","user":"demo","uid":1002,"memory":{"total":{"bytes":1073741824,"human":"1.0G"},"used":{"bytes":314572800,"human":"300.0M"},"free":{"bytes":759169024,"human":"724.0M"},"buff_cache":{"bytes":52428800,"human":"50.0M"},"available":{"bytes":759169024,"human":"724.0M"},"usage_pct":29},"cpu":{"usage":{"pct":100,"human":"1.0 cores"},"total":{"pct":4,"human":"0.0 cores"},"server":{"pct":400,"human":"4.0 cores"}},"tasks":{"current":37,"limit":100,"usage_pct":37},"bandwidth":{"limit":{"bits":100000000,"human":"100.0Mbit"},"total_sent":{"bytes":1048576,"human":"8.4Mbit"},"usage_pct":8},"warning":null}
```
</details>

For a single user:
```bash
opencli docker-collect_stats <USERNAME>
```

<details>
  <summary>Example output</summary>

```bash
# opencli docker-collect_stats stefan
{"timestamp":"2026-09-24T10:30:00Z","user":"stefan","uid":1001,"memory":{"total":{"bytes":1073741824,"human":"1.0G"},"used":{"bytes":314572800,"human":"300.0M"},"free":{"bytes":759169024,"human":"724.0M"},"buff_cache":{"bytes":52428800,"human":"50.0M"},"available":{"bytes":759169024,"human":"724.0M"},"usage_pct":29},"cpu":{"usage":{"pct":100,"human":"1.0 cores"},"total":{"pct":4,"human":"0.0 cores"},"server":{"pct":400,"human":"4.0 cores"}},"tasks":{"current":37,"limit":100,"usage_pct":37},"bandwidth":{"limit":{"bits":100000000,"human":"100.0Mbit"},"total_sent":{"bytes":1048576,"human":"8.4Mbit"},"usage_pct":8},"warning":null}
```
</details>

`collect_stats` script will also rotate data according to [`resource_usage_retention` setting](/docs/articles/opencli/config/#resource_usage_retention)
