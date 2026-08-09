---
name: run-website
description: Start the OpenPanel Docusaurus website locally with yarn so changes under website/ can be visually verified at localhost:3000 before considering them done.
---

# Run the OpenPanel website locally

This is the Docusaurus site under `website/`. It uses **yarn**, not npm,
even though a stray `package-lock.json` may also be present — always run
commands with `yarn`. Whenever you change anything under `website/`
(pages, components, docs, styles, config), start the dev server and check
the result in a browser before reporting the change as done. Type-checking
or a successful build is not a substitute for actually looking at the page.

## Step 1 — Install dependencies (only if `node_modules` is missing/stale)

```bash
cd website
yarn install
```

Skip this if `website/node_modules` already exists and no dependency
changed.

## Step 2 — Start the dev server

```bash
cd website
yarn start
```

- Serves at **http://localhost:3000** by default (Docusaurus' standard
  port). Docusaurus hot-reloads on file save, so leave this running rather
  than restarting it after every edit.
- Run it with `run_in_background: true` (Bash tool) since it's a
  long-lived process — don't block on it.
- If port 3000 is already taken (e.g. another session's dev server is
  still up), Docusaurus will prompt to use the next free port, or pass
  `PORT=3001 yarn start` explicitly.
- Faster partial variants exist if you only need one section:
  `yarn dev:docs` (docs only, blog disabled) and `yarn dev:blog` (blog
  only, docs disabled) — both skip versioning/example generation too.

## Step 3 — Visually check the change

Use whatever browser-automation tool/skill is available in this
environment (e.g. `claude-in-chrome`) to open `http://localhost:3000` (or
the specific changed page/route) and confirm the change actually renders
as intended — check the golden path and, for visual/layout changes, both
light and dark mode. If no browser tool is available, say so explicitly
rather than claiming the change was verified.

## Step 4 — Stop the server when done

If you started the dev server as a background process for this task, stop
it once verification is complete rather than leaving it running
indefinitely.

## Other useful commands

```bash
yarn build   # full production build — catches broken links/MDX errors that dev mode tolerates
yarn clear   # clears Docusaurus' local cache, useful if the dev server is behaving strangely after a config change
```
