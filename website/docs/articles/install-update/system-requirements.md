---
sidebar_label: "System Requirements & Sizing"
description: "OpenPanel minimum and recommended system requirements, and how much RAM, CPU and disk you need for 1, 10, 50 or 100+ websites - with a link to the resource calculator."
---

# OpenPanel System Requirements: How Much RAM, CPU and Disk Do You Need?

This page lists the minimum requirements for installing OpenPanel and explains how to size a server for the number of websites and users you plan to host.

For an exact estimate for your setup, use the **[OpenPanel Resource Calculator](/calculator)**.

---

## Minimum Requirements

| | Minimum | Recommended |
|---|---|---|
| **RAM** | 1 GB | 4 GB or more |
| **Disk** | 20 GB (the installer needs 5 GB free) | 50 GB or more, SSD/NVMe |
| **CPU** | 1 core | 2 cores or more |
| **Architecture** | x86_64 (AMD64) or ARM64 (AArch64) | |
| **Network** | Public IPv4 address | |
| **Virtualization** | Full VM (KVM, VMware, Hyper-V, Xen HVM) or bare metal | |

**Supported operating systems:** Ubuntu 22.04, 24.04 and 26.04, Debian 10-13, AlmaLinux 9.5 and 10, Rocky Linux 9.6 and 10, CentOS 9.5. Ubuntu 24.04 is recommended. See the [full list and known issues](/docs/admin/intro/#requirements).

**Not supported:** Windows, macOS, shared hosting accounts, and container-based VPS (OpenVZ, LXC, Docker). The server must be a **clean** install with no other control panel or web server.

---

## How OpenPanel Uses Resources

OpenPanel runs every user's services (web server, PHP, database, caching) in that user's own isolated containers. This makes sizing more predictable than on traditional panels:

1. **A fixed system base.** The panel itself, the Caddy proxy, MySQL for panel data, Redis and optional services such as DNS, phpMyAdmin, email and malware scanning. Around **2-3 GB of RAM** for a typical setup without email, around **1 GB more** with email.
2. **A small cost per idle user.** An account whose sites get no traffic uses only about **15 MB of RAM** and **~100 MB of disk** (plus its website files, databases and emails).
3. **Shared images, downloaded once.** Web server, database and PHP images are stored once on disk and shared by every user - adding users doesn't multiply them.
4. **Real usage from traffic.** Busy sites use more CPU and RAM. Each user can be capped with [plan limits](/docs/articles/containers/are-plan-limits-hard-or-soft-limits/), so one busy site can't take down the whole server.

Because most hosted sites are idle most of the time, one server can host far more accounts than the sum of their plan limits.

---

## Quick Sizing Guide

Rough starting points for **RAM**, assuming typical small websites (WordPress, PHP sites), Nginx + MariaDB, one PHP version and DNS enabled. The disk figures exclude your websites' own files, databases and emails - add those on top.

| Users / websites | RAM without email | RAM with email | Disk (system) | Try it |
|---|---|---|---|---|
| 1-3 (personal) | 3 GB | 4 GB | 10 GB | [Calculate](/calculator?users=3) |
| 10 | 3 GB | 4 GB | 10 GB | [Calculate](/calculator?users=10) |
| 25 | 3.5 GB | 4.5 GB | 10 GB | [Calculate](/calculator?users=25) |
| 50 | 3.5 GB | 4.5 GB | 15 GB | [Calculate](/calculator?users=50) |
| 100 | 4.5 GB | 5.5 GB | 20 GB | [Calculate](/calculator?users=100) |
| 200 | 6 GB | 7 GB | 25 GB | [Calculate](/calculator?users=200) |

These are **typical usage** figures - most users idle, a few busy at the same time. Add headroom for busy sites: WooCommerce shops, forums and sites with heavy traffic can need 1 GB of RAM or more each on their own.

:::tip Rule of thumb
Start with **4 GB of RAM** for up to ~50 small sites, **8 GB** for up to ~200, and add RAM as you add busy sites. Upgrading RAM later is easy on most cloud providers.
:::

---

## Using the Resource Calculator

The [calculator](/calculator) estimates **CPU, RAM and disk** from the options you choose:

- **Number of users** and the **CPU / RAM limit per user** from your hosting plans. The limits give the *maximum* if every user maxes out at once; the recommendation is based on typical usage.
- **Database** (MySQL or MariaDB), **web server** (Nginx, Apache, OpenResty, OpenLiteSpeed) and number of **PHP versions** - each adds shared disk space, and extra PHP versions add a little RAM.
- **Caching** (Redis, Valkey, Memcached) and **cron**.
- **Add-ons**: DNS, phpMyAdmin, email, FTP, ClamAV and ImunifyAV. **ClamAV** is the heaviest (about 4 GB of RAM) - use ImunifyAV if you're short on memory.

It also shows whether the setup fits the free **Community** edition or needs **Enterprise**. You can share the result - the settings are saved in the URL.

---

## Other Things to Consider

- **Disk speed matters more than size.** Use SSD or NVMe storage; databases and PHP are slow on HDDs.
- **Swap.** The installer creates a swap file automatically on servers with 32 GB of RAM or less. Swap prevents crashes during memory spikes but is not a replacement for RAM. See [How to Add Swap](/docs/articles/server/how-to-add-swap/).
- **Backups** need extra disk space (or remote storage) - see [Configuring Backups](/docs/articles/backups/configuring-backups/).
- **Email** storage grows over time. Plan mailbox quotas, or [store email on a remote server](/docs/articles/email/how-to-use-remote-server-for-emails-in-openpanel/).

---

## Related

- [Install OpenPanel](/docs/articles/install-update/install/)
- [How to find which site is using high CPU or RAM](/docs/articles/server/high-cpu-ram-usage/)
- [Suggested hosting plan limits](/docs/articles/hosting-business/hosting-plan-templates/)
