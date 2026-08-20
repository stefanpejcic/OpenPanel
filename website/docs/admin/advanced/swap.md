---
sidebar_position: 4
---

# Swap

*OpenAdmin > Server > Swap* lets Administrators view current swap usage, change the swap file allocation, and drop (clear) swap.

### Current usage

Shows total/used/free swap (as reported by `free -m`) and a table of every active swap device (name, type, size, used, priority, from `swapon --show`). The alert threshold shown here is the same `swap=` value used by Sentinel's own swap check.

### Change allocation

Recreates the managed swap file (`/swapfile` by default) at the requested size and re-enables it, adding it to `/etc/fstab` if it isn't already there. This briefly disables swap while the file is recreated.

### Drop swap

Runs `swapoff -a; swapon -a`, the same cleanup Sentinel performs automatically when swap usage crosses its threshold. This moves swapped-out pages back into RAM without changing the swap size.
