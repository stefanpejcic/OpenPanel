# Migrate server

OpenPanel is a truly OS-agnostic hosting panel, meaning it runs seamlessly on any Linux distribution. This makes migrating the panel along with all user data - from one server to another straightforward and efficient.

To migrate all data from a server to another:

## Prepare the New Server

The new server should have a clean environment with **only OpenPanel installed**-no existing users or additional configurations.

Install OpenPanel by running:

```bash
bash <(curl -sSL https://openpanel.org)
```

Wait for the installation to complete, then proceed to the source server.

## Migrate from the Old Server

Migration can be performed via OpenAdmin UI or from the terminal:

### Using OpenAdmin

On the old server, login to **OpenAdmin** and navigate to *Advanced > Migration*, provide the new server’s IP address and root credentials:

Click on the 'Start Migration' button and wait for the process to complete:

### Using Terminal

```bash
opencli server-migrate -h NEW_SERVER_IP --user root --password NEW_SERVER_ROOT_PASSWORD
```

Allow the migration process to finish.

## Verify Migration

After completion, log in to the OpenAdmin panel on the new server using the same credentials you used on the old server.

Verify that all services have started correctly in OpenAdmin under Service Status, and test a few user websites to ensure they are functioning properly.

If everything looks good, proceed to update the DNS records for the domains to point to the new server.

If something is not running, check the migration log on source server for any errors: `/tmp/server_migrate.log`
