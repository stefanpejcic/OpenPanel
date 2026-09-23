// OpenAdmin screenshot manifest: one entry per docs/admin page, same conventions as shots.mjs
export const BASE_URL = process.env.ADMIN_URL || 'https://demo.openpanel.com:2087';
export const STATE_PATH = '.auth/admin_state.json';
export const OUT_DIR = '../../static/img/openadmin-screenshots';
// below 1200px the admin header's resource bar is pinned to the bottom of the screen
export const VIEWPORT_WIDTH = 1280;

import { rename, all, click, fill, fillVisible, selectFirst, maskIPs, hideNotices, hideText, markCard, tab, markSection, markSectionId, clickAlpine } from './helpers.mjs';

// demo host names and addresses -> example ones
const DEMO_NAMES = {
  'website-builder.tests.openpanel.org': 'portfolio.example.com',
  'redirect.tests.openpanel.org': 'old.example.com',
  'python.tests.openpanel.org': 'api.example.com',
  'nodejs.tests.openpanel.org': 'app.example.com',
  'files.tests.openpanel.org': 'files.example.com',
  'php.tests.openpanel.org': 'shop.example.com',
  'to-be-removed.com': 'example.net',
  'anotheruser.com': 'maria-design.com',
  testinguser: 'john',
  anotheruser: 'maria',
  dorotea: 'agency',
  wpxss: 'shopowner',
  'test@test.com': 'john@example.com',
  musing_johnson7: 'owner',
  'dorotea.': 'agency.',
  'wp.tests.openpanel.org': 'blog.example.com',
  'tests.openpanel.org': 'example.com',
  'demo.openpanel.com': 'server.example.com',
};

// runs on every shot after its own prepare(), so selectors there can still use the real names
// API keys and tokens shown in settings fields
const maskSecrets = async page => {
  await page.evaluate(() => {
    document.querySelectorAll('main input, main textarea').forEach(i => {
      if (/AIza[\w-]{20,}|\b(sk|pk|key|tok)[_-][\w-]{16,}/.test(i.value)) i.value = i.value.replace(/AIza[\w-]{20,}/g, 'AIzaSy-your-api-key').replace(/\b(sk|pk|key|tok)[_-][\w-]{16,}/g, '$1_xxxxxxxxxxxxxxxx');
    });
  });
};

export const globalPrepare = all(rename(DEMO_NAMES, 'body'), maskIPs('main'), hideNotices, maskSecrets);

