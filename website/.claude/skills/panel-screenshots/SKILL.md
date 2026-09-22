---
name: panel-screenshots
description: Generate or update OpenPanel UI screenshots for the docs/panel/ pages with Playwright against the demo panel, then place them in the markdown. Use when asked to add, replace or fix screenshots on a docs/panel page, or to screenshot a section, button, dialog or tab of the OpenPanel user panel.
---

# OpenPanel docs screenshots

Screenshots for `docs/panel/**` are generated, never hand-made. Everything lives in
`scripts/panel-screenshots/`:

- `shots.mjs` – one entry per docs page: URL, `prepare()` steps, and a crop per shot.
- `shoot.mjs` – runs the shots: `node shoot.mjs <page-key> [...]` (no key = all, ~15 min).
- `helpers.mjs` – `rename`, `maskIPs`, `fillVisible`, `selectFirst`, `click`, `tab`,
  `markCard`, `markSection`, `hideOnboarding`, `hideNotices`, ...
- `login.mjs` – saves the session to `.auth/state.json` (gitignored).
- `PLAN.md` – page → route → shots map and conventions.

Output goes to `static/img/openpanel-screenshots/<section>/<page>-<shot>.png` and is
referenced as `/img/openpanel-screenshots/...`. Do not touch `/img/panel/v2/*`, the
versioned 1.X docs still use it.

## Login: try the demo first, ask only if it fails

1. `cd scripts/panel-screenshots && npm i` if `node_modules` is missing.
2. Run `node login.mjs`. It uses the credentials prefilled on
   `https://demo.openpanel.com:2083/login` (user `testinguser`).
3. Only if that fails, ask the user. Say exactly what failed and offer:
   - disable the Cloudflare Turnstile on the demo login (login.mjs then works headless), or
   - solve the captcha in the browser window login.mjs opens, or
   - give a different panel URL (`PANEL_URL=https://host:2083 node login.mjs`) and credentials.
4. The session expires during long runs. When shots fail with "session expired",
   rerun `node login.mjs` and then only the failed page keys.

The demo is in demo mode: only GET requests work, and `shoot.mjs` also aborts every
non-GET request. Nothing can be installed, created or deleted, so shoot the state just
before the submit (the Delete → Confirm countdown, a filled form, an open drawer).
Some demo services are off (MongoDB, Web Terminal) – report those pages as skipped.
Cloudflare returns 403 for some query strings (e.g. `wp-config.php`), so pick a search
term that gets through.

## What the user wants

- One screenshot per docs section that describes something on screen, placed under
  that section's heading or after its first paragraph – never between a sentence
  ending in `:` and its list/table, and never inside a numbered list.
- Crop to what the section is about: a single card, section, dialog, drawer, menu or
  button (with a little padding, no slivers of neighbouring elements). Full window only
  when the docs talk about navigation or the whole page state.
- Buttons from a page get their own small crop (e.g. Scan, Refresh Data, Themes/Plugins).
- Readable fake data: `rename()` demo names (`wp.tests.openpanel.org` → `blog.example.com`,
  random DB names → `wp_blog`), `maskIPs()` for addresses, never show tokens or passwords.
  Fake output (e.g. a cron run) must match what the real command would print.
- Light theme, 1100px wide at 2x (a shot can set `viewportWidth`/`viewportHeight`).
- Alt text describes what the image shows, not the file name.
- Fix doc text that doesn't match the UI when you see it (button names, missing
  options), and say so in the reply. Use `:::info` blocks, not "NOTE:" lines.
- After each page is done: `git add` the doc, the images and the script changes.
  Never commit. Don't stage `static/llms*.txt` (the dev server regenerates them).

## Workflow per page

1. Read the doc page and list its sections.
2. Find the route and element ids in `../openpanel/internal/web/templates/` and
   `../openpanel/static/js/` (dialogs are often Alpine `x-data` or `fmdrawer` drawers –
   opening them from JS via `Alpine.$data(el)` or `element.click()` is fine).
3. Probe the live page for selectors/layout before writing the entry.
4. Add the entry to `shots.mjs`, run `node shoot.mjs <key>`, and look at every image.
5. Place images with the placement rules above, keeping existing text.
6. Keep the dev server running (`yarn dev:docs`, see the run-website skill) and check
   the page in a browser: all images load, notes render.
7. `git add`, then tell the user what was added per section and anything skipped.
