---
sidebar_position: 5
---

# PHP.INI Editor  

The PHP.INI Editor allows you to view and modify the configuration file for each installed PHP version.  

> Note: PHP.INI Editor is [only available if `php_ini` feature is enabled by Administrator](/docs/admin/settings/modules#phpini-editor).

You can increase limits, enable new defaults, or adjust settings as needed. Changes will restart the corresponding PHP service.

:::tip Unsaved changes
Unsaved edits in this editor are kept as a copy in your browser. If you leave the page, close the tab or the save fails, the next time you open the php.ini of that PHP version a notice shows when that copy was made, with three options:

- **View changes** shows what's different from the saved file, line by line.
- **Restore** puts the copy back in the editor, then you can save it.
- **Discard** deletes the copy.

The copy stays only in that browser and is removed once the file is saved. Leaving the page with unsaved edits also asks you to confirm first.
:::

![PHP 8.5 INI Editor showing the php.ini file with the Save Changes button](/img/openpanel-screenshots/php/php_ini_editor-editor.png#gh-light-mode-only)
![PHP 8.5 INI Editor showing the php.ini file with the Save Changes button](/img/openpanel-screenshots/php/php_ini_editor-editor_dark.png#gh-dark-mode-only)

---

To modify the total CPU or Memory resource limits for PHP service (container) use the [**Containers** page](/docs/panel/containers/).
