---
sidebar_position: 4
---

# Service Limits

The **Service Limits** section allows Administrators to view and adjust the allocated CPU and Memory resources for system services.

---

To update limits, enter new values in the fields for the appropriate service.  

> **Note:** The service must be stopped and restarted for the new limits to take effect.

**Available Services**:

- openpanel
- caddy
- mysql
- clamav
- redis
- bind9
- ftp

---

To edit limits via the terminal, modify the `/root/.env` file directly.
