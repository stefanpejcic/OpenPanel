---
sidebar_label: "Themes"
---

# How to Create and Install Themes for OpenPanel

A theme changes how the OpenPanel user panel looks: page layouts, partials, CSS, JavaScript and images. It doesn't change what the panel can do. Themes are plain folders, so you can build your own, install one from the theme list in OpenAdmin, or copy one onto the server by hand.

## Where themes live

Each theme is a folder in `/etc/openpanel/openpanel/templates/`:

```
/etc/openpanel/openpanel/templates/
├── themes.yml        # the list OpenAdmin offers for install, not a theme
└── modern/           # a theme
    ├── readme.txt    # required
    ├── screenshot.png
    ├── base.html
    ├── partials/
    ├── user/
    ├── ...
    └── static/
```

The active theme is set with `template=` in `/etc/openpanel/openpanel/conf/openpanel.config`, using the folder name. Set it with:

```bash
opencli config update template modern
```

OpenPanel restarts on its own and loads the theme. To go back to the templates built into OpenPanel, set it to an empty value:

```bash
opencli config update template ""
```

If the value doesn't match a folder that has a `readme.txt`, OpenPanel also uses the built-in templates.

## Pick a theme from OpenAdmin

Go to **OpenAdmin > Settings > OpenPanel > Display > Theme**. The cards show:

- **Default**: the built-in templates.
- **Installed themes**: every folder in `/etc/openpanel/openpanel/templates/` with a `readme.txt`, with its `screenshot.png` and details from `readme.txt`.
- **Available themes**: themes from the official list that aren't installed yet, with an **Install** button.

Select a theme, click **Save**, and restart OpenPanel when OpenAdmin asks you to. See [OpenPanel settings](/docs/admin/settings/openpanel#display).

## Let users pick their own theme

Under the theme cards, **Users can switch theme** (`allow_user_theme` in `openpanel.config`) is off by default, so everyone sees the theme selected above. Turn it on and each user can pick any installed theme for themselves from the account menu at the bottom left of the sidebar (on Saturn, from the General Information widget on the dashboard).

The choice is saved per user in `/home/<context>/theme`. If that theme is uninstalled, invalid, or switching is turned off again, the user gets the default theme.

## Create a theme

### 1. Start from the built-in templates

A theme folder has the same layout as the panel's `internal/web/templates/` folder in the [OpenPanel repository](https://github.com/stefanpejcic/OpenPanel/tree/main/openpanel/internal/web/templates). Copy that folder as your starting point, or only copy the files you want to change. **Any file your theme doesn't include is taken from the built-in templates**, so a theme can be as small as one restyled `base.html`.

Templates use Go's [html/template](https://pkg.go.dev/html/template) syntax. Keep the `{{define "..."}}` names and the data the built-in file uses, because the panel fills the same data no matter which theme is active.

### 2. Add your own partials and static files

- `partials/*.html` files that the panel doesn't have are added to every page, so you can define new blocks there and use them from any template.
- Files in `static/` replace the panel's `/static/` files with the same path. For example, `static/css/theme.css` is served as `/static/css/theme.css`. Load it from your `base.html` with `{{static "css/theme.css"}}`.

### 3. Add readme.txt

`readme.txt` is required and is what makes a folder a theme. It uses the same headers as a WordPress theme's `style.css`, one per line:

```
Theme Name: Modern
Theme URI: https://github.com/your-name/your-theme
Author: Your Name
Author URI: https://github.com/your-name
Version: 1.0.0
License: GPLv2 or later
License URI: https://www.gnu.org/licenses/gpl-2.0.html
Description: A short description shown on the theme card in OpenAdmin.
```

Only `Theme Name` is really needed; without it the folder name is shown. Links must start with `http://` or `https://`.

### 4. Add screenshot.png

`screenshot.png` in the theme folder is shown on the theme card. It's optional, and OpenAdmin shows a broken image without it. A 1200×900 screenshot of the dashboard works well.

### 5. Test it

Copy the folder to `/etc/openpanel/openpanel/templates/<your-theme>/`, select it in OpenAdmin or run `opencli config update template <your-theme>`.

If a page in your theme fails to parse or render, OpenPanel shows the built-in version of that page instead of a broken page, and logs the reason:

```bash
podman logs openpanel 2>&1 | grep "theme "
```

```
theme mytheme: templates/domains/domains.html failed to render, using the built-in page: template: domains.html:56:124: executing "content" at <.Something>: can't evaluate field Something
```

That usually means the template reads a field the panel doesn't pass to that page. Check what the built-in template of that page uses.

## Package a theme

Pack the theme as `.tar.gz`, `.tgz` or `.zip`, with `readme.txt` either at the root of the archive or inside one top folder:

```bash
tar czf mytheme.tar.gz mytheme/
```

A GitHub repository with the theme files at its root works well, since its release archives already have that layout.

## Submit your theme

Themes offered for install in OpenAdmin come from the list in the [openpanel-configuration](https://github.com/stefanpejcic/openpanel-configuration) repository: `openpanel/templates/themes.yml` for the user panel and `openadmin/templates/themes.yml` for the admin panel.

When your theme is ready:

1. Publish it, for example as a GitHub repository with a release that has the archive attached.
2. Open a pull request to [openpanel-configuration](https://github.com/stefanpejcic/openpanel-configuration) that adds your theme to `openpanel/templates/themes.yml`:

```yaml
themes:
  - folder: mytheme
    name: My Theme
    description: One line about the theme.
    author: Your Name
    author_uri: https://github.com/your-name
    version: 1.0.0
    license: GPLv2 or later
    license_uri: https://www.gnu.org/licenses/gpl-2.0.html
    uri: https://github.com/your-name/mytheme
    screenshot: https://raw.githubusercontent.com/your-name/mytheme/main/screenshot.png
    download: https://github.com/your-name/mytheme/releases/download/v1.0.0/mytheme.tar.gz
```

- `folder` is the folder the theme installs into: lowercase letters, digits, `-` and `_`.
- `download` must be an `https` link to the archive.
- `screenshot` is shown before the theme is installed. Once installed, the theme's own `screenshot.png` is used.

Once the pull request is merged, every OpenAdmin shows your theme under **Available themes** (the list is refreshed once a day).

You don't need to be on the list to use a theme: anything copied into `/etc/openpanel/openpanel/templates/` shows up as installed.
