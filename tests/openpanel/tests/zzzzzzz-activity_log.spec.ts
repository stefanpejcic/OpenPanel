import { test, expect } from '@playwright/test';

// covers all record_user_action(username, action) calls in all core .py files

const actionTemplates = [
  // account
  'changed password',
  'changed email address to {email}',
  'changed username from {old} to {new}',
  'Added {link} to Favorites',
  'removed {link} from Favorites',
  'password reset requested',
  'password changed using email confirmation',
  'logged in via admin API',
  'logged in with a password',
  'logged in with 2FA code',
  'logged out.',
  'changed notification preferences for the account',
  'terminated session {token}',
  'enabled 2FA for account',
  'disabled 2FA for account',

  // api
  'logged in via user API',

  // cache (varnish)
  'enabled Varnish',
  'disabled Varnish',
  'enabled Varnish caching for domain {domain}',
  'disabled Varnish caching for domain {domain}',

  // crons
  'added a new cron job',
  'edited cron file',
  'edited cron job',
  'deleted a cron job',

  // dns
  'updated DNS record to {content} for domain {domain}',
  'deleted DNS record {record} for domain {domain}',
  'reset DNS zone for domain {domain}',
  'exported DNS zone for domain {domain}',
  'added DNS record {record} for domain {domain}',
  'edited DNS zone for domain {domain}',

  // docker
  'added container {service}',
  'edited container {service}',
  'deleted container {service}',
  'switched mysql type to: {type}',
  'switched webserver type to: {type}',
  'changed image tag for {service} to {tag}',
  'updated cpu limit for {container} to: {value}',
  'updated ram limit for {container} to: {value}',
  'pulled image for {container} and activated',
  'pulled image for {container} and deactivated',
  'activated {container}',
  'deactivated {container}',
  'opened interactive terminal for service {container}',

  // domains / vhost / redirects
  'added domain {domain}',
  'suspended domain {domain}',
  'unsuspended domain {domain}',
  'deleted domain {domain}',
  'edited Virtual Host file for domain: {domain}',
  'deleted the redirect link {url} for domain {domain}',
  'created a redirect link {url} for domain {domain}',

  // emails
  'edited filter (GUI) for {email}',
  'edited filter (raw) for {email}',
  'accessed webmail login',
  'accessed webmail for {email}',
  'exported mailbox list ({count} addresses)',
  'imported {count} email accounts',
  'created email {email}',
  'downloaded {type} configuration for email {username}',
  'added alias {email} -> {target}',
  'deleted alias {email}',
  'deleted alias {email} -> {target}',
  'created alias {source} -> {target}',
  'suspended incoming emails',
  'unsuspended incoming emails',
  'suspended outgoing emails',
  'unsuspended outgoing emails',
  'updated password',

  // file manager / ftp / malware scan
  'created a new file {filename} via File Manager',
  'created a new folder {foldername} via File Manager',
  'uploaded a new file via File Manager',
  'Extracted {file} into {destination} using File Manager',
  'downloaded file from URL {url} via File Manager',
  'Created archive {archive} with File Manager',
  'renamed file /var/www/html/{old} to /var/www/html/{new} using File Manager',
  'moved {type} {path} to Trash using File Manager',
  'deleted {type} {path} using File Manager',
  'changed permissions to {perms} for {path} using File Manager',
  'copied directory {source} to {destination} using File Manager',
  'copied file {source} to {destination} using File Manager',
  'moved {source} to {destination} using File Manager',
  'edited file {path}',
  'restored {destination} from Trash using File Manager',
  'permanently deleted {type} {path} using File Manager',
  'Emptied trash using File Manager',
  'restored {path} from Trash using Restore All',
  'created FTP account {username}',
  'deleted FTP account {username}',
  'changed password for FTP account {username}',
  'changed path for FTP account {username}',
  'downloaded {type} configuration for FTP account {account}',

  // ip blocker
  'blocked IP addresses using IP Blocker',
  'removed all blocked IPs using IP Blocker',

  // mautic
  'uninstalled Mautic website for {domain}',
  'detached Mautic website',
  'installed Mautic on {domain}',

  // mysql
  'created a MySQL database {name}',
  'created a MySQL database user {user}',
  'used wizard to create MySQL database {db}, user {user}, privileges {privileges}',
  'assigned privileges {privileges} to MySQL user {user} on database {db}',
  'exported MYSQL database {db} to browser',
  'exported MYSQL database {db} to folder {path}',
  'deleted a MYSQL database {db}',
  'deleted a MySQL database user {user}',
  'changed password for MySQL database user {user}',
  'revoked all privileges for MySQL user {user} from database {db}',
  'edited MySQL configuration',
  'changed MySQL root user password',
  'enabled remote MySQL',
  'disabled remote MySQL',

  // postgresql
  'created a PostgreSQL database {db}',
  'created a PostgreSQL user {user}',
  'assigned all privileges to PostgreSQL user {user} on database {db}',
  'deleted a PostgreSQL database {db}',
  'deleted a PostgreSQL database user {user}',
  'changed password for PostgreSQL user {user}',
  'revoked all privileges for PostgreSQL user {user} from database {db}',
  'edited PostgreSQL configuration',
  'enabled remote PostgreSQL',
  'disabled remote PostgreSQL',

  // pgadmin / phpmyadmin
  'enabled pgAdmin',
  'disabled pgAdmin',
  'opened phpMyAdmin',

  // php
  'changed default PHP version for new domains to {version}',
  'changed PHP version for domain {domain} from {old} to {new}',
  'edited PHP {version} configuration using PHP Selector',

  // pm2
  'created a new {type} application on domain {domain}',
  'stopped container for application {site}',
  'started container for application {site}',
  'edited container for application {site}',
  'restarted container for application {site}',
  'deleted application {site}',

  // process manager / services / waf
  'terminated process {pid} using Process Manager',
  'enabled service {service}',
  'disabled service {service}',
  'updated WAF rules for domain {domain}',
  'enabled WAF for domain {domain}',
  'disabled WAF for domain {domain}',

  // website builder / websites / wordpress
  'removed Website Builder for {domain}',
  'detached Website Builder for {domain}',
  'installed Website Builder on {domain}',
  'initiated refresh for PageSpeed data on website {site}',
  'initiated refresh for WP vulnerability data on website {site}',
  'executed {type} install for application {site}',
  'updated debug options for WordPress website {site}',
  'updated site information for WordPress website {site}',
  'started core update for WordPress website in {docroot}',
  'edited auto-update preferences for WordPress website in {docroot}',
  'restored WordPress database backup from {path} on {domain}',
  'restored WordPress files backup from {path} on {domain}',
  'generated full backup for WordPress website {domain}',
  'generated database backup for WordPress website {domain}',
  'generated files backup for WordPress website {domain}',
  'enabled hardened rules for domain {domain} using WP Manager',
  'disabled hardened rules for domain {domain} using WP Manager',
  'cloned WordPress website from {source} to {destination}',
  'uninstalled WordPress website for {domain}',
  'detached WordPress website',
  'initiated scan for WordPress installations',
  'installed WordPress on domain {domain}',
  'generated auto-login link for wp-admin: {link}',
  'enabled maintenance mode',
  'disabled maintenance mode',
  'flushed cache',
  'Executed command: wp {command} for website {domain} using WP Manager',
];

