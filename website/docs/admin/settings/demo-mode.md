---
sidebar_position: 8
---

# Demo Mode

Demo mode allows users to log in with read-only access—no changes can be made. It's designed for hosting providers to showcase their features and plans in a public demo environment. This applies to both the admin and user panels.

Once enabled, it can't be turned off from the admin panel. To disable it, run the following command in the terminal:
```
opencli config update demo_mode off
```

Be sure to set up the demo content beforehand and properly secure the server before enabling this feature.
