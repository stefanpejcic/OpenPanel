---
sidebar_label: "Add or Resize Swap"
description: "How to add, resize or clear swap on an OpenPanel server: from OpenAdmin, during installation or from the terminal, and how the Sentinel monitoring script checks swap usage and clears it automatically."
---

# How to Add or Resize Swap on Your Server

**Swap** is disk space the server uses as overflow memory when RAM is full. It keeps the server from killing processes (and taking websites down) during short memory spikes. It is much slower than RAM, so it's a safety net - not a replacement for more memory.

OpenPanel creates a swap file automatically during installation and monitors it afterwards. This guide shows how to check, resize and clear swap.

---

## How Much Swap Do You Need?

The installer creates `/swapfile` automatically, sized by RAM:

| Server RAM | Swap created |
|---|---|
| up to 2 GB | 2 × RAM |
| 2-4 GB | same as RAM |
| 4-8 GB | 4 GB |
| 8-32 GB | 8 GB |
| more than 32 GB | none |

That's a good starting point. If the server regularly uses a lot of swap, add RAM instead of more swap - see [System requirements and sizing](/docs/articles/install-update/system-requirements/).

To choose the size yourself when installing, use the `--swap` flag (1-10 GB):

```bash
bash <(curl -sSL https://openpanel.org) --swap=4
```

---

## From OpenAdmin

Go to **OpenAdmin → System → Swap**:

- **Current usage** - total, used and free swap, and every active swap device.
- **Change allocation** - enter a new size to recreate `/swapfile` at that size. It's added to `/etc/fstab` so it survives reboots. Swap is briefly turned off while the file is recreated.
- **Drop swap** - moves swapped-out data back into RAM (`swapoff -a; swapon -a`) without changing the size.

See [Swap](/docs/admin/system/swap/).

---

## From the Terminal

Check the current swap:

```bash
free -h
swapon --show
```

Create or resize the swap file (example: 4 GB):

```bash
swapoff /swapfile 2>/dev/null
fallocate -l 4G /swapfile
chmod 600 /swapfile
mkswap /swapfile
swapon /swapfile
grep -q '^/swapfile' /etc/fstab || echo '/swapfile none swap sw 0 0' >> /etc/fstab
```

Clear swap (only when enough RAM is free to hold what's in swap):

```bash
swapoff -a && swapon -a
```

### Swappiness

`vm.swappiness` controls how eagerly Linux moves memory to swap (default `60`). On a web server, a lower value keeps active data in RAM longer:

```bash
sysctl vm.swappiness=10
echo 'vm.swappiness=10' > /etc/sysctl.d/99-swappiness.conf
```

---

## How OpenPanel Monitors Swap

The **Sentinel** monitoring script (`opencli sentinel`) runs every 5 minutes and checks swap usage against the **SWAP %** threshold in **OpenAdmin → Settings → Notifications → Resource Usage** (default **85%**):

1. **Swap is below the threshold** - nothing happens.
2. **Swap is above the threshold** - Sentinel sends a *High SWAP usage* notification, then clears the page cache and swap (`swapoff -a; swapon -a`), and reports the new usage.
3. **Not enough free RAM to clear it safely** - Sentinel skips the cleanup and notifies you instead, so clearing swap can never run the server out of memory.
4. **Swap is turned off but configured in `/etc/fstab`** - Sentinel turns it back on (`swapon -a`) and notifies you.

Change the threshold from OpenAdmin, or from the terminal:

```bash
opencli admin notifications update swap 70
```

Run the checks manually with:

```bash
opencli sentinel
```

See [Notifications settings](/docs/admin/settings/notifications/).

---

## Swap Keeps Filling Up?

Frequent high swap usage means the server needs more memory than it has. Find out what's using it:

```bash
grep VmSwap /proc/*/status 2>/dev/null | sort -k2 -hr | head
```

Then see [Find which site is using high CPU or RAM](/docs/articles/server/high-cpu-ram-usage/) - often a single account, a leaking app or a service like ClamAV. Lower that account's memory limit, or upgrade the server's RAM.

---

## Related

- [Free up disk space](/docs/articles/server/how-to-free-up-disk-space-on-linux/) - swap files use disk space too
- [System cron jobs](/docs/articles/server/openadmin-system-crons/)