// Two call sites in the source don't follow the "{placeholder}" convention, so they
// need their own regex rather than the generic template conversion below:
//   - malware_scan.py builds the message WITHOUT an f-prefix, so "{requested_path}"
//     is logged literally instead of being interpolated
//   - pgadmin.py concatenates "edited " + "; ".join(...) + "for pgAdmin" with no
//     space before "for", e.g. "edited theme; modefor pgAdmin"
const literalActionPatterns = [
  { label: 'initiated a Malware Scan for {requested_path} (logged literally, not interpolated)', regex: /initiated a Malware Scan for \{requested_path\}/ },
  { label: 'edited <settings>for pgAdmin (note: missing space before "for")', regex: /^edited .+for pgAdmin$/ },
];

function templateToRegex(template: string): RegExp {
  const escaped = template.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
  const pattern = escaped.replace(/\\\{[^}]*\\\}/g, '.*?');
  return new RegExp(pattern);
}

test('activity log contains all known recorded user actions', async ({ page }) => {
  await page.goto('/account/activity?show_all=true');

  const actionCells = page.locator('table tbody tr td:nth-child(2)');
  const count = await actionCells.count();

  const loggedActions: string[] = [];
  for (let i = 0; i < count; i++) {
    loggedActions.push((await actionCells.nth(i).innerText()).trim());
  }

  expect(loggedActions.length).toBeGreaterThan(0);
  console.log(`Found ${loggedActions.length} entries in the activity log table.\n`);

  const matchers = [
    ...actionTemplates.map((t) => ({ label: t, regex: templateToRegex(t) })),
    ...literalActionPatterns,
  ];

  let foundCount = 0;
  for (const { label, regex } of matchers) {
    const found = loggedActions.some((entry) => regex.test(entry));
    if (found) foundCount++;
    console.log(`[${found ? 'FOUND  ' : 'MISSING'}] ${label}`);
  }

  console.log(`\n${foundCount}/${matchers.length} known action types were present in the activity log.`);
});
