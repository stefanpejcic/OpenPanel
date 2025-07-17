# Transfer account

This feature allows administrators to transfer (copy) individual accounts from one OpenPanel server to another.

To migrate all users and files at once, please refer to the, please refer to the [Server Migration Documentation](/docs/articles/transfers/migrate-openadmin-to-new-server/).

### How to Transfer an Account

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
