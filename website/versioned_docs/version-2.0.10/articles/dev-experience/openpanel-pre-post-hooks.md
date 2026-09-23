# OpenCLI Hooks

OpenPanel supports **pre** and **post hooks** for all commands, that run bash scripts before or after running opencli commands.

To create a hook, first create a new directory: `/etc/openpanel/openpanel/hooks/` and inside create a file based on the desired command:

- `pre_` prefix for script to run **before** a command.
- `post_` prefix for a script to run **after** executing an opencli command.

## Examples

To run a custom script before the user creation process (opencli *user-add*) you would create a new file:
```bash
/etc/openpanel/openpanel/hooks/pre_user-add.sh
````


Another example to run **after a domain is added**:

When a user adds a new domain via the UI, the underlying command executed is:

```bash
/etc/openpanel/openpanel/hooks/post_domains-add.sh
```

These scripts will be executed automatically when the command runs, depending on whether you're hooking into the pre- or post-execution phase.

## Passing Arguments to Hooks

All arguments passed to the `opencli` command are also forwarded to your hook script.

For example, if the following command is executed:

```bash
opencli domains-add example.com stefan --docroot /var/www/example --php_version 8.1 --skip_caddy --skip_vhost --skip_containers --skip_dns --debug
```

Your hook script will receive the same arguments:

```bash
bash /etc/openpanel/openpanel/hooks/post_domains-add.sh example.com johndoe --docroot /var/www/example --php_version 8.1 --skip_caddy --skip_vhost --skip_containers --skip_dns --debug
```
