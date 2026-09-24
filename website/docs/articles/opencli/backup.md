# Backup

`opencli backup` backs up (and restores) OpenPanel/OpenAdmin **system configuration** — panel settings, per-domain webserver/DNS config, the root `docker-compose.yml`/`.env` stacks, PHP/MySQL/FTP/SSH config, etc.

It deliberately excludes real user data: website files, databases, mailbox contents, mail accounts/DKIM keys, and large downloadable software assets (CMS installer archives, the GeoIP database, the vendored Caddy WAF ruleset, ionCube/composer caches) are never touched.

Destination and retention are read from `/etc/openpanel/openadmin/config/backups.ini` — this is the same command OpenAdmin's [System Backups](/docs/admin/backups/system/) page uses under the hood.

Create a backup:
```bash
opencli backup
```

<details>
  <summary>Example output</summary>

```bash
# opencli backup
[2026-09-24 03:00:01] === System backup started ===
[2026-09-24 03:00:01] Including 42/45 configured paths that exist on this server.
[2026-09-24 03:00:09] Backup archive created: /backup/system-backup_2026-09-24_03-00-01.tar.gz (18M, 8s)
[2026-09-24 03:00:09] Pruned old backup: /backup/system-backup_2026-09-17_03-00-01.tar.gz
[2026-09-24 03:00:09] === System backup finished (8s, pruned 1 old backup(s)) ===
```
</details>

Restore from an archive:
```bash
opencli backup --restore <archive_filename>
```

<details>
  <summary>Example output</summary>

```bash
# opencli backup --restore system-backup_2026-09-24_03-00-01.tar.gz
[2026-09-24 10:15:42] === Restoring from system-backup_2026-09-24_03-00-01.tar.gz ===
[2026-09-24 10:15:45] === Restore finished (3s) -- affected services may need a restart to pick up restored config ===
```
</details>

Additional flags:

- `--quiet` - only log to file, don't print progress to stdout.
