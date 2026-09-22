# OpenPanel docs screenshot plan

Screenshots for `docs/panel/**` are generated with Playwright against the demo
(`https://demo.openpanel.com:2083/`, user `testinguser`) by `shoot.mjs`.
Every shot is defined in `shots.mjs`, so a UI change can be followed by a re-run.

## Conventions

- Viewport 1100×900 at deviceScaleFactor 2 (the content column is then close to the docs column width, so UI text stays readable), light theme, English locale.
- Output goes to `static/img/openpanel-screenshots/<section>/<page>-<shot>.png`, replacing
  the old ad-hoc `/img/panel/v2/*` and `/img/docs-content/*` images.
- Kinds of shot:
  - **page**: the main content column only (no header or sidebar), one per page, placed under the H1.
  - **detail**: a crop around one element (form, modal, dropdown, row actions) with 24px padding, placed under the H2 it explains.
  - **window**: the full window including the sidebar. Used only when the docs talk about navigation.
- Transient state (open dropdowns, the Delete→Confirm countdown, modals) is set up in the shot's `prepare()`.
- Destructive actions are never submitted. The shot is taken right before the submit click.
- Alt text describes what the image shows. The file name is not used as alt text.

## Login

Run `node login.mjs` once. It signs in with the credentials pre-filled on the demo
login form and saves the session to `.auth/state.json` (gitignored). If a captcha
is enabled, it opens a visible browser so you can solve it by hand.
`shoot.mjs` reuses that session until it expires.

The demo runs in demo mode, so it only answers GET requests. `shoot.mjs` also blocks
every non-GET request. Demo data has random names, so each page's `prepare()` swaps
in readable ones (like `wp_blog`) in the browser before the shot.

## Page → route → shots

Legend: ✅ route exists · ⚠️ route missing or differs · — no screenshot needed

### Getting started / Dashboard
| Doc page | Route | Shots |
|---|---|---|
| 000_intro | `/login`, `/reset_password` | ✅ login form · reset form · reset sent (the email shot stays manual) |
| dashboard/dashboard | `/dashboard` | ✅ page · detail: usage widgets · favorites · search open · tour popover |
| dashboard/dark-mode | any page | ✅ window light · window dark (`data-theme` toggle) |
| dashboard/onboarding | `/dashboard` (first login) | ⚠️ needs a user with no `onboarding.completed` · steps 1–N |

### Account
| Doc page | Route | Shots |
|---|---|---|
| account/account | `/account` | ✅ page |
| account/2fa | `/account/2fa` | ✅ page · detail: QR + code form |
| account/passkeys | `/account/passkeys` | ✅ page · detail: passkey list |
| account/active_sessions | `/account/sessions` | ✅ page |
| account/login_history | `/account/login-history` | ✅ page |
| account/account_activity | `/account/activity` | ✅ page |
| account/notifications | `/account/notifications` | ✅ page · detail: email notification settings |
| account/favorites | `/account/favorites` | ✅ page |
| account/language | `/account/language` | ✅ page · detail: language dropdown open |
| account/api | `/account/api` | ✅ page · detail: token list |
| account/mcp | `/account/mcp` | ✅ page · detail: generate token |

### Domains
| Doc page | Route | Shots |
|---|---|---|
| domains/domains | `/domains` | ✅ page · detail: row actions |
| domains/new | `/domains/new` | ✅ page (form) |
| domains/dns | `/domains/edit-dns-zone/{domain}` | ✅ page · detail: edit record · create record · advanced editor |
| domains/ssl | `/domains/ssl` | ✅ page · detail: custom SSL form |
| domains/redirects | `/domains/redirect` | ✅ page |
| domains/docroot | `/domains/docroot` | ✅ page |
| domains/vhosts | `/domains/vhosts` | ✅ page (editor) |
| domains/logs | `/domains/log/{domain}` | ✅ page |
| domains/goaccess | `/domains/stats/{domain}` | ✅ page |
| domains/dynamic-dns | `/domains/dynamic-dns` | ✅ page · detail: create entry |
| domains/capitalize | `/domains/capitalize` | ✅ page |
| domains/suspend / unsuspend | `/domains/suspend`, `/domains/unsuspend` | ✅ detail: confirm dialog |

