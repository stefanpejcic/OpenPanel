// screenshot manifest: one entry per docs page, see PLAN.md for conventions
export const BASE_URL = process.env.PANEL_URL || 'https://demo.openpanel.com:2083';
export const STATE_PATH = '.auth/state.json';
export const OUT_DIR = '../../static/img/openpanel-screenshots';
export const VIEWPORT_WIDTH = 1100;

import { rename, all, click, fill, fillVisible, selectFirst, maskIPs, hideNotices, hideOnboarding, hideText, markCard, tab, markSection } from './helpers.mjs';

// readable names for the demo's random mysql users and databases
const MYSQL_USERS = { u1pd6u5s: 'wp_blog_user', stefan: 'shop_admin' };
const MYSQL_DBS = { bd37zzjy: 'wp_blog', stefan: 'shop_db' };
const PG_NAMES = { postgredb: 'shop_db', postgreuser: 'shop_admin' };
const PASSWORD = 'hT9#vQ2m!Lx7pR4z';
const DEMO_IP = { '185.241.214.25': '203.0.113.10' };
// demo domains -> example domains, most specific first
const DOMAINS = {
  'wp.demo.openpanel.org': 'blog.example.com',
  'website-builder.tests.openpanel.org': 'portfolio.example.com',
  'redirect.tests.openpanel.org': 'old.example.com',
  'python.tests.openpanel.org': 'api.example.com',
  'nodejs.tests.openpanel.org': 'app.example.com',
  'files.tests.openpanel.org': 'files.example.com',
  'php.tests.openpanel.org': 'shop.example.com',
  'wp.tests.openpanel.org': 'blog.example.com',
  'to-be-removed.com': 'example.net',
  'tests.openpanel.org': 'example.com',
  'aliasexample@': 'sales@',
  'something@gmail.com': 'team@example.org',
  'test1@': 'john@',
  'stefan@': 'deploy@',
  'testing@': 'designer@',
  '28cp-memekid3-superJumbo.jpg': 'team-photo.jpg',
  nodeaplikacija: 'nodeapp',
  pythonaplikacija: 'pythonapp',
  bd37zzjy: 'wp_blog',
  u1pd6u5s: 'wp_blog_user',
};
const MAIL_HOST = { 'demo.openpanel.com': 'mail.example.com' };

// runs on every shot after its own prepare(), so selectors there can still use the real names
export const globalPrepare = all(rename(DOMAINS, 'body'), maskIPs('main'), hideNotices);

const DEMO_HOST = { 'demo.openpanel.com': 'panel.example.com', 'test@test.com': 'john@example.com' };

// demo users have random db/user names, swap them for readable ones before shooting
const mysqlFixtures = async page => {
  await page.evaluate(() => {
    const names = [['wp_blog', 'wp_blog_user', '48.20'], ['shop_db', 'shop_admin', '126.75']];
    document.querySelectorAll('#databases-table tbody tr').forEach((row, i) => {
      const [db, user, size] = names[i] || [];
      if (!db) return;
      const [dbCell, userCell] = row.querySelectorAll('td:not(.db_size_cell)');
      const dbText = [...dbCell.childNodes].find(n => n.nodeType === 3 && n.textContent.trim());
      if (dbText) dbText.textContent = db;
      else dbCell.firstElementChild.textContent = db;
      const link = userCell.querySelector('a') || userCell;
      link.textContent = user;
      const sizeCell = row.querySelector('.db_size_cell');
      if (sizeCell && sizeCell.style.display !== 'none') sizeCell.textContent = size;
    });
  });
};

