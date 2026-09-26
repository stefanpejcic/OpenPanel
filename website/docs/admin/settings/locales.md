---
sidebar_position: 9
---

# Locales (Languages)

Manage the languages available to OpenPanel users.

![Languages page listing locales with their provider, install status, default and Set as Default buttons](/img/openadmin-screenshots/settings/locales-list.png#gh-light-mode-only)
![Languages page listing locales with their provider, install status, default and Set as Default buttons](/img/openadmin-screenshots/settings/locales-list_dark.png#gh-dark-mode-only)

## Bulk Actions

Tick the checkbox of one or more locales, or the checkbox in the table header to select all locales shown by the current search. A bar appears at the bottom of the page with the number selected, a **Clear** link and these actions:

![Two locales selected with the bulk actions bar offering Install, Update and Delete](/img/openadmin-screenshots/settings/locales-bulk.png#gh-light-mode-only)
![Two locales selected with the bulk actions bar offering Install, Update and Delete](/img/openadmin-screenshots/settings/locales-bulk_dark.png#gh-dark-mode-only)

| Action | What it does |
|---|---|
| **Install** | Downloads and installs the selected locales. |
| **Update** | Downloads the latest translations for the selected locales that have an update. |
| **Delete** | Deletes the selected locales, the default locale can't be deleted. |

Click an action, confirm it, and it runs on the selected locales one after another. When it's done the page reloads with a notice listing the locales it worked for, or which ones failed and why.

## Install Locale

By default, only the **EN** locale is installed. To enable other locales, they must be installed first.

<Tabs>
  <TabItem value="openadmin-install-locale" label="With OpenAdmin" default>

To install a locale, go to **OpenAdmin > Settings > Locales** and click the **Install** button next to the desired locale.

To install every available locale at once, click **Install All** at the top of the page. This downloads all locales from GitHub (about 11 MB in total) and re-downloads ones that are already installed, so it also updates them. Once finished, a notification lists which locales were installed and which failed.

  </TabItem>
  <TabItem value="API-install-locale" label="With API">

Install a single locale with [`POST /api/settings/locales`](/docs/admin/settings/api):

```bash
curl -X POST https://panel.example.com:2087/api/settings/locales \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"locale": "tr-tr"}'
```

Pass `"all"` as the locale to install every available locale:

```bash
curl -X POST https://panel.example.com:2087/api/settings/locales \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"locale": "all"}'
```

The response lists what was installed and what failed. `success` is `false` if any locale failed:

```json
{
  "success": true,
  "message": "Installed 17 locales: bg-bg, de-de, en-us, ...",
  "installed": ["bg-bg", "de-de", "en-us", "..."],
  "failed": []
}
```

  </TabItem>
  <TabItem value="CLI-install-locale" label="With OpenCLI">

To install a locale from the terminal, use its locale prefix from [Github](https://github.com/stefanpejcic/openpanel-translations/tree/main) and run:

```bash
opencli locale <LOCALE_HERE>
```

Example: Install Turkish locale:

```bash
opencli locale tr-tr
```

Example: Install multiple locales:

```bash
opencli locale sr-rs tr-tr zh-cn
```

Example: Install all available locales:

```bash
opencli locale $(curl -s "https://api.github.com/repos/stefanpejcic/openpanel-translations/contents" | jq -r '.[] | select(.type=="dir" and (.name|test("^[a-z]{2}-[a-z]{2}$"))) | .name' | tr '\n' ' ')
```

  </TabItem>
</Tabs>


## Update Locale

Each time the **Locales** page is opened, OpenAdmin compares every installed locale file with the latest version on GitHub. If a newer version exists, a notification lists the locales with updates, the locale shows an **Update available** badge, and an **Update** button appears in its **Actions** column.

<Tabs>
  <TabItem value="openadmin-update-locale" label="With OpenAdmin" default>

Click **Update** next to a locale to download its latest version, or click **Update All** at the top of the page to update every outdated locale at once. **Update All** is only shown when at least one update is available.

  </TabItem>
  <TabItem value="API-update-locale" label="With API">

`GET /api/settings/locales` returns `update_available` for each locale. To update one locale:

```bash
curl -X POST https://panel.example.com:2087/api/settings/locales \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"update": "tr-tr"}'
```

Pass `"all"` to update every locale that has an update available:

```bash
curl -X POST https://panel.example.com:2087/api/settings/locales \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"update": "all"}'
```

The response lists updated locales under `installed` and failed ones under `failed`.

  </TabItem>
  <TabItem value="CLI-update-locale" label="With OpenCLI">

Installing a locale again downloads its latest version:

```bash
opencli locale tr-tr
```

  </TabItem>
</Tabs>

## Delete Locale

<Tabs>
  <TabItem value="openadmin-delete-locale" label="With OpenAdmin" default>

Click **Delete** next to an installed locale and confirm. The default locale can't be deleted; set another locale as default first.

  </TabItem>
  <TabItem value="API-delete-locale" label="With API">

```bash
curl -X POST https://panel.example.com:2087/api/settings/locales \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"delete": "tr-tr"}'
```

  </TabItem>
  <TabItem value="CLI-delete-locale" label="From Terminal">

Remove the locale's directory, named after its two-letter prefix:

```bash
rm -rf /etc/openpanel/openpanel/translations/tr
```

  </TabItem>
</Tabs>

:::info
Users who had selected a deleted locale will see untranslated (English) text instead.
:::

## Default Locale


<Tabs>
  <TabItem value="openadmin-default-locale" label="With OpenAdmin" default>

To make a specific locale the default, go to **OpenAdmin > Settings > Locales** and click the **Set as Default** button next to the desired locale.

  </TabItem>
  <TabItem value="CLI-default-locale" label="From Terminal">

To set a default locale from the terminal:

```bash
echo <LOCALE_HERE> > /etc/openpanel/openpanel/default_locale
```

Example: Set Turkish as the default:

```bash
echo tr > /etc/openpanel/openpanel/default_locale
```

  </TabItem>
</Tabs>

:::info
Changing the default will **not** automatically update existing users’ settings; their browser preferences and account configurations will take precedence.
For more details, see [How-to Guides > Setting the Default Locale](/docs/articles/accounts/default-user-locales/#setting-the-default-locale).
:::

## Edit Locale

To edit a locale, click the GitHub icon next to it in the table. This opens the source on GitHub, where you can fork the repository and edit the translation file.

## Create a Locale

To create a new locale:

1. Fork [the translations repository](https://github.com/stefanpejcic/openpanel-translations/).
2. Copy the `en_us` folder to a new locale folder, e.g., `es_es`.
3. Translate the `messages.pot` file.
4. Submit a pull request.
