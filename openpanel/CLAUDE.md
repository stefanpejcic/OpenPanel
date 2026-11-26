# Code comments

Write comments like a human dev jotting a quick note, not like documentation.

- Single line only. Don't wrap a comment across multiple lines unless the thing is genuinely too complex to say in one sentence.
- Keep it short and plain. No formal tone, no "Note:" prefixes, no " - " hyphen clauses stacked together, no restating what the code already makes obvious.
- Apply this whenever you write or edit comments in this repo, not just when asked to clean up comments specifically.

# Git

- After making changes, `git add` them and stop there. Don't `git commit` - the user commits themselves.

# Heavy commands

Commands the panel runs in its own container are shared by every user, so anything new that's disk or CPU heavy goes through `tenantfs.LowPrio(idle, name, args...)` in `internal/core/tenantfs`: `idle` true adds `ionice -c3` for scans of a whole home (find, du of everything), false gives only `nice -n19` for a single folder the user is waiting on. Don't hand-roll nice/ionice or use `ionice -c2 -n7`, mq-deadline ignores best-effort levels. `ionice` comes from `util-linux-misc` in the Dockerfile. Work done through `podman exec` in the user's own containers already counts against the user's limits and doesn't need it.
