# License

`opencli license` command is used to set a license key for [OpenPanel Enterprise](https://openpanel.com/enterprise/), verify it, display information and delete the key to downgrade to Community edition.


View available options:
```bash
opencli license
```

Adding a license key:
```bash
opencli license <KEY>
```

Additional flags are available:

- `--json` - displays response as json.
- `--no-restart` - skips restarting OpenAdmin interface after adding a key.

View license key:
```bash
opencli license key
```

View license information:
```bash
opencli license info
```
Delete license key:
```bash
opencli license delete
```
