# License

`opencli license` command is used to set a license key for [OpenPanel Enterprise](https://openpanel.com/enterprise/), verify it, display information and delete the key to downgrade to Community edition.


View available options:
```bash
opencli license
```

<details>
  <summary>Example output</summary>

```bash
# opencli license
Usage: opencli license [options]
Commands:
  key                                           View current license key.
  enterprise-XXXXXXXXXX                         Save an Enterprise license key.
  noc-XXXXXXXXXX                                Save a NOC license key.
  lifetime-XXXXXXXXXX                           Save a Lifetime license key.
  verify                                        Verify the license key.
  info                                          Display information about the license owner and expiration.
  delete                                        Delete the license key and downgrade OpenPanel to Community edition.
```
</details>

Adding a license key (keys start with `enterprise-`, `noc-` or `lifetime-`):
```bash
opencli license <KEY>
```

<details>
  <summary>Example output</summary>

```bash
# opencli license enterprise-a1b2c3d4e5
License key enterprise-a1b2c3d4e5 added.
OpenPanel and OpenAdmin are restarted to apply Enterprise features.
```
</details>

Additional flags are available:

- `--json` - displays response as json.
- `--no-restart` - skips restarting OpenAdmin interface after adding a key.

View license key:
```bash
opencli license key
```

<details>
  <summary>Example output</summary>

```bash
# opencli license key
enterprise-a1b2c3d4e5
```
</details>

Verify the current license key:
```bash
opencli license verify
```

<details>
  <summary>Example output</summary>

```bash
# opencli license verify
License is valid
OpenPanel and OpenAdmin are restarted to apply Enterprise features.
```
</details>

View license information:
```bash
opencli license info
```

<details>
  <summary>Example output</summary>

```bash
# opencli license info
Owner: Stefan Pejcic
Company Name: OpenPanel
Email: stefan@example.com
License Type: OpenPanel Enterprise
Registration Date: 2026-01-15
Next Due Date: 2026-10-15
Billing Cycle: Monthly
Valid IP: 203.0.113.10
```
</details>

Delete license key:
```bash
opencli license delete
```
The key is removed and OpenAdmin is restarted. The command prints no output.
