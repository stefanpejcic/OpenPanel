---
sidebar_position: 6
---

# Manage Servers

The **Servers** table on **System > Multi-Server** lists the master and every slave with its users, load, RAM, disk and OpenPanel version.

![Servers table with Open, Make master, Pause and Remove for each slave](/img/openadmin-screenshots/multi-server/servers-table.png#gh-light-mode-only)
![Servers table with Open, Make master, Pause and Remove for each slave](/img/openadmin-screenshots/multi-server/servers-table_dark.png#gh-dark-mode-only)

- **Open** switches to the server.
- **Update to** appears when a newer OpenPanel version is available, and updates that server.
- **Pause** (pause icon) leaves a slave out of everything until you click **Resume**: lists, search, server pickers, settings applied to all servers, sync and moves. It still shows in the sidebar menu, greyed out with "paused", and its load stays in the header resource usage. Use it to change settings on all other servers while keeping one as it is. After resuming, sync it if it should pick up what changed meanwhile.
- **Remove** (trash icon) takes a slave out of the cluster. The slave becomes standalone and keeps all its accounts and settings.

## Make another server the master

Click **Make master** next to a slave. That server takes over the cluster, and the old master becomes one of its slaves. Afterwards you manage the cluster from the new master with its own admin login, so make sure you have one there first.

Both servers need to be online. If the master is lost, use **Leave cluster** on each slave and add them to a new master.

## Changing ports or addresses

If the admin port of the master changes, the slaves learn the new address on the next check, within a minute. If a slave's admin port changes, remove it on the master and add it again with the new address, **Settings > General** on the slave reminds you of this.

## Notifications

Cluster events show on the **Notifications** page of the server they happen on:

- On the master: a server added, removed, paused or resumed, settings synced to servers, account moves, updates started, a new master, and a server going offline or coming back.
- On a slave: joining or leaving the cluster, and settings synced from the master.

Only problems, like a server going offline or a failed sync, are also sent by email or webhook.

## Security

- Use a strong password and 2FA for every admin on the master, since it can reach all servers.
- Removing a server or leaving the cluster deletes the trust between them right away.

## From the terminal

On the master, OpenCLI lists accounts, domains and plans from all servers, runs account commands on the server that has the account, and takes `--server=<NAME>` and `--all-servers` for the rest. See [OpenCLI on Multi-Server](/docs/articles/opencli/#multi-server).

```bash
opencli user-list
opencli user-add bob generate bob@example.com "Standard plan" --server=server2
opencli version --all-servers
```
