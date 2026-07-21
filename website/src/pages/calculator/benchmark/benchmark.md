# Benchmark Methodology & Raw Results

The numbers behind the [Resource Calculator](/calculator/) aren't guesses - they come from an actual OpenPanel 2.0 (Podman) install, pushed to real user counts and measured with `opencli`. This page documents exactly how the test was run and what came back, so you can verify the assumptions yourself.

## Test environment

- VPS: [AltusHost VM-4](https://altushost.com/vps-hosting/): 8 cores, 9GB RAM
- Enterprise license, **without** the premium services enabled (no Email/mailserver, no FTP, no DNS)
- Fresh OpenPanel install, no prior accounts

## How the test was run

1. **Fresh install** of OpenPanel on the VPS above.
2. Ran [`max_users.sh`](https://github.com/stefanpejcic/openpanel-tests/blob/main/opencli/max_users.sh), which:
   - Creates hosting accounts **one at a time**, each on the **Standard plan** (2 CPU cores / 2GB RAM cap).
   - Enables **5 services per account**: MariaDB, Apache, PHP 8.5, Cron, Memcached.
   - **Waits for every service in the account to report started** before moving on to create the next account - so accounts are never created faster than the server can actually bring them up.
3. Left the server running **idle for 1 hour** after the last account was created (no synthetic traffic).
4. Ran `opencli docker-collect_stats --all` to capture actual, per-user resource usage.
5. Cross-checked against `podman image ls`, `du -sh`, `free -mh`, `df -h`, `uptime`, and `opencli user-list`.

## Run it yourself

Any admin can reproduce this on their own test VPS with the following commands.

Fresh install (see [install command generator](/install) for other options):

```bash
bash <(curl -sSL https://openpanel.org/) --key=enterprise-xxxx
```

Download and run the account-creation script (creates users one at a time on the Standard plan, 5 services each, waiting for each account to finish starting before the next):

```bash
curl -sSLO https://raw.githubusercontent.com/stefanpejcic/openpanel-tests/main/opencli/max_users.sh
chmod +x max_users.sh
./max_users.sh
```

Let the server sit idle for about an hour, then collect real per-user usage:

```bash
opencli docker-collect_stats --all
```

And cross-check with the system-level commands used for this test:

```bash
opencli user-list
podman image ls
du -sh /var/lib/containers/
du -sh /home
free -mh
df -h /
uptime
```

## Shared images (downloaded once, reused by every user)

This is the basis for the calculator's "Services" section - every image below is pulled once, regardless of how many users select that service.

| Service | Image | Size |
|---|---|---|
| Database | `mariadb:10-focal` | 347 MB |
| Database | `mysql` | 962 MB |
| Web server | `nginx:stable-alpine` | 64 MB |
| Web server | `httpd:alpine` (Apache) | 69 MB |
| Web server | `openlitespeed` | 682 MB |
| Web server | `openresty:alpine` | 158 MB |
| Caching | `memcached:1.6.41-alpine` | 13 MB |
| Caching | `valkey:9.0-alpine` | 44 MB |
| Caching | `redis:8.6.2-alpine` | 98 MB |
| PHP (each version) | `shinsenter/php:*-fpm` | ~250-360 MB per version |
| Cron | `mcuadros/ofelia` | 25 MB |

## System images (shared infrastructure, not chosen per user)

These aren't services a user selects - they're the always-on containers every install runs to serve and administer all accounts.

| Service | Image | Size |
|---|---|---|
| OpenPanel UI | `openpanel/openpanel-ui` | 2.41 GB (includes Playwright) |
| phpMyAdmin | `phpmyadmin` | 589 MB |
| Caddy | `openpanel/caddy-coraza` | 116 MB |
| DNS | `ubuntu/bind9` | 99 MB |

## Disk usage

With 15 users running the 5 services above:

- **User images (all versions/services combined):** 11.18 GB
- **System images** (UI, mysql, redis, caddy, phpmyadmin): 4.60 GB - the UI image alone is 2.41GB of that, mostly Playwright
- **Actual container storage on disk:** `du -sh /var/lib/containers/` → **5.1 GB** - far below the naive 11.18 + 4.60 GB sum, because shared images are stored once and reused
- **Per-additional-user disk delta observed:** ~100 MB/user (writable layer + account data, on top of the shared images)

## RAM usage

**Bare/idle account** (no services or domain yet): ~15MB RAM.

**Real account, 5 services running, 1 hour idle** (from `opencli docker-collect_stats --all`, sampled across the test accounts):

```text
user: testhdcw6lom   memory used: 182.8M   cpu: 0.1 cores (pct: 5)
user: test8z2c9li0    memory used: 96.2M    cpu: 0.1 cores (pct: 9)
user: testyldvawbz   memory used: 176.6M   cpu: 0.0 cores (pct: 4)
user: testm420v4mo   memory used: 189.6M   cpu: 0.0 cores (pct: 3)
user: testdk4e6qcb   memory used: 180.8M   cpu: 0.1 cores (pct: 5)
user: testns1hzu92   memory used: 114.2M   cpu: 0.0 cores (pct: 4)
user: test16bl1nh3   memory used: 192.6M   cpu: 0.1 cores (pct: 5)
user: testnpz6ktcj   memory used: 26.3M    cpu: 0.0 cores (pct: 0)
user: testkm20itbn   memory used: 23.9M    cpu: 0.0 cores (pct: 0)
user: test2yv96x7q   memory used: 23.7M    cpu: 0.0 cores (pct: 0)
user: testkwzwlkok   memory used: 24.2M    cpu: 0.0 cores (pct: 0)
user: testybg395dc   memory used: 24.1M    cpu: 0.0 cores (pct: 0)
user: testy02m4ux5   memory used: 74.5M    cpu: 0.0 cores (pct: 4)
user: testh7xjd0gv   memory used: 78.4M    cpu: 0.1 cores (pct: 5)
user: test9xjhqukr   memory used: 89.4M    cpu: 0.1 cores (pct: 6)
user: testrbgigc7e   memory used: 191.6M   cpu: 0.0 cores (pct: 4)
user: testu0xdlajh   memory used: 193.3M   cpu: 0.1 cores (pct: 8)
user: testajuhx8yf   memory used: 193.4M   cpu: 0.0 cores (pct: 4)
```

That's **18 accounts, ~2.03 GB RAM combined**, averaging **~115MB/account** - some fresh accounts (no site provisioned yet) sat as low as ~24MB, accounts with an actual site running closer to ~180-203MB.

The `cpu` field in that same output is each container's own attributed usage, and only covers the container's own workload - it doesn't include the kernel/scheduling overhead of running 100+ containers on one host (network namespaces, rootless networking proxies, context switching). Real per-user CPU cost is measured system-wide instead, below.

**System containers** (mysql, redis, caddy, UI, phpmyadmin combined): with `free -mh` showing 2.9GB used system-wide against ~2.03GB from the accounts above, that leaves roughly **~0.87GB RAM** for all 5 system containers together - well below the sum of their individually-assumed costs, which is why the calculator's system baseline is set low and treated as a floor, not a hard measured ceiling.

## Real CPU usage (system-wide)

`vmstat 1 5`, sampled ~6 hours after the last account was created, with all 20 accounts idle:

```text
procs -----------memory---------- ---swap-- -----io---- -system-- -------cpu-------
 r  b   swpd   free   buff  cache   si   so    bi    bo   in   cs us sy id wa st gu
27  1 1045664 489412 269500 5809848    1   33    85 50498 72510  132 39 42 17  2  0  0
25  1 1045664 497804 269500 5809996    0    0     0 49276 76445 110036 42 43 13  2  0  0
12  0 1045664 499536 269500 5814108    0    0     0 113760 78239 103005 37 41 15  6  0  0
17  1 1045664 455716 269500 5814112    0    0     0 49224 75054 107971 40 45 14  2  0  0
14  2 1045664 472212 269500 5814112    0    0     0 48312 75814 109682 41 43 14  2  0  0
```

`b` (processes blocked on I/O) sits at 0-2 - this isn't disk contention. `r` (processes runnable, waiting for a CPU) runs 12-27 on an 8-core box, `us`+`sy` average ~82.6% CPU busy, with `sy` (kernel time) nearly matching `us` (actual process time), and context switches running 72k-110k/sec.

That combination - high `sy`, huge context-switch rate, near-zero `wa` - is kernel/scheduling overhead from running 105 containers. Rootless Podman gives each container its own network namespace plus a userspace networking proxy (slirp4netns/pasta), and that cost lands in kernel time and context switches, not in any individual container's own cgroup accounting.

Backing out the real per-user CPU cost from system-wide utilization:

```
busy cores  = (us% + sy%) × 8 cores       = 0.826 × 8   ≈ 6.6 cores
per user    = (busy cores − system baseline) ÷ 20 users
            = (6.6 − 1.5) ÷ 20             ≈ 0.25 cores/user
```

**~0.25 cores/user** is what the calculator uses for its idle CPU estimate.

## Scale test

- **15 users × 5 services:** ran with zero issues on the 8-core/9GB VPS.
- **20 users × 5 services = 105 containers total** (100 rootless user containers + 5 rootful system containers: mysql, UI, redis, phpmyadmin, caddy).
- Those 20 users were all on the **Standard plan (2 CPU / 2GB RAM cap each)** - meaning **40 cores / 40GB RAM worth of plan commitments** oversold on an 8-core/9GB box.
- `uptime` load average stayed at **13.48-24.45 for hours**, including a re-check nearly 6 hours after account creation finished - consistent with the ~6.6 busy cores measured above on an 8-core box.
- `free -mh` at 20 users: **2.9GiB used / 8.7GiB total**, 454MB free, 5.8GiB in buff/cache.
- `df -h`: **22GB used / 99GB total** (24%).

## For comparison: the old Docker-based 1.x release

The same 40-user, 5-services-each scenario on the previous Docker-based OpenPanel 1.x needed **2 cores and 5GB RAM at idle**, with each user costing roughly **~250MB RAM, ~0.1% CPU, and ~1GB storage** at idle - versus the **~15MB RAM and ~100MB disk per user** measured here on Podman/2.0. That difference is the whole reason the [Resource Calculator](/calculator/) recommends dramatically smaller servers than the old 1.x sizing did.
