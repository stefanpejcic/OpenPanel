---
sidebar_position: 10
---

# Remote PostgreSQL

Remote PostgreSQL access gives you the ability to connect to a PostgreSQL database on this server from another (remote) device or location over the internet.

Allowing remote PostgreSQL access opens your database to connections from the entire internet, which may pose a security risk. Please consider the following:

- Security Vulnerabilities: Allowing access from the web can expose your database to potential security vulnerabilities, increasing the risk of unauthorized access, data breaches, and data loss.
- Data Privacy: Your sensitive data may be at risk if not properly secured. Make sure to use strong passwords and encryption to protect your information.
- Regular Backups: Ensure that you have regular database backups in place to recover data in case of any security incidents.
- Before enabling remote PostgreSQL access, please review your security settings, and consider the potential risks carefully.

If you're unsure about the security implications or need assistance, consult with your hosting administrator.

---

## Enable remote PostgreSQL access

Remote PostgreSQL access is disabled by default.

To enable remote access to your databases, click on the "Enable Remote PostgreSQL Access" button in the "PostgreSQL > Remote Access" page.

Once enabled, the page will show two sets of connection details:

- **Remote** - the server IP and port to use when connecting **from a remote server** over the internet.
- **Local** - the internal hostname (`postgres`) and default port (`5432`) to use when connecting **from a local server** inside the same account.

:::info
The remote port is unique to your PostgreSQL instance. Avoid using the standard port `5432` for remote access, as it will not function.
:::

---

## Test remote PostgreSQL access

To test remote PostgreSQL access from another server, you can use the `psql` command-line client or a database management tool.

### psql command-line client

Ensure that you have the PostgreSQL client installed on the server from which you want to access the PostgreSQL database.

Use the following command to connect to the PostgreSQL server from the remote server:
```bash
psql -h <IP_ADDRESS> -p <PORT> -U <USERNAME> -d <DATABASE_NAME>
```

### PHP

To establish a PHP-based connection from a remote server to a PostgreSQL instance hosted on OpenPanel, here's an example code snippet:

```php
<?php
// PostgreSQL server details
$hostname = 'IP_ADDRESS_HERE';
$port = 'CUSTOM_PORT'; // Custom port from the OpenPanel interface
$username = 'YOUR_POSTGRESQL_USERNAME';
$password = 'YOUR_POSTGRESQL_PASSWORD';
$database = 'YOUR_POSTGRESQL_DATABASE';

// Create a connection
$conn = pg_connect("host=$hostname port=$port dbname=$database user=$username password=$password");

// Check the connection
if (!$conn) {
   die("Connection failed.");
}

echo "Connected to the PostgreSQL server on $hostname:$port successfully.";

// Close the connection
pg_close($conn);
?>
```

---

## Disable remote PostgreSQL access

If you wish to disable access, simply click on the "Disable Remote PostgreSQL Access" button, and it will immediately deactivate remote access in your PostgreSQL configuration. Please be aware that this action will also necessitate a PostgreSQL service restart to apply the new setting.
