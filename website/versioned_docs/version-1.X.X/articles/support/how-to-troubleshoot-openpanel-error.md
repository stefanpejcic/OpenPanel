# Troubleshooting OpenPanel UI Errors

## 500 Error

If a **500 error** occurs in the OpenPanel UI, a unique error code will be displayed on the page.

> ⚠️ This error code is specific to your machine. Only the server administrator can access detailed information about it.

To view the full error details, run the following command in your terminal:

```bash
opencli error ERROR_ID_HERE
```

Check the output for the error message. If you need assistance, you can copy the message to our [support forums](https://community.openpanel.org/) or [Discord channel](https://discord.openpanel.com/) for help troubleshooting.

---

## UI Not Responding

If a feature isn’t working as expected (like clicking a button with no response) it is likely a **front-end issue**.

To troubleshoot:

1. Open your browser’s **Developer Tools** (usually `F12` or `Ctrl+Shift+I` / `Cmd+Option+I`).
2. Navigate to the **Network** tab.
3. Repeat the action that’s not working.
4. Check for requests in the **Network** tab and any errors in the **Console** log.

* Selecting a request in the Network tab allows you to view the response returned by the backend.
* If the response doesn’t contain enough information to diagnose the issue, enable **dev_mode** on the server and check the Docker logs.

---

## Dev Mode

Enabling **dev_mode** allows OpenPanel and OpenAdmin interfaces to log detailed debugging information for every request. This helps administrators see the exact commands run by the panel and the responses received.

To enable dev_mode:

```bash
opencli config update dev_mode on
```

Then restart the panel.

When dev_mode is enabled, detailed logs for the user panel are available via:

```bash
docker logs -f openpanel
```

These logs provide verbose debugging information for troubleshooting.

