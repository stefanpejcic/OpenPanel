---
sidebar_position: 5
---

# Tasks

The **Server > Tasks** page lists the background tasks of OpenPanel: the `opencli` commands that are running right now and the ones scheduled to run from [Scheduled Actions](/docs/admin/advanced/crons/). The list refreshes on its own every 5 seconds.

Each row shows:

- **Command** – the `opencli` command.
- **Status** – **Running** for a command that is running now, **Scheduled** for one that runs from cron.
- **Time** – how long a running command has been running, or when a scheduled one runs next (for example `in 3h 30m`).
- **Actions** – **Trace** shows what a running command is doing and **Kill** stops it. **Edit** opens the job on the Scheduled Actions page, where its schedule can be changed.

Click the **Command**, **Status** or **Time** column to sort the list by it, and click again to reverse the order. The sort stays when the list refreshes. Sorting by **Time** puts the running commands first, then the scheduled ones in the order they run.

:::info
Only `opencli` commands are listed. To see every process on the server, use the [Process Manager](/docs/admin/advanced/processes/).
:::

## Multi-Server

On a [multi-server](/docs/admin/system/multi-server/) master, the page lists the tasks of every server and adds a **Server** column:

- A cron job that runs the same command on the same schedule on several servers is shown once, with all of those servers in the **Server** column. If the schedule differs on a server, that server gets its own row.
- **Trace** and **Kill** act on the server the command runs on.
- Use the **Server** filter above the table to show the tasks of one server, or sort by the **Server** column.
