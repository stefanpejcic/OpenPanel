---
sidebar_position: 9
---

# Change Root Password

This page lets you change the MySQL/MariaDB **root** user's password for your account's database service.

## Usage

1. Open **OpenPanel** and navigate to **MySQL > Change root password**.
2. Enter the new password.
3. Click **Save** (or equivalent) to apply it.

Behind the scenes, the panel updates the `root` user for both `%` (remote) and `localhost` hosts, flushes privileges, saves the new password to your account's `my.cnf`, and restarts the database container so the new credentials take effect immediately.

:::danger
There's no way to view the current root password from the panel — this page only sets a new one. If the root password is lost, this is the way to recover access.
:::
