---
title: Why OpenPanel 2.0 Is Switching to Podman
description: OpenPanel 2.0 moves new installations from Docker to Podman - here's the real RAM, disk, and image-storage math behind the switch.
slug: podman-switch-openpanel-2
authors: stefanpejcic
tags: [OpenPanel, podman, docker, containers, performance]
image: https://openpanel.com/img/blog/OPENPANEL BLOG 1500x800.png
hide_table_of_contents: true
is_featured: false
---

OpenPanel 2.0, scheduled for the end of this year, changes the container engine under the hood: **new installations will run on Podman instead of Docker.**

<!--truncate-->

Existing installations are not affected. If you're already running OpenPanel on Docker, you stay on Docker - there is no forced migration. This is purely a change for fresh installs going forward, and it's the single biggest performance change we've shipped since launch.

The short version: Podman gets rid of the Docker daemon overhead, and that overhead was being paid **per user, on every server, all the time.**

---

## Why we're doing this

Docker's architecture assumes one daemon (`dockerd`) per host, running as root, managing every container for every user through a single long-lived process. OpenPanel runs Docker rootless per user to keep tenants isolated, which means every single user account on a server gets its own `dockerd` instance, its own containerd shim, and its own copy of every image it needs.

That's a lot of duplicated weight for a hosting panel where a server might have 10, 100, or 200+ user accounts, each running a handful of lightweight services.

Podman doesn't have a daemon at all. Containers are launched as regular child processes (fork/exec) directly by the podman command, supervised by systemd. No background service sitting there per user, no daemon socket, no container manager process idling in RAM waiting for work.

We tested this directly on idle, freshly created accounts with zero running containers, on otherwise identical servers.

---

## The idle-user cost: RAM and disk

Here's the raw output from two brand-new, completely idle user accounts - no services started, just the account created:

**Podman:**
```
root@stefan:/# bash 1 dorotea
user:        dorotea (uid 1007)
home:        /home/dorotea
disk usage:  1.3M
cpu:         0.0%   (sum of live procs)
ram:         20.2 MB   (RSS)
containers:  0 total, 0 running
```

**Docker (rootless):**
```
root@srv54:~# bash 1 prazan
user:        prazan (uid 1010)
home:        /home/prazan
disk usage:  269M
cpu:         1.0%   (sum of live procs)
ram:         166.8 MB   (RSS)
containers:  0 total, 0 running
```

Both users have **zero containers running.** This is pure engine overhead - the cost of the container runtime existing on the account at all.

| | Docker (rootless) | Podman | Saved per user |
|---|---|---|---|
| Disk | 269 MB | 1.3 MB | **267.7 MB** |
| RAM | 166.8 MB | 20.2 MB | **146.6 MB** |
| CPU | 1.0% | 0.0% | idle daemon gone |

That's before a single website, database, or cron job is even running. It's just the cost of Docker's rootless daemon sitting on the account, versus Podman having nothing to sit there.

Multiply that across a real user base:

| Users | Docker disk (idle overhead) | Podman disk (idle overhead) | Disk saved | Docker RAM (idle overhead) | Podman RAM (idle overhead) | RAM saved |
|---|---|---|---|---|---|---|
| 10 | 2.63 GB | 13 MB | **~2.6 GB** | 1.63 GB | 202 MB | **~1.4 GB** |
| 100 | 26.3 GB | 130 MB | **~26.1 GB** | 16.3 GB | 2.0 GB | **~14.3 GB** |
| 200 | 52.5 GB | 254 MB | **~52.3 GB** | 32.6 GB | 3.9 GB | **~28.6 GB** |

On a 200-user server, that's roughly **28.6 GB of RAM** and **52 GB of disk** you get back before anyone has started a single container - just from removing per-user daemons. On a memory-constrained VPS host, that's the difference between needing to upgrade the box and not.

---

## The bigger number: image storage

