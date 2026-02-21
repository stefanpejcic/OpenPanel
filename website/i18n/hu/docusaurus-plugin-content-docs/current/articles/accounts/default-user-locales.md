# Default Locale

**OpenPanel** is fully localization-ready, but currently ships only with English (`en`) as the default language.
Additional languages can be installed by the System Administrator from ***OpenAdmin → Settings → Locales***, and they will become immediately available in the user interface.

If your preferred locale isn’t available yet, you can help by contributing a translation: 👉 [OpenPanel Translations on GitHub](https://github.com/stefanpejcic/openPanel-translations)

---

## Setting the Default Locale

In OpenPanel, there are **five** different sources that determine which locale is applied to a user.
They are checked in the following priority order:

| Priority | Source                        | Example Path / Data                                   |
| -------- | ----------------------------- | ----------------------------------------------------- |
| 1️⃣      | Session (`session['locale']`) | `'fr'`                                                |
| 2️⃣      | User-specific file            | `/home/<context>/locale`                              |
| 3️⃣      | Browser best match            | `Accept-Language: es-ES,es;q=0.9`                     |
| 4️⃣      | Default locale file           | `/etc/openpanel/openpanel/default_locale` → e.g. `de` |
| 5️⃣      | Hardcoded final fallback      | `'en'`                                                |

---

### 1. Session Locale

When a user visits OpenPanel, the system first checks the session for a stored locale.
A session locale is created when the user logs in.
For example, if the user selects `'de'` (German) on the login page, that locale is set for the current session.

> **Note:** This setting overrides all other locale sources.

---

### 2. User-Specific File

After logging in, users can change their preferred language — if the `locale` module and feature are enabled for them.
Navigate to: ***OpenPanel → Account → Change Language*** ([Docs link](/docs/panel/account/language/)) - this view lists all languages installed by the Administrator.
The user’s chosen locale overrides their browser settings and the system default.

The user’s preference is stored in a per-user file:

```
/home/<context>/locale
```

---

### 3. Browser Best Match

If the user hasn’t set a locale yet, OpenPanel checks the browser’s [`Accept-Language` header](https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Accept-Language).
If the preferred language is installed on the server, it will be temporarily applied for that session.

> **NOTE**: This only affects the **current session** and does not persist as a user preference.
> It overrides the Administrator’s default locale for that session only.

Example header:

```
Accept-Language: es-ES,es;q=0.9
```

---

### 4. Default Locale

The Administrator can set a global default locale by creating a configuration file at:

```
/etc/openpanel/openpanel/default_locale
```

For example, to set Spanish (`es`) as the default:

```bash
echo 'es' > /etc/openpanel/openpanel/default_locale
```

---

### 5. Fallback Locale

If no other locale setting is found, OpenPanel defaults to English (`en`), which is included by default.

---

## Checking Which Locale Is Active

To verify which locale is being used for a user, you can temporarily enable developer mode and check the logs.

1. Enable `dev_mode`:

   ```bash
   opencli config update dev_mode on
   ```

2. Repeat the user action in the browser.

3. Tail the logs:

   ```bash
   docker logs -f openpanel
   ```

Look for log lines similar to:

```
APP - Using locale..
```
