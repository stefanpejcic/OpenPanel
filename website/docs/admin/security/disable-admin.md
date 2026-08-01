---
sidebar_position: 7
---

# Disable OpenAdmin

As an advanced security measure, OpenPanel allows you to completely disable the OpenAdmin interface and its related admin service.

By enabling this option, access to OpenAdmin will be blocked, preventing any further administrative changes through the UI. This is particularly useful for hardened production environments where direct UI access is unnecessary or unwanted.

Navigate to **Security > Disable OpenAdmin**, then click **Confirm** to disable OpenAdmin, or **Cancel** to return to the dashboard without making changes.

Use this feature with caution — re-enabling OpenAdmin access will require manual intervention via server-side configuration. To enable access to the OpenAdmin: `opencli admin on`
