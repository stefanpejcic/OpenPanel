---
sidebar_position: 5
---

# Composer App Manager

Install and manage a [Composer](https://getcomposer.org/)-based PHP project (e.g. Laravel) inside an existing domain's docroot, run inside whichever shared `php-fpm-<version>` container that domain's PHP version already points to.

Unlike the Node.js/Python app installer, this never creates a dedicated container — the domain's existing vhost already routes to the right PHP-FPM container.

## Install a PHP Application

Navigate to **OpenPanel > AutoInstaller** and click **Setup PHP Application**.

On the install page, configure:

* **Domain** – The domain (and optional subfolder) to install into. Leave the subfolder blank to install at the domain root.
* **Initial Project** – A `composer create-project` package name (e.g. `laravel/laravel`) or a URL to a `.zip` archive of an existing project.
* **Automatically run composer install** – Runs `composer install` right after the project files are in place.
* **Optimize autoloader** – Adds `--optimize-autoloader` to that initial install (only available when auto-install is checked).

Progress is shown as the project is fetched and (if selected) Composer dependencies are installed.

## Manage a PHP Application

Once installed, open the app from its listing to manage it. The manage page has four tabs:

### Dashboard

Basic info about the app: domain, docroot, and PHP version in use.

### Composer

Re-run Composer against the project's `composer.json` at any time, without reinstalling the project:

* **Composer install** – Installs dependencies as currently declared in `composer.json`/`composer.lock`.
* **Composer update** – Updates dependencies to their latest allowed versions.
* **Optimize autoloader** – Adds `--optimize-autoloader` to whichever action you run.
* **Edit composer.json** – Opens `composer.json` directly in the File Manager's editor.

Output from the last run is shown in a live output pane below the buttons.

### Logs

Full history of every Composer run (install/update) for this app, including timestamps and command output.

### Remove

**Delete Application** removes the app's tracking entry from the Composer App Manager only — docroot files and any database are left untouched, so you can still access them manually or re-add the app later (same behavior as removing a Node.js/Python application).
