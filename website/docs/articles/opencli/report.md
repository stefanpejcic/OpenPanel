# Check server info

To check current server info you can use the following command:

```bash
opencli report
```

<details>
  <summary>Example output</summary>

```bash
# opencli report
Information collected successfully. Please provide content of the following file to the support team:
/var/log/openpanel/admin/reports/system_info_20260924101542.txt
```
</details>

To upload the report to support.openpanel.org and get a key to share when asking for support:
```bash
opencli report --public
```

<details>
  <summary>Example output</summary>

```bash
# opencli report --public
Information collected successfully. Please provide the following key to the support team:
XXXXXXXXXX
```
</details>

`--link` and `--upload` are aliases for `--public`.

For plain output without progress messages, screen clearing or colors (useful in scripts):
```bash
opencli report --non-interactive
```

<details>
  <summary>Example output</summary>

```bash
# opencli report --non-interactive
Information collected successfully. Please provide content of the following file to the support team:
/var/log/openpanel/admin/reports/system_info_20260924101542.txt
```
</details>

The report is saved to `/var/log/openpanel/admin/reports/`.
