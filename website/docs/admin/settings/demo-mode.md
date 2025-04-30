---
sidebar_position: 8
---

# Demo Mode

Enable Demo Mode to lock both the OpenPanel and OpenAdmin interfaces in read-only mode.

This mode is ideal for hosting providers who want to showcase OpenPanel in a secure, public demo environment. Users will be able to explore the UI, but no changes can be made—all actions are disabled across both the admin and user panels.

Once enabled, Demo Mode cannot be turned off via the admin panel.
To disable it, run the following command in your terminal:
```
opencli config update demo_mode off
```

Make sure to configure your demo content and secure the server before enabling this mode. 📘 [Learn more](https://dev.openpanel.com/cli/config.html#Demo-mode)

