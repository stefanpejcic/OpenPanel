---
sidebar_position: 4
---

# Process Manager

The Process Manager allows you to view and manage running system processes directly from the OpenPanel interface. It's a powerful tool for monitoring system activity and taking quick action when needed.

The process table includes the following details:

- **PID** – Process ID. Options are available to trace the process or kill it directly.
- **Owner** – The user that owns the process.
- **Priority** – The process's current scheduling priority.
- **CPU** – Percentage of CPU usage.
- **Memory** – Percentage of Memory usage.
- **Command** – The command or service that launched the process.

![Process Manager listing system processes with a checkbox, PID, owner, priority, CPU and memory usage, command, and Trace and Kill links](/img/openadmin-screenshots/advanced/processes-list.png#gh-light-mode-only)
![Process Manager listing system processes with a checkbox, PID, owner, priority, CPU and memory usage, command, and Trace and Kill links](/img/openadmin-screenshots/advanced/processes-list_dark.png#gh-dark-mode-only)

Use this tool to identify resource-heavy or suspicious processes, manage system load, or terminate unresponsive services.

Processes using 90% or more CPU, or 20% or more RAM, are listed above the table:

- for a process in an OpenPanel user's container (owned by the user, or by a UID in their range in `/etc/subuid`), it links to the user and recommends checking their container limits for a misconfiguration or their sites for unusual traffic.
- for a system process, it links to trace the process to see what it's doing, or recommends restarting its service.

CPU usage is measured between page loads, so a process using a lot of CPU is listed from the second time the page is opened.

## Bulk Actions

Tick the checkbox of one or more processes, or the checkbox in the table header to select all processes shown by the current search. A bar appears at the bottom of the page with the number selected, a **Clear** link and these actions:

![Two processes selected with the bulk actions bar offering Kill](/img/openadmin-screenshots/advanced/processes-bulk.png#gh-light-mode-only)
![Two processes selected with the bulk actions bar offering Kill](/img/openadmin-screenshots/advanced/processes-bulk_dark.png#gh-dark-mode-only)

| Action | What it does |
|---|---|
| **Kill** | Kills the selected processes. |

The same **Kill** bulk action is on the **Tasks** page for running background tasks.

Click an action, confirm it, and it runs on the selected processes one after another. When it's done the page reloads with a notice listing the processes it worked for, or which ones failed and why.
