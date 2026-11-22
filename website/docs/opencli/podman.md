---
sidebar_position: 12
---

# Podman

`opencli podman` runs `podman` or `podman-compose` against an OpenPanel user's rootless podman, so you don't need to find the user's socket yourself.

Run a podman command as the user:
```bash
opencli podman <USERNAME> <ARGS...>
```

<details>
  <summary>Example output</summary>

```bash
# opencli podman stefan ps --format '{{.Names}}'
php-fpm-8.6
mariadb
cron
```
</details>

Run podman-compose for the user's services, from the user's home directory:
```bash
cd /home/<USERNAME> && opencli podman <USERNAME> compose <ARGS...>
```

<details>
  <summary>Example output</summary>

```bash
# cd /home/stefan && opencli podman stefan compose restart cron
cron
```
</details>

Print the user's podman socket, for tools that read `CONTAINER_HOST`:
```bash
opencli podman <USERNAME> --socket
```

<details>
  <summary>Example output</summary>

```bash
# opencli podman stefan --socket
unix:///hostfs/run/user/1003/podman/podman.sock
```
</details>
