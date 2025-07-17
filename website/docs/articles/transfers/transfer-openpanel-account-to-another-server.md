# Transfer account

This feature allows administrators to transfer (copy) individual accounts from one OpenPanel server to another.

To migrate all users and files at once, please refer to the, please refer to the [Server Migration Documentation](/docs/articles/transfers/migrate-openadmin-to-new-server/).

## How to Transfer an Account

### Using OpenAdmin

1. Navigate to **OpenAdmin > Users**.
2. Click on the user you want to transfer.
3. Under the **Transfer** section, fill in the remote server details:
   
   * **Server**: IP address or domain name of the remote server
   * **Port**: SSH port (usually `22`)
   * **Username**: SSH username for the remote server (usually `root`)
   * **Password**: SSH password for the specified user

   [![2025-07-17-18-32.png](https://i.postimg.cc/BvbcRp2k/2025-07-17-18-32.png)](https://postimg.cc/Y4cWW1nz)

5. (Optional) Enable the **Live Transfer** option.
   If checked, the account will be suspended on the current server after migration, and DNS will be forwarded to the new server.

6. Click **Start Transfer**.

Once initiated, the transfer process will begin. A log file will be generated - click on its name to view live progress.

---

Here’s a rewritten and polished version of your text:

---

### Using Terminal

To transfer an OpenPanel account to another server from the terminal, use the following command:

```bash
opencli user-transfer --account <OPENPANEL_USER> --host <DESTINATION_IP> --username <SSH_USERNAME> --password <DESTINATION_SSH_PASSWORD> [--live-transfer]
```

If the destination server uses a custom SSH port, include the `--port` flag, for example:

```bash
--port 2222
```

To use SSH key authentication instead of a password, simply omit the `--password` flag. Ensure that your SSH key is already configured on the destination server and verify the connection by running:

```bash
ssh root@<DESTINATION_IP>
```
