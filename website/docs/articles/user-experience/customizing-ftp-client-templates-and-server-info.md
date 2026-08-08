# Customizing FTP Server Info and Client Templates

The **FTP Accounts** page (`OpenPanel > Files > FTP Accounts`) shows users the FTP server address/port and lets them download ready-made config files for **FileZilla** and **Cyberduck**.

By default:
- The server address shown is the user's dedicated IP (if one is assigned) or the server's public IP, and the port is always `21`.
- The FileZilla/Cyberduck files are generated in the background with hardcoded XML.

Administrators can override both of these, without editing any code and without restarting OpenPanel — the files below are read fresh on every request.

---

## 1. Customize the displayed server address and port

**File path:** `/etc/openpanel/ftp/server.conf`

This file doesn't exist by default. If it's missing, or a key is missing/invalid, OpenPanel falls back to its normal auto-detected behavior for that value.

**Format** (plain `key=value`, one per line):

```env
hostname=ftp.example.com
port=2121
```

- `hostname` — shown instead of the dedicated/public IP, and used as the `<Host>`/`<hostname>` value in downloaded client configs.
- `port` — shown as the FTP port, and used as the `<Port>` value in the FileZilla config. Must be a number between `1` and `65535`; anything else is ignored and `21` is used instead.

Either key can be set on its own — you don't need to provide both.

---

## 2. Customize the FileZilla / Cyberduck download files

**File paths:**
- `/etc/openpanel/ftp/filezilla.conf`
- `/etc/openpanel/ftp/cyberduck.conf`

If either file is present, OpenPanel uses it as the template for that client's downloadable config instead of its built-in one. If a file is missing, empty, or doesn't contain well-formed XML after substitution, OpenPanel silently falls back to the built-in generator — users are never shown an error because of a bad template.

**Available placeholders**, substituted before the file is validated and served:

| Placeholder    | Value                                                            |
| -------------- | ----------------------------------------------------------------- |
| `{host}`       | The FTP server address (respects `server.conf`'s `hostname`, if set) |
| `{port}`       | The FTP port (respects `server.conf`'s `port`, if set)            |
| `{username}`   | The FTP account's username                                        |
| `{path}`       | The FTP account's home/root path                                  |

### Example: `filezilla.conf`

```xml
<?xml version="1.0" encoding="UTF-8"?>
<FileZilla3>
    <Servers>
        <Server>
            <Host>{host}</Host>
            <Port>{port}</Port>
            <Protocol>0</Protocol>
            <Type>0</Type>
            <User>{username}</User>
            <Logontype>2</Logontype>
            <EncodingType>Auto</EncodingType>
            <RemoteDir>{path}</RemoteDir>
            <UsePassive>1</UsePassive>
        </Server>
    </Servers>
</FileZilla3>
```

### Example: `cyberduck.conf`

```xml
<?xml version="1.0" encoding="UTF-8"?>
<bookmark>
    <hostname>{host}</hostname>
    <port>{port}</port>
    <username>{username}</username>
    <protocol>ftp</protocol>
    <path>{path}</path>
</bookmark>
```

:::note
FileZilla site-export files and Cyberduck bookmarks don't carry a password field, so neither the built-in generator nor a custom template can pre-fill one — users are prompted for their password on first connect either way.
:::

---

## Scope

- Both files under `/etc/openpanel/ftp/` are shared across **all users and hosting plans** — there's no per-user or per-plan override.
- Everything above — the `server.conf` host/port override and the custom `filezilla.conf`/`cyberduck.conf` templates — applies equally to the FTP Accounts page's download links and to the [REST API](/docs/panel/api/) equivalent (`GET /api/ftp/configuration/{type}/{account}`).