export const pages = {
  'mysql/databases': {
    url: '/mysql',
    prepare: mysqlFixtures,
    shots: [
      {
        name: 'list',
        alt: 'MySQL Databases page listing databases with their assigned users and action buttons',
        crop: { from: 'main section > div:first-child', to: '#databases-table tbody tr:last-of-type' },
      },
      {
        name: 'sizes',
        alt: 'Databases table with Show database sizes enabled and the size unit dropdown set to MB',
        async prepare(page) {
          await page.check('#showSizesCheckbox', { force: true });
          await page.selectOption('#display-size', 'mb').catch(() => {});
          // changing the unit refetches sizes, let that finish before fixtures overwrite the cells
          await page.waitForTimeout(2500);
          await page.waitForFunction(() => [...document.querySelectorAll('.db_size_cell')].every(c => /\d|N\/A/.test(c.textContent)), null, { timeout: 15_000 }).catch(() => {});
        },
        crop: { from: 'main section > div:first-child', to: '#databases-table tbody tr:last-of-type' },
      },
      {
        name: 'export',
        alt: 'Export panel for a database with the SQL or GZIP format and Browser or Files destination options',
        async prepare(page) {
          await page.locator('#databases-table tbody tr').first().locator('button[title="Export"]').click();
          await page.locator('.export-section').first().waitFor({ state: 'visible' });
          await page.waitForTimeout(400);
        },
        crop: { from: '#databases-table thead', to: '.export-section >> visible=true', pad: 16 },
      },
      {
        name: 'delete',
        alt: 'Delete button turned into a Confirm button with a countdown after the first click',
        async prepare(page) {
          await page.locator('#databases-table tbody tr').first().locator('form[action="/mysql/delete"] button').click();
          await page.waitForTimeout(300);
        },
        crop: { from: '#databases-table thead', to: '#databases-table tbody tr:first-child', pad: 0 },
      },
    ],
  },
  'mysql/new_db': {
    url: '/mysql/new',
    shots: [
      {
        name: 'form',
        alt: 'Create Database form with the database name field and the Create Database button',
        prepare: fill({ 'input[name=database_name]': 'shop_db' }),
        crop: 'content',
      },
    ],
  },
  'mysql/users': {
    url: '/mysql/users',
    prepare: rename(MYSQL_USERS, '#users-table'),
    shots: [
      {
        name: 'list',
        alt: 'MySQL Users page listing database users with Change Password and Delete buttons',
        crop: 'content',
      },
      {
        name: 'delete',
        alt: 'Delete button turned into a Confirm button with a countdown after the first click',
        prepare: click('#users-table tbody tr:first-of-type button.btn-danger', 300),
        crop: { from: '#users-table thead', to: '#users-table tbody tr:first-of-type' },
      },
    ],
  },
  'mysql/new_user': {
    url: '/mysql/user',
    shots: [
      {
        name: 'form',
        alt: 'Create MySQL User form with username and password fields and a password strength bar',
        prepare: fill({ 'main input[name=db_user]': 'shop_admin', '#password': 'hT9#vQ2m!Lx7pR4z' }),
        crop: 'content',
      },
    ],
  },
  'mysql/assign': {
    url: '/mysql/assign',
    shots: [
      {
        name: 'form',
        alt: 'Assign User to Database form with user and database dropdowns and a grid of privilege checkboxes',
        prepare: all(
          async page => {
            await page.waitForFunction(() => document.querySelectorAll("select[name='db_user'] option").length > 1);
            await page.selectOption("select[name='db_user']", 'stefan');
            await page.selectOption("select[name='database_name']", 'stefan');
            await page.waitForTimeout(1200);
          },
          rename(MYSQL_USERS, "select[name='db_user']"),
          rename(MYSQL_DBS, "select[name='database_name']"),
        ),
        crop: 'content',
      },
    ],
  },
  'mysql/remove': {
    url: '/mysql/remove',
    shots: [
      {
        name: 'form',
        alt: 'Remove User from Database form with user and database dropdowns',
        prepare: all(
          async page => {
            await page.waitForFunction(() => document.querySelectorAll('main select option').length > 2).catch(() => {});
            const selects = page.locator('main select');
            for (let i = 0; i < (await selects.count()); i++) {
              const opts = await selects.nth(i).locator('option').evaluateAll(os => os.map(o => o.value).filter(Boolean));
              if (opts.includes('stefan')) await selects.nth(i).selectOption('stefan');
            }
          },
          rename(MYSQL_USERS, "main select[name='db_user']"),
          rename(MYSQL_DBS, "main select:not([name='db_user'])"),
        ),
        crop: 'content',
      },
    ],
  },
  'mysql/password': {
    url: '/mysql/password/stefan',
    shots: [
      {
        name: 'form',
        alt: 'Change User Password form with the username prefilled and a new password field',
        prepare: all(rename(MYSQL_USERS), fill({ '#new_password': 'hT9#vQ2m!Lx7pR4z' })),
        crop: 'content',
      },
    ],
  },
  'mysql/wizard': {
    url: '/mysql/wizard',
    shots: [
      {
        name: 'form',
        alt: 'Database Wizard with steps to create a database, create a user with a password, and a preview of the GRANT statement',
        prepare: async page => {
          const inputs = page.locator('main input[type=text]:visible');
          const values = ['shop_db', 'shop_admin', 'hT9#vQ2m!Lx7pR4z'];
          for (let i = 0; i < Math.min(3, await inputs.count()); i++) await inputs.nth(i).fill(values[i]);
          await page.waitForTimeout(300);
        },
        crop: 'content',
      },
    ],
  },
  'mysql/import': {
    url: '/mysql/import',
    shots: [
      {
        name: 'form',
        alt: 'Import into Database form with a database dropdown and a file picker for the .sql file',
        prepare: all(
          async page => {
            const sel = page.locator('main select').first();
            await page.waitForFunction(() => document.querySelectorAll('main select option').length > 1).catch(() => {});
            await sel.selectOption({ index: 1 }).catch(() => {});
          },
          rename(MYSQL_DBS, 'main select'),
        ),
        crop: 'content',
      },
    ],
  },
  'mysql/phpmyadmin': {
    url: '/mysql',
    prepare: mysqlFixtures,
    shots: [
      {
        name: 'button',
        alt: 'phpMyAdmin button with its tooltip in the actions of a database row',
        prepare: async page => {
          await page.locator('#databases-table tbody tr:first-of-type a[href^="/phpmyadmin"]').hover();
          await page.waitForTimeout(400);
        },
        crop: { from: '#databases-table thead', to: 'main div.absolute:has-text("phpMyAdmin") >> visible=true', pad: 12 },
      },
    ],
  },
  'mysql/processlist': {
    url: '/mysql/processlist',
    shots: [
      { name: 'list', alt: 'MySQL Processes table with the Refresh Processes button and the currently running queries', crop: 'content' },
    ],
  },
  'mysql/remote': {
    url: '/mysql/remote-mysql',
    prepare: rename({ '185.241.214.25': '203.0.113.10' }),
    shots: [
      {
        name: 'page',
        alt: 'Remote Access page showing remote access as Disabled with the Enable Remote Database Access button',
        crop: 'content',
      },
      {
        name: 'status',
        alt: 'Remote access Status row showing Disabled with the Click to Enable button',
        crop: { from: 'main .m-6 section:nth-of-type(1)', pad: 24 },
      },
      {
        name: 'connection',
        alt: 'Remote and Local connection details with the server address and port for each',
        crop: { from: 'main .m-6 section:nth-of-type(2)', to: 'main .m-6 section:nth-of-type(3)', pad: 24 },
      },
    ],
  },
  'mysql/root-password': {
    url: '/mysql/root-password',
    shots: [
      {
        name: 'form',
        alt: 'Change MySQL root password form with a new password field and a password strength bar',
        prepare: async page => {
          await page.locator('main input:not([type=hidden])').first().fill('hT9#vQ2m!Lx7pR4z');
          await page.waitForTimeout(300);
        },
        crop: 'content',
      },
    ],
  },
  'mysql/configuration': {
    url: '/mysql/configuration',
    shots: [
      {
        name: 'page',
        alt: 'MySQL Configuration page with a table of settings, editable values and the Save Changes button',
        crop: { from: 'main section > div:first-child', to: 'main table tbody tr:nth-of-type(9)' },
      },
    ],
  },
  'postgresql/databases': {
    url: '/postgresql',
    prepare: rename(PG_NAMES),
    shots: [
      { name: 'list', alt: 'PostgreSQL Databases page listing databases with their size, assigned users and actions', crop: 'content' },
      {
        name: 'delete',
        alt: 'Delete button turned into a Confirm button with a countdown after the first click',
        prepare: click('#databases-table tbody tr:first-of-type button.btn-danger', 300),
        crop: { from: '#databases-table thead', to: '#databases-table tbody tr:first-of-type' },
      },
    ],
  },
  'postgresql/new_db': {
    url: '/postgresql/new',
    shots: [{ name: 'form', alt: 'Create PostgreSQL Database form with the database name field', prepare: fillVisible(['shop_db']), crop: 'content' }],
  },
  'postgresql/users': {
    url: '/postgresql/users',
    prepare: rename(PG_NAMES),
    shots: [
      { name: 'list', alt: 'PostgreSQL Users page listing database users with Change Password and Delete buttons', crop: 'content' },
      {
        name: 'delete',
        alt: 'Delete button turned into a Confirm button with a countdown after the first click',
        prepare: click('#users-table tbody tr:first-of-type button.btn-danger', 300),
        crop: { from: '#users-table thead', to: '#users-table tbody tr:first-of-type' },
      },
    ],
  },
  'postgresql/new_user': {
    url: '/postgresql/user',
    shots: [{ name: 'form', alt: 'Create PostgreSQL User form with username and password fields and a password strength bar', prepare: fillVisible(['shop_admin', PASSWORD]), crop: 'content' }],
  },
  'postgresql/assign': {
    url: '/postgresql/assign',
    shots: [{ name: 'form', alt: 'Assign User to Database form with user and database dropdowns', prepare: all(selectFirst(1200), rename(PG_NAMES)), crop: 'content' }],
  },
  'postgresql/remove': {
    url: '/postgresql/remove',
    shots: [{ name: 'form', alt: 'Remove User from Database form with user and database dropdowns', prepare: all(selectFirst(), rename(PG_NAMES)), crop: 'content' }],
  },
  'postgresql/password': {
    url: '/postgresql/password/postgreuser',
    shots: [
      {
        name: 'form',
        alt: 'Change User Password form with the username prefilled and a new password field',
        prepare: all(rename(PG_NAMES), async page => {
          await page.locator('main input:visible').last().fill(PASSWORD);
          await page.waitForTimeout(300);
        }),
        crop: 'content',
      },
    ],
  },
  'postgresql/wizard': {
    url: '/postgresql/wizard',
    shots: [
      {
        name: 'form',
        alt: 'PostgreSQL Database Wizard with steps to create a database, create a user with a password, and a preview of the GRANT statement',
        prepare: fillVisible(['shop_db', 'shop_admin', PASSWORD]),
        crop: 'content',
      },
    ],
  },
  'postgresql/import': {
    url: '/postgresql/import',
    shots: [{ name: 'form', alt: 'Import into PostgreSQL Database form with a database dropdown and a file picker', prepare: all(selectFirst(), rename(PG_NAMES)), crop: 'content' }],
  },
  'postgresql/processlist': {
    url: '/postgresql/processlist',
    shots: [{ name: 'list', alt: 'PostgreSQL Processes table listing the active backend processes', crop: 'content' }],
  },
  'postgresql/remote': {
    url: '/postgresql/remote-postgresql',
    prepare: rename(DEMO_IP),
    shots: [
      { name: 'page', alt: 'Remote PostgreSQL Access page showing remote access as Disabled with the Enable Remote PostgreSQL Access button', crop: 'content' },
      { name: 'status', alt: 'Remote access Status row showing Disabled with the Click to Enable button', crop: { from: 'main .m-6 section:nth-of-type(1)', pad: 24 } },
      {
        name: 'connection',
        alt: 'Remote and Local PostgreSQL connection details with the server address and port for each',
        crop: { from: 'main .m-6 section:nth-of-type(2)', to: 'main .m-6 section:nth-of-type(3)', pad: 24 },
      },
    ],
  },
  'postgresql/configuration': {
    url: '/postgresql/configuration',
    shots: [
      {
        name: 'page',
        alt: 'PostgreSQL Configuration page with a table of settings, editable values and the Save Changes button',
        crop: { from: 'main section > div:first-child', to: 'main table tbody tr:nth-of-type(9)' },
      },
    ],
  },
  'account/account': {
    url: '/account',
    prepare: rename(DEMO_HOST),
    shots: [{ name: 'form', alt: 'Email & Password page with the email address, new password and confirm password fields', crop: 'content' }],
  },
  'account/2fa': {
    url: '/account/2fa',
    shots: [{ name: 'page', alt: 'Two-Factor Authentication page showing 2FA as disabled with the Click to enable 2FA button', crop: 'content' }],
  },
  'account/passkeys': {
    url: '/account/passkeys',
    shots: [{ name: 'page', alt: 'Passkeys page with the Add a passkey button and the list of registered passkeys', crop: 'content' }],
  },
  'account/sessions': {
    url: '/account/sessions',
    prepare: maskIPs(),
    shots: [{ name: 'list', alt: 'Active Sessions table with the IP address, creation time, online status and a Terminate action for each session', crop: 'content' }],
  },
  'account/login-history': {
    url: '/account/login-history',
    prepare: maskIPs(),
    shots: [
      {
        name: 'list',
        alt: 'Login History table with the country flag, IP address and login time of each successful login',
        crop: { from: 'main section > div:first-child', to: 'main table tbody tr:nth-of-type(6)' },
      },
    ],
  },
  'account/activity': {
    url: '/account/activity',
    prepare: maskIPs(),
    shots: [
      {
        name: 'list',
        alt: 'Activity Log table listing recent account actions with the user, action and IP address',
        crop: { from: 'main section > div:first-child', to: 'main table tbody tr:nth-of-type(8)' },
      },
    ],
  },
  'account/notifications': {
    url: '/account/notifications',
    shots: [{ name: 'form', alt: 'Email Notifications page with a checkbox for each event and the Save Preferences button', crop: 'content' }],
  },
  'account/favorites': {
    url: '/account/favorites',
    shots: [{ name: 'list', alt: 'Favorites table listing saved pages with their title, link and a Delete action', crop: 'content' }],
  },
  'account/language': {
    url: '/account/language',
    shots: [{ name: 'page', alt: 'Change Language page with the language dropdown showing the current locale', crop: 'content' }],
  },
  'account/api': {
    url: '/account/api',
    prepare: async page => {
      await page.locator('.swagger-ui .opblock').first().waitFor({ timeout: 15000 }).catch(() => {});
    },
    shots: [{ name: 'page', alt: 'API Reference page with the interactive OpenPanel API documentation and the Download spec button', crop: { content: 'main', maxHeight: 900 } }],
  },
  'account/mcp': {
    url: '/account/mcp',
    prepare: rename(DEMO_HOST),
    shots: [{ name: 'page', alt: 'MCP page with the token generator and connection snippets for Claude Code and Claude Desktop', crop: 'content' }],
  },
  'domains/domains': {
    url: '/domains',
    shots: [
      { name: 'list', alt: 'Domains page listing domains with their status, document root and PHP version', crop: 'content' },
      {
        name: 'actions',
        alt: 'Actions menu of a domain with Edit DNS Zone, Manage WAF, Change docroot, Edit VirtualHosts, Capitalize, Suspend and Delete',
        prepare: async page => {
          // hide the wide columns so the Actions column fits without horizontal scrolling
          await page.evaluate(() => {
            const el = document.querySelector('[x-data^="createTableController"]');
            const data = window.Alpine.$data(el);
            for (const k of ['docroot', 'php', 'site_count', 'redirect']) data.columns[k] = false;
          });
          await page.waitForTimeout(300);
          await page.locator('#dropdownHoverButton-0').hover();
          await page.waitForTimeout(500);
        },
        crop: { from: 'main table thead', to: '#dropdownHover-0', pad: 12 },
      },
    ],
  },
  'domains/new': {
    url: '/domains/new',
    shots: [
      {
        name: 'form',
        alt: 'New Domain form with the domain name and document root fields',
        prepare: async page => {
          await page.locator('main input:visible').first().fill('mycoffeeshop.com');
          await page.waitForTimeout(500);
        },
        crop: 'content',
      },
    ],
  },
  'domains/dns': {
    url: '/domains/edit-dns-zone/wp.tests.openpanel.org',
    shots: [
      {
        name: 'list',
        alt: 'DNS zone editor listing A, MX, CNAME and TXT records with Edit and Delete buttons',
        crop: { from: 'main', to: 'main table tbody tr:nth-of-type(9)', fromTop: true },
      },
    ],
  },
  'domains/ssl': {
    url: '/domains/ssl?domain_name=wp.tests.openpanel.org',
    shots: [
      { name: 'page', alt: 'SSL page for a domain with the AutoSSL status, the Generate now button, and the certificate details and files', crop: { from: 'main', to: 'main .m-6 section:nth-of-type(3)', fromTop: true, pad: 24 } },
      { name: 'custom', alt: 'Configure custom SSL form with fields for the certificate and the private key', crop: { from: 'main .m-6 section:nth-of-type(4)', pad: 24 } },
      { name: 'status', alt: 'SSL Status row showing Auto SSL with the Generate now button', crop: { from: 'main .m-6 section:nth-of-type(1)', pad: 24 } },
    ],
  },
  'domains/redirects': {
    url: '/domains/redirect?domain=redirect.tests.openpanel.org',
    shots: [{ name: 'form', alt: 'Redirect Domain to URL form with the destination URL field and the Save Redirect button', crop: 'content' }],
  },
  'domains/docroot': {
    url: '/domains/docroot?domain_name=wp.tests.openpanel.org',
    shots: [{ name: 'form', alt: 'Change docroot form with the document root folder field', crop: 'content' }],
  },
  'domains/vhosts': {
    url: '/domains/vhosts?domain=wp.tests.openpanel.org',
    shots: [{ name: 'editor', alt: 'VirtualHosts file editor showing the domain configuration with the Save Changes button', crop: { content: 'main', maxHeight: 760 } }],
  },
  'domains/logs': {
    url: '/domains/log/wp.tests.openpanel.org',
    shots: [
      {
        name: 'list',
        alt: 'Domain Logs table with the timestamp, client IP, method and requested URI of each request',
        crop: { from: 'main section > div:first-child', to: 'main table tbody tr:nth-of-type(10)' },
      },
    ],
  },
  'domains/dynamic-dns': {
    url: '/domains/dynamic-dns',
    prepare: async page => {
      // the update token is a live credential on the demo
      await page.evaluate(() => document.querySelectorAll('main *').forEach(el => {
        for (const n of el.childNodes) if (n.nodeType === 3) n.textContent = n.textContent.replace(/token=[^\s&"]+/g, 'token=a1b2c3d4e5f6…');
      }));
    },
    shots: [{ name: 'list', alt: 'Dynamic DNS page listing entries with their subdomain, record type, IP, TTL and update URL', crop: 'content' }],
  },
  'domains/capitalize': {
    url: '/domains/capitalize/to-be-removed.com',
    prepare: async page => {
      await page.evaluate(() => {
        const box = document.getElementById('letters-container');
        const tpl = box.firstElementChild;
        const word = 'MyCoffeeShop.com';
        box.innerHTML = '';
        for (const ch of word) {
          const b = tpl.cloneNode();
          b.textContent = ch;
          b.classList.remove('border-blue-500', 'border-transparent');
          b.classList.add(/[A-Z]/.test(ch) ? 'border-blue-500' : 'border-transparent');
          box.appendChild(b);
        }
        document.querySelectorAll('main .punycode').forEach(e => (e.textContent = 'mycoffeeshop.com'));
      });
    },
    shots: [{ name: 'page', alt: 'Capitalize page with one button per letter, the capitalized letters of MyCoffeeShop.com underlined', crop: 'content' }],
  },
  'domains/suspend': {
    url: '/domains/suspend?domain=wp.tests.openpanel.org',
    shots: [{ name: 'form', alt: 'Suspend domain confirmation with an optional reason field and the Confirm Suspend button', crop: 'content' }],
  },
  'domains/unsuspend': {
    url: '/domains/unsuspend?domain=website-builder.tests.openpanel.org',
    shots: [{ name: 'form', alt: 'Unsuspend domain confirmation with the Confirm Unsuspend button', crop: 'content' }],
  },
  'domains/goaccess': {
    url: '/domains/stats/wp.tests.openpanel.org',
    prepare: async page => {
      await page.evaluate(() => window.scrollTo(0, 0));
      await page.waitForTimeout(1500);
    },
    shots: [{ name: 'report', alt: 'GoAccess report for a domain with request totals, unique visitors and a chart of daily visitors', crop: { from: '#overall', pad: 16, clamp: false, height: 760 } }],
  },
  'domains/delete': {
    url: '/domains/delete?domain=to-be-removed.com',
    shots: [{ name: 'form', alt: 'Delete domain confirmation page with the Delete Domain button', crop: 'content' }],
  },
  'emails/emails': {
    url: '/emails',
    shots: [
      { name: 'list', alt: 'Email Accounts page listing mailboxes with their storage usage and the Webmail, Manage and Connect Devices buttons', crop: 'content' },
      {
        name: 'buttons',
        alt: 'Webmail, Manage and Connect Devices buttons in the row of an email account',
        crop: { from: 'main table thead', to: 'main table tbody tr:first-of-type' },
      },
    ],
  },
  'emails/new': {
    url: '/emails/new',
    shots: [
      {
        name: 'form',
        alt: 'Create an Email form with the domain, username, password and storage quota fields',
        prepare: async page => {
          await page.locator('main input[placeholder="info"], main input[name=email_username], main input[name=username]').first().fill('john').catch(() => {});
          await page.waitForTimeout(300);
        },
        crop: 'content',
      },
    ],
  },
  'emails/edit': {
    url: '/emails/edit/info@wp.tests.openpanel.org',
    shots: [{ name: 'form', alt: 'Edit Email account page with storage quota, incoming and outgoing mail toggles and a password field', crop: 'content' }],
  },
  'emails/connect': {
    url: '/emails/info/info@wp.tests.openpanel.org',
    prepare: rename(MAIL_HOST),
    shots: [{ name: 'page', alt: 'Connect Devices page with the username, incoming and outgoing servers and the IMAP and SMTP ports', crop: 'content' }],
  },
  'emails/aliases': {
    url: '/emails/aliases',
    shots: [{ name: 'list', alt: 'Aliases page listing alias addresses and the addresses they deliver to', crop: 'content' }],
  },
  'emails/aliases_new': {
    url: '/emails/aliases/new',
    shots: [
      {
        name: 'form',
        alt: 'Create an Alias form with the domain, alias address and destination address fields',
        prepare: all(selectFirst(400), fillVisible(['sales', 'team@example.org'])),
        crop: 'content',
      },
    ],
  },
  'emails/default': {
    url: '/emails/default/wp.tests.openpanel.org',
    shots: [
      {
        name: 'form',
        alt: 'Default Email Address page with the current catch-all configuration and the destination address field',
        prepare: fillVisible(['info@wp.tests.openpanel.org']),
        crop: 'content',
      },
    ],
  },
  'emails/filters': {
    url: '/emails/filter/info@wp.tests.openpanel.org/gui',
    shots: [
      { name: 'page', alt: 'Email Filters page with no filters yet and the Create first filter button', crop: 'content' },
      {
        name: 'rule',
        alt: 'An email filter that moves messages whose subject contains newsletter into the Newsletters folder',
        prepare: all(click('main button:has-text("Create first filter")', 600), fillVisible(['Newsletters', 'newsletter', 'Newsletters'])),
        crop: 'content',
      },
    ],
  },
  'emails/deliverability': {
    url: '/emails/deliverability/wp.tests.openpanel.org',
    prepare: async page => {
      await page.waitForFunction(() => !document.querySelector('main').innerText.includes('Checking DNS records'), null, { timeout: 30000 }).catch(() => {});
      await page.waitForTimeout(500);
    },
    shots: [{ name: 'page', alt: 'Email Deliverability details for a domain comparing the current and expected SPF, DKIM and DMARC records', crop: 'content' }],
  },
  'emails/deliverability_list': {
    url: '/emails/deliverability',
    prepare: async page => {
      await page.waitForFunction(() => !document.querySelector('main').innerText.includes('Checking'), null, { timeout: 30000 }).catch(() => {});
      await page.waitForTimeout(500);
    },
    shots: [{ name: 'list', alt: 'Email Deliverability overview listing domains with the status of their email DNS records', crop: 'content' }],
  },
  'emails/import': {
    url: '/emails/import',
    shots: [{ name: 'form', alt: 'Import Emails form with the file picker for a CSV, XLSX or XLS file and the Upload button', crop: 'content' }],
  },
  'emails/delete': {
    url: '/emails/delete/info@wp.tests.openpanel.org',
    shots: [{ name: 'form', alt: 'Delete email account confirmation for the selected address', crop: 'content' }],
  },
  'emails/aliases_manage': {
    url: '/emails/aliases/aliasexample@wp.tests.openpanel.org',
    shots: [{ name: 'page', alt: 'Manage alias page listing its destination addresses with Remove buttons and a field to add a destination', crop: 'content' }],
  },
  'files/files': {
    url: '/files',
    shots: [
      { name: 'list', alt: 'File Manager listing the folders and files in /var/www/html with their size, modification date and permissions', crop: 'content' },
      {
        name: 'new-file',
        alt: 'New File drawer with the file name field',
        prepare: all(click('main button:has-text("New File")', 600), async page => { await page.locator('#newfiDrawer input:visible').first().fill('robots.txt').catch(() => {}); }),
        viewportHeight: 620,
        crop: { from: '#newfiDrawer', clamp: false },
      },
      {
        name: 'new-folder',
        alt: 'New Folder drawer with the folder name field',
        prepare: all(click('main button:has-text("New Folder")', 600), async page => { await page.locator('#newfoDrawer input:visible').first().fill('images').catch(() => {}); }),
        viewportHeight: 620,
        crop: { from: '#newfoDrawer', clamp: false },
      },
      {
        name: 'context-menu',
        alt: 'Right-click menu on a file with Copy, Move, Rename, Download, View, Edit, Permissions, Compress and Delete',
        prepare: async page => {
          await page.locator('tr[data-file="index.php"]').click({ button: 'right' });
          await page.waitForTimeout(500);
        },
        crop: { from: 'main table thead', to: '#fmContextMenu', pad: 16 },
      },
      {
        name: 'rename',
        alt: 'Rename drawer with the new name field for the selected file',
        prepare: async page => {
          await page.locator('tr[data-file="index.php"]').click();
          await page.waitForTimeout(400);
          await page.evaluate(() => document.getElementById('renameButton').click());
          await page.waitForTimeout(600);
        },
        viewportHeight: 520,
        crop: { from: '#renameDrawer', clamp: false },
      },
      {
        name: 'permissions',
        alt: 'Change file permissions drawer with the octal permission value for the selected file',
        prepare: async page => {
          await page.locator('tr[data-file="index.php"]').click();
          await page.waitForTimeout(400);
          await page.evaluate(() => document.getElementById('permButton').click());
          await page.waitForTimeout(600);
        },
        viewportHeight: 520,
        crop: { from: '#permDrawer', clamp: false },
      },
      {
        name: 'compress',
        alt: 'Compress items drawer listing the selected files and the archive path and extension',
        prepare: async page => {
          await page.locator('tr[data-file="index.php"]').click();
          await page.waitForTimeout(400);
          await page.evaluate(() => document.getElementById('compressButton').click());
          await page.waitForTimeout(600);
        },
        viewportHeight: 520,
        crop: { from: '#compressDrawer', clamp: false },
      },
    ],
  },
  'files/edit': {
    url: '/file-manager/edit-file/index.php',
    shots: [{ name: 'editor', alt: 'File editor with syntax highlighting and the Save button', viewportHeight: 420 }],
  },
  'files/upload': {
    url: '/file-manager/upload',
    shots: [{ name: 'form', alt: 'File Upload page with the drag and drop area and the Upload button', crop: 'content' }],
  },
  'files/wget': {
    url: '/file-manager/upload?method=download',
    shots: [{ name: 'form', alt: 'Download from URL page with the URL field and the Download button', prepare: fillVisible(['https://wordpress.org/latest.zip']), crop: 'content' }],
  },
  'files/trash': {
    url: '/files.trash',
    shots: [{ name: 'list', alt: 'Trash page listing deleted files with Restore and Delete actions', crop: { from: 'main', to: 'main table tbody tr:last-of-type', fromTop: true, pad: 24 } }],
  },
  'files/ftp': {
    url: '/ftp',
    shots: [{ name: 'list', alt: 'FTP Accounts page with the FTP server and port, and a table of accounts with their path and configuration downloads', crop: 'content' }],
  },
  'files/ftp_new': {
    url: '/ftp/new',
    shots: [{ name: 'form', alt: 'New FTP Account form with the username, domain, password and path fields', prepare: fillVisible(['deploy']), crop: 'content' }],
  },
  'files/ftp_connections': {
    url: '/ftp/connections',
    shots: [{ name: 'list', alt: 'FTP Connections page listing current FTP sessions', crop: 'content' }],
  },
  'files/ftp_password': {
    url: '/ftp/password/stefan@wp.tests.openpanel.org',
    shots: [{ name: 'form', alt: 'Change FTP password form with a new password field', crop: 'content' }],
  },
  'files/ftp_path': {
    url: '/ftp/path/stefan@wp.tests.openpanel.org',
    shots: [{ name: 'form', alt: 'Change FTP path form with the folder field', crop: 'content' }],
  },
  'files/backups': {
    url: '/backups',
    shots: [{ name: 'page', alt: 'Backups page with the three setup steps: connect a destination, configure settings, and set limits', crop: 'content' }],
  },
  'files/backups_destination': {
    url: '/backups/destination',
    shots: [{ name: 'page', alt: 'Backup destination picker with S3-compatible, WebDAV, SSH, Azure and Dropbox options', crop: 'content' }],
  },
  'files/backups_list': {
    url: '/backups/list',
    shots: [{ name: 'list', alt: 'Restore & Download page with the backup search and the Reindex Destination button', crop: 'content' }],
  },
  'files/backup-wizard': {
    url: '/backup-wizard',
    shots: [{ name: 'page', alt: 'Backup Wizard with the list of included items, the Generate Backup button and the existing backups', crop: 'content' }],
  },
  'files/disk_usage': {
    url: '/disk-usage/',
    prepare: async page => { await page.waitForTimeout(1000); },
    shots: [{ name: 'page', alt: 'Disk Usage page with a table of folder sizes and a bar chart of disk usage per folder', crop: 'content' }],
  },
  'files/inodes': {
    url: '/inodes-explorer/',
    prepare: async page => { await page.waitForTimeout(1000); },
    shots: [{ name: 'page', alt: 'Inodes Explorer page with a table of inode counts per folder and a bar chart', crop: 'content' }],
  },
  'files/fix_permissions': {
    url: '/fix-permissions',
    shots: [{ name: 'form', alt: 'Fix Permissions page with the directory field and the Fix Permissions button', crop: 'content' }],
  },
  'files/malware': {
    url: '/malware-scanner',
    shots: [{ name: 'form', alt: 'ClamAV Scanner page with the directory to scan and the Start Scan button', crop: 'content' }],
  },
  'files/quarantine': {
    url: '/malware-scanner/quarantine',
    shots: [{ name: 'list', alt: 'Quarantine page listing files that ClamAV flagged as malicious', crop: 'content' }],
  },
  'php/domains': {
    url: '/php/domains',
    shots: [{ name: 'list', alt: 'PHP version for domains page listing each domain with its current PHP version and a dropdown to change it', crop: 'content' }],
  },
  'php/default': {
    url: '/php/default',
    shots: [{ name: 'form', alt: 'Default PHP version page with the version dropdown and the current default version', prepare: selectFirst(300), crop: 'content' }],
  },
  'php/extensions': {
    url: '/php/8.5/extensions',
    shots: [
      { name: 'list', alt: 'PHP 8.5 Extensions page listing extensions with their status and Enable or Disable buttons', crop: { from: 'main', to: 'main table tbody tr:nth-of-type(8)', fromTop: true } },
      {
        name: 'install',
        alt: 'Install PHP extensions dialog with checkboxes for the extensions that can be installed, imagick and memcached selected',
        prepare: async page => {
          await page.locator('main button:has-text("Install")').first().click();
          await page.waitForTimeout(2500);
          for (const ext of ['imagick', 'memcached']) await page.locator(`label:has-text("${ext}") input[type=checkbox]`).first().check({ force: true }).catch(() => {});
        },
        viewportHeight: 520,
      },
    ],
  },
  'php/options': {
    url: '/php/8.5/options',
    shots: [{ name: 'page', alt: 'PHP 8.5 Options page with editable values such as memory_limit, max_execution_time and upload limits', crop: { from: 'main', to: 'main table tbody tr:nth-of-type(11)', fromTop: true } }],
  },
  'php/php_ini_editor': {
    url: '/php/8.5/editor',
    shots: [{ name: 'editor', alt: 'PHP 8.5 INI Editor showing the php.ini file with the Save Changes button', crop: { content: 'main', maxHeight: 640 } }],
  },
  'caching/redis': {
    url: '/cache/redis',
    shots: [
      { name: 'page', alt: 'Redis page with the service status, TCP server and port, container resource usage and logs', crop: 'content' },
      {
        name: 'limits',
        alt: 'Containers page opened from Edit limits, with the CPU and memory limits of the Redis container',
        url: '/containers?s=redis',
        crop: 'content',
      },
      { name: 'status', alt: 'Redis Status row with the Click to Enable or Click to Disable button', crop: { from: 'main', to: 'main .m-6 > div:first-child', fromTop: true, pad: 24 } },
      { name: 'log-button', alt: 'Logs section of the Redis page with the View container log button', crop: { from: 'main .m-6 > section', pad: 16 } },
    ],
  },
  'caching/memcached': {
    url: '/cache/memcached',
    shots: [
      { name: 'page', alt: 'Memcached page with the service status, TCP server and port, container resource usage and logs', crop: 'content' },
      {
        name: 'limits',
        alt: 'Containers page opened from Edit limits, with the CPU and memory limits of the Memcached container',
        url: '/containers?s=memcached',
        crop: 'content',
      },
      { name: 'status', alt: 'Memcached Status row with the Click to Enable or Click to Disable button', crop: { from: 'main', to: 'main .m-6 > div:first-child', fromTop: true, pad: 24 } },
      { name: 'log-button', alt: 'Logs section of the Memcached page with the View container log button', crop: { from: 'main .m-6 > section', pad: 16 } },
    ],
  },
  'caching/valkey': {
    url: '/cache/valkey',
    shots: [
      { name: 'page', alt: 'Valkey page with the service status, TCP server and port', crop: 'content' },
      { name: 'status', alt: 'Valkey Status row with the Click to Enable or Click to Disable button', crop: { from: 'main', to: 'main .m-6 > div:first-child', fromTop: true, pad: 24 } },
    ],
  },
  'caching/elasticsearch': {
    url: '/cache/elasticsearch',
    shots: [
      { name: 'page', alt: 'ElasticSearch page with the service status, TCP server and port', crop: 'content' },
    ],
  },
  'caching/opensearch': {
    url: '/cache/opensearch',
    shots: [
      { name: 'page', alt: 'OpenSearch page with the service status, TCP server and port', crop: 'content' },
    ],
  },
  'caching/varnish': {
    url: '/cache/varnish',
    shots: [{ name: 'page', alt: 'Varnish page with the service status and a per-domain toggle to enable the Varnish cache', crop: 'content' }],
  },
  'containers/containers': {
    url: '/containers',
    shots: [
      { name: 'list', alt: 'Containers page listing services with their image, CPU and memory usage, PIDs and status', crop: { from: 'main', to: 'main table tbody tr:nth-of-type(9)', fromTop: true } },
    ],
  },
  'containers/edit': {
    url: '/containers/edit/redis',
    shots: [{ name: 'form', alt: 'Edit service form with the image tag, environment variables, CPU, memory and PID limits, volumes and networks', crop: { content: 'main', maxHeight: 1000 } }],
  },
  'containers/new': {
    url: '/containers/new',
    shots: [
      {
        name: 'form',
        alt: 'Add service form with the name, Docker image and tag, environment variables and resource limits',
        prepare: fillVisible(['uptime-kuma', 'louislam/uptime-kuma:1']),
        crop: { content: 'main', maxHeight: 1000 },
      },
    ],
  },
  'containers/change': {
    url: '/containers/image/change/redis',
    shots: [{ name: 'form', alt: 'Change image tag for redis form with the new image tag field', crop: 'content' }],
  },
  'containers/logs': {
    url: '/containers/logs?container=mariadb&lines=100',
    prepare: async page => {
      await page.waitForFunction(() => !/Select a container/.test(document.getElementById('log-content')?.textContent || ''), null, { timeout: 15000 }).catch(() => {});
    },
    shots: [{ name: 'page', alt: 'Container Logs page showing the last lines of the mariadb container log', crop: { content: 'main', maxHeight: 760 } }],
  },
  'containers/mysql': {
    url: '/containers/mysql',
    shots: [{ name: 'page', alt: 'Switch MySQL type page with the conditions for switching between MariaDB and MySQL', crop: 'content' }],
  },
  'containers/webserver': {
    url: '/containers/webserver',
    shots: [{ name: 'page', alt: 'Switch web server page with the conditions for switching to another web server', crop: 'content' }],
  },
  'advanced/cronjobs': {
    url: '/cronjobs',
    shots: [
      { name: 'list', alt: 'Cron Jobs page listing scheduled jobs with their schedule, container, command and comment', crop: { from: 'main', to: 'main table tbody tr:last-of-type', fromTop: true, pad: 24 } },
      {
        name: 'run',
        alt: 'Run now dialog of a cron job showing the command output and Finished successfully',
        prepare: async page => {
          // open the dialog with a finished run instead of starting the job over the websocket
          await page.evaluate(() => {
            const el = [...document.querySelectorAll('[x-data]')].find(e => /runOpen/.test(e.getAttribute('x-data')));
            const d = window.Alpine.$data(el);
            d.runOutput = [
              'Connecting to ip.openpanel.com (203.0.113.25:443)',
              "saving to 'index.html'",
              'index.html           100% |********************************|    13  0:00:00 ETA',
              "'index.html' saved",
              '',
            ].join('\n');
            d.runStatus = 'done';
            d.runExitCode = 0;
            d.runOpen = true;
          });
          await page.waitForTimeout(600);
        },
        viewportHeight: 720,
      },
      {
        name: 'edit',
        alt: 'A cron job row in edit mode with editable schedule, container, command and comment fields',
        prepare: click('main table tbody tr:first-of-type button:has-text("Edit")', 500),
        crop: { from: 'main table thead', to: 'main table tbody tr:first-of-type', pad: 8 },
      },
      {
        name: 'delete',
        alt: 'Delete button of a cron job turned into a Confirm button with a countdown after the first click',
        prepare: click('main table tbody tr:first-of-type button:has-text("Delete")', 300),
        crop: { from: 'main table thead', to: 'main table tbody tr:first-of-type', pad: 8 },
      },
    ],
  },
  'advanced/cronjobs_new': {
    url: '/cronjobs/new',
    shots: [
      {
        name: 'form',
        alt: 'Create Cron Job form with the container, schedule, common schedules, command and comment fields',
        prepare: async page => {
          await page.locator('main select').first().selectOption({ label: 'php-fpm-8.4' }).catch(() => {});
          await page.locator('main input[name=schedule], main input[placeholder*="@daily"]').first().fill('0 3 * * *').catch(() => {});
          await page.locator('main input[name=command], main input[placeholder*="php /var/www"]').first().fill('php /var/www/html/blog.example.com/wp-cron.php').catch(() => {});
          await page.locator('main input[name=comment], main input[placeholder=optional]').first().fill('wp-cron').catch(() => {});
        },
        crop: 'content',
      },
      {
        name: 'container',
        alt: 'Select Container dropdown of the Create Cron Job form',
        prepare: async page => { await page.locator('main select').first().selectOption({ label: 'php-fpm-8.4' }).catch(() => {}); },
        crop: { from: 'main label:has-text("Select Container")', to: 'main select >> nth=0', pad: 16 },
      },
      {
        name: 'common',
        alt: 'Common schedules dropdown set to Hourly, which fills in @hourly as the schedule',
        prepare: async page => {
          const sel = page.locator('main select').nth(1);
          const opts = await sel.locator('option').evaluateAll(os => os.map(o => o.value).filter(Boolean));
          if (opts.length) await sel.selectOption(opts[Math.min(5, opts.length - 1)]);
          await page.waitForTimeout(300);
        },
        crop: { from: 'main label:has-text("Schedule")', to: 'main select >> nth=1', pad: 16 },
      },
    ],
  },
  'advanced/cronjobs_editor': {
    url: '/cronjobs?view=code',
    shots: [{ name: 'editor', alt: 'Cron jobs File Editor showing the jobs in crons.ini format', crop: { content: 'main', maxHeight: 700 } }],
  },
  'advanced/services': {
    url: '/services/redis',
    shots: [{ name: 'page', alt: 'Service page for redis with its status, container resource usage and logs', crop: 'content' }],
  },
  'advanced/services_list': {
    url: '/services/',
    shots: [{ name: 'page', alt: 'Choose Service page with the service dropdown', crop: 'content' }],
  },
  'advanced/ip-blocker': {
    url: '/security/ip-blocker',
    shots: [{ name: 'form', alt: 'IP Blocker page with a text area for IP addresses and CIDR ranges to block', prepare: async page => { await page.locator('main textarea').first().fill('198.51.100.23\n203.0.113.0/24').catch(() => {}); }, crop: { content: 'main', maxHeight: 560 } }],
  },
  'advanced/process_manager': {
    url: '/process-manager',
    shots: [{ name: 'list', alt: 'Process Manager listing processes per container with their user, PID, CPU, time and command', crop: { from: 'main', to: 'main table tbody tr:nth-of-type(12)', fromTop: true } }],
  },
  'advanced/webserver_settings': {
    url: '/server/webserver_conf',
    shots: [{ name: 'editor', alt: 'Web server configuration editor with the Restore Default and Save Changes buttons', crop: { content: 'main', maxHeight: 700 } }],
  },
  'advanced/waf': {
    url: '/server/waf',
    shots: [{ name: 'list', alt: 'WAF page listing domains with a toggle to enable the firewall and Manage Rules and View Logs buttons', crop: 'content' }],
  },
  'advanced/waf_domain': {
    url: '/server/waf/wp.tests.openpanel.org',
    shots: [
      { name: 'page', alt: 'WAF settings for a domain with the status toggle and fields for disabled rule IDs and tags', prepare: fillVisible(['942100 920350', 'attack-sqli']), crop: 'content' },
      { name: 'ids', alt: 'Disabled IDs field of the WAF settings with rule IDs 942100 and 920350', prepare: fillVisible(['942100 920350', '']), crop: { from: 'main :text-is("Disabled IDs")', to: 'main input:visible >> nth=0', pad: 24 } },
      { name: 'tags', alt: 'Disabled Tags field of the WAF settings with the attack-sqli tag', prepare: fillVisible(['', 'attack-sqli']), crop: { from: 'main :text-is("Disabled Tags")', to: 'main input:visible >> nth=1', pad: 24 } },
    ],
  },
  'advanced/waf_logs': {
    url: '/server/waf/log',
    shots: [{ name: 'page', alt: 'WAF Logs page with the dropdown to choose the domain whose firewall log to view', crop: 'content' }],
  },
  'advanced/resource_usage': {
    url: '/server/usage',
    prepare: async page => { await page.waitForTimeout(1500); },
    shots: [{ name: 'page', alt: 'Resource Usage page with gauges for current CPU and RAM usage', crop: 'content' }],
  },
  'advanced/resource_history': {
    url: '/server/usage/history',
    prepare: async page => { await page.waitForTimeout(1500); },
    shots: [{ name: 'page', alt: 'Historical CPU and Memory Usage charts and a table of past usage', crop: { from: 'main', to: 'main table tbody tr:nth-of-type(6)', fromTop: true } }],
  },
  'advanced/server_info': {
    url: '/server/info',
    shots: [
      { name: 'page', alt: 'Server Information page with hostname, load, uptime, IP address, ports and operating system', crop: { from: 'main', to: 'main table tbody > tr:nth-child(11)', fromTop: true } },
      { name: 'plan', alt: 'Hosting Plan section of the Server Information page with the plan name and its limits for CPU, memory, disk, domains, websites, databases, email and FTP', crop: { from: 'main table tbody > tr:nth-child(12)', to: 'main table tbody > tr:nth-child(27)', right: true } },
      { name: 'panel', alt: 'Panel Information section with the panel version and the list of features enabled for the account', crop: { from: 'main table tbody > tr:nth-child(28)', to: 'main table tbody > tr:last-child', right: true } },
    ],
  },
  'dashboard/dashboard': {
    url: '/dashboard',
    prepare: hideOnboarding,
    shots: [
      { name: 'window', alt: 'OpenPanel dashboard with the sidebar, feature shortcuts grouped by section, and the 2FA, Information and Usage widgets', viewportHeight: 820 },
      { name: 'twofa', alt: 'Two-Factor Authentication widget showing 2FA as disabled with the Click to Enable button', prepare: all(hideOnboarding, markCard('Two-Factor Authentication')), crop: { from: '[data-shot=card]', pad: 12 } },
      { name: 'information', alt: 'Information widget with the username, plan, IP address and last login IP address', prepare: all(hideOnboarding, markCard('Information')), crop: { from: '[data-shot=card]', pad: 12 } },
      { name: 'usage', alt: 'Usage widget with bars for websites, domains, databases, email and FTP accounts, storage, inodes, CPU and memory', prepare: all(hideOnboarding, markCard('Usage')), crop: { from: '[data-shot=card]', pad: 12 } },
      { name: 'howto', alt: 'General How-to widget with links to knowledge base articles', prepare: all(hideOnboarding, markCard('General How-to')), crop: { from: '[data-shot=card]', pad: 12 } },
      {
        name: 'favorites',
        alt: 'Star icon in the top-right corner of the page header, used to add the current page to favorites',
        prepare: hideOnboarding,
        viewportHeight: 820,
        crop: { from: '#addFavoriteBtn', clamp: false, pad: 14 },
      },
      {
        name: 'favorites-sidebar',
        alt: 'Favorites list at the top of the sidebar with a saved page',
        prepare: hideOnboarding,
        viewportHeight: 820,
        crop: { from: '#sidebar', clamp: false, height: 150 },
      },
      {
        name: 'service',
        alt: 'Service badge in the page header with its tooltip showing the container name and resource usage',
        url: '/mysql',
        prepare: async page => {
          await page.locator('#service-status span[x-on\\:mouseenter]').first().hover();
          await page.waitForFunction(() => document.querySelector('#service-status [x-show="showTooltip"]')?.innerText.trim().length > 20, null, { timeout: 15000 }).catch(() => {});
          await page.waitForTimeout(500);
        },
        viewportHeight: 820,
        crop: { from: '#service-status', to: '#service-status [x-show="showTooltip"]', clamp: false, pad: 12 },
      },
      {
        name: 'search',
        alt: 'Search box opened from the magnifying glass icon in the header, with results for mysql',
        prepare: async page => {
          await hideOnboarding(page);
          await page.locator('#openSearchBtn').click();
          await page.locator('#searchInput').fill('mysql');
          await page.waitForTimeout(2000);
        },
        viewportHeight: 820,
        crop: { from: '#searchGroup', to: '#filteredDropdown', clamp: false, pad: 8 },
      },
    ],
  },
  'dashboard/dark-mode': {
    url: '/dashboard',
    init: () => localStorage.setItem('color-theme', 'dark'),
    prepare: hideOnboarding,
    shots: [
      { name: 'window', alt: 'OpenPanel dashboard in dark mode', viewportHeight: 820 },
    ],
  },
  'dashboard/onboarding': {
    url: '/dashboard',
    shots: [
      {
        name: 'wizard',
        alt: 'Onboarding wizard on its first step, choosing the webserver: Apache, Nginx, OpenLiteSpeed or OpenResty',
        viewportHeight: 820,
        crop: { from: 'div.max-w-xl:has(h2:has-text("set up your account"))', clamp: false },
      },
    ],
  },
  'intro/login': {
    url: '/login',
    anonymous: true,
    prepare: all(hideText('Sign in with testinguser', 1), rename({ testinguser: 'john' }, 'body')),
    shots: [{ name: 'form', alt: 'OpenPanel login form with username and password fields, the Forgot password link and passkey sign-in', viewportHeight: 820 }],
  },
  'intro/reset': {
    url: '/reset_password',
    anonymous: true,
    shots: [{ name: 'form', alt: 'Password reset form asking for the account email address', prepare: fillVisible(['john@example.com']), viewportHeight: 820 }],
  },
  'applications/sites': {
    url: '/sites',
    shots: [
      { name: 'list', alt: 'Site Manager listing websites grouped by type with their version, creation date and PageSpeed scores', crop: 'content' },
      {
        name: 'scan',
        alt: 'Scan button in Site Manager and the confirmation box that explains the scan, with Start Scan and Cancel buttons',
        prepare: click('#scanForSitesButton', 500),
        crop: { from: 'main', to: '#scanForSitesExplain', fromTop: true, pad: 16 },
      },
      {
        name: 'bulk',
        alt: 'Site Manager with one site selected and the bulk action bar with Update, Backup, Detach and Delete buttons',
        prepare: async page => {
          await page.locator('main .site-select-box').nth(1).check();
          await page.waitForTimeout(600);
        },
        viewportHeight: 900,
        crop: { from: 'main', to: 'main div[x-show="selectedCount > 0"]', fromTop: true, clamp: false, pad: 16 },
      },
    ],
  },
  'applications/autoinstaller': {
    url: '/auto-installer',
    shots: [{ name: 'page', alt: 'Auto Installer page with cards for WordPress, Joomla, Drupal, Website Builder, PrestaShop, OpenCart and other applications', crop: { content: 'main', maxHeight: 1000 } }],
  },
  'applications/wp_install': {
    url: '/wordpress/install',
    shots: [{ name: 'form', alt: 'Install WordPress form with site details, domain and location, and admin credentials', prepare: all(fillVisible(['My Coffee Shop', 'Fresh coffee, every day']), async page => { await page.locator('main input[type=email]').first().fill('owner@example.com').catch(() => {}); }), crop: 'content' }],
  },
  'applications/wordpress': {
    url: '/website?domain=wp.tests.openpanel.org',
    prepare: async page => { await page.waitForTimeout(2500); },
    shots: [
      { name: 'site', alt: 'WordPress site manager with the screenshot, versions, files and database details', viewportHeight: 900, crop: { content: 'main', maxHeight: 900 } },
      { name: 'header', alt: 'Site header with the Live Preview and Login as Admin buttons', crop: { from: 'main', to: 'main a:has-text("Live Preview")', fromTop: true, pad: 24 } },
      { name: 'versions', alt: 'WordPress, PHP and MariaDB version cards and the creation date', crop: { from: 'main :text-is("WordPress:")', to: 'main :text-is("Created:")', pad: 64, right: true } },
      { name: 'speed', alt: 'Speed card with desktop and mobile PageSpeed scores and First Contentful Paint, Speed Index and Time to Interactive', prepare: markCard('Speed'), crop: { from: '[data-shot=card]', pad: 12 } },
      { name: 'firewall', alt: 'Firewall card showing the firewall as active with denied and challenged request counts', prepare: markCard('Firewall'), crop: { from: '[data-shot=card]', pad: 12 } },
      { name: 'cache', alt: 'Cache card with the cache type and the Clear Cache button', prepare: markCard('Cache'), crop: { from: '[data-shot=card]', pad: 12 } },
      { name: 'options', alt: 'Options tab with the site URL, site name, email, registration, SEO visibility and pingback settings', prepare: tab('Options'), crop: { from: 'main', to: '[data-shot=tabs]', fromTop: true, pad: 16 } },
      {
        name: 'files-database',
        alt: 'Files and Database cards of the Overview tab with the folder path and size, disk usage, and the database name, user, host, size and phpMyAdmin link',
        prepare: async page => {
          await page.evaluate(() => {
            const card = title => {
              const h = [...document.querySelectorAll('main h2, main h3, main p, main span')].find(e => e.childElementCount === 0 && e.textContent.trim() === title && e.offsetParent);
              let el = h;
              while (el && !(/\brounded/.test(el.className) && /\bborder\b/.test(el.className) && el.getBoundingClientRect().height > 100)) el = el.parentElement;
              return el;
            };
            card('Files')?.setAttribute('data-shot', 'files');
            card('Database')?.setAttribute('data-shot', 'database');
          });
        },
        crop: { from: '[data-shot=files]', to: '[data-shot=database]', pad: 12 },
      },
      {
        name: 'vulnerabilities',
        alt: 'WP Vulnerabilities section of the Security tab with the number of detected vulnerabilities, the last check time and the Scan for vulnerabilities button',
        prepare: all(tab('Security'), async page => { await page.waitForTimeout(1500); }, markSection('WP Vulnerabilities')),
        crop: { from: '[data-shot=section]', pad: 16 },
      },
      {
        name: 'safe-browsing',
        alt: 'Google Safe Browsing section of the Security tab with the result of the check against the Google Safe Browsing list',
        prepare: all(tab('Security'), async page => { await page.waitForTimeout(2500); }, markSection('Google Safe Browsing')),
        crop: { from: '[data-shot=section]', pad: 16 },
      },
      { name: 'maintenance', alt: 'Maintenance tab with the maintenance mode toggle', prepare: tab('Maintenance'), crop: { from: 'main', to: '[data-shot=tabs]', fromTop: true, pad: 16 } },
      { name: 'security', alt: 'Security tab with vulnerability report, Safe Browsing, salts, integrity check, malware scan and reinstall', prepare: tab('Security'), crop: { from: 'main', to: '[data-shot=tabs]', fromTop: true, pad: 16 } },
      { name: 'updates', alt: 'Updates tab listing core, plugin and theme update status', prepare: tab('Updates'), crop: { from: 'main', to: '[data-shot=tabs]', fromTop: true, pad: 16 } },
      { name: 'debugging', alt: 'Debugging tab with toggles for WP_DEBUG, WP_DEBUG_LOG, WP_DEBUG_DISPLAY, SCRIPT_DEBUG and SAVEQUERIES', prepare: tab('Debugging'), crop: { from: 'main', to: '[data-shot=tabs]', fromTop: true, pad: 16 } },
      { name: 'backups', alt: 'Backups tab with the Create a Backup and Restore from Backups sections', prepare: tab('Backups'), crop: { from: 'main', to: '[data-shot=tabs]', fromTop: true, pad: 16 } },
      { name: 'clone', alt: 'Clone tab with the destination location and database for the copy', prepare: tab('Clone'), crop: { from: 'main', to: '[data-shot=tabs]', fromTop: true, pad: 16 } },
      { name: 'remove', alt: 'Remove tab with the Detach and Uninstall options', prepare: tab('Remove'), crop: { from: 'main', to: '[data-shot=tabs]', fromTop: true, pad: 16 } },
    ],
  },
  'applications/python': {
    url: '/website?domain=python.tests.openpanel.org',
    prepare: async page => { await page.waitForTimeout(2500); },
    shots: [
      { name: 'site', alt: 'Python application page with its status, runtime version, CPU and memory limits, and Stop and Restart actions', viewportHeight: 900, crop: { content: 'main', maxHeight: 900 } },
    ],
  },
  'applications/python_install': {
    url: '/python/install',
    shots: [{ name: 'form', alt: 'Install Python Application form with the application details, domain, startup file and advanced options', crop: 'content' }],
  },
  'applications/nodejs': {
    url: '/website?domain=nodejs.tests.openpanel.org',
    prepare: async page => { await page.waitForTimeout(2500); },
    shots: [
      { name: 'site', alt: 'Node.js application page with its status, runtime version, CPU and memory limits, and Stop and Restart actions', viewportHeight: 900, crop: { content: 'main', maxHeight: 900 } },
    ],
  },
  'applications/nodejs_install': {
    url: '/nodejs/install',
    shots: [{ name: 'form', alt: 'Install Node.js Application form with the application details, domain, startup file and advanced options', crop: 'content' }],
  },
  'applications/ruby_install': {
    url: '/ruby/install',
    shots: [{ name: 'form', alt: 'Install Ruby Application form with the application details, domain, startup command and advanced options', crop: 'content' }],
  },
  'applications/java_install': {
    url: '/java/install',
    shots: [{ name: 'form', alt: 'Install Java Application form with the application details, domain, startup command and advanced options', crop: 'content' }],
  },
  'applications/builder': {
    url: '/website-builder/install',
    shots: [{ name: 'form', alt: 'Website Builder page with the domain and folder to create the website in', crop: 'content' }],
  },
  'applications/php_install': {
    url: '/php/install',
    shots: [{ name: 'form', alt: 'Install PHP Application form with the domain and folder, and an optional Composer project to create', crop: 'content' }],
  },
  'applications/drupal': {
    url: '/drupal/install',
    shots: [{ name: 'form', alt: 'Install Drupal form with the site details, domain and location, and admin credentials', crop: 'content' }],
  },
  'applications/joomla': {
    url: '/joomla/install',
    shots: [{ name: 'form', alt: 'Install Joomla form with the site details, domain and location, and admin credentials', crop: 'content' }],
  },
  'applications/moodle': {
    url: '/moodle/install',
    shots: [{ name: 'form', alt: 'Install Moodle form with the site details, domain and location, and admin credentials', crop: 'content' }],
  },
  'applications/nextcloud': {
    url: '/nextcloud/install',
    shots: [{ name: 'form', alt: 'Install Nextcloud form with the site details, domain and location, and admin credentials', crop: 'content' }],
  },
  'applications/matomo': {
    url: '/matomo/install',
    shots: [{ name: 'form', alt: 'Install Matomo form with the site details, domain and location, and admin credentials', crop: 'content' }],
  },
  'applications/mediawiki': {
    url: '/mediawiki/install',
    shots: [{ name: 'form', alt: 'Install MediaWiki form with the site details, domain and location, and admin credentials', crop: 'content' }],
  },
  'applications/dokuwiki': {
    url: '/dokuwiki/install',
    shots: [{ name: 'form', alt: 'Install DokuWiki form with the site details, domain and location, and admin credentials', crop: 'content' }],
  },
  'applications/flarum': {
    url: '/flarum/install',
    shots: [{ name: 'form', alt: 'Install Flarum form with the site details, domain and location, and admin credentials', crop: 'content' }],
  },
  'applications/ojs': {
    url: '/ojs/install',
    shots: [{ name: 'form', alt: 'Install OJS form with the site details, domain and location, and admin credentials', crop: 'content' }],
  },
  'applications/opencart': {
    url: '/opencart/install',
    shots: [{ name: 'form', alt: 'Install OpenCart form with the site details, domain and location, and admin credentials', crop: 'content' }],
  },
  'applications/phpbb': {
    url: '/phpbb/install',
    shots: [{ name: 'form', alt: 'Install phpBB form with the site details, domain and location, and admin credentials', crop: 'content' }],
  },
  'applications/prestashop': {
    url: '/prestashop/install',
    shots: [{ name: 'form', alt: 'Install PrestaShop form with the site details, domain and location, and admin credentials', crop: 'content' }],
  },
  'applications/sofawiki': {
    url: '/sofawiki/install',
    shots: [{ name: 'form', alt: 'Install SofaWiki form with the site details, domain and location, and admin credentials', crop: 'content' }],
  },
  'applications/tinyfilemanager': {
    url: '/tinyfilemanager/install',
    shots: [{ name: 'form', alt: 'Install TinyFileManager form with the site details, domain and location, and admin credentials', crop: 'content' }],
  },
  'applications/tinyphotogallery': {
    url: '/tinyphotogallery/install',
    shots: [{ name: 'form', alt: 'Install TinyPhotoGallery form with the site details, domain and location, and admin credentials', crop: 'content' }],
  },
  'applications/wp_manager': {
    url: '/wordpress',
    shots: [
      { name: 'scan', alt: 'Scan for Existing Installations button next to New Installation on the WordPress Manager page', crop: { from: 'main a:has-text("New Installation")', to: '#scanButton', pad: 16 } },
      { name: 'sets', alt: 'Themes and Plugins buttons for managing the sets that are installed on every new WordPress site', crop: { from: 'main div:has(> a:text-is("Themes"))', pad: 3 } },
      { name: 'refresh', alt: 'Refresh Data button on the WordPress Manager page', crop: { from: '#refreshData', pad: 3 } },
      { name: 'view', alt: 'Switch to Table view button on the WordPress Manager page', crop: { from: 'main a:has-text("Switch to Table view")', pad: 3 } },
    ],
  },
  'domains/dns_sections': {
    url: '/domains/edit-dns-zone/wp.tests.openpanel.org',
    shots: [
      {
        name: 'edit',
        alt: 'A DNS record row in edit mode with editable fields and the Save and Cancel buttons',
        prepare: async page => {
          await page.locator('main #filemanager_table tbody tr:visible >> nth=2').locator('button:has-text("Edit")').click();
          await page.waitForTimeout(600);
        },
        crop: { from: 'main #filemanager_table thead', to: 'main #filemanager_table tbody tr:visible >> nth=2', pad: 8 },
      },
      {
        name: 'create',
        alt: 'New Record row added at the top of the DNS table with the record type, name, TTL and value fields',
        prepare: async page => {
          await page.locator('#AddDNSRecord').click();
          await page.waitForTimeout(600);
        },
        crop: { from: 'main', to: '#addRecordRow', fromTop: true, pad: 16 },
      },
      {
        name: 'delete',
        alt: 'Delete button of a DNS record turned into a Confirm button after the first click',
        prepare: async page => {
          await page.locator('main #filemanager_table tbody tr:visible >> nth=2').locator('button:has-text("Delete")').click();
          await page.waitForTimeout(400);
        },
        crop: { from: 'main #filemanager_table thead', to: 'main #filemanager_table tbody tr:visible >> nth=2', pad: 8 },
      },
      {
        name: 'toolbar',
        alt: 'DNS Zone Editor menu with the Edit DNS File, Export Zone and Reset options',
        prepare: async page => {
          await page.locator('#dropdownHoverButton').hover();
          await page.waitForTimeout(600);
        },
        crop: { from: '#dropdownHover', pad: 6 },
      },
      {
        name: 'reset',
        alt: 'Reset Zone panel asking for confirmation before the DNS zone is restored to its default records',
        prepare: async page => {
          await page.locator('#dropdownHoverButton').hover();
          await page.waitForTimeout(400);
          await page.locator('#dropdownHover a:has-text("Reset")').click();
          await page.waitForTimeout(800);
        },
        viewportHeight: 620,
        crop: { from: '#drawer-right-restart-zone', clamp: false },
      },
    ],
  },
  'domains/dns_editor': {
    url: '/domains/edit-dns-zone/wp.tests.openpanel.org?view=code',
    shots: [{ name: 'editor', alt: 'Advanced Editor showing the DNS zone file of the domain for direct editing', crop: { content: 'main', maxHeight: 760 } }],
  },
  'domains/dynamic-dns_create': {
    url: '/domains/dynamic-dns',
    shots: [
      {
        name: 'form',
        alt: 'Add Entry form for a new Dynamic DNS record with the domain, subdomain and initial IP fields',
        prepare: async page => {
          await page.locator('#add-entry').click();
          await page.waitForTimeout(600);
          await page.locator("[x-show=\"panel === 'create'\"] input[name=subdomain]").fill('home').catch(() => {});
        },
        crop: { from: 'main', to: "main div[x-show=\"panel === 'create'\"]", fromTop: true, pad: 16 },
      },
    ],
  },
  'files/files_drop': {
    url: '/files',
    shots: [
      {
        name: 'drop',
        alt: 'File Manager while dragging files over the page, with the drag and drop upload area shown above the file list',
        prepare: async page => {
          // no real file is dragged, the page only needs the dragenter events to show the drop area
          await page.evaluate(() => {
            const dt = new DataTransfer();
            document.dispatchEvent(new DragEvent('dragenter', { bubbles: true, cancelable: true, dataTransfer: dt }));
            const area = document.getElementById('dropArea') || document.querySelector('label[for="fileUpload"]');
            area?.dispatchEvent(new DragEvent('dragenter', { bubbles: true, cancelable: true, dataTransfer: dt }));
          });
          await page.waitForTimeout(500);
        },
        crop: { from: 'main', to: '#upload_files_form', fromTop: true, pad: 24 },
      },
    ],
  },
  'files/ftp_actions': {
    url: '/ftp',
    shots: [
      {
        name: 'delete',
        alt: 'Delete button of an FTP account turned into a Confirm button with a countdown after the first click',
        prepare: async page => {
          // the actions column is past the right edge, so scroll the table to it first
          await page.evaluate(() => document.querySelectorAll('main .overflow-x-auto, main .overflow-auto').forEach(e => (e.scrollLeft = e.scrollWidth)));
          await page.locator('main table tbody tr:first-of-type button:has-text("Delete")').click();
          await page.waitForTimeout(300);
        },
        crop: { from: 'main table thead', to: 'main table tbody tr:first-of-type', pad: 8 },
      },
      {
        name: 'config',
        alt: 'FileZilla and Cyberduck buttons in the Configuration column of an FTP account',
        crop: { from: 'main table tbody tr:first-of-type a:has-text("Filezilla")', to: 'main table tbody tr:first-of-type a:has-text("Cyberduck")', pad: 12 },
      },
    ],
  },
  'files/files_more': {
    url: '/files',
    shots: [
      {
        name: 'delete',
        alt: 'Delete drawer listing the selected files with the Delete button',
        prepare: async page => {
          await page.locator('tr[data-file="index.php"]').click();
          await page.locator('tr[data-file="index.html"]').click({ modifiers: ['Control'] });
          await page.waitForTimeout(300);
          await page.evaluate(() => document.getElementById('deleteButton').click());
          await page.waitForTimeout(800);
        },
        viewportHeight: 620,
        crop: { from: '#deleteDrawer', clamp: false },
      },
      {
        name: 'copy',
        alt: 'Copy drawer listing the selected files with the destination folder picker and the Copy button',
        prepare: async page => {
          await page.locator('tr[data-file="index.php"]').click();
          await page.locator('tr[data-file="index.html"]').click({ modifiers: ['Control'] });
          await page.waitForTimeout(300);
          await page.evaluate(() => document.getElementById('copyButton').click());
          await page.waitForTimeout(800);
        },
        viewportHeight: 620,
        crop: { from: '#copyDrawer', clamp: false },
      },
      {
        name: 'move',
        alt: 'Move drawer listing the selected files with the destination folder picker and the Move button',
        prepare: async page => {
          await page.locator('tr[data-file="index.php"]').click();
          await page.locator('tr[data-file="index.html"]').click({ modifiers: ['Control'] });
          await page.waitForTimeout(300);
          await page.evaluate(() => document.getElementById('moveButton').click());
          await page.waitForTimeout(800);
        },
        viewportHeight: 620,
        crop: { from: '#moveDrawer', clamp: false },
      },
      {
        name: 'search',
        alt: 'File Manager search box opened from its magnifying glass icon, with results for wp-config',
        prepare: async page => {
          await page.locator('#searchIcon').click();
          // Cloudflare in front of the demo blocks the literal wp-config.php in a query string, so search for wp-config
          await page.locator('#searchFilesInput').pressSequentially('wp-config', { delay: 60 });
          await page.waitForFunction(() => document.querySelectorAll('#filteredFilesDropdown li').length > 0, null, { timeout: 15000 }).catch(() => {});
          await page.waitForTimeout(800);
        },
        crop: { from: 'main', to: '#filteredFilesDropdown', fromTop: true, pad: 4 },
      },
    ],
  },
  'files/files_empty': {
    url: '/files/test',
    shots: [{ name: 'empty', alt: 'File Manager showing an empty folder with the No items found message', crop: 'content' }],
  },
  'files/files_modern': {
    url: '/files?view=modern',
    shots: [
      {
        name: 'modern',
        alt: 'File Manager in Modern button style with three items selected, the floating action bar at the bottom, and the account menu open showing the Buttons style switch',
        prepare: async page => {
          // open the account menu first, its click would otherwise clear the selection
          await page.locator('#user-btn-info').click();
          await page.waitForTimeout(500);
          // ctrl-clicks that don't bubble, so the document-level deselect handler never sees them
          await page.evaluate(() => {
            const rows = [...document.querySelectorAll('tr.clickable-row')].slice(-3);
            rows.forEach(r => r.dispatchEvent(new MouseEvent('click', { ctrlKey: true, bubbles: false })));
          });
          await page.waitForTimeout(600);
        },
        viewportHeight: 820,
        // the floating action bar is wider than the default window
        viewportWidth: 1440,
      },
    ],
  },
};
