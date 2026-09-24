# Error 

`opencli error` command allows Administrators to trace OpenPanel error IDs.

```bash
opencli error ID_HERE
```

The command searches the last 60 minutes of the OpenPanel container logs (`podman logs openpanel`) for the error ID and displays the related log lines.

<details>
  <summary>Example output</summary>

```bash
# opencli error 9RIdD2X6aGaIzCzAEh0YL57U

=== LOGS FOR ERROR ID: '9RIdD2X6aGaIzCzAEh0YL57U' ===

[2026-09-24 14:14:47] ERROR Exception on /emails/info@example.com [GET]
...
```
</details>

If the ID is not found in the logs:

<details>
  <summary>Example output</summary>

```bash
# opencli error 9RIdD2X6aGaIzCzAEh0YL57U

=== NO LOGS FOR ERROR ID ===
Error Code '9RIdD2X6aGaIzCzAEh0YL57U' not found in the OpenPanel UI logs.
```
</details>
