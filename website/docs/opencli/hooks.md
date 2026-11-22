## Hooks

OpenPanel supports hooks that run bash scripts before or after running opencli commands.

To create a hook, first create a new directory: `/etc/openpanel/openpanel/hooks/` and inside create a file based on the desired command:

- `pre_` prefix for script to run **before** a command.
- `post_` prefix for a script to run **after** executing an opencli command.

Example, to run a custom script before the user creation process (opencli *user-add*) you would create a new file:
```bash
pre_user-add.sh
```

another example, script that executes after adding a domain (opencli *domains-add*) you would create file:
```bash
post_domains-add.sh
```

any command-line attributes passed to opencli script will also be to your custom scripts.
