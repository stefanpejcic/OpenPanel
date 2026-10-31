---
sidebar_position: 3
---

# Work With All Servers

## Switch servers

Click the server name at the top of the sidebar and pick a server. Every page you open after that is the page of that server, served through the master. A red dot next to a server means the master can't reach it.

Pages that only make sense for one server, like logs, services or resource usage, also have a server picker, so you can look at another server without switching.

![Server menu at the top of the sidebar listing the master and two slaves](/img/openadmin-screenshots/multi-server/server-menu.png#gh-light-mode-only)
![Server menu at the top of the sidebar listing the master and two slaves](/img/openadmin-screenshots/multi-server/server-menu_dark.png#gh-dark-mode-only)

## Data from all servers

On the master, lists show the rows of every server:

- Tables like users, domains, websites, plans, email accounts, FTP accounts and notifications get a **Server** column and a filter next to the search box. Click the column header to sort by server.
- Counts above tables, like "8 plans across 3 servers", count all servers.
- The dashboard totals, the daily summary and the **Servers** widget cover all servers.
- **Services** shows the status of each service on every server, with **RESTART NEEDED** in the column of the server that needs it. When several servers need the same restart, **Restart on all servers** under the service name restarts it everywhere at once.
- The resource usage in the header shows the highest value of all servers, and the tooltip lists each server.
- Search finds pages and accounts on every server. Accounts from another server are shown with its name, like `john (server2)`.
- Notifications from all servers are listed together, and **Acknowledge all** applies to every server.

![Users list with accounts from all servers and a Server column and filter](/img/openadmin-screenshots/multi-server/users-all-servers.png#gh-light-mode-only)
![Users list with accounts from all servers and a Server column and filter](/img/openadmin-screenshots/multi-server/users-all-servers_dark.png#gh-dark-mode-only)

![Services page with a status column for each server](/img/openadmin-screenshots/multi-server/services-all-servers.png#gh-light-mode-only)
![Services page with a status column for each server](/img/openadmin-screenshots/multi-server/services-all-servers_dark.png#gh-dark-mode-only)

![Search results with users from every server and the server name under each](/img/openadmin-screenshots/multi-server/search-all-servers.png#gh-light-mode-only)
![Search results with users from every server and the server name under each](/img/openadmin-screenshots/multi-server/search-all-servers_dark.png#gh-dark-mode-only)

![Cluster widget on the dashboard with each server, its users, domains and version](/img/openadmin-screenshots/multi-server/dashboard-cluster-widget.png#gh-light-mode-only)
![Cluster widget on the dashboard with each server, its users, domains and version](/img/openadmin-screenshots/multi-server/dashboard-cluster-widget_dark.png#gh-dark-mode-only)

## Actions on rows of other servers

Links, menus and buttons in a row act on the server the row belongs to. Opening a user from another server switches to that server. Bulk actions split the selection and run on each row's own server.

When creating something, like a user, administrator or reseller, pick the server in the form.

## Apply on all servers

Pages that save settings have an **Apply on all servers** box next to the **Save** button. When it's checked, the master saves the change on the current server first and then applies the same change on the others. A notice at the top of the page shows the result for each server.

The box is checked by default. To use a different value on one server, clear the box, or log in on that server directly and change it there.

![Server picker and Apply on all servers box next to the Save button](/img/openadmin-screenshots/multi-server/apply-all-servers.png#gh-light-mode-only)
![Server picker and Apply on all servers box next to the Save button](/img/openadmin-screenshots/multi-server/apply-all-servers_dark.png#gh-dark-mode-only)

Some settings always stay on one server:

- Domain, ports and SSL under **Settings > General**.
- License.
- 2FA and passkeys of your own account.

The timezone, root password, SSH and cron pages start with the box cleared, since those usually differ per server.

These changes always go to every server:

- WAF profiles and rules under **Security > WAF**.
- Updating or rolling back from **Settings > Updates**.
