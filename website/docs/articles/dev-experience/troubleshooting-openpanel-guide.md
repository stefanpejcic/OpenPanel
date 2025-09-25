# Troubleshooting OpenPanel

When troubleshooting OpenPanel, the first step is to **reproduce the issue**. If you can confirm it’s a bug (and like all software, bugs do happen), try to identify a consistent way to trigger it.

* If the issue is **reproducible**, please [report it on GitHub Issues](https://github.com/stefanpejcic/OpenPanel/issues/new/choose) so we can prioritize it for the next update.
* If you’d like to investigate further on your own, follow the steps below.

---

## Step 1: Enable Developer Mode

Run the following command to turn on verbose debugging for both **OpenPanel** and **OpenAdmin** interfaces:

```bash
opencli config update dev_mode on
```

---

## Step 2: Check Logs

Reproduce the issue in the UI, then review the logs at the same time to spot errors.

* **OpenPanel logs:**

  ```bash
  docker logs -f openapanel
  ```
* **OpenAdmin logs:**

  ```bash
  tail -f /var/log/openpanel/admin/error.log
  ```

---

## Step 3: Disable Developer Mode

After finishing, disable developer mode to prevent unnecessary log growth:

```bash
opencli config update dev_mode off
```