### Emails
| Doc page | Route | Shots |
|---|---|---|
| emails/emails | `/emails` | ✅ page · detail: row actions · connect-devices config |
| emails/new | `/emails/new` | ✅ page (form) |
| emails/aliases | `/emails/aliases` | ✅ page · detail: new alias |
| emails/default_address | `/emails/default/{domain}` | ✅ page |
| emails/filters | `/emails/filter/{email}/gui` | ✅ page |
| emails/email_deliverability | `/emails/deliverability/{domain}` | ✅ page · detail: record details |
| emails/import | `/emails/import` | ✅ page |
| emails/delete | `/emails` | ✅ detail: delete confirm |
| emails/webmail | `/webmail/` | ✅ detail: webmail button (external app is not shot) |

### Files
| Doc page | Route | Shots |
|---|---|---|
| files/files | `/files` | ✅ page · detail: new file · new folder · right-click menu · editor · permissions dialog (≈8 shots, one per main H2) |
| files/upload | `/file-manager/upload` | ✅ page |
| files/wget | `/file-manager/upload?method=download` | ✅ page |
| files/trash | `/files.trash` | ✅ page · detail: row actions |
| files/FTP | `/ftp` | ✅ page · detail: create user · connections (`/ftp/connections`) |
| files/backups | `/backups`, `/backups/destination`, `/backups/settings`, `/backups/list` | ✅ one page shot per tab · detail: restore dialog |
| files/backup-wizard | `/backup-wizard` | ✅ page · detail: existing backups |
| files/disk_usage | `/disk-usage/` | ✅ page · detail: chart |
| files/inodes_explorer | `/inodes-explorer/` | ✅ page · detail: chart |
| files/fix_permissions | `/fix-permissions` | ✅ page |
| files/malware-scanner | `/malware-scanner`, `/malware-scanner/quarantine` | ✅ page · quarantine |

### MySQL
| Doc page | Route | Shots |
|---|---|---|
| **mysql/databases** | `/mysql` | ✅ **page · detail: sizes on · export panel · delete confirm** ← first run |
| mysql/new_db | `/mysql/new` | ✅ page (form) |
| mysql/users | `/mysql/users` | ✅ page · detail: new user · assign · password change |
| mysql/new_user | `/mysql/user` | ✅ page (form) |
| mysql/assign | `/mysql/assign` | ✅ page (form) |
| mysql/remove | `/mysql/remove` | ✅ page (form) |
| mysql/wizard | `/mysql/wizard` | ✅ page |
| mysql/import | `/mysql/import` | ✅ page |
| mysql/phpmyadmin | `/mysql/phpmyadmin` | ✅ detail: phpMyAdmin button (the external app stays manual) |
| mysql/processlist | `/mysql/processlist` | ✅ page |
| mysql/remote | `/mysql/remote-mysql` | ✅ page enabled · page disabled · detail: per-user access |
| mysql/root-password | `/mysql/root-password` | ✅ page |
| mysql/configuration | `/mysql/configuration` | ✅ page |

### PostgreSQL / MongoDB
Same set of shots as MySQL at `/postgresql/*` and `/mongodb/*`. All routes exist ✅.
MongoDB has no processlist, remote or configuration pages, and the docs don't have them either.

### PHP
| Doc page | Route | Shots |
|---|---|---|
| php/domains | `/php/domains` | ✅ page |
| php/default | `/php/default` | ✅ page |
| php/extensions | `/php/extensions` | ✅ page · detail: install modal |
| php/options | `/php/options` | ✅ page |
| php/php_ini_editor | `/php/php_ini_editor` | ✅ page |

### Caching
| Doc page | Route | Shots |
|---|---|---|
| caching/caching | `/cache/` | ✅ page |
| Redis / valkey / Memcached | `/cache/redis` etc. | ✅ page · detail: memory limit · connect info · logs |
| elasticsearch / opensearch | `/cache/elasticsearch` etc. | ✅ page · detail: logs |
| varnish | `/cache/varnish`, `/cache/varnish/stats` | ✅ page · detail: domains · stats |

