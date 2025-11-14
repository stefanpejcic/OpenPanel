---
sidebar_position: 5
---

# Locales (Languages)

Manage the languages available to OpenPanel users.

![openadmin admin panel locales](/img/admin/locales.png)

## Install Locale

By default, only the **EN** locale is installed. To enable other locales, they must be installed first.

<Tabs>
  <TabItem value="openadmin-install-locale" label="With OpenAdmin" default>

To install a locale, go to **OpenAdmin > Settings > Locales** and click the **Install** button next to the desired locale.

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

Example: Install multiple locales at once:

```bash
opencli locale sr-sr tr-tr zh-cn
```

  </TabItem>
</Tabs>


## Default Locale

To make a specific locale the default, click **Set Default** next to it in the table.

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
