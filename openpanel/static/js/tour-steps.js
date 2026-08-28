/* OpenPanel UI product tour - step definitions, grouped by page in walkthrough order.
   Add new pages by appending more entries with the matching `path`. Steps whose
   `element` selector isn't found on the page (feature disabled for this user,
   Enterprise-only, widget hidden, etc.) are skipped automatically by tour.js. */
window.TOUR_HOOKS = {
    openProfileMenu: function () {
        var btn = document.getElementById('user-btn-info');
        var menu = document.getElementById('popup-menu');
        if (btn && menu && menu.classList.contains('hidden')) btn.click();
    },
    closeProfileMenu: function () {
        var btn = document.getElementById('user-btn-info');
        var menu = document.getElementById('popup-menu');
        if (btn && menu && !menu.classList.contains('hidden')) btn.click();
    },
    openSearchBox: function () {
        var btn = document.getElementById('openSearchBtn');
        if (btn && btn.offsetParent !== null) btn.click();
    }
};

window.TOUR_STEPS = [
    {
        path: '/dashboard',
        element: '#dashboard-sortable-area',
        title: 'Your Shortcuts',
        description: 'Quick-access icons for every feature you have access to, grouped into sections like Files, Domains, MySQL and more.',
        side: 'top',
        align: 'start'
    },
    {
        path: '/dashboard',
        element: '#section-title-files button',
        title: 'Show/Hide a Section',
        description: 'Click this button to collapse or expand a section - handy once you know its icons by heart.',
        side: 'bottom',
        align: 'end'
    },
    {
        path: '/dashboard',
        element: '#section-title-files',
        title: 'Reorder Sections',
        description: 'Click and drag a section by its title to reorder it. Your custom order is remembered for next time.',
        side: 'top',
        align: 'start'
    },
    {
        path: '/dashboard',
        element: '#tour-icons-position',
        title: 'Icons Position',
        description: 'Click your profile at the bottom of the sidebar, then use this toggle to switch between showing the icon label below the icon (Top) or beside it (Start), whichever is easier to scan.',
        side: 'top',
        align: 'start',
        beforeShow: 'openProfileMenu',
        beforeHide: 'closeProfileMenu'
    },
    {
        path: '/dashboard',
        element: '#favorites-list',
        title: 'Favorites',
        description: 'Pages you star are pinned here at the top of the sidebar for one-click access.',
        side: 'right',
        align: 'start'
    },
    {
        path: '/dashboard',
        element: '#tour-main-menu',
        title: 'Main Menu',
        description: 'Files, Domains, MySQL, Emails and the rest of your enabled features live here, grouped by category.',
        side: 'right',
        align: 'start'
    },
    {
        path: '/dashboard',
        element: '#theme-toggle',
        title: 'Dark Mode',
        description: 'Click your profile at the bottom of the sidebar, then use this toggle to switch between light and dark theme.',
        side: 'top',
        align: 'start',
        beforeShow: 'openProfileMenu',
        beforeHide: 'closeProfileMenu'
    },
    {
        path: '/dashboard',
        element: '#searchInput',
        title: 'Search',
        description: 'Quickly find any feature or one of your websites by typing its name here.',
        side: 'bottom',
        align: 'end',
        beforeShow: 'openSearchBox'
    },
    {
        path: '/dashboard',
        element: '#addFavoriteBtn',
        title: 'Add to Favorites',
        description: 'Left-click to add the current page to your Favorites, right-click an already-favorited page to remove it.',
        side: 'bottom',
        align: 'end'
    },
    {
        path: '/dashboard',
        element: '#dashboard_twofa',
        title: 'Two-Factor Authentication',
        description: 'Add an extra layer of security to your account with 2FA. Click the button here to enable it.',
        side: 'left',
        align: 'start'
    },
    {
        path: '/dashboard',
        element: '#dashboard_info',
        title: 'Information',
        description: 'Your username, hosting plan, IP address and last login details at a glance.',
        side: 'left',
        align: 'start'
    },
    {
        path: '/dashboard',
        element: '#dashboard_usage',
        title: 'Usage',
        description: 'Tracks how much of your plan’s resources - websites, domains, databases, storage, CPU, RAM and more - you’re currently using.',
        side: 'left',
        align: 'start'
    },
    {
        path: '/dashboard',
        element: '#dashboard_howto',
        title: 'General How-to',
        description: 'Handy links to knowledge base articles for common tasks - opens in a new tab.',
        side: 'left',
        align: 'start'
    },

    {
        path: '/files',
        element: '#tour-fm-breadcrumbs',
        title: 'File Manager',
        description: 'Browse your website files. Click any part of the path to jump back to that folder.',
        side: 'bottom',
        align: 'start'
    },
    {
        path: '/files',
        element: '#newFileButton',
        title: 'New File / Folder / Upload',
        description: 'Create a new file or folder, or upload files from your device, directly into the current folder.',
        side: 'bottom',
        align: 'start'
    },
    {
        path: '/files',
        element: '#dropdownToggleButton',
        title: 'Filter',
        description: 'Filter the listing to show only files or only folders.',
        side: 'bottom',
        align: 'start'
    },
    {
        path: '/files',
        element: '#searchIcon',
        title: 'Search',
        description: 'Search for a file or folder by name within the current path, or a specific subfolder.',
        side: 'bottom',
        align: 'end'
    },
    {
        path: '/files',
        element: '#filemanager_table',
        title: 'Files & Folders',
        description: 'Select one or more rows to reveal actions like copy, move, rename, compress, set permissions or delete. Click a folder to open it, or a file to preview/edit it.',
        side: 'top',
        align: 'start'
    },

    {
        path: '/domains',
        element: '#domains-table',
        title: 'Domains',
        description: 'View all your domains - their status, PHP version, SSL and more, depending on which columns you show.',
        side: 'top',
        align: 'start'
    },
    {
        path: '/domains',
        element: '#dropdownToggleButton',
        title: 'Show Columns',
        description: 'Choose which columns to display, like HSTS status or PHP version.',
        side: 'bottom',
        align: 'start'
    },
    {
        path: '/domains',
        element: '#tour-domains-actions',
        title: 'Actions',
        description: 'Each domain row has actions to edit DNS, manage SSL, suspend/unsuspend, edit the vhost, or delete it.',
        side: 'left',
        align: 'start'
    },
    {
        path: '/domains',
        element: 'a[href="/domains/new"]',
        title: 'Add Domain',
        description: 'Click here to add a new domain to your account.',
        side: 'bottom',
        align: 'end'
    },

    {
        path: '/mysql',
        element: '#databases-table',
        title: 'Databases',
        description: 'Lists your MySQL/MariaDB databases. Click a database name to manage its tables.',
        side: 'top',
        align: 'start'
    },
    {
        path: '/mysql',
        element: '#showAllCheckbox',
        title: 'System Databases',
        description: 'Show system databases (like information_schema) in the list - usually only needed for troubleshooting.',
        side: 'bottom',
        align: 'start'
    },
    {
        path: '/mysql',
        element: '#showSizesCheckbox',
        title: 'Database Sizes',
        description: 'Show the size of each database. This can be slower to load on servers with many databases.',
        side: 'bottom',
        align: 'start'
    },

    {
        path: '/postgresql',
        element: '#databases-table',
        title: 'Databases',
        description: 'Lists your PostgreSQL databases. Click a database name to manage it.',
        side: 'top',
        align: 'start'
    },
    {
        path: '/postgresql',
        element: '#showAllCheckbox',
        title: 'System Databases',
        description: 'Show system databases (like postgres or template0/template1) in the list.',
        side: 'bottom',
        align: 'start'
    },

    {
        path: '/sites',
        element: '#sites-table',
        title: 'Site Manager',
        description: 'All your websites - WordPress, Node.js, Python and others - grouped by type, with their admin email, version, speed score and actions.',
        side: 'top',
        align: 'start'
    },
    {
        path: '/sites',
        element: '#sites-search-input',
        title: 'Search',
        description: 'Filter the list by site type or name.',
        side: 'bottom',
        align: 'start'
    },
    {
        path: '/sites',
        element: 'a[href="/auto-installer"]',
        title: 'Add a Website',
        description: 'Click New to launch the Auto Installer and set up WordPress or another supported app.',
        side: 'bottom',
        align: 'end'
    }
];

