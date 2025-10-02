# Update LF_ALERT_TO

How to Set the Email Address for CSF/LFD Alerts

---

Updating the notification email through **OpenAdmin** will also update the email used by **CSF/LFD** for alerts.

To do this via command line, simply run:

```bash
opencli config update email youremail@yourdomain.com
```

This updates both the OpenAdmin notifications and the `LF_ALERT_TO` setting for CSF/LFD.

---

## Using a Different Email for CSF/LFD

If you want **CSF/LFD** to send alerts to a different address than the one set in OpenAdmin:

1. Open the CSF configuration file in your preferred text editor:

   ```bash
   nano /etc/csf/csf.conf
   ```

2. Locate the following line:

   ```
   LF_ALERT_TO = ""
   ```

3. Replace it with your desired email:

   ```
   LF_ALERT_TO = "youremail@yourdomain.com"
   ```

4. Save the file and restart CSF to apply changes:

   ```bash
   csf -r
   ```

That's it! CSF/LFD will now send notifications to your specified email address.

---

> **Note:** CSF/LFD is third-party software not developed or maintained by OpenPanel.
> For support, please contact Sentinel Firewall directly:
> [https://sentinelfirewall.org/](https://sentinelfirewall.org/)
