# Keyboard Shortcuts

OpenAdmin UI can be navigated using keyboard shortcuts.

## Available shortcuts

To view all available shortcuts: `Ctrl` + `K` 

| Shortcut                     | Action                          | URL Path                      |
|-----------------------------|----------------------------------|-------------------------------|
| `Ctrl` + `Escape`           | Logout                           | `/logout`                    |
| `Ctrl` + `Shift` + `H`      | Dashboard                        | `/dashboard`                 |
| `Ctrl` + `Shift` + `N`      | Notifications                    | `/view_notifications`        |
| `Ctrl` + `Shift` + `U`      | Users                            | `/users`                     |
| `Ctrl` + `Shift` + `R`      | Resellers                        | `/resellers`                 |
| `Ctrl` + `Shift` + `A`      | Administrators                   | `/administrators`            |
| `Ctrl` + `Shift` + `P`      | Plans                            | `/plans`                     |
| `Ctrl` + `Shift` + `F`      | Feature Manager                  | `/features`                  |
| `Ctrl` + `Shift` + `D`      | Domains                          | `/domains`                   |
| `Ctrl` + `Shift` + `E`      | Email Accounts                   | `/emails/accounts`           |
| `Ctrl` + `Shift` + `S`      | Services Status                  | `/services`                  |
| `Ctrl` + `Shift` + `C`      | ConfigServer Firewall (CSF)      | `/security/firewall`         |
| `Ctrl` + `Shift` + `W`      | CorazaWAF Settings               | `/security/waf`              |
| `Ctrl` + `Shift` + `L`      | View Logs                        | `/services/logs`             |
| `Ctrl` + `Shift` + `B`      | Basic Auth                       | `/security/basic_auth`       |
| `Ctrl` + `Shift` + `1` (Numpad only) | License                     | `/license`                   |

> ℹ️ Use `Cmd` on macOS in place of `Ctrl`.

---

## Edit shortcuts

From OpenPanel 1.7.41 Administrators can edit the shortcuts and customize them. To edit shortcuts, edit file: `/etc/openpanel/openadmin/config/shortcuts.json`.

Default content:

```json
{
  "ctrl+shift+h": "/dashboard",
  "ctrl+shift+n": "/notifications",
  "ctrl+shift+u": "/users",
  "ctrl+shift+r": "/resellers",
  "ctrl+shift+a": "/administrators",
  "ctrl+shift+p": "/plans",
  "ctrl+shift+f": "/features",
  "ctrl+shift+d": "/domains",
  "ctrl+shift+e": "/emails/accounts",
  "ctrl+shift+s": "/services",
  "ctrl+shift+c": "/security/firewall",
  "ctrl+shift+w": "/security/waf",
  "ctrl+shift+l": "/services/logs",
  "ctrl+shift+b": "/security/basic_auth",
  "ctrl+escape": "/logout"
}
```

After editing the file, restart OpenAdmin: `service admin restart`.

---

## Disable shortcuts

To completelly disable the shortcuts, delete the file: `/etc/openpanel/openadmin/config/shortcuts.json`