(function () {
    // Redis, Valkey, Memcached, Varnish, OpenSearch and ElasticSearch all share the
    // same cache/service.html template, so the same step set applies to every path.
    var cachePaths = ['/cache/redis', '/cache/valkey', '/cache/memcached', '/cache/varnish', '/cache/opensearch', '/cache/elasticsearch'];
    cachePaths.forEach(function (path) {
        window.TOUR_STEPS.push(
            {
                path: path,
                element: 'form button[type="submit"]',
                title: 'Enable / Disable',
                description: 'Click here to turn this caching service on or off for your account.',
                side: 'bottom',
                align: 'end'
            },
            {
                path: path,
                element: '#service-page-status',
                title: 'Status',
                description: 'Shows whether the service is currently running.',
                side: 'bottom',
                align: 'start'
            },
            {
                path: path,
                element: '#port',
                title: 'Connection Details',
                description: 'Use this server and port (not localhost or 127.0.0.1) to connect from your application.',
                side: 'right',
                align: 'start'
            },
            {
                path: path,
                element: '#domains',
                title: 'Per-Domain Toggle',
                description: 'Turn this cache on or off individually for each of your domains.',
                side: 'top',
                align: 'start'
            }
        );
    });
})();

window.TOUR_STEPS.push(
    {
        path: '/emails',
        element: '#email-accounts',
        title: 'Email Accounts',
        description: 'Lists your email accounts and their quota usage. Click an address to manage it.',
        side: 'top',
        align: 'start'
    },
    {
        path: '/emails',
        element: '#emails-search-input',
        title: 'Search',
        description: 'Filter accounts by address.',
        side: 'bottom',
        align: 'start'
    },
    {
        path: '/emails',
        element: '#showAddEmailFormBtn',
        title: 'New Email',
        description: 'Click here to create a new email account.',
        side: 'bottom',
        align: 'end'
    },
    {
        path: '/emails',
        element: '#export_emails_csv',
        title: 'Export',
        description: 'Export your email accounts to a CSV file.',
        side: 'bottom',
        align: 'end'
    },

    {
        path: '/php/domains',
        element: '#php-domains-table',
        title: 'Select PHP Version',
        description: 'Shows the PHP version each of your domains is currently using, and flags outdated or insecure versions.',
        side: 'top',
        align: 'start'
    },
    {
        path: '/php/domains',
        element: '#new_php_version',
        title: 'Change Version',
        description: 'Pick a new PHP version for a domain here. Changing it restarts PHP for that site, which takes a few seconds.',
        side: 'left',
        align: 'start'
    },

    {
        path: '/php/default',
        element: '#current_default_version',
        title: 'Default PHP Version',
        description: 'The PHP version used for any new domain you create, unless you pick a different one for it.',
        side: 'bottom',
        align: 'start'
    },
    {
        path: '/php/default',
        element: '#new_php_version',
        title: 'Change Default',
        description: 'Select a new default PHP version, then click Change to apply it.',
        side: 'bottom',
        align: 'start'
    },

    {
        path: '/php/options',
        element: '#php_version',
        title: 'Select Version',
        description: 'Pick which PHP version to edit options for, then click Edit Options.',
        side: 'bottom',
        align: 'start'
    },
    {
        path: '/php/options',
        element: '#save-changes-button',
        title: 'Save Changes',
        description: 'Toggle options like upload limits or execution time, then click here to save.',
        side: 'bottom',
        align: 'end'
    },

    {
        path: '/php/extensions',
        element: '#php_version',
        title: 'Select Version',
        description: 'Pick which PHP version to manage extensions for.',
        side: 'bottom',
        align: 'start'
    },
    {
        path: '/php/extensions',
        element: '#open-install-modal',
        title: 'Install Extension',
        description: 'Click here to browse and install additional PHP extensions for the selected version.',
        side: 'bottom',
        align: 'end'
    },

    {
        path: '/php/php_ini_editor',
        element: '#php_version',
        title: 'Select Version',
        description: 'Pick which PHP version’s php.ini file to open.',
        side: 'bottom',
        align: 'start'
    },
    {
        path: '/php/php_ini_editor',
        element: '#editor',
        title: 'php.ini Content',
        description: 'Edit the raw php.ini content directly here.',
        side: 'top',
        align: 'start'
    },
    {
        path: '/php/php_ini_editor',
        element: '#save-changes-button',
        title: 'Save',
        description: 'Click Save to apply your changes to this php.ini file.',
        side: 'bottom',
        align: 'end'
    },

    {
        path: '/containers',
        element: '#containers-table',
        title: 'Containers',
        description: 'Lists the Docker containers behind your account - webserver, database, cache and other services - with live CPU/RAM and status.',
        side: 'top',
        align: 'start'
    },
    {
        path: '/containers',
        element: '#dropdownToggleButton',
        title: 'Show Columns',
        description: 'Choose which columns to display, like Block I/O, Network I/O or PIDs.',
        side: 'bottom',
        align: 'start'
    },
    {
        path: '/containers',
        element: 'a[href="/containers/new"]',
        title: 'New Service',
        description: 'Click here to add a new container/service to your account.',
        side: 'bottom',
        align: 'end'
    },

    {
        path: '/services',
        element: '#service-select',
        title: 'Services',
        description: 'Pick a service to view its real-time resource usage, status, and to enable/disable it.',
        side: 'bottom',
        align: 'start'
    },

    {
        path: '/cronjobs',
        element: '#cronjobs-table',
        title: 'Cron Jobs',
        description: 'Lists your scheduled cron jobs and their schedule, command and last run status.',
        side: 'top',
        align: 'start'
    },
    {
        path: '/cronjobs',
        element: 'a[href="/cronjobs/new"]',
        title: 'New Cron Job',
        description: 'Click here to schedule a new cron job.',
        side: 'bottom',
        align: 'end'
    },
    {
        path: '/cronjobs',
        element: '#editForm',
        title: 'Edit Raw Crontab',
        description: 'Prefer editing the crontab directly? Use this editor to make changes to the raw file.',
        side: 'top',
        align: 'start'
    },

    {
        path: '/security/ip-blocker',
        element: '#blocked_ips',
        title: 'IP Blocker',
        description: 'List IP addresses (one per line) that should be blocked from accessing your websites.',
        side: 'top',
        align: 'start'
    },
    {
        path: '/security/ip-blocker',
        element: '#save-ips',
        title: 'Save',
        description: 'Click Save to apply the blocklist.',
        side: 'bottom',
        align: 'end'
    },

    {
        path: '/server/usage',
        element: '#cpuGauge',
        title: 'CPU Usage',
        description: 'Real-time CPU usage for your account.',
        side: 'right',
        align: 'start'
    },
    {
        path: '/server/usage',
        element: '#ramGauge',
        title: 'RAM Usage',
        description: 'Real-time memory usage for your account.',
        side: 'left',
        align: 'start'
    },

    {
        path: '/process-manager',
        element: '#process-manager-table',
        title: 'Process Manager',
        description: 'Lists running processes with their CPU/memory usage. Sort by a column, or click a process to view its full command or take action.',
        side: 'top',
        align: 'start'
    },

    {
        path: '/server/waf',
        element: '#waf-domains-table',
        title: 'WAF',
        description: 'Turn the Coraza Web Application Firewall on or off for each of your domains individually.',
        side: 'top',
        align: 'start'
    },

    {
        path: '/server/info',
        element: '#server',
        title: 'Server Information',
        description: 'General info about the server your account runs on - hostname, load, uptime, IP and OS.',
        side: 'top',
        align: 'start'
    },

    {
        path: '/server/webserver_conf',
        element: '#editor-container',
        title: 'Webserver Configuration',
        description: 'Edit the main webserver configuration file directly. Click Save Changes when done.',
        side: 'top',
        align: 'start'
    },

    {
        path: '/account',
        element: '#email',
        title: 'Email & Password',
        description: 'Update the email address and password used to log into your account.',
        side: 'right',
        align: 'start'
    },
    {
        path: '/account',
        element: '#save-button',
        title: 'Update',
        description: 'Click Update to save your changes.',
        side: 'bottom',
        align: 'end'
    },

    {
        path: '/account/language',
        element: '#locale-select',
        title: 'Change Language',
        description: 'Pick the language used across the panel interface.',
        side: 'bottom',
        align: 'start'
    },

    {
        path: '/account/notifications',
        element: 'form button[type="submit"]',
        title: 'Email Notifications',
        description: 'Choose which email notifications you want to receive, then click Save Preferences.',
        side: 'bottom',
        align: 'end'
    },

    {
        path: '/account/2fa',
        element: '#qrcode',
        title: 'Two-Factor Authentication',
        description: 'Scan this QR code with an authenticator app, then enter the generated code to enable 2FA on your account.',
        side: 'right',
        align: 'start'
    },

    {
        path: '/account/sessions',
        element: '#sessions-table',
        title: 'Active Sessions',
        description: 'Shows every device/browser currently logged into your account. End a session here if you don’t recognize it.',
        side: 'top',
        align: 'start'
    },

    {
        path: '/account/favorites',
        element: '#favorites-table',
        title: 'Favorite Pages',
        description: 'Pages you’ve starred for quick access from the sidebar. Remove one from here, or by right-clicking its star.',
        side: 'top',
        align: 'start'
    },

    {
        path: '/account/activity',
        element: '#activitySearchForm',
        title: 'Account Activity',
        description: 'Search and filter the full activity log for your account.',
        side: 'bottom',
        align: 'start'
    },

    {
        path: '/account/login-history',
        element: '#login-log-table',
        title: 'Login History',
        description: 'Shows past logins to your account, including the IP address and time.',
        side: 'top',
        align: 'start'
    },

    {
        path: '/ftp',
        element: '#ftp-table',
        title: 'FTP Accounts',
        description: 'Lists your FTP accounts and the folder each one has access to.',
        side: 'top',
        align: 'start'
    },
    {
        path: '/ftp',
        element: 'a[href="/ftp/new"]',
        title: 'New Account',
        description: 'Click here to create a new FTP account.',
        side: 'bottom',
        align: 'end'
    },

    {
        path: '/backups',
        element: '#backups-steps',
        title: 'Backups',
        description: 'Walks you through connecting a backup destination, then scheduling and restoring backups.',
        side: 'top',
        align: 'start'
    },

    {
        path: '/disk-usage',
        element: '#folders_to_navigate',
        title: 'Disk Usage',
        description: 'Browse folders to see how much space each one uses. Click a folder to drill into it.',
        side: 'top',
        align: 'start'
    },
    {
        path: '/disk-usage',
        element: '#folderChart',
        title: 'Usage Breakdown',
        description: 'A visual breakdown of disk usage for the current folder.',
        side: 'left',
        align: 'start'
    },

    {
        path: '/inodes-explorer',
        element: '#folders_to_navigate',
        title: 'Inodes Explorer',
        description: 'Browse folders to see how many inodes (files) each one contains.',
        side: 'top',
        align: 'start'
    },
    {
        path: '/inodes-explorer',
        element: '#folderChart',
        title: 'Inodes Breakdown',
        description: 'A visual breakdown of inode usage for the current folder.',
        side: 'left',
        align: 'start'
    },

    {
        path: '/malware-scanner',
        element: '#directory-select',
        title: 'ClamAV Scanner',
        description: 'Pick the folder you want to scan for malware.',
        side: 'bottom',
        align: 'start'
    },
    {
        path: '/malware-scanner',
        element: '#start-scan-btn',
        title: 'Start Scan',
        description: 'Click here to start scanning the selected folder. Infected files get quarantined.',
        side: 'bottom',
        align: 'end'
    },

    {
        path: '/fix-permissions',
        element: '#directory-select',
        title: 'Fix Permissions',
        description: 'Pick the folder whose file and folder permissions should be reset to the recommended defaults.',
        side: 'bottom',
        align: 'start'
    },
    {
        path: '/fix-permissions',
        element: '#start-scan-btn',
        title: 'Fix Permissions',
        description: 'Click here to apply the fix.',
        side: 'bottom',
        align: 'end'
    },

    {
        path: '/files.trash',
        element: '#SelectAll-button',
        title: 'Trash',
        description: 'Deleted files are kept here. Select one or more items to restore or permanently delete them.',
        side: 'bottom',
        align: 'start'
    },
    {
        path: '/files.trash',
        element: '#restoreAllButton',
        title: 'Restore / Delete All',
        description: 'Restore everything in Trash back to its original location, or empty the Trash completely.',
        side: 'bottom',
        align: 'start'
    },

    {
        path: '/domains/redirect',
        element: '#save-redirect',
        title: 'Redirects',
        description: 'Select a domain, set the URL it should redirect to, then click Save Redirect.',
        side: 'bottom',
        align: 'end'
    },

    {
        path: '/domains/dynamic-dns',
        element: '#add-entry',
        title: 'Dynamic DNS',
        description: 'Add an entry that automatically keeps a DNS record updated with a changing IP address.',
        side: 'bottom',
        align: 'start'
    },

    {
        path: '/domains/ssl',
        element: '#certData',
        title: 'SSL',
        description: 'Shows the SSL certificate currently installed for the selected domain - issuer, validity dates and more.',
        side: 'top',
        align: 'start'
    },
    {
        path: '/domains/ssl',
        element: '#view-cert',
        title: 'Certificate Files',
        description: 'View the raw certificate and key files for this domain.',
        side: 'top',
        align: 'start'
    },

    {
        path: '/domains/docroot',
        element: '#new_docroot',
        title: 'Change Docroot',
        description: 'Point the selected domain to a different folder on disk to serve its website from.',
        side: 'bottom',
        align: 'start'
    },

    {
        path: '/domains/vhosts',
        element: '#editor-container',
        title: 'Edit Virtual Hosts',
        description: 'Edit the raw virtual host (vhost/Caddyfile) configuration for the selected domain.',
        side: 'top',
        align: 'start'
    },

    {
        path: '/domains/log',
        element: '#logs-table',
        title: 'Raw Access Logs',
        description: 'View the raw access log entries for the selected domain.',
        side: 'top',
        align: 'start'
    },
    {
        path: '/domains/log',
        element: '#showAllCheckbox',
        title: 'Show All',
        description: 'Toggle to show all log entries instead of a limited preview.',
        side: 'bottom',
        align: 'start'
    },

    {
        path: '/mysql/users',
        element: '#users-table',
        title: 'MySQL Users',
        description: 'Lists your MySQL users and which databases each one has access to.',
        side: 'top',
        align: 'start'
    },
    {
        path: '/mysql/wizard',
        element: '#databaseWizardForm',
        title: 'Database Wizard',
        description: 'Create a database, a user and grant access between them - all in one step.',
        side: 'top',
        align: 'start'
    },
    {
        path: '/mysql/processlist',
        element: '#processlist-table',
        title: 'Process List',
        description: 'Shows currently running MySQL queries/connections, useful for spotting slow or stuck queries.',
        side: 'top',
        align: 'start'
    },
    {
        path: '/mysql/remote-mysql',
        element: '#remote',
        title: 'Remote Access',
        description: 'Connection details for accessing this MySQL server remotely, and which users are allowed to.',
        side: 'right',
        align: 'start'
    },
    {
        path: '/mysql/configuration',
        element: '#mysql-config-table',
        title: 'MySQL Configuration',
        description: 'Tune MySQL server configuration values directly.',
        side: 'top',
        align: 'start'
    },
    {
        path: '/mysql/root-password',
        element: '#password',
        title: 'Change Root Password',
        description: 'Set a new root password for this MySQL server.',
        side: 'bottom',
        align: 'start'
    },

    {
        path: '/postgresql/users',
        element: '#users-table',
        title: 'PostgreSQL Users',
        description: 'Lists your PostgreSQL users and which databases each one has access to.',
        side: 'top',
        align: 'start'
    },
    {
        path: '/postgresql/wizard',
        element: '#databaseWizardForm',
        title: 'Database Wizard',
        description: 'Create a database, a user and grant access between them - all in one step.',
        side: 'top',
        align: 'start'
    },
    {
        path: '/postgresql/processlist',
        element: '#processlist-table',
        title: 'Process List',
        description: 'Shows currently running PostgreSQL queries/connections.',
        side: 'top',
        align: 'start'
    },
    {
        path: '/postgresql/remote-postgresql',
        element: '#remote',
        title: 'Remote Access',
        description: 'Connection details for accessing this PostgreSQL server remotely.',
        side: 'right',
        align: 'start'
    },
    {
        path: '/postgresql/configuration',
        element: '#postgresql-config-table',
        title: 'PostgreSQL Configuration',
        description: 'Tune PostgreSQL server configuration values directly.',
        side: 'top',
        align: 'start'
    },

    {
        path: '/auto-installer',
        element: '#autoinstaller-apps',
        title: 'Auto Installer',
        description: 'Pick an application to install - WordPress and others depending on what’s enabled for your plan.',
        side: 'top',
        align: 'start'
    },

    {
        path: '/wordpress',
        element: '#sites-table',
        title: 'WordPress Manager',
        description: 'Lists your installed WordPress sites with their version and quick actions.',
        side: 'top',
        align: 'start'
    },
    {
        path: '/wordpress',
        element: '#scanButton',
        title: 'Scan for Existing Installations',
        description: 'Already have WordPress installed outside the panel? Scan to detect and manage it from here.',
        side: 'bottom',
        align: 'end'
    },

    {
        path: '/emails/aliases',
        element: '#aliases-table',
        title: 'Aliases',
        description: 'Lists email aliases - addresses that forward mail to another address - for your domains.',
        side: 'top',
        align: 'start'
    },
    {
        path: '/emails/filter',
        element: '#filter-email-select',
        title: 'Filters',
        description: 'Pick an email account to manage its filters (sieve rules) for sorting or rejecting incoming mail.',
        side: 'bottom',
        align: 'start'
    },
    {
        path: '/emails/deliverability',
        element: '#deliverabilityTableBody',
        title: 'Email Deliverability',
        description: 'Checks SPF, DKIM and DMARC records for your domains to help your mail avoid the spam folder.',
        side: 'top',
        align: 'start'
    },
    {
        path: '/emails/default',
        element: '#domains',
        title: 'Default Address',
        description: 'Pick a domain, then set a catch-all address for mail sent to addresses that don’t exist on it.',
        side: 'bottom',
        align: 'start'
    },
    {
        path: '/emails/import',
        element: '#toAddActive',
        title: 'Address Importer',
        description: 'Bulk-import email accounts from a CSV file.',
        side: 'top',
        align: 'start'
    },

    {
        path: '/containers/terminal',
        element: '#shell',
        title: 'Terminal',
        description: 'Choose sh or bash, then use the terminal below to run commands interactively inside the container.',
        side: 'bottom',
        align: 'start'
    },
    {
        path: '/containers/logs',
        element: '#log-select',
        title: 'Select a Container',
        description: 'Pick which container’s logs to view.',
        side: 'bottom',
        align: 'start'
    },
    {
        path: '/containers/logs',
        element: '#log-content',
        title: 'Log Content',
        description: 'Once a container is selected, its log output is shown here.',
        side: 'top',
        align: 'start'
    },
    {
        path: '/containers/image/',
        element: '#check-updates-btn',
        title: 'Check for Updates',
        description: 'Click here to check whether newer images are available for your monitored containers.',
        side: 'bottom',
        align: 'end'
    },
    {
        path: '/containers/image/change',
        element: '#service',
        title: 'Change Image Tag',
        description: 'Pick a service and a new image tag to switch it to.',
        side: 'bottom',
        align: 'start'
    },
    {
        path: '/containers/webserver',
        element: '#toAddActive',
        title: 'Change Webserver',
        description: 'Switch the webserver used across your account - this restarts all affected domains.',
        side: 'top',
        align: 'start'
    },
    {
        path: '/containers/mysql',
        element: '#toAddActive',
        title: 'Change MySQL Type',
        description: 'Switch between MySQL and MariaDB for your account.',
        side: 'top',
        align: 'start'
    },

    {
        path: '/mysql/import',
        element: '#file',
        title: 'Import Database',
        description: 'Upload a .sql file to import into an existing database.',
        side: 'bottom',
        align: 'start'
    },
    {
        path: '/postgresql/import',
        element: '#file',
        title: 'Import Database',
        description: 'Upload a .sql file to import into an existing database.',
        side: 'bottom',
        align: 'start'
    },

    {
        path: '/domains/suspend',
        element: '#confirm-suspend',
        title: 'Suspend a Domain',
        description: 'Select a domain, then confirm to suspend it - visitors will see a suspended page instead of the site.',
        side: 'bottom',
        align: 'start'
    },
    {
        path: '/domains/unsuspend',
        element: '#confirm-unsuspend',
        title: 'Unsuspend a Domain',
        description: 'Select a suspended domain, then confirm to bring it back online.',
        side: 'bottom',
        align: 'start'
    },

    {
        path: '/domains/new',
        element: '#domain_url',
        title: 'Domain Name',
        description: 'Enter the domain name you want to add, e.g. example.com.',
        side: 'bottom',
        align: 'start'
    },
    {
        path: '/domains/new',
        element: '#docroot',
        title: 'Document Root',
        description: 'The folder this domain will serve its website from. Defaults to a folder named after the domain.',
        side: 'bottom',
        align: 'start'
    },
    {
        path: '/domains/new',
        element: '#installButton',
        title: 'Add Domain',
        description: 'Click here to create the domain. Progress is streamed below - use Show Details to follow along.',
        side: 'top',
        align: 'end'
    },

    {
        path: '/domains/edit-dns-zone',
        element: '#domains',
        title: 'DNS Zone Editor',
        description: 'Pick a domain to view and edit its DNS records - as a table, or as a raw zone file.',
        side: 'bottom',
        align: 'start'
    },

    {
        path: '/domains/stats',
        element: '#domains',
        title: 'GoAccess',
        description: 'Pick a domain to view its GoAccess report, generated from that domain\'s access logs.',
        side: 'bottom',
        align: 'start'
    },

    {
        path: '/backup-wizard',
        element: '#generate-backup-btn',
        title: 'Generate Backup',
        description: 'Click here to create a compressed backup of all your account files. This runs in the background - the page refreshes automatically when it finishes.',
        side: 'top',
        align: 'start'
    },
    {
        path: '/backup-wizard',
        element: '#existing-backups-section',
        title: 'Existing Backups',
        description: 'Previously generated backups are listed here - click Download to grab a copy.',
        side: 'top',
        align: 'start'
    },

    {
        path: '/file-manager/upload',
        element: '#fileUpload',
        title: 'Upload from Device',
        description: 'Click to browse, or drag and drop files here to upload them into the current folder.',
        side: 'bottom',
        align: 'start'
    },
    {
        path: '/file-manager/upload',
        element: '#change_to_wget',
        title: 'Download from URL Instead',
        description: 'Prefer to fetch a file from a URL instead of uploading from your device? Click here to switch.',
        side: 'top',
        align: 'start'
    },
    {
        path: '/file-manager/upload?method=download',
        element: '#fileWget',
        title: 'File URL',
        description: 'Paste the URL of the file you want the server to download directly into the current folder.',
        side: 'bottom',
        align: 'start'
    },
    {
        path: '/file-manager/upload?method=download',
        element: '#change_to_upload',
        title: 'Upload from Device Instead',
        description: 'Prefer to upload a file from your own device instead? Click here to switch.',
        side: 'top',
        align: 'start'
    },

    {
        path: '/account/passkeys',
        element: '#add-passkey-btn',
        title: 'Add a Passkey',
        description: 'Register a fingerprint, face scan, screen lock, or hardware security key to sign in without a password.',
        side: 'bottom',
        align: 'end'
    },
    {
        path: '/account/passkeys',
        element: '#passkeys-list',
        title: 'Your Passkeys',
        description: 'Lists the passkeys registered to your account. Remove one here if you no longer use that device.',
        side: 'top',
        align: 'start'
    },

    {
        path: '/account/mcp',
        element: '#mcp-generate-form',
        title: 'Generate a Token',
        description: 'Name the token, optionally set an expiry or make it read-only, then generate it to let an MCP client (like Claude) manage this account.',
        side: 'top',
        align: 'start'
    },
    {
        path: '/account/mcp',
        element: '#mcp-connect-section',
        title: 'Connecting a Client',
        description: 'Copy-paste snippets for pointing Claude Code or Claude Desktop at this panel using your generated token.',
        side: 'top',
        align: 'start'
    },

    {
        path: '/account/api',
        element: '#swagger-ui-container',
        title: 'API Reference',
        description: 'Browse every API endpoint and try requests live against this server - pre-authorized with a token for your account.',
        side: 'top',
        align: 'start'
    },
    {
        path: '/account/api',
        element: '#download-spec-btn',
        title: 'Download Spec',
        description: 'Download the raw OpenAPI spec file, e.g. to import into Postman or another API client.',
        side: 'bottom',
        align: 'end'
    }
);
