---
sidebar_position: 25
---

# PHP Applications

Deploy a [Composer](https://getcomposer.org/)-based PHP application into an existing domain. Unlike the containerized [Node.js](/docs/panel/applications/nodejs), Python, Ruby, and Java application installers, a PHP application never gets a dedicated container: it never creates a new container, edits `docker-compose.yml`, or configures a reverse proxy. The domain's existing vhost already routes to its PHP-FPM (or LiteSpeed) container, so installing a PHP application is really just deploying Composer-managed code into that domain's docroot and, optionally, running Composer against it.

---

## Install a PHP Application

Navigate to **OpenPanel > AutoInstaller** and click **Install PHP Application**.

On the install page, configure:

* **Domain / Subfolder** – The domain (and optional subfolder) to install into. Leave the subfolder blank to install at the domain root.
* **Initial project** – Optional. Leave blank to deploy into a directory that already exists and already has its own `composer.json`, or provide one of:
  * An `https://` URL to an archive ending in `.zip`, `.tar.gz`, `.tgz`, or `.tar` — downloaded (capped at 200MB) and extracted directly into the docroot.
  * A Composer package name (`vendor/package`, e.g. `laravel/laravel`) — installed via `composer create-project`.
* **Run composer install** – Runs `composer install` after deployment. Required if the directory doesn't already contain a built `vendor/` folder.
* **Optimize autoloader** – Adds `--optimize-autoloader` to the composer run.

There's no Name, Port, Startup File/Command, Type, Version picker, or CPU/Memory allocation — none of that applies, since the app doesn't get its own container. The domain must already have a PHP version assigned (via PHP Selector) before you can install; otherwise the install fails immediately.

Behind the scenes, this starts the domain's `php-fpm-<version>` container if it isn't already running, then either downloads and extracts the archive host-side, or runs `composer create-project <package> <path> --no-interaction` inside that container. If **Run composer install** is checked, it then runs `composer install --no-interaction` (with `--optimize-autoloader` if selected) — both composer calls use `--working-dir` rather than `podman exec`'s own working-directory flag, since the container's `composer` binary is itself a wrapper that otherwise resets its working directory regardless. The app's settings (initial project, composer flags, working directory) are recorded as `.env` keys under a synthetic prefix, since — unlike the containerized app types — there's no dedicated container to key them off of. The site is then added to Site Manager with type `PHP`. Progress is streamed live as each step completes.

## Manage a PHP Application

Every PHP application shows up on the general **Site Manager** page (`/sites`) alongside your other websites, listed by site name. There's no dedicated overview/status page the way there is for the containerized app types — there's no container to report status, resource usage, or a version for — so management is limited to a few focused actions, run against the domain's *current* PHP-FPM container (following whatever version PHP Selector currently has set):

* **Composer Install** – Re-runs `composer install --no-interaction` against the app's working directory.
* **Composer Update** – Runs `composer update --no-interaction` (optionally with `--optimize-autoloader`).
* **Logs** – Shows the accumulated output of every past Composer install/update run for the site, each entry timestamped, or "No Composer runs recorded yet." if none have run.
* **Remove** – Removes the site from Site Manager and clears its stored `.env` settings. Docroot files and any database are left untouched — the same "all website data remains" behavior the Node.js/Python installers use on delete — since there's no dedicated container to tear down.

---

## Not included

Because a PHP application isn't containerized the way Node.js/Python/Ruby/Java applications are, the following don't apply:

- A dedicated container — so also no Start/Stop/Restart actions, no CPU/Memory/PIDs limits, and no per-app Docker image/version picker
- A custom port, or a startup file/command — the domain's existing vhost and PHP-FPM already handle routing and execution
- Changing the PHP version from the app itself — it always follows whatever version the domain has set in PHP Selector
- Screenshot/status overview cards on a dedicated management page — a PHP application is managed inline from Site Manager instead
