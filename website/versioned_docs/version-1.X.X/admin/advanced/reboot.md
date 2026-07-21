---
sidebar_position: 4
---

# Server Reboot

Initiate a reboot of the server. Depending on the situation, you can perform either a **graceful reboot** or a **forceful reboot**.

- **Graceful Server Reboot** attempts to safely stop running processes and allow the operating system to cleanly shut down services before restarting.
- **Forceful Server Reboot** immediately restarts the system at the kernel level and should only be used if the server is unresponsive.

## Graceful

A graceful reboot tells the operating system to restart normally. Services and user processes are given a chance to shut down cleanly.

Command used:

```bash
reboot
```

## Forceful

A forceful reboot bypasses the normal shutdown procedure and immediately reboots the kernel. This may result in **data loss** or **filesystem inconsistencies**, so it should only be used when the server is completely unresponsive.

Command used:
```bash
echo b > /proc/sysrq-trigger
```

:::danger
⚠️ Use with caution. This command forces an immediate reboot without syncing disks or safely stopping processes.
:::