export const pages = {
  'accounts/users': {
    url: '/users',
    shots: [
      { name: 'list', alt: 'Users page listing OpenPanel accounts with their status, plan, limits, usage and the Impersonate button', crop: 'content' },
      {
        name: 'columns',
        alt: 'Show Columns menu of the Users table with a toggle for each column',
        prepare: click('#dropdownToggleButton', 600),
        crop: { from: '#dropdownToggleButton', to: '#dropdownToggle', pad: 12 },
      },
      { name: 'new', url: '/user/new', alt: 'Create New user form with username, email, password, webserver, database type and hosting plan', prepare: fillVisible(['john', 'john@example.com']), crop: 'content' },
      { name: 'stats', url: '/users/testinguser#stats', alt: 'Overview tab of a user with gauges for storage, inodes, CPU and memory usage and the View Past Usage button', prepare: async page => { await page.waitForTimeout(2500); }, crop: { from: 'main', to: '#get_docker_history_usage', fromTop: true, pad: 24 } },
      {
        name: 'history',
        url: '/users/testinguser',
        alt: 'Past resource usage table of a user with the date, CPU and memory usage and the number of tasks',
        prepare: async page => {
          await page.waitForTimeout(1500);
          await page.locator('#get_docker_history_usage').click();
          await page.waitForTimeout(2500);
        },
        crop: { from: 'main h4:text-is("Last updated:")', to: '#docker_stats_history tbody tr:nth-of-type(6)', pad: 16, right: true },
      },
      {
        name: 'info',
        url: '/users/testinguser',
        alt: 'User details on the Overview tab: username, email, plan, locale, 2FA status, IP address, location, server, docker context and setup time',
        prepare: async page => { await page.waitForTimeout(2000); },
        crop: { from: 'main span:text-is("Username:")', to: 'main span:text-is("Setup time:")', pad: 16, right: true },
      },
      {
        name: 'transfer',
        url: '/users/testinguser#export',
        alt: 'Transfer to another server form on the Export tab with the remote server, SSH credentials and the Live Transfer option',
        prepare: async page => {
          await page.waitForTimeout(2000);
          await page.locator('main button:has-text("Transfer to another server")').click();
          await page.waitForTimeout(600);
        },
        crop: 'content',
      },
      { name: 'services', url: '/users/testinguser#services', alt: 'Services tab listing the user containers with their CPU and memory usage, PIDs and actions', prepare: async page => { await page.waitForTimeout(2500); }, crop: { from: 'main', to: `main [x-show="activeTab === 'services'"] table tbody tr:nth-of-type(6)`, fromTop: true } },
      { name: 'storage', url: '/users/testinguser#storage', alt: 'Storage tab with the user volumes, containers and images from docker system df', prepare: async page => { await page.waitForTimeout(2500); }, crop: { content: 'main', maxHeight: 1000 } },
      { name: 'permissions', url: '/users/testinguser#permissions', alt: 'Permissions tab where the enabled OpenPanel features follow the plan or are set per user', prepare: async page => { await page.waitForTimeout(2500); }, crop: { from: 'main', to: `main [x-show="activeTab === 'permissions'"] table tbody tr:nth-of-type(6)`, fromTop: true } },
      { name: 'activity', url: '/users/testinguser#activity', alt: 'Activity Log tab with the date and the action performed', prepare: async page => { await page.waitForTimeout(2500); }, crop: { from: 'main', to: `main [x-show="activeTab === 'activity'"] table tbody tr:nth-of-type(8)`, fromTop: true } },
      { name: 'logins', url: '/users/testinguser#logins', alt: 'Login Log tab with the date, country and IP address of each login', prepare: async page => { await page.waitForTimeout(2500); }, crop: { from: 'main', to: `main [x-show="activeTab === 'logins'"] table tbody tr:nth-of-type(8)`, fromTop: true } },
      { name: 'edit', url: '/users/testinguser#edit', alt: 'Edit tab with the username, email, password, IP address, reseller and hosting package fields', prepare: async page => { await page.waitForTimeout(2500); }, crop: { from: 'main', to: `main [x-show="activeTab === 'edit'"] a[href="/users"]`, fromTop: true, pad: 24 } },
      { name: 'export', url: '/users/testinguser#export', alt: 'Export tab with the Generate full account backup option and the list of existing backups', prepare: async page => { await page.waitForTimeout(2500); }, crop: 'content' },
      { name: 'suspend', url: '/users/testinguser#suspend', alt: 'Suspend tab asking to type the username before suspending the account', prepare: all(async page => { await page.waitForTimeout(2500); }, fillVisible(['testinguser'])), crop: 'content' },
      { name: 'delete', url: '/users/testinguser#delete', alt: 'Delete tab asking to type the username before deleting the account permanently', prepare: all(async page => { await page.waitForTimeout(2500); }, fillVisible(['testinguser'])), crop: 'content' },
    ],
  },
  '001_dashboard': {
    url: '/dashboard',
    shots: [
      { name: 'window', alt: 'OpenAdmin dashboard with the summary bar and the User Activity, Latest News, System Information and Resource Usage widgets', viewportHeight: 860 },
      { name: 'summary', alt: 'Summary bar with the number of nodes, containers, users, domains, websites, packages and emails', crop: { from: '#tour-quick-summary', pad: 12 } },
      { name: 'activity', alt: 'User Activity widget with the latest actions of all OpenPanel users', crop: { from: '#tour-user-activity', pad: 12 } },
      { name: 'news', alt: 'Latest News widget with articles from the OpenPanel blog', prepare: async page => { await page.waitForTimeout(2000); }, crop: { from: '#tour-latest-news', pad: 12 } },
      { name: 'sysinfo', alt: 'System Information widget with hostname, IPv4 address, OS, OpenPanel version, server time, kernel, CPU, uptime, processes and package updates', crop: { from: '#tour-system-info', pad: 12 } },
      { name: 'usage', alt: 'Resource Usage widget with a chart of CPU and RAM usage over the last hour and the View history link', prepare: async page => { await page.waitForTimeout(7000); }, viewportHeight: 1600, crop: { from: '#tour-usage-graphs', pad: 12 } },
      {
        name: 'sse',
        alt: 'Resource usage bar in the top-right corner of OpenAdmin with the load tooltip open',
        prepare: async page => {
          await page.waitForTimeout(2500);
          await page.locator('#sse_load').hover();
          await page.waitForTimeout(700);
        },
        viewportHeight: 700,
        crop: { from: 'header a[role=tab] >> nth=0', to: '#tooltip-sse_load', clamp: false, pad: 12, right: true },
      },
      {
        name: 'search',
        alt: 'Search box in the OpenAdmin sidebar with matching pages, users and websites',
        prepare: async page => {
          await page.locator('#searchInput').pressSequentially('wp', { delay: 80 });
          await page.waitForTimeout(2500);
        },
        viewportHeight: 860,
        crop: { from: '#sidebar', clamp: false, height: 620 },
      },
      { name: 'menu', alt: 'OpenAdmin sidebar menu with the Accounts, Hosting Plans, Domains, Emails, Services, Security, Settings and Advanced sections', viewportHeight: 900, crop: { from: '#sidebar', clamp: false } },
    ],
  },
  '001_dashboard_dark': {
    url: '/dashboard',
    init: () => localStorage.setItem('color-theme', 'dark'),
    shots: [{ name: 'window', alt: 'OpenAdmin dashboard in dark mode', viewportHeight: 860 }],
  },
  '002_notifications': {
    url: '/notifications',
    shots: [{ name: 'list', alt: 'Notifications page listing recorded system alerts with the time, notification and details, and the Edit Settings, Pause notifications, Acknowledge All and Delete All buttons', crop: { from: 'main', to: 'main table tbody tr:nth-of-type(10)', fromTop: true } }],
  },
  license: {
    url: '/license',
    // the demo shows the real license owner, key and billing email
    prepare: async page => {
      await page.evaluate(() => {
        const w = document.createTreeWalker(document.querySelector('main'), NodeFilter.SHOW_TEXT);
        for (let n; (n = w.nextNode()); ) {
          n.textContent = n.textContent
            .replace(/[\w.+-]+@[\w-]+\.[\w.]+/g, 'billing@example.com')
            .replace(/(Owner:\s*)?Stefan Pejcic/g, (m, o) => (o || '') + 'John Doe');
        }
        document.querySelectorAll('main input').forEach(i => { if (/^enterprise/i.test(i.value)) i.value = 'enterprise-0123456789abcdef'; });
      });
    },
    shots: [
      { name: 'key', alt: 'License key section with the key field, the Save key, Re-Verify and Downgrade buttons and the license details', crop: { from: 'main', to: 'main :text("Valid IP")', fromTop: true, pad: 60, right: true } },
      { name: 'support', alt: 'Contact Support section with the Generate report button', crop: { from: 'main h2:has-text("Contact Support")', to: '#tour-generate-report-btn', pad: 24, right: true } },
    ],
  },
  'accounts/administrators': {
    url: '/administrators',
    shots: [
      { name: 'list', alt: 'Administrators page listing OpenAdmin users with their status, role, 2FA, passkeys and last login', crop: 'content' },
      {
        name: 'menu',
        alt: 'Edit menu of an administrator with the Rename and Change Password options',
        prepare: click('main button#administrator', 600),
        crop: { from: 'main table thead', to: '#dropdown-administrator', pad: 12 },
      },
      { name: 'new', alt: 'Create New administrator form with the username and password fields', prepare: all(click('#tour-create-admin-btn', 600), fillVisible(['support_admin'])), crop: 'content' },
      { name: 'rename', url: '/administrators/rename/administrator', alt: 'Rename administrator form with the new username field and the Change Username button', crop: 'content' },
      { name: 'password', url: '/administrators/password/administrator', alt: 'Change Password form for an administrator', crop: 'content' },
    ],
  },
  'accounts/resellers': {
    url: '/resellers',
    shots: [{ name: 'list', alt: 'Resellers page with the reseller table columns and the Enable Resellers button', crop: 'content' }],
  },
  'plans/hosting_plans': {
    url: '/plans',
    shots: [
      { name: 'list', alt: 'User Packages page listing hosting plans with their memory, CPU, disk, inodes, port speed and other limits', crop: { from: 'main', to: 'main table tbody tr:last-of-type', fromTop: true, pad: 24 } },
      { name: 'menu', alt: 'Row menu of a hosting plan that is in use, with the Edit option', prepare: click('main button#\\31', 600), crop: { from: 'main table thead', to: '#dropdown-1', pad: 12, right: true } },
      { name: 'new', url: '/plans/new', alt: 'New Package form with name, description, disk, inodes, CPU, memory, port speed, domains, websites, databases, email, FTP and feature set fields', prepare: fill({ 'input[name=name]': 'Business plan', 'input[name=description]': 'For small business sites', 'input[name=disk_limit]': '20', 'input[name=inodes_limit]': '500000', 'input[name=cpu]': '2', 'input[name=ram]': '4', 'input[name=bandwidth]': '200', 'input[name=domains_limit]': '5', 'input[name=websites_limit]': '10', 'input[name=db_limit]': '10', 'input[name=email_limit]': '50', 'input[name=max_email_quota]': '5G', 'input[name=max_hourly_email]': '200', 'input[name=ftp_limit]': '10' }), crop: 'content' },
      { name: 'edit', url: '/plans/1', alt: 'Edit plan page with the limits of an existing hosting package', crop: 'content' },
      { name: 'usage', url: '/users?plan=Standard%20plan', alt: 'Users page filtered to the users on one hosting plan', crop: 'content' },
    ],
  },
  'plans/feature-manager': {
    url: '/features',
    shots: [
      { name: 'index', alt: 'Feature Manager page with the Create and Manage sections', crop: 'content' },
      { name: 'edit', url: '/features/default', alt: 'Feature set edit page listing features with a toggle, name, description and type, and the Enable All, Disable All and Save buttons', crop: { from: 'main', to: 'main table tbody tr:nth-of-type(10)', fromTop: true } },
    ],
  },
  'domains/domains': {
    url: '/domains',
    shots: [
      { name: 'list', alt: 'Domains page listing all domains with their status, PHP version, webserver, SSL, WAF and owner', crop: { from: 'main', to: 'main table tbody tr:nth-of-type(10)', fromTop: true } },
      {
        name: 'add',
        alt: 'Add Domain form with the domain name and the user to add it to',
        prepare: all(click('#tour-add-domain-btn', 600), fillVisible(['mycoffeeshop.com'])),
        crop: { from: 'main', to: 'main table thead', fromTop: true },
      },
      {
        name: 'actions',
        alt: 'Actions menu of a domain with Edit DNS Zone, Manage SSL, Edit Virtual Host, Edit Apache Config, Edit Caddyfile, Suspend domain and Delete domain',
        prepare: async page => {
          await page.evaluate(() => document.querySelectorAll('main .overflow-x-auto, main .overflow-auto').forEach(e => (e.scrollLeft = e.scrollWidth)));
          await page.locator('main tbody button[data-dropdown-toggle]').first().click();
          await page.waitForTimeout(600);
        },
        crop: { from: 'main table thead', to: 'main [id^="dropdown-"]:visible >> nth=0', pad: 12, right: true },
      },
      {
        name: 'delete',
        alt: 'Delete domain confirmation asking to type the domain name before deleting it permanently',
        prepare: async page => {
          await page.evaluate(() => document.querySelectorAll('main .overflow-x-auto, main .overflow-auto').forEach(e => (e.scrollLeft = e.scrollWidth)));
          await page.locator('main tbody button[data-dropdown-toggle]').first().click();
          await page.waitForTimeout(400);
          await page.locator('main button:has-text("Delete domain") >> visible=true').first().click();
          await page.waitForTimeout(600);
          await page.evaluate(() => document.querySelectorAll('main [id^="dropdown-"]').forEach(d => d.classList.add('hidden')));
        },
        viewportHeight: 820,
        crop: { from: 'main h3:has-text("Delete Domain"), main h2:has-text("Delete Domain")', to: 'main [x-show="showDeleteModal"] input:visible', clamp: false, pad: 10 },
      },
    ],
  },
  'domains/dns': {
    url: '/domains/dns',
    shots: [
      { name: 'select', alt: 'DNS Zone Editor with the domain dropdown', crop: 'content' },
      { name: 'edit', url: '/domains/dns/wp.tests.openpanel.org', alt: 'DNS Zone Editor showing the zone file of a domain with the Save button', crop: { content: 'main', maxHeight: 820 } },
    ],
  },
  'domains/dns_templates': {
    url: '/domains/zone-templates',
    shots: [{ name: 'page', alt: 'Edit Zone Templates page with the IPv4 and IPv6 zone templates and the Restore Default and Save Files buttons', crop: { content: 'main', maxHeight: 900 } }],
  },
  'domains/file_templates': {
    url: '/domains/file-templates',
    shots: [
      { name: 'default', url: '/domains/file-templates', alt: 'Default Page template editor with a live preview of the page shown on new domains', prepare: markSection('Default Page'), crop: { from: '[data-shot=section]', pad: 16 } },
      { name: 'suspended-website', url: '/domains/file-templates', alt: 'Suspended Website template editor with a live preview', prepare: markSection('Suspended Website'), crop: { from: '[data-shot=section]', pad: 16 } },
      { name: 'suspended-user', url: '/domains/file-templates', alt: 'Suspended User template editor with a live preview', prepare: markSection('Suspended User'), crop: { from: '[data-shot=section]', pad: 16 } },
      { name: 'apache', url: '/domains/file-templates', alt: 'Apache VirtualHost template editor', prepare: markSection('Apache VirtualHost'), crop: { from: '[data-shot=section]', pad: 16 } },
      { name: 'nginx', url: '/domains/file-templates', alt: 'Nginx VirtualHost template editor', prepare: markSection('Nginx VirtualHost'), crop: { from: '[data-shot=section]', pad: 16 } },
      { name: 'openresty', url: '/domains/file-templates', alt: 'OpenResty VirtualHost template editor', prepare: markSection('OpenResty VirtualHost'), crop: { from: '[data-shot=section]', pad: 16 } },
      { name: 'varnish', url: '/domains/file-templates', alt: 'Varnish template editor', prepare: markSection('Varnish Template'), crop: { from: '[data-shot=section]', pad: 16 } },
    ],
  },
  'domains/dns-cluster': {
    url: '/domains/dns-cluster',
    shots: [{ name: 'page', alt: 'DNS Cluster Management page with the Enable DNS Clustering button', crop: 'content' }],
  },
  'security/2fa': {
    url: '/security/2fa',
    // the demo shows a live setup secret and its QR code
    prepare: async page => {
      await page.evaluate(() => {
        document.querySelectorAll('main img[alt="2FA QR code"]').forEach(i => (i.style.filter = 'blur(7px)'));
        const w = document.createTreeWalker(document.querySelector('main'), NodeFilter.SHOW_TEXT);
        for (let n; (n = w.nextNode()); ) n.textContent = n.textContent.replace(/\b[A-Z2-7]{16,}\b/g, 'JBSWY3DPEHPK3PXP');
      });
    },
    shots: [{ name: 'page', alt: 'Two-Factor Authentication page with the QR code to scan, the secret key and the field for the authentication code', crop: 'content' }],
  },
  'security/basic_auth': {
    url: '/security/basic_auth',
    shots: [{ name: 'page', alt: 'Basic Authentication page with the Enable dropdown and the username and password fields', prepare: fillVisible(['admin']), crop: 'content' }],
  },
  'security/firewall': {
    url: '/security/firewall',
    shots: [{ name: 'csf', alt: 'Sentinel Security and Firewall (CSF) page with the firewall status and server information actions', crop: { content: 'main', maxHeight: 900 } }],
  },
  'security/imunify': {
    url: '/security/imunify/',
    shots: [{ name: 'not-running', alt: 'ImunifyAV page telling that the GUI is not running yet, with the command to start it', crop: 'content' }],
  },
  'security/passkeys': {
    url: '/security/passkeys',
    shots: [{ name: 'page', alt: 'Passkeys page with the registered passkeys and the Add a passkey form', prepare: fillVisible(['Work laptop']), crop: 'content' }],
  },
  'security/waf': {
    url: '/security/waf',
    shots: [
      { name: 'page', alt: 'WAF page with the CorazaWAF Enable dropdown and the number of active rule sets', crop: 'content' },
      { name: 'rules', url: '/security/waf/rules', alt: 'WAF rule sets page listing each rule set with its number of rules, status and the View and Disable actions', crop: { from: 'main', to: 'main table tbody tr:nth-of-type(10)', fromTop: true } },
    ],
  },
  'backups/system': {
    url: '/backups/system',
    shots: [
      { name: 'backups', url: '/backups/system#backups', alt: 'System Backups page on the Backups tab with the Run Backup Now button', crop: 'content' },
      { name: 'runs', url: '/backups/system#runs', alt: 'Runs tab of System Backups with the log of past backup runs', crop: 'content' },
      { name: 'settings', url: '/backups/system#settings', alt: 'Settings tab of System Backups with the destination directory and retention in days', crop: 'content' },
    ],
  },
  'backups/user': {
    url: '/backups/user',
    shots: [
      { name: 'settings', url: '/backups/user#settings', alt: 'User Backups Settings tab with the backup schedule dropdown', crop: 'content' },
      { name: 'configuration', url: '/backups/user#configuration', alt: 'User Backups Configuration tab with the default backup.env editor and the Restore Default button', crop: 'content' },
      { name: 'runs', url: '/backups/user#runs', alt: 'User Backups Runs tab with the log of past backup runs', crop: 'content' },
    ],
  },
  'services/status': {
    url: '/services',
    shots: [
      { name: 'list', alt: 'Services page listing system services with their status, version, real name, type, port, monitoring and actions', crop: { from: 'main', to: 'main table tbody tr:nth-of-type(12)', fromTop: true } },
      { name: 'edit', url: '/services/edit', alt: 'Edit Services page for changing which services are shown and monitored', crop: { content: 'main', maxHeight: 900 } },
    ],
  },
  'services/ftp': {
    url: '/services/ftp',
    shots: [
      { name: 'accounts', alt: 'FTP page on the Accounts tab with the FTP service status and the table of FTP accounts', crop: 'content' },
      { name: 'configuration', url: '/services/ftp/settings', alt: 'FTP Configuration tab with the FTP server settings', crop: 'content' },
    ],
  },
  'services/limits': {
    url: '/services/limits',
    shots: [{ name: 'page', alt: 'Service Limits page with CPU and memory limits for each system service', crop: { content: 'main', maxHeight: 1000 } }],
  },
  'services/logs': {
    url: '/services/logs',
    prepare: async page => {
      // pick the first log that actually has content on the demo
      const sel = page.locator('main select').first();
      const opts = await sel.locator('option').evaluateAll(os => os.map(o => o.value).filter(Boolean));
      for (const o of opts) {
        await sel.selectOption(o);
        await page.waitForTimeout(2000);
        const t = await page.evaluate(() => (document.querySelector('main pre, main code, main textarea')?.innerText || '').trim());
        if (t.length > 200 && !/not found|Select a log/i.test(t)) break;
      }
    },
    shots: [{ name: 'page', alt: 'Log Viewer with a log file selected, its content, and the Delete and Download buttons', crop: { content: 'main', maxHeight: 900 } }],
  },
  'services/podman': {
    url: '/services/podman',
    shots: [
      { name: 'info', alt: 'Podman page on the Info tab with the output of podman info', crop: { content: 'main', maxHeight: 800 } },
      { name: 'images', alt: 'Podman Images tab listing container images with their tag, size and actions', prepare: clickAlpine("tab = 'images'", 2500), crop: { content: 'main', maxHeight: 900 } },
      { name: 'volumes', alt: 'Podman Volumes tab listing volumes', prepare: clickAlpine("tab = 'volumes'", 2500), crop: { content: 'main', maxHeight: 900 } },
      { name: 'diskusage', alt: 'Podman Disk Usage tab with the space used by images, containers and volumes', prepare: clickAlpine("tab = 'diskusage'", 2500), crop: { content: 'main', maxHeight: 900 } },
    ],
  },
  'settings/general': {
    url: '/settings/general',
    shots: [
      { name: 'domain', alt: 'Domain section of General Settings with the hostname field', prepare: markSectionId('Domain'), crop: { from: '[data-shot=section]', pad: 16 } },
      { name: 'ssl', alt: 'SSL section of General Settings with the current SSL status and the Manage SSL button', prepare: markSectionId('SSL'), crop: { from: '[data-shot=section]', pad: 16 } },
      { name: 'ports', alt: 'Ports section of General Settings with the OpenAdmin and OpenPanel ports', prepare: markSectionId('Ports'), crop: { from: '[data-shot=section]', pad: 16 } },
      { name: 'redirect', alt: 'Redirect section of General Settings with the /openpanel redirect name', prepare: markSectionId('Redirect'), crop: { from: '[data-shot=section]', pad: 16 } },
      { name: 'debug', alt: 'Debugging (Dev Mode) section of General Settings with the Dev mode toggle', prepare: markSectionId('Debugging'), crop: { from: '[data-shot=section]', pad: 16 } },
    ],
  },
  'settings/openpanel': {
    url: '/settings/open-panel',
    shots: [
      { name: 'branding', alt: 'Branding section with the brand name, logo, favicon and logout URL', prepare: markSectionId('Branding'), crop: { from: '[data-shot=section]', pad: 16 } },
      { name: 'nameservers', alt: 'Nameservers section with the ns1 to ns4 fields', prepare: markSectionId('Nameservers'), crop: { from: '[data-shot=section]', pad: 16 } },
      { name: 'users', alt: 'Users section with toggles for what OpenPanel users can change', prepare: markSectionId('Users'), crop: { from: '[data-shot=section]', pad: 16 } },
      { name: 'display', alt: 'Display section with the OpenPanel interface options', prepare: markSectionId('Display'), crop: { from: '[data-shot=section]', pad: 16 } },
      { name: 'filemanager', alt: 'File Manager section with the file manager options', prepare: markSectionId('File Manager'), crop: { from: '[data-shot=section]', pad: 16 } },
      { name: 'databases', alt: 'Databases section with the database options', prepare: markSectionId('Databases'), crop: { from: '[data-shot=section]', pad: 16 } },
      { name: 'security', alt: 'Security section with the OpenPanel login and session options', prepare: markSectionId('Security'), crop: { from: '[data-shot=section]', pad: 16 } },
      { name: 'statistics', alt: 'Statistics section with the resource usage and statistics options', prepare: markSectionId('Statistics'), crop: { from: '[data-shot=section]', pad: 16 } },
    ],
  },
  'settings/modules': {
    url: '/settings/modules',
    shots: [{ name: 'page', alt: 'Modules page with a card per module showing its path, type, description and Activate toggle', crop: { content: 'main', maxHeight: 1000 } }],
  },
  'settings/defaults': {
    url: '/settings/defaults',
    shots: [
      { name: 'page', alt: 'Edit Defaults page with the default webserver, database, Varnish cache, PHP version and autostart services for new users', crop: { content: 'main', maxHeight: 1100 } },
      { name: 'services', alt: 'Service limits on the Edit Defaults page for new user accounts', prepare: markSectionId('Services'), crop: { from: '[data-shot=section]', pad: 16, height: 900 } },
    ],
  },
  'settings/custom_code': {
    url: '/settings/custom-code',
    shots: [
      { name: 'css', alt: 'Custom CSS editor with the Insert Example button', prepare: markSectionId('Custom CSS'), crop: { from: '[data-shot=section]', pad: 16 } },
      { name: 'js', alt: 'Custom JS editor', prepare: markSectionId('Custom JS'), crop: { from: '[data-shot=section]', pad: 16 } },
      { name: 'header', alt: 'Code in Header editor for the head tag of every OpenPanel page', prepare: markSectionId('Code in Header'), crop: { from: '[data-shot=section]', pad: 16 } },
      { name: 'footer', alt: 'Code in Footer editor', prepare: markSectionId('Code in Footer'), crop: { from: '[data-shot=section]', pad: 16 } },
      { name: 'section', alt: 'Custom Section editor for the OpenPanel dashboard', prepare: markSectionId('Custom Section'), crop: { from: '[data-shot=section]', pad: 16 } },
      { name: 'plugins', alt: 'WordPress Plugins Set editor', prepare: markSectionId('WordPress Plugins Set'), crop: { from: '[data-shot=section]', pad: 16 } },
      { name: 'themes', alt: 'WordPress Themes Set editor', prepare: markSectionId('WordPress Themes Set'), crop: { from: '[data-shot=section]', pad: 16 } },
      { name: 'after', alt: 'After update script editor', prepare: markSectionId('After update'), crop: { from: '[data-shot=section]', pad: 16 } },
      { name: 'before', alt: 'Before startup script editor', prepare: markSectionId('Before startup'), crop: { from: '[data-shot=section]', pad: 16 } },
    ],
  },
  'settings/locales': {
    url: '/settings/locales',
    shots: [{ name: 'list', alt: 'Languages page listing locales with their provider, install status, default and Set as Default buttons', crop: { from: 'main', to: 'main table tbody tr:nth-of-type(12)', fromTop: true } }],
  },
  'settings/notifications': {
    url: '/settings/notifications',
    shots: [
      { name: 'email', alt: 'Email section with the address for notifications and daily usage reports', prepare: markSectionId('email'), crop: { from: '[data-shot=section]', pad: 16 } },
      { name: 'webhook', alt: 'Webhook section with the webhook URL field', prepare: markSectionId('webhook_url'), crop: { from: '[data-shot=section]', pad: 16 } },
      { name: 'services', alt: 'Services section with a toggle per monitored service', prepare: markSectionId('services'), crop: { from: '[data-shot=section]', pad: 16 } },
      { name: 'thresholds', alt: 'Resource Usage section with the load, CPU, memory, disk and swap thresholds', prepare: markSectionId('thresholds'), crop: { from: '[data-shot=section]', pad: 16 } },
      { name: 'server', alt: 'Server actions section with toggles for reboot, OOM, DNS and other server events', prepare: markSectionId('server-actions'), crop: { from: '[data-shot=section]', pad: 16 } },
      { name: 'users', alt: 'User actions section with toggles for account and domain change notifications', prepare: markSectionId('user-actions'), crop: { from: '[data-shot=section]', pad: 16 } },
      { name: 'ssh', alt: 'SSH Allowlist section with the allowed IP addresses', prepare: markSectionId('ssh_allowlist'), crop: { from: '[data-shot=section]', pad: 16 } },
      { name: 'smtp', alt: 'SMTP section with the mail server settings and the Test SMTP connection button', prepare: markSectionId('smtp'), crop: { from: '[data-shot=section]', pad: 16 } },
    ],
  },
  'settings/php': {
    url: '/settings/php',
    shots: [
      { name: 'default', alt: 'Default version section linking to the User Defaults page', prepare: markSectionId('Default version'), crop: { from: '[data-shot=section]', pad: 16 } },
      { name: 'options', alt: 'Available Options section with the PHP options users can edit', prepare: markSectionId('Available Options'), crop: { from: '[data-shot=section]', pad: 16 } },
      { name: 'ini', alt: 'Default PHP.INI Files section with a php.ini per PHP version and Restore Default and Save buttons', prepare: markSectionId('Default PHP.INI Files'), crop: { from: '[data-shot=section]', pad: 16 } },
    ],
  },
  'settings/updates': {
    url: '/settings/updates',
    shots: [
      { name: 'current', alt: 'Current version section with the installed and latest version and the changelog link', prepare: markSectionId('Current version'), crop: { from: '[data-shot=section]', pad: 16 } },
      { name: 'auto', alt: 'Auto Updates section with the update preference dropdown', prepare: markSectionId('Auto Updates'), crop: { from: '[data-shot=section]', pad: 16 } },
      { name: 'logs', alt: 'Update Logs section listing the update log files', prepare: markSectionId('Update Logs'), crop: { from: '[data-shot=section]', pad: 16 } },
      { name: 'rollback', alt: 'Rollback section for returning to a previous version', prepare: markSectionId('Rollback'), crop: { from: '[data-shot=section]', pad: 16 } },
    ],
  },
  'settings/api': {
    url: '/settings/api',
    shots: [{ name: 'page', alt: 'API Access page showing API access as disabled with the Enable API access button', crop: 'content' }],
  },
  'advanced/crons': {
    url: '/server/crons',
    shots: [{ name: 'page', alt: 'Scheduler page listing OpenPanel system jobs with their cron schedule fields and command', crop: { content: 'main', maxHeight: 1000 } }],
  },
  'advanced/demo-mode': {
    url: '/server/demo-mode',
    shots: [{ name: 'page', alt: 'Demo Mode page saying demo mode is active and can only be disabled from the terminal', crop: 'content' }],
  },
  'advanced/migrate': {
    url: '/server/migrate',
    shots: [{ name: 'form', alt: 'Server Migration form with the remote host, username and password fields and the Start Migration button', prepare: fillVisible(['203.0.113.50', 'root']), crop: 'content' }],
  },
  'advanced/processes': {
    url: '/server/processes',
    shots: [{ name: 'list', alt: 'Process Manager listing system processes with their PID, owner, priority, CPU and memory usage, command, and Trace and Kill links', crop: { from: 'main', to: 'main table tbody tr:nth-of-type(14)', fromTop: true } }],
  },
  'advanced/reboot': {
    url: '/server/reboot',
    shots: [{ name: 'page', alt: 'Server Reboot page explaining graceful and forceful reboots, with the reboot type dropdown and the Reboot Server button', crop: 'content' }],
  },
  'advanced/resource-usage': {
    url: '/server/resource-usage',
    shots: [
      { name: 'page', alt: 'Resource Usage page with the Load, RAM, CPU, Disk and Network tabs and a live chart of the average load', prepare: async page => { await page.waitForTimeout(4000); }, viewportHeight: 900, crop: { content: 'main', maxHeight: 700 } },
      { name: 'history', url: '/server/resource-usage/history', alt: 'Resource usage history with charts and a table of past snapshots', prepare: async page => { await page.waitForTimeout(4000); }, viewportHeight: 1400, crop: { content: 'main', maxHeight: 1000 } },
    ],
  },
  'advanced/ssh': {
    url: '/server/ssh',
    shots: [
      { name: 'basic', alt: 'SSH Access page on the Basic tab with the SSH port, root login, password and public key authentication settings', crop: 'content' },
      { name: 'advanced', alt: 'SSH Access Advanced tab with the sshd configuration editor', prepare: clickAlpine("tab = 'advanced'", 1200), crop: { content: 'main', maxHeight: 900 } },
    ],
  },
  'advanced/swap': {
    url: '/server/swap',
    shots: [{ name: 'page', alt: 'Swap page with the current swap usage, the swap size field with Apply, and the Drop Swap button', crop: 'content' }],
  },
  'advanced/timezone': {
    url: '/server/timezone',
    shots: [{ name: 'page', alt: 'Change TimeZone page with the current timezone and the timezone dropdown', crop: 'content' }],
  },
  'advanced/cpanel': {
    url: '/import/cpanel',
    shots: [{ name: 'page', alt: 'Account Imports page with the import logs and the Import Account button', crop: 'content' }],
  },
  '000_intro': {
    url: '/login',
    anonymous: true,
    // the demo prefills its real credentials
    prepare: fill({ 'input[name=username]': 'admin', 'input[name=password]': '' }),
    shots: [{ name: 'login', alt: 'OpenAdmin login form with the username and password fields, Remember me, passkey sign-in and the Switch to OpenPanel button', viewportHeight: 760 }],
  },
};
