---
sidebar_position: 5
---

# Swap

*OpenAdmin > System > Swap* lets Administrators view current swap usage, change the swap file allocation, and drop (clear) swap.

Step-by-step guide: [How to Add or Resize Swap](/docs/articles/server/how-to-add-swap/)

### Current usage

Shows total, used (with the percentage) and free swap in GB (as reported by `free -m`) and a table of every active swap device (name, type, size, used, priority, from `swapon --show`). The alert threshold shown here is the same `swap=` value used by Sentinel's own swap check.

When swap is more than 90% used, a warning is shown: the server is short on memory, so add RAM, lower memory usage on [Processes](/docs/admin/advanced/processes), or increase swap.

![Swap page with the current swap usage, the swap size field with Apply, and the Drop Swap button](/img/openadmin-screenshots/advanced/swap-page.png#gh-light-mode-only)
![Swap page with the current swap usage, the swap size field with Apply, and the Drop Swap button](/img/openadmin-screenshots/advanced/swap-page_dark.png#gh-dark-mode-only)

### What's in swap

Lists the 10 processes with the most memory swapped out: process name, PID, owner and how much of it is in swap (the `VmSwap` value from `/proc/<pid>/status`). When a process belongs to an OpenPanel user's containers, the owner links to that user. The section is hidden when nothing is swapped out.

### Change allocation

Enter the new size in GB (at least 0.125 GB) under **Swap size** and click **Apply**. OpenAdmin asks you to confirm, then turns off the managed swap file (`/swapfile` by default), recreates it at the requested size and re-enables it, adding it to `/etc/fstab` if it isn't already there.

Turning the swap file off moves everything in it back into RAM first, so this takes longer the more is swapped out, and it's refused when what's in the swap file wouldn't fit in the available RAM (with 256 MB to spare). Free some memory first, or use [Processes](/docs/admin/advanced/processes) to find what's using it.

The largest size you can set is shown under the field: the current swap file plus the free disk space, minus 1 GB that always stays free. A larger size is refused both on the page and by the server.

A recommended swap size is shown when the size is smaller than it, or no swap is set up: 2× RAM on servers with up to 2 GB RAM, otherwise the square root of the RAM in GB (at least 2 GB), for example 4 GB for 16 GB RAM.

### Drop swap

Runs `swapoff -a; swapon -a`, the same cleanup Sentinel performs automatically when swap usage crosses its threshold. This moves swapped-out pages back into RAM without changing the swap size.

The section shows how much is in swap next to the available RAM. When the swapped-out memory wouldn't fit in the available RAM (with 256 MB to spare), the **Drop Swap** button is disabled and a warning is shown, and the server refuses the request too.

See also: [High CPU or RAM usage: find the culprit](/docs/articles/server/high-cpu-ram-usage/)
