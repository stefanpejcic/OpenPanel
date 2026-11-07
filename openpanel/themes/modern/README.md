# Modern theme

The design source for the Modern theme of the user panel, in the openpanel.com style. Each screen is a static page here first, then ported to the Go templates using the same `op-` classes.

## What's here

- `src/app.css`: tokens in `@theme` (dark values under `.dark`, the class the panel already toggles) and the `op-` component layer. Both the theme pages and the panel use it.
- `src/panel.css`: panel-only rules. These are helpers that stand in for Tailwind utilities, the shared partials restyled inside `.op-scope`, and one `/* area: ... */` section per ported area. The theme pages don't load it.
- `dist/app.css`: the theme pages' built stylesheet (Tailwind v4 output, committed so the pages open from disk).
- `export-kit.mjs`: writes `../../static/css/op.css`, the stylesheet the panel loads, from `src/app.css` + `src/panel.css`. It also copies the Inter fonts to `../../static/vendor/fonts/inter/`.
- `*.html`: one page per screen. `index.html` links them all. Add `?theme=dark` or `?theme=light` to force a theme.
- `img/`, `fonts/`, `vendor/`: assets the pages use when opened from disk (Alpine is MIT, Inter is OFL, see `fonts/LICENSE.txt`).

## Build

Run both after changing either CSS file or a page:

```bash
cd openpanel/themes/modern
npx @tailwindcss/cli@4.1.6 -i src/app.css -o dist/app.css --minify
cd ../..
node themes/modern/export-kit.mjs
```

Commit `dist/app.css` and `static/css/op.css` with the source change. Never edit `op.css` by hand.

## Adding a page

1. Build it here first: copy a page with the same layout, use the kit classes from `components.html`, and link it from `index.html`. Keep page-only CSS in the page's `<style>`.
2. Port it: rewrite the template in `internal/web/templates/` on the same classes. Wrap the page in `{{define "main_class"}}op-scope{{end}}`, take icons from `{{template "op_icon" "name"}}` (`partials/_op_icons.html`), and put every visible string through `{{.T.Get "..."}}` (JS strings through `{{toJSON (.T.Get "...")}}`).
3. Move the page's CSS into `src/panel.css` under a new `/* area: <name> */` section, then run the build. Before adding a rule, search both CSS files for the class, because a second definition in a later area silently overrides the first.
4. Add or extend a render test for the template. Then follow `../ui-tests/README.md` to add it to `migrated.txt` and check it at all six sizes.

Breakpoints are Tailwind's defaults: phone under 768px (drawer sidebar, rows as cards, filters in a sheet, bulk bar docked at the bottom), tablet 768 to 1023px (drawer sidebar, full table with the document root folded under the name), desktop from 1024px (fixed sidebar, stat pills from 1280px).

## Classic theme

Modern is the default. Server owners can bring back the old interface for every user with `ui_theme` in the `[PANEL]` section of `/etc/openpanel/openpanel/conf/openpanel.config`:

```ini
[PANEL]
ui_theme=classic
```

Allowed values are `modern` (the default; anything else also counts as modern) and `classic`. OpenAdmin sets it under Settings > OpenPanel > Panel theme (`/settings/open-panel`) and restarts the panel, which reads it once at start-up (`cmd/openpanel/main.go`).

How it works (`internal/web/theme.go`):

- `internal/web/templates_classic/` is a copy of the templates from before the redesign. Under `classic`, a page uses its classic copy when one exists. A page that only exists in the Modern theme stays modern, and new `_partials` get mixed in, since they only add defines.
- `static/classic/` holds the old versions of the static files the Modern theme changed. Classic pages link those through `static`. After changing a file under `static/` that existed before the Modern theme, copy its old version there.
- Some defaults depend on the theme, such as the dashboard opening as a list (modern) or an icon grid (classic).

To run the Go tests against the classic templates, use `OPENPANEL_UI_THEME=classic go test ./internal/...`. Tests that assert Modern markup call `webtest.ModernOnly(t)` and skip. The rest still render their pages, and `ParseAllPages` parses every classic template.
