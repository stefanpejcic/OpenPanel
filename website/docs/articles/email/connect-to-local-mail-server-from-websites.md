---
sidebar_label: "Connect websites to the local mail server"
---

# How to Connect Websites to the Mail Server on the Same Server

Some PHP applications log in to a mailbox or send mail through SMTP on their own: for example **WHMCS** ticket importing (POP3/IMAP), WordPress SMTP plugins, or contact form plugins. If the mail server runs on the same server as the website, the connection can fail even though the same settings work from an email client on your computer.

Common errors:

```
cannot connect to host ; error = stream_socket_client(): Unable to connect to ssl://mail.example.com:995 (Connection refused)
```

```
stream_socket_client(): Unable to connect to ssl://host.containers.internal:995 (Unknown error) (errno = 0)
```

## Why this happens

Each OpenPanel user's services (PHP, web server, database) run in rootless containers. From inside these containers, the server's own public IP address does not lead back to the host, so connecting to `mail.example.com` (which resolves to the server's public IP) is refused on every mail port.

The host can be reached from inside the containers at the special name `host.containers.internal`. But the mail server's SSL certificate is issued for `mail.example.com`, not for `host.containers.internal`, so applications that check the certificate (like WHMCS) reject that connection with `Unknown error`.

The fix is to keep using `mail.example.com` in the application, and make that name point to the host from inside the user's PHP containers.

## Step 1: Find the host address

Replace `USERNAME` with the OpenPanel username and `php-fpm-8.3` with the PHP version the website uses:

```bash
USERNAME=USERNAME
sudo -u $USERNAME XDG_RUNTIME_DIR=/run/user/$(id -u $USERNAME) \
  podman exec php-fpm-8.3 getent hosts host.containers.internal
```

Output:

```
169.254.1.2     host.containers.internal host.docker.internal
```

Note the IP address (usually `169.254.1.2`).

## Step 2: Add the mail hostname to all PHP services

First check that the PHP services don't already have an `extra_hosts` section:

```bash
grep -n "extra_hosts" /home/$USERNAME/docker-compose.yml
```

If there is no output, back up the file and add the mail hostname to every `php-fpm-*` service. Replace `mail.example.com` with your mail server hostname and `169.254.1.2` with the IP address from Step 1:

```bash
cp -a /home/$USERNAME/docker-compose.yml /home/$USERNAME/docker-compose.yml.bak-mailhosts

awk -v entry="mail.example.com:169.254.1.2" '
  { print }
  /^[[:space:]]+container_name: php-fpm-/ {
    match($0, /^[[:space:]]+/); i = substr($0, 1, RLENGTH)
    print i "extra_hosts:"
    print i "    - \"" entry "\""
  }' /home/$USERNAME/docker-compose.yml > /tmp/compose.new \
  && cat /tmp/compose.new > /home/$USERNAME/docker-compose.yml && rm /tmp/compose.new
```

Each PHP service now contains:

```yaml
    php-fpm-8.3:
        command: php-fpm --allow-to-run-as-root
        container_name: php-fpm-8.3
        extra_hosts:
            - "mail.example.com:169.254.1.2"
```

:::info
If a PHP service already has an `extra_hosts` section, don't run the command. Add the `- "mail.example.com:169.254.1.2"` line to the existing list by hand instead, so the service doesn't end up with two `extra_hosts` keys.
:::

## Step 3: Recreate the PHP container

Recreate the PHP version the website uses so it picks up the change:

```bash
sudo -u $USERNAME XDG_RUNTIME_DIR=/run/user/$(id -u $USERNAME) \
  bash -c "cd /home/$USERNAME && podman-compose up -d --force-recreate --no-deps php-fpm-8.3"
```

Other PHP versions get the change the next time they are recreated, for example when a domain is switched to that version.

## Step 4: Test the connection

Check that the hostname now points to the host:

```bash
sudo -u $USERNAME XDG_RUNTIME_DIR=/run/user/$(id -u $USERNAME) \
  podman exec php-fpm-8.3 getent hosts mail.example.com
```

The first line should show `169.254.1.2`.

Then test an SSL connection from PHP, with certificate checking turned on (use `993` for IMAP, `995` for POP3, `465` for SMTP):

```bash
sudo -u $USERNAME XDG_RUNTIME_DIR=/run/user/$(id -u $USERNAME) \
  podman exec php-fpm-8.3 php -r '$s = stream_socket_client("ssl://mail.example.com:995", $e, $es, 5); echo $s ? fgets($s) : "FAIL: $es\n";'
```

A working connection prints the mail server greeting:

```
+OK Dovecot (Debian) ready.
```

## Step 5: Use the settings in your application

In the application, use your normal mail settings with the real hostname. For WHMCS ticket importing (**Setup > Support > Support Departments**), for example:

| Field    | Value                                       |
|----------|---------------------------------------------|
| Hostname | `mail.example.com`                          |
| Port     | `995` (POP3 over SSL)                       |
| Email / Username | the full email address, for example `info@example.com` |
| Password | the mailbox password                        |

Use `995` with WHMCS: its standard ticket importing uses POP3, so it connects to port `995` even if you enter the IMAP port `993`.

If the connection now works but login fails, check the password. Failed logins appear in the mail server log as `Password mismatch`:

```bash
podman exec openadmin_mailserver grep "Password mismatch" /var/log/mail/mail.log | tail
```

## Apply to all new users (optional)

To add the mail hostname to the PHP services of every new account, run the same `awk` command from Step 2 on the compose template for new users, `/etc/openpanel/docker/compose/1.0/docker-compose.yml`. Existing users are not affected.
