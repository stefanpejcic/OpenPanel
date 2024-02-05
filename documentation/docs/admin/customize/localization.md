---
sidebar_position: 9
---

# Localization

OpenPanel is localization ready and can easily be translated into any language.

## Available Locales

OpenPanel is shipped with the following locales:

- English (EN)

## How to translate

To translate OpenPanel to another language, for example Serbian:

1. Navigate to the OpenPanel directory using the "cd" command:
```bash
cd /usr/local/panel
```

2. Initialize a new translations directory for your locale:
```bash
pybabel init -i messages.pot -d translations -l sr
```

3. Enter the newly created locale directory and edit the messages.po file. For each `msgid` write the translation in the `msgtr` tag, for example:


```bash title="/usr/local/panel/translations/sr/LC_MESSAGES/messages.po"
#: templates/base.html:237
msgid "Websites"
msgstr "Sajtovi"
``` 


4. After you are finished translating, you need to compile the edited `.po` file to `.mo` in order to be used by the panel:

```bash
pybabel compile -d translations
```

Thats it, your new locale is added and ready to be used.

:::info
If you would like to share your translation with the OpenPanel community, please email the content of your locale folder `/usr/local/panel/translations/<LOCALE-HERE>/LC_MESSAGES/` to support@openpanel.co
:::
