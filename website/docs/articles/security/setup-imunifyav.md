# Setting Up ImunifyAV

ImunifyAV regularly scans user files and alerts you if any malware is detected.

---

To configure ImunifyAV via the control panel:

1. Go to **OpenAdmin > Security > ImunifyAV**.
2. Click the **cog icon** to access settings.
3. Adjust the resource limits for scans and set the scan schedule — by default, scans run monthly.

[![imunifyav.png](https://i.postimg.cc/PqmmF0JV/imunifyav.png)](https://postimg.cc/f3RtV2xY)


Alternatively, you can perform these configurations directly from the terminal:

**Setting Resource Limits for ImunifyAV**:

```bash
imunify-antivirus config update '{"MALWARE_SCAN_INTENSITY": {"cpu": 2}}'
imunify-antivirus config update '{"MALWARE_SCAN_INTENSITY": {"io": 2}}'
imunify-antivirus config update '{"MALWARE_SCAN_INTENSITY": {"ram": 1024}}'
imunify-antivirus config update '{"MALWARE_SCAN_INTENSITY": {"user_scan_cpu": 2}}'
imunify-antivirus config update '{"MALWARE_SCAN_INTENSITY": {"user_scan_io": 2}}'
imunify-antivirus config update '{"MALWARE_SCAN_INTENSITY": {"user_scan_ram": 1024}}'
imunify-antivirus config update '{"RESOURCE_MANAGEMENT": {"cpu_limit": 1}}'
imunify-antivirus config update '{"RESOURCE_MANAGEMENT": {"io_limit": 1}}'
imunify-antivirus config update '{"RESOURCE_MANAGEMENT": {"ram_limit": 500}}'
```

**Adjusting the Scan Schedule**:

```bash
imunify-antivirus config update '{"MALWARE_SCAN_SCHEDULE": {"day_of_month": 1}}'
imunify-antivirus config update '{"MALWARE_SCAN_SCHEDULE": {"hour": 3}}'
imunify-antivirus config update '{"MALWARE_SCAN_SCHEDULE": {"interval": "none"}}'
```

---

## Upgrade to ImunifyAV+

Execute the following command on your server to activate the Imunify AV+ license.  **If you plan on using Imunify AV free version you can skip this step**.

**Activation using an activation key**:
To activate your ImunifyAV+ license using the activation key, run the following commands:

```
imunify-antivirus unregister
imunify-antivirus register YOUR_KEY
```

Where `YOUR_KEY` is your license key, replace it with the actual key – trial or purchased. Key format is: `IMAVPXXXXXXXXXXXXXXX`.

**Activation with IP-based license**:
If you have an IP-based license, run the following commands:

```
imunify-antivirus unregister
imunify-antivirus register IPL
```

----


For more information refer to ImunifyAV Documentation 