### Applications
| Doc page | Route | Shots |
|---|---|---|
| applications/applications | none (index page) | ✅ window: "new site" popup (`/sites`) |
| applications/sites | `/sites` | ✅ page · detail: bulk actions |
| applications/autoinstaller | `/auto-installer` | ✅ page |
| applications/wordpress | `/wordpress`, `/website?domain=` | ✅ install · site manager overview (currently 20 images, cut to ~6 key ones) |
| nodejs / python / ruby / java | `/{lang}/install`, `/sites` | ✅ install form · manage page |
| builder | `/website-builder/install`, `/website-builder/edit` | ✅ create · editor |
| 15 app managers (drupal, joomla, moodle…) | `/{app}/install`, `/website?domain=` | ✅ install form · manage page. They share one layout, so 2 shots each and the loop is cheap |
| n8n | `/n8n/install` | ⚠️ module exists but has no doc page yet |

### Containers
| Doc page | Route | Shots |
|---|---|---|
| containers/containers | `/containers` | ✅ page · detail: edit resources (`/containers/edit/{svc}`) · new service |
| containers/change | `/containers/image/change` | ✅ page |
| containers/logs | `/containers/logs` | ✅ page |
| containers/terminal | `/containers/terminal` | ✅ page (after the shell connects) |
| containers/mysql | `/containers/mysql` | ✅ page |
| containers/webserver | `/containers/webserver` | ✅ page |

### Advanced
| Doc page | Route | Shots |
|---|---|---|
| advanced/advanced | none (index page) | — |
| advanced/services | `/services/` | ✅ page |
| advanced/cronjobs | `/cronjobs` | ✅ page · detail: add · common presets · run now · logs · editor. Replaces the two GIFs with still images of the end state |
| advanced/ip-blocker | `/security/ip-blocker` | ✅ page |
| advanced/process_manager | `/process-manager` | ✅ page · detail: kill confirm |
| advanced/webserver_settings | `/server/webserver_conf` | ✅ page |
| advanced/waf | `/server/waf`, `/server/waf/{domain}`, `/server/waf/log` | ✅ page · manage domain · logs |
| advanced/resource_usage | `/server/usage`, `/server/usage/history` | ✅ page · history |
| advanced/server_info | `/server/info` | ✅ page |

### Not covered
- `999_api` is the API reference and needs no screenshots.
- Routes with no doc page: `/n8n/install`, `/containers/new` (partly covered in containers.md), `/sites/updates`, `/malware-scanner/quarantine` (covered in malware-scanner.md).

## Rollout
1. mysql/databases as the pilot. Review how it looks.
2. The DB sections (MySQL, PostgreSQL, MongoDB). Their layouts are the same, so the patterns carry over.
3. Account, Domains, Emails, Files, PHP, Caching, Containers, Advanced.
4. Applications last. The demo is read-only, so only apps already installed on it can have manage-page shots. The install forms work for every app.
5. Keep `/img/panel/v2/*`: the versioned 1.X docs still use those files.

## Status (2026-09-22)

All sections are shot and placed. Pages the read-only demo can't show, left as they were:

- MongoDB (8 pages): the MongoDB service isn't running on the demo, so every page redirects to an error.
- containers/terminal: Web Terminal is disabled on the demo.
- Caching overview (`/cache/`): the page has no content to crop.
- Kept as-is because they show things outside the panel or need a POST: the password reset "sent", email and new-password
  steps in 000_intro, the notification emails in account/notifications, the Discord setup steps in files/backups, and
  the dark-mode toggle GIF.
- Log shots were dropped where the demo has no log data (Redis/Memcached container logs, cron job logs, WAF logs).

Re-run everything with `node login.mjs && node shoot.mjs` (about 15 minutes). The session expires after a while, so
if shots start failing with "session expired", log in again and re-run just the failed pages.
