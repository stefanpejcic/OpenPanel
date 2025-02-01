# How to fix: failed to create runc console socket: stat /tmp

When running command: 

```bash
opencli user-login stefan
```

If you receive an error:
> failed to create runc console socket: stat /tmp: no such file or directory: unknown

It indicates that docker runtime files from the /tmp are missing (deleted by some program/user), so the quickest solution is to reboot the server.
