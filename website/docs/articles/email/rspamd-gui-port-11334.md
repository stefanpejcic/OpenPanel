# RSPAMD GUI

[Rspamd](https://rspamd.com/) is an advanced spam filtering system that offers many different ways to filter messages, including regular expressions and statistical analysis. Each message is analyzed by rspamd and given a spam score.

![rspamd gui](https://i.postimg.cc/wHkc2j15/2025-08-11-17-56-1.png)

[OpenPanel Enterprise edition](https://openpanel.com/enterprise/#compare) includes RSPAMD by default.

## How to enable Rspamd

To enable RSPAMD:

1. Go to **OpenAdmin > Emails > Settings**.
2. Enable the 'RSPAMD' option.
  ![rspamd enable openadmin](https://i.postimg.cc/M6gTbRyZ/2025-08-11-18-04.png)
3. Save changes and restart the email server.
  ![rspamd restart openadmin](https://i.postimg.cc/6KPyXsDs/2025-08-11-18-05.png)

> **NOTE**: If Rspamd is enabled, you should disable the following services (as they provide overlapping functionality): *Amavis, SpamAssassin, OpenDKIM, OpenDMARC*.

### Setting a Password for the Rspamd GUI

You must set a controller password before logging in.

1. Open a terminal and run:

   ```bash
   docker exec -it openadmin_mailserver rspamadm pw
   ```

   * Enter your new password.
   * You'll receive a hashed value, for example:

     ```
     $2$mu7yqw9bn9heied5aeh8utec173umcub$oesyhcdpayqob6emzctn76c3dfrr1ipsi1hmht4a9sm7ytui8wjy
     ```

2. Enter the mailserver container:

   ```bash
   docker exec -it openadmin_mailserver sh
   ```

3. Add the hashed password to the configuration:

   ```bash
   echo 'password = "HASH_HERE";' >> /etc/rspamd/local.d/worker-controller.inc
   ```

4. Exit the container and restart the mailserver:

   ```bash
   exit
   docker restart openadmin_mailserver
   ```

## How to access the RSPAMD GUI

Rspamd includes a web-based GUI that displays statistics and filtering data.

The interface is accessible via port `11334`.

![rspamd login](https://i.postimg.cc/HWSMgjMf/2025-08-11-17-56.png)

1. Open your browser and visit:

   ```
   http://<SERVER_IP_ADDRESS>:11334
   ```
2. Enter the password you configured.
3. You now have access to the Rspamd dashboard.
