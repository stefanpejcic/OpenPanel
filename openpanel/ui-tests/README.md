# UI tests

Screenshot and layout checks for the Modern theme at three sizes (phone 390px, tablet 820px, desktop 1440px) in dark and light, so six runs per page.

- **Theme pages** (`specs/theme.spec.mjs`): every page in `../themes/modern`. These are the signed-off designs, so each one must load without console errors or missing files and must never scroll sideways. A few interaction checks too (phone drawer, tool search).
- **Gallery** (`specs/gallery.spec.mjs`): the real Go templates, rendered with the fixtures of the existing render tests. Pages listed in `migrated.txt` must pass the same layout check. The gallery has no backend, so on those pages a JSON parse error from a fetch that came back 404 is noted in the report instead of failing; any other console error still fails; the rest are screenshot-only and get an `overflow` annotation in the report until they move to the new design.
- **Coverage** (`scripts/coverage.mjs`): every page template the Go code loads has to land in the gallery. `coverage-allow.txt` lists the ones that have no render test yet; a new page without one fails the check.

## Run it

```bash
npm install
npm run gallery
npm run coverage
npm test
npm run report
```

`npm run gallery` runs `go test -tags gallery ./internal/...` with `OPENPANEL_GALLERY_DIR=ui-tests/gallery`. The `gallery` build tag turns on the capture in `internal/web/gallery_on.go`; normal builds compile it out. Without Go on the machine, run the same command in the `golang` image from the `openpanel` directory:

```bash
docker run --rm -v "$PWD:/src" -w /src -e OPENPANEL_GALLERY_DIR=/src/ui-tests/gallery golang:1.26 go test -tags gallery -count=1 ./internal/...
```

Screenshots land in `out/<project>/<theme|gallery>/` and are attached to the HTML report. `GALLERY_ALL=1` checks every render instead of one per template. `PW_CHANNEL=msedge` (or `chrome`) reuses an installed browser; otherwise run `npx playwright install chromium` once.

Playwright 1.48 hangs on Node 26; use Node 22 or 24.

## Classic theme

`ui_theme=classic` serves the pre-redesign templates (see `../themes/modern/README.md`). Set `OPENPANEL_UI_THEME=classic` to run the Go render tests against them:

```bash
OPENPANEL_UI_THEME=classic go test -count=1 ./internal/...
```

Render tests that check Modern markup start with `webtest.ModernOnly(t)` and get skipped in that run. Call it at the top of any new test that asserts `op-` markup.

## Moving a page to the Modern theme

1. Change the template to the `op-` classes from `../themes/modern`, and add `{{define "main_class"}}op-scope{{end}}` and `{{define "page_tabs"}}{{template "op_page_tabs" .}}{{end}}` so `base.html` gives it the new tokens and tab bar. Icons come from `{{template "op_icon" "name"}}` (`partials/_op_icons.html`).
2. Add its template path to `migrated.txt`.
3. `npm run gallery && npm test`: it now has to pass at all six sizes.

The panel loads the kit as `static/css/op.css`, generated from `../themes/modern/src/app.css` and `../themes/modern/src/panel.css` by `node ../themes/modern/export-kit.mjs`. The colour tokens sit on `:root` so the shell (sidebar, top bar) gets them on every page. Radius and font tokens are renamed to `--op-radius-*` and `--op-font-*` on export, so Tailwind's `rounded-*` and `font-*` on pages that haven't moved keep their values. Run the export after changing either source file.
