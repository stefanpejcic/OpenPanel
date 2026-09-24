# Check server info

To check current server info you can use the following command:

```bash
opencli report 
```

To upload the report to support.openpanel.org and get a key to share when asking for support:
```bash
opencli report --public
```
`--link` and `--upload` are aliases for `--public`.

For plain output without progress messages, screen clearing or colors (useful in scripts):
```bash
opencli report --non-interactive
```

The report is saved to `/var/log/openpanel/admin/reports/`.