The idle-account overhead above is just the runtime itself. The bigger cost shows up once users actually start services, and it comes down to how each engine stores images.

**Docker rootless duplicates images per user.** Every user has their own isolated Docker root, so if user A and user B both run PHP 8.2, the PHP 8.2 image is downloaded and stored twice - once per user. There's no sharing between rootless Docker instances by design. If a user spins up every service OpenPanel offers - all 30+ available images (multiple PHP versions, webservers, databases, mail, DNS, monitoring, etc.) - that user alone can account for **~11 GB of duplicated images**, entirely on their own.

**Podman uses shared image storage.** Images are pulled once into a shared, read-only store and every user's containers reference the same layers instead of re-downloading and re-storing their own copy. The full image catalog for the whole system - every version of every service OpenPanel supports - comes to about **10 GB total**, shared across every user on the box, regardless of how many accounts exist.

For a typical active user - not a power user running all 30+ services, just a normal site with 2 PHP versions, a webserver, cron, MySQL, and Redis - that duplication costs about **2 GB per user** on Docker that Podman simply doesn't need, because those images are already sitting in the shared store.

### Typical active user (2 PHP versions + webserver + cron + MySQL + Redis)

| Users | Docker (duplicated images) | Podman (shared store) | Saved |
|---|---|---|---|
| 10 | 20 GB | ~10 GB (shared, flat) | **~10 GB** |
| 100 | 200 GB | ~10 GB (shared, flat) | **~190 GB** |
| 200 | 400 GB | ~10 GB (shared, flat) | **~390 GB** |

### Power user scenario (all 30+ services running)

| Users | Docker (duplicated images) | Podman (shared store) | Saved |
|---|---|---|---|
| 10 | 110 GB | ~10 GB (shared, flat) | **~100 GB** |
| 100 | 1.1 TB | ~10 GB (shared, flat) | **~1.09 TB** |
| 200 | 2.2 TB | ~10 GB (shared, flat) | **~2.19 TB** |

The Podman number barely moves as user count grows, because the images are shared, not duplicated. Docker's number scales linearly with every account you add. This is the part of the switch that matters most for hosts running dense, multi-tenant servers - disk that used to grow with every new signup now mostly doesn't.

---

## Podman vs Docker: the rest of the picture

The RAM and disk numbers are what pushed us to make the switch, but they're not the only reason Podman is the better fit for a rootless, multi-tenant hosting panel:

- **No root daemon.** Docker's daemon traditionally runs as root and is a single point of failure and a single point of attack surface for the entire host. Podman has no daemon - each container is a direct child process of the user that started it, supervised by systemd.
- **True rootless isolation.** Podman was built rootless-first. User namespaces, container processes, and storage are isolated per user without needing a privileged background service to broker access.
- **systemd-native.** Podman generates systemd unit files (`podman generate systemd` / Quadlet) so containers restart, log, and get managed the same way as every other service on the box, instead of through a separate container supervisor.
- **Drop-in CLI compatibility.** `podman` accepts the same commands and Compose-style workflows as `docker`, so the switch doesn't mean relearning tooling.
- **OCI-compliant images.** Podman uses standard OCI images and registries - nothing proprietary, nothing that locks images to one engine.
- **No single point of failure.** Because there's no central daemon, one user's container problems can't take down container management for every other user on the server the way a wedged `dockerd` can.

For a panel where hundreds of independent user accounts share one host, "no shared root daemon" isn't a nice-to-have, it's the whole point.

---

## Still in beta

To be upfront: Podman support in OpenPanel is still in **beta**. It's what new installations will default to once OpenPanel 2.0 ships at the end of the year, but we're actively hardening it against edge cases before then.

Existing OpenPanel installations keep running on Docker - nothing changes for you, and there's no forced migration path being pushed on current servers. If you want to try Podman ahead of the 2.0 release to kick the tires, or you hit something that doesn't behave the way Docker did, we want to hear about it. That feedback loop is exactly how we get this out of beta.
