---
sidebar_position: 28
---

import HookGenerator from "@site/src/components/hook-generator";

# Hooks

OpenPanel supports hooks: your own bash scripts that run before or after an opencli command.

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

Any arguments passed to the opencli command are also passed to your hook script, in the same order: `$1` is the first argument, `$2` the second, and `"$@"` holds all of them, flags included.

- A **pre** hook runs with `bash` and the command waits for it to finish. Its exit code doesn't stop the command.
- A **post** hook runs in the background after the command, so the command doesn't wait for it.
- The hook's output isn't shown, write to a log file if you need it.
- When the command is started from OpenPanel (the user panel), the hook runs inside the `openpanel` container, where `/etc/openpanel` is mounted.

## Generator

Pick the command and when the hook should run, write what it should do, and copy the result into the file.

<HookGenerator />
