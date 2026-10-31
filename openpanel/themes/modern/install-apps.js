// theme pages data for install-wizard.html: every installer's real fields, run checkpoints and log lines (from internal/modules/*/install.go and the install templates)
(function () {
  // the test account's domains with their docroot and PHP version, read from /php/install on the test VM
  const DOMAINS = [
    ['shared.tests.openpanel.org', '8.5'], ['wp.tests.openpanel.org', '8.5'], ['waf.tests.openpanel.org', '8.5'], ['capitalize.tests.openpanel.org', '8.5'],
    ['wordpress.tests.openpanel.org', '8.5'], ['php.tests.openpanel.org', '8.3'], ['files.tests.openpanel.org', '7.3'], ['website-builder.tests.openpanel.org', '8.5'],
    ['drupal.tests.openpanel.org', '8.5'], ['joomla.tests.openpanel.org', '7.4'], ['opencart.tests.openpanel.org', '8.5'], ['nextcloud.tests.openpanel.org', '8.5'],
    ['prestashop.tests.openpanel.org', '8.5'], ['moodle.tests.openpanel.org', '8.5'], ['mediawiki.tests.openpanel.org', '8.5'], ['matomo.tests.openpanel.org', '8.5'],
    ['java.tests.openpanel.org', '8.5'], ['ruby.tests.openpanel.org', '8.5'], ['n8n.tests.openpanel.org', '8.5'], ['nodejs.tests.openpanel.org', '8.5'],
    ['python.tests.openpanel.org', '8.5'], ['domains.tests.openpanel.org', '5.6', '/var/www/html/folder_wrzy7'], ['redirect.tests.openpanel.org', '8.5']
  ].map((d, i) => ({ DomainID: i + 1, DomainURL: d[0], Docroot: d[2] || '/var/www/html/' + d[0], PHPVersion: d[1] }));

  const later = (list) => () => new Promise((ok) => setTimeout(() => ok(list), 500));
  const VERSIONS = {
    wordpress: ['7.1.4', '7.0.6', '6.9.4', '6.8.5', '6.7.5', '6.6.4'],
    joomla: ['6.2.0', '6.1.3', '6.0.5', '5.4.10', '5.4.9', '5.3.4'],
    drupal: ['11.3.2', '11.3.1', '11.2.8', '11.1.9', '10.6.2', '10.5.6'],
    prestashop: ['8.2.8', '8.2.7', '8.2.6', '8.2.5', '8.1.7', '8.0.5'],
    opencart: ['4.1.0.3', '4.1.0.2', '4.1.0.1', '4.0.2.3', '3.0.4.1'],
    phpbb: ['3.3.15', '3.3.14', '3.3.13', '3.3.12'],
    flarum: ['2.0.1', '2.0.0', '1.8.10', '1.8.9'],
    mediawiki: ['1.46.2', '1.46.1', '1.46.0', '1.45.6', '1.45.5', '1.44.6'],
    matomo: ['5.14.2', '5.14.1', '5.14.0'],
    nextcloud: ['35.0.2', '35.0.1', '35.0.0', '34.0.5', '34.0.4', '33.0.10'],
    moodle: ['5.1.3', '5.1.2', '5.0.6', '4.5.10', '4.4.12'],
    ojs: ['3.5.0-3', '3.5.0-2', '3.4.0-10', '3.3.0-22'],
    nodejs: ['26.13.0', '25.9.0', '24.22.2', '23.11.1', '22.23.4', '21.7.3'],
    python: ['3.15.2', '3.14.10', '3.13.18', '3.12.16', '3.11.18', '3.10.22'],
    ruby: ['4.0.9', '4.0.8', '4.0.7', '4.0.6', '4.0.5', '4.0.4'],
    java: ['26-jdk-jammy', '25-jdk-jammy', '21-jdk-jammy', '17-jdk-jammy', '11-jdk-jammy', '8-jdk-jammy'],
    n8n: ['3.0.1', '3.0.0', '2.46.2', '2.46.1', '2.46.0', '2.45.6']
  };

  // steps
  const WHERE = { id: 'where', label: 'Location', title: 'Where to install', desc: 'Pick one of your domains and, if you like, a folder under it.' };
  const SITE = { id: 'site', label: 'Site & admin', title: 'Site and admin account', desc: 'The admin account is what you log in to the new site with.' };
  const DB = { id: 'db', label: 'Database', title: 'Database', desc: 'Names and a password are generated for you. Change them only if you need specific ones.' };
  const REVIEW = { id: 'review', label: 'Review', title: 'Review and install', desc: 'Check the details, then start the install. It takes a minute or two.' };

  // fields
  const version = (slug, latestOption) => ({ name: slug + '_version', label: 'Version', type: 'select', versions: later(VERSIONS[slug]), latestOption: latestOption, half: true, review: 'Version' });
  const email = (o) => Object.assign({ name: 'admin_email', label: 'Admin email', type: 'email', required: true, validate: 'email', auto: (d) => 'admin@' + d, half: true, placeholder: 'admin@example.com' }, o || {});
  const user = (o) => Object.assign({ name: 'admin_username', label: 'Admin username', value: 'admin', required: true, validate: 'username', half: true }, o || {});
  const pass = (o) => Object.assign({ name: 'admin_password', label: 'Admin password', type: 'password', gen: 'password', required: true, validate: 'adminpass', strength: true }, o || {});
  const dbFields = (slug) => [
    { name: 'db_name', step: 'db', label: 'Database name', prefix: true, gen: () => (slug ? slug + '_' : '') + Math.random().toString(36).slice(2, 8), required: true, validate: 'dbname', mono: true, half: true },
    { name: 'db_user', step: 'db', label: 'Database user', prefix: true, gen: 'lower10', required: true, validate: 'dbname', mono: true, half: true },
    { name: 'db_password', step: 'db', label: 'Database password', type: 'password', gen: 'alnum16', required: true, validate: 'dbpass' }
  ];
  const credsCms = (userField) => [
    { label: 'Username', field: userField || 'admin_username', from: 'user' },
    { label: 'Password', field: 'admin_password', from: 'pass', secret: true },
    { label: 'Email', field: 'admin_email' },
    { label: 'Database', field: 'db_name', prefix: true },
    { label: 'Database user', field: 'db_user', prefix: true }
  ];

  // run checkpoints, same labels and triggers as the panel's install pages
  const R = {
    pre: { label: 'Running pre-flight checks', triggers: ['Checking if existing installation', 'Validating provided data'] },
    php: { label: 'Starting PHP container', triggers: ['Starting PHP container'] },
    dl: (n) => ({ label: 'Downloading ' + n + ' archive', triggers: ['Downloading', 'Using existing archive'] }),
    ex: (more) => ({ label: 'Extracting files to document root', triggers: ['Extracting files', 'Setting files permissions'].concat(more || []) }),
    db: { label: 'Creating database', triggers: ['Testing database connection', 'Checking', 'Creating database'] },
    cli: (label, t) => ({ label: label, triggers: t }),
    save: (more) => ({ label: 'Saving to Site Manager', triggers: (more || []).concat(['Saving website information', 'installation completed']) })
  };

  // log lines the real handlers stream, values filled from the form
  function cmsLog(vm, name, mid) {
    const c = 'php-fpm-' + (vm.dom().PHPVersion || '8.3'), p = vm.installPath(), u = 'testinguser';
    return ['Checking if existing installation processes are running..', 'Validating provided data', 'Starting PHP container: ' + c]
      .concat(mid.map((l) => l.replace('{path}', p).replace('{user}', u).replace('{db}', vm.v.db_name).replace('{dbuser}', vm.v.db_user).replace('{ver}', vm.v[vm.cfg.slug + '_version'] || VERSIONS[vm.cfg.slug] && VERSIONS[vm.cfg.slug][0])))
      .concat(['Saving website information to Site Manager', name + ' installation completed!']);
  }
  const dbLines = ['Testing database connection..', 'Creating database {db} and user {dbuser}'];

  const CMS = {
    wordpress: {
      name: 'WordPress', desc: 'WordPress content management system (CMS).', adminPath: '/wp-admin/', manager: 'wordpress.html', managerLabel: 'Open WP Manager',
      steps: [WHERE, SITE, DB, REVIEW], success: ['WordPress installation completed'], summary: ['website_name', 'wordpress_version', 'admin_username', 'db_name'],
      fields: [
        { name: 'website_name', label: 'Website name', required: true, placeholder: 'My Blog', half: true },
        { name: 'site_description', label: 'Tagline', placeholder: 'Just another WordPress site', half: true },
        { name: 'wordpress_version', label: 'WordPress version', type: 'select', versions: later(VERSIONS.wordpress), half: true, review: 'Version' },
        email(), user({ value: undefined, gen: 'alnum8' }), pass(),
        { name: 'db_name', step: 'db', label: 'Database name', prefix: true, gen: () => Math.random().toString(36).slice(2, 10), required: true, validate: 'dbname', mono: true, half: true },
        { name: 'db_prefix', step: 'db', label: 'Table prefix', gen: 'prefix', required: true, validate: 'dbprefix', mono: true, half: true },
        { name: 'db_user', step: 'db', label: 'Database user', prefix: true, gen: () => Math.random().toString(36).slice(2, 10), required: true, validate: 'dbname', mono: true, half: true },
        { name: 'db_password', step: 'db', label: 'Database password', type: 'password', gen: 'alnum12', required: true, validate: 'dbpass', half: true },
        { name: 'sets_note', step: 'db', type: 'note', label: 'Plugin and theme sets your hosting provider set up are installed right after WordPress, nothing to pick here.' }
      ],
      run: [
        { label: 'Starting PHP container', triggers: ['Starting PHP container'] },
        { label: 'Downloading WordPress archive', triggers: ['Downloading', 'Using existing archive'] },
        { label: 'Extracting files to document root', triggers: ['Extracting files'] },
        { label: 'Creating .htaccess and wp-config.php', triggers: ['Creating .htaccess', 'Writing database logins', 'Added mu-plugin', 'Set FS_METHOD'] },
        { label: 'Generating security keys and salts', triggers: ['Shuffling WordPress Salts'] },
        { label: 'Setting file permissions and ownership', triggers: ['Setting files permissions and owner', 'Fixing wp-content permissions'] },
        { label: 'Testing WP-CLI', triggers: ['WP-CLI is functional', 'WP-CLI not available'] },
        { label: 'Creating database and MySQL user', triggers: ['Testing database connection', 'Checking', 'container is healthy', 'Creating database'] },
        { label: 'Importing tables into the database', triggers: ['Importing WordPress tables'] },
        { label: 'Configuring permalinks and site settings', triggers: ['Setting pretty permalinks', 'Setting site tagline'] },
        { label: 'Enabling auto-login from Site Manager', triggers: ['Enabling auto-login'] },
        { label: 'Installing plugin and theme sets', triggers: ['Checking for server-wide', 'Installing plugin', 'Installing theme'] },
        { label: 'Saving to Site Manager', triggers: ['Saving website information', 'Reloading'] }
      ],
      creds: [{ label: 'Username', field: 'admin_username', from: 'user' }, { label: 'Password', field: 'admin_password', from: 'pass', secret: true }, { label: 'Email', field: 'admin_email' }, { label: 'Database', field: 'db_name', prefix: true }, { label: 'Database user', field: 'db_user', prefix: true }, { label: 'Table prefix', field: 'db_prefix' }],
      log: (vm) => {
        const c = 'php-fpm-' + vm.dom().PHPVersion, p = vm.installPath();
        return ['Starting PHP container: ' + c, 'Downloading https://wordpress.org/wordpress-' + vm.v.wordpress_version + '.tar.gz', 'Extracting files to ' + p, 'Creating .htaccess file for WordPress permalinks', 'Writing database logins to wp-config.php', 'Shuffling WordPress Salts in wp-config.php file', 'Added mu-plugin to force HTTPS for wp-admin behind Varnish.', "Setting files permissions and owner to 'testinguser'", 'PHP container is running and WP-CLI is functional, starting process..', 'Testing database connection..', 'Creating database ' + vm.v.db_name + ' and user ' + vm.v.db_user, 'Importing WordPress tables in the database', 'Setting pretty permalinks for WordPress', "Setting site tagline to '" + (vm.v.site_description || 'My WordPress Blog') + "'", 'Enabling auto-login from SiteManager', 'Checking for server-wide or user provided themes and plugins sets to be installed', 'Saving website information to SiteManager', 'Reloading ' + c + ' to apply rewrite rules from .htaccess file', 'WordPress installation completed!'];
      }
    },
    joomla: {
      name: 'Joomla', desc: 'Joomla content management system (CMS).', adminPath: '/administrator/',
      steps: [WHERE, SITE, DB, REVIEW], summary: ['site_name', 'joomla_version', 'admin_username', 'db_name'],
      fields: [{ name: 'site_name', label: 'Website name', required: true, placeholder: 'My Joomla Site', half: true }, version('joomla', true), email(), user(), { name: 'admin_name', label: 'Full name', value: 'Administrator', required: true }, pass()].concat(dbFields('joomla')),
      run: [R.pre, R.php, R.dl('Joomla'), R.ex(), R.db, R.cli('Running Joomla CLI installer', ['Running Joomla CLI installer']), R.save(['Deploying admin login helper'])],
      creds: credsCms(), log: (vm) => cmsLog(vm, 'Joomla', ['Downloading https://github.com/joomla/joomla-cms/releases/download/{ver}/Joomla_{ver}-Stable-Full_Package.tar.gz', 'Extracting files to {path}', "Setting files permissions and owner to '{user}'"].concat(dbLines, ['Running Joomla CLI installer', 'Deploying admin login helper']))
    },
    drupal: {
      name: 'Drupal', desc: 'Drupal content management system (CMS).', adminPath: '/user/login',
      steps: [WHERE, SITE, DB, REVIEW], summary: ['site_name', 'drupal_version', 'admin_username', 'db_name'],
      fields: [{ name: 'site_name', label: 'Website name', required: true, placeholder: 'My Drupal Site', half: true }, version('drupal', true), email(), user(), pass()].concat(dbFields('drupal')),
      run: [R.pre, R.php, R.cli('Creating Composer project', ['Creating Composer project', 'Requiring drush', 'Linking web root', 'Setting files permissions']), R.db, R.cli('Running drush site:install', ['Running drush site:install']), R.save()],
      creds: credsCms(), log: (vm) => cmsLog(vm, 'Drupal', ['Creating Composer project drupal/recommended-project:{ver}', 'Requiring drush/drush', 'Linking web root into docroot', "Setting files permissions and owner to '{user}'"].concat(dbLines, ['Running drush site:install']))
    },
    prestashop: {
      name: 'PrestaShop', desc: 'PrestaShop e-commerce platform.', doneNote: 'PrestaShop gets a renamed admin folder during the install. Use Log in on the site in Site Manager to open the back office.',
      steps: [WHERE, SITE, DB, REVIEW], summary: ['prestashop_version', 'admin_email', 'db_name'],
      fields: [version('prestashop', true), email(), { name: 'admin_firstname', label: 'First name', value: 'Admin', required: true, half: true }, { name: 'admin_lastname', label: 'Last name', value: 'User', required: true, half: true }, pass()].concat(dbFields('prestashop')),
      run: [R.pre, R.php, R.dl('PrestaShop'), R.ex(), R.db, R.cli('Running PrestaShop CLI installer', ['Running PrestaShop CLI installer']), R.save(['Removing installer folder', 'Securing admin directory', 'Deploying admin login helper'])],
      creds: [{ label: 'Email (login)', field: 'admin_email' }, { label: 'Password', field: 'admin_password', from: 'pass', secret: true }, { label: 'Database', field: 'db_name', prefix: true }, { label: 'Database user', field: 'db_user', prefix: true }],
      log: (vm) => cmsLog(vm, 'PrestaShop', ['Downloading https://github.com/PrestaShop/PrestaShop/releases/download/{ver}/prestashop_{ver}.zip', 'Extracting files to {path}', "Setting files permissions and owner to '{user}'", 'Ensuring PHP container temp directory is writable across filesystems'].concat(dbLines, ['Running PrestaShop CLI installer', 'Warming up production cache', 'Removing installer folder', 'Setting permissions on writable directories', 'Securing admin directory', 'Deploying admin login helper']))
    },
    opencart: {
      name: 'OpenCart', desc: 'OpenCart e-commerce platform.', adminPath: '/admin/',
      steps: [WHERE, SITE, DB, REVIEW], summary: ['opencart_version', 'admin_username', 'db_name'],
      fields: [version('opencart', true), email(), user(), pass()].concat(dbFields('opencart')),
      run: [R.pre, R.php, R.dl('OpenCart'), R.ex(['Preparing config.php']), R.db, R.cli('Running OpenCart CLI installer', ['Running OpenCart CLI installer']), R.save(['Removing installer folder', 'Deploying admin login helper'])],
      creds: credsCms(), log: (vm) => cmsLog(vm, 'OpenCart', ['Downloading https://github.com/opencart/opencart/releases/download/{ver}/opencart-{ver}.zip', 'Extracting files to {path}', 'Preparing config.php files', "Setting files permissions and owner to '{user}'"].concat(dbLines, ['Running OpenCart CLI installer', 'Removing installer folder', 'Deploying admin login helper']))
    },
    phpbb: {
      name: 'phpBB', desc: 'phpBB forum software.', adminPath: '/ucp.php?mode=login',
      steps: [WHERE, Object.assign({}, SITE, { title: 'Board and admin account' }), DB, REVIEW], summary: ['board_name', 'phpbb_version', 'admin_username', 'db_name'],
      fields: [{ name: 'board_name', label: 'Board name', value: 'My Board', required: true, half: true }, { name: 'board_description', label: 'Board description', value: 'Powered by phpBB', half: true }, version('phpbb', true), email(), user(), pass()].concat(dbFields('phpbb')),
      run: [R.pre, R.php, R.cli('Downloading and extracting phpBB', ['Downloading and extracting phpBB', 'Setting files permissions']), R.db, R.cli('Running phpBB installer', ['Running phpBB installer', 'Removing install/']), R.save()],
      creds: credsCms(), log: (vm) => cmsLog(vm, 'phpBB', ['Downloading and extracting phpBB', "Setting files permissions and owner to '{user}'"].concat(dbLines, ['Running phpBB installer', 'Removing install/ directory']))
    },
    flarum: {
      name: 'Flarum', desc: 'Flarum forum software.', adminPath: '/admin', php: { min: '8.1', block: true },
      steps: [WHERE, Object.assign({}, SITE, { title: 'Forum and admin account' }), DB, REVIEW], summary: ['site_name', 'flarum_version', 'admin_username', 'db_name'],
      fields: [{ name: 'site_name', label: 'Forum title', value: 'Flarum', required: true, placeholder: 'My Flarum Forum', half: true }, version('flarum', true), email(), user(), pass()].concat(dbFields('flarum')),
      run: [R.pre, R.php, R.cli('Creating Composer project', ['Creating Composer project', 'Linking public root', 'Setting files permissions']), R.db, R.cli('Running Flarum installer', ['Running Flarum installer']), R.save(['Registering cron job'])],
      creds: credsCms(), log: (vm) => cmsLog(vm, 'Flarum', ['Creating Composer project flarum/flarum:{ver}', 'Linking public root into docroot', "Setting files permissions and owner to '{user}'"].concat(dbLines, ['Running Flarum installer', 'Registering cron job (flarum schedule:run, every minute)']))
    },
    mediawiki: {
      name: 'MediaWiki', desc: 'MediaWiki wiki platform.', adminPath: '/index.php?title=Special:UserLogin',
      steps: [WHERE, Object.assign({}, SITE, { title: 'Wiki and admin account' }), DB, REVIEW], summary: ['site_name', 'mediawiki_version', 'admin_username', 'db_name'],
      fields: [{ name: 'site_name', label: 'Site name', value: 'MediaWiki', required: true, half: true }, version('mediawiki', true), user(), email(), pass()].concat(dbFields('mediawiki')),
      run: [R.pre, R.php, R.dl('MediaWiki'), R.ex(), R.db, R.cli('Running MediaWiki CLI installer', ['Running MediaWiki CLI installer']), R.save(['Deploying admin login helper', 'Registering cron job'])],
      creds: credsCms(), log: (vm) => cmsLog(vm, 'MediaWiki', ['Downloading https://releases.wikimedia.org/mediawiki/1.46/mediawiki-{ver}.tar.gz', 'Extracting files to {path}', "Setting files permissions and owner to '{user}'"].concat(dbLines, ['Running MediaWiki CLI installer (maintenance/install.php)', 'Deploying admin login helper', 'Registering cron job (maintenance/runJobs.php, every minute)']))
    },
    dokuwiki: {
      name: 'DokuWiki', desc: 'File-based wiki software with versioned releases and no database required.', adminPath: '/doku.php?do=login', php: { min: '7.4', block: true },
      note: 'No database: the admin account and site settings are written straight into DokuWiki\'s config, so the wiki is ready right after the install.',
      steps: [WHERE, Object.assign({}, SITE, { title: 'Wiki and admin account' }), REVIEW], summary: ['site_title', 'admin_user'],
      fields: [{ name: 'site_title', label: 'Site title', auto: (d) => d, placeholder: 'My Wiki' }, { name: 'admin_user', label: 'Admin username', value: 'admin', validate: 'username', lower: true, half: true }, email({ required: false, half: true }), pass({ required: false, hint: 'Leave it empty to get a generated one.' })],
      run: [R.pre, R.php, R.cli('Downloading and extracting DokuWiki', ['Downloading and extracting DokuWiki']), R.cli('Writing admin account and configuration', ['Hashing admin password', 'Writing DokuWiki configuration', 'Setting files permissions']), R.save()],
      creds: [{ label: 'Username', field: 'admin_user', from: 'user' }, { label: 'Password', field: 'admin_password', from: 'pass', secret: true }],
      log: (vm) => cmsLog(vm, 'DokuWiki', ['Downloading and extracting DokuWiki', 'Hashing admin password', 'Writing DokuWiki configuration', "Setting files permissions and owner to '{user}'"])
    },
    sofawiki: {
      name: 'SofaWiki', desc: 'File-based wiki software, no database required.', php: { max: '7.4', block: true },
      note: 'SofaWiki has no database and no CLI installer. After the files are in place, open the site to finish its own setup (folder rights, then a login).',
      doneNote: 'Open the site to finish SofaWiki\'s own setup wizard: folder rights first, then the login and main page.',
      steps: [WHERE, { id: 'site', label: 'Details', title: 'Details', desc: 'SofaWiki creates its admin account in its own setup wizard after the install.' }, REVIEW], summary: ['admin_email'],
      fields: [email({ required: false, label: 'Email', hint: 'For your own records in Site Manager only, SofaWiki has no admin account to create yet.' })],
      run: [R.pre, R.php, R.cli('Downloading and extracting SofaWiki', ['Downloading and extracting SofaWiki', 'Setting files permissions']), R.save()],
      creds: [],
      log: (vm) => cmsLog(vm, 'SofaWiki', ['Downloading and extracting SofaWiki', "Setting files permissions and owner to '{user}'"]).map((l) => l === 'SofaWiki installation completed!' ? 'SofaWiki installation completed! Visit the site to finish setup (folder rights, then the SofaWiki setup wizard).' : l)
    },
    matomo: {
      name: 'Matomo', desc: 'Matomo web analytics platform.', adminPath: '/',
      steps: [WHERE, SITE, DB, REVIEW], summary: ['matomo_version', 'admin_login', 'db_name'],
      fields: [version('matomo', true), user({ name: 'admin_login', label: 'Admin login' }), email(), pass()].concat(dbFields('matomo')),
      run: [R.pre, R.php, R.dl('Matomo'), R.ex(['Setting permissions on writable']), R.db, R.cli('Running Matomo installation wizard', ['Running Matomo installation wizard']), R.save(['Deploying admin login helper', 'Storing admin credentials'])],
      creds: credsCms('admin_login'), log: (vm) => cmsLog(vm, 'Matomo', ['Downloading https://builds.matomo.org/matomo-{ver}.zip', 'Extracting files to {path}', "Setting files permissions and owner to '{user}'", 'Setting permissions on writable directories'].concat(dbLines, ['Running Matomo installation wizard', 'Deploying admin login helper']))
    },
    nextcloud: {
      name: 'Nextcloud', desc: 'Nextcloud file sync and collaboration platform.', adminPath: '/',
      steps: [WHERE, SITE, DB, REVIEW], summary: ['nextcloud_version', 'admin_username', 'db_name'],
      fields: [version('nextcloud', true), email(), user(), pass()].concat(dbFields('nextcloud')),
      run: [R.pre, R.php, R.dl('Nextcloud'), R.ex(), R.db, R.cli('Running Nextcloud CLI installer', ['Running Nextcloud CLI installer']), R.save(['Configuring trusted domain', 'Deploying admin login helper'])],
      creds: credsCms(), log: (vm) => cmsLog(vm, 'Nextcloud', ['Downloading https://download.nextcloud.com/server/releases/nextcloud-{ver}.tar.bz2', 'Extracting files to {path}', "Setting files permissions and owner to '{user}'"].concat(dbLines, ['Running Nextcloud CLI installer (occ maintenance:install)', 'Configuring trusted domain and site URL', 'Deploying admin login helper']))
    },
    tinyphotogallery: {
      name: 'TinyPhotoGallery', desc: 'Single-file photo gallery, no database required.', php: { min: '8.1' },
      note: 'No database and no admin account: this downloads index.php and creates an empty photos/ folder. Needs PHP 8.1+ with the GD extension.',
      doneNote: 'Open the site to start uploading photos.',
      steps: [WHERE, REVIEW], creds: [],
      run: [R.pre, R.php, R.cli('Downloading TinyPhotoGallery', ['Downloading TinyPhotoGallery', 'Setting files permissions']), R.save()],
      log: (vm) => cmsLog(vm, 'TinyPhotoGallery', ['Downloading TinyPhotoGallery', "Setting files permissions and owner to '{user}'"]).map((l) => l === 'TinyPhotoGallery installation completed!' ? 'TinyPhotoGallery installation completed! Visit the site to start uploading photos.' : l)
    },
    tinyfilemanager: {
      name: 'TinyFileManager', desc: 'Single-file web-based file manager with login protection, no database required.', adminPath: '/tinyfilemanager.php',
      note: 'No database: the login you set here is written into tinyfilemanager.php.',
      steps: [WHERE, { id: 'site', label: 'Login', title: 'Login', desc: 'The account that protects the file manager.' }, REVIEW], summary: ['admin_username'],
      fields: [user(), pass()],
      run: [R.pre, R.php, R.cli('Downloading TinyFileManager', ['Downloading TinyFileManager', 'Hashing admin password', 'Writing admin credentials', 'Setting files permissions']), R.save()],
      creds: [{ label: 'Username', field: 'admin_username', from: 'user' }, { label: 'Password', field: 'admin_password', from: 'pass', secret: true }],
      log: (vm) => cmsLog(vm, 'TinyFileManager', ['Downloading TinyFileManager', 'Hashing admin password', 'Writing admin credentials', "Setting files permissions and owner to '{user}'"])
    },
    moodle: {
      name: 'Moodle', desc: 'Moodle learning management system (LMS).', adminPath: '/login/index.php',
      steps: [WHERE, SITE, DB, REVIEW], summary: ['moodle_version', 'admin_username', 'db_name'],
      fields: [version('moodle', true), user(), email(), pass()].concat(dbFields('moodle')),
      run: [R.pre, R.php, R.dl('Moodle'), R.ex(['Linking web root']), R.db, R.cli('Running Moodle CLI installer', ['Running Moodle CLI installer']), R.save(['Setting permissions on moodledata', 'Registering cron job'])],
      creds: credsCms(), log: (vm) => cmsLog(vm, 'Moodle', ['Downloading https://download.moodle.org/download.php/direct/stable501/moodle-{ver}.tgz', 'Extracting files to /var/www/moodle', "Setting files permissions and owner to '{user}'", "Linking web root to Moodle's public/ directory"].concat(dbLines, ['Running Moodle CLI installer (admin/cli/install.php)', 'Setting permissions on moodledata directory', 'Registering cron job (admin/cli/cron.php, every minute)']))
    },
    ojs: {
      name: 'OJS', desc: 'Open source software for managing and publishing scholarly journals.', adminPath: '/index.php/index/login',
      steps: [WHERE, SITE, DB, REVIEW], summary: ['ojs_version', 'admin_username', 'db_name'],
      fields: [version('ojs', true), user(), email(), pass()].concat(dbFields('ojs')),
      run: [R.pre, R.php, R.dl('OJS'), R.ex(['Linking web root']), R.db, R.cli('Running OJS CLI installer', ["Ensuring PHP 'ftp' extension", 'Running OJS CLI installer', 'Finalizing configuration', 'Setting permissions on OJS files directory', 'Deploying admin login helper']), R.save(['Registering cron job'])],
      creds: credsCms(), log: (vm) => cmsLog(vm, 'OJS', ['Downloading https://pkp.sfu.ca/ojs/download/ojs-{ver}.tar.gz', 'Extracting files to /var/www/ojs', "Setting files permissions and owner to '{user}'", 'Linking web root to OJS app directory'].concat(dbLines, ["Ensuring PHP 'ftp' extension is installed", 'Running OJS CLI installer (tools/install.php)', 'Finalizing configuration (base_url, time_zone)', 'Setting permissions on OJS files directory', 'Deploying admin login helper', 'Registering cron job (lib/pkp/tools/scheduler.php, every minute)']))
    }
  };
  Object.keys(CMS).forEach((k) => {
    const c = CMS[k];
    c.slug = k;
    c.success = c.success || ['installation completed'];
    c.manager = c.manager || 'site-cms.html';
    if (c.slug === 'wordpress' || c.fields && c.fields.some((f) => f.name === 'db_name')) c.mysqlPrefix = '';
  });

  // runtimes, from manager/app_install.html and appinstall/install.go
  const RT_RUN = [
    R.pre,
    { label: 'Adding the service to docker-compose.yml', triggers: ['Generating docker-compose entry', 'Starting the actual install', "Added '", 'Service added successfully'] },
    { label: 'Starting the application container', triggers: ['Starting docker container', 'Checking if container started', 'is running successfully', 'n8n owner account', 'Waiting for n8n'] },
    { label: 'Configuring the reverse proxy', triggers: ['Detecting installed webserver', 'Editing Apache configuration', 'Editing Nginx configuration', 'Editing Litespeed configuration', 'Editing OpenResty configuration'] },
    { label: 'Restarting the web server', triggers: ['Reloading', 'Restarting'] },
    { label: 'Saving to Site Manager', triggers: ['Installation finished', 'application setup completed'] }
  ];
  const RES = { id: 'res', label: 'Resources', title: 'Resources', desc: 'Limits for the app\'s container. You can change them later on the app\'s page.' };
  const resFields = [
    { name: 'res_presets', step: 'res', type: 'presets', label: 'Presets', presets: [{ label: 'Small', values: { cpu_limit: '0.25', mem_limit: '0.25', pids_limit: '100' } }, { label: 'Medium', values: { cpu_limit: '0.5', mem_limit: '1', pids_limit: '500' } }, { label: 'Large', values: { cpu_limit: '1', mem_limit: '2', pids_limit: '1000' } }] },
    { name: 'cpu_limit', step: 'res', label: 'CPU cores', type: 'number', value: '0.5', min: 0.1, max: 9.99, inc: 0.01, required: true, half: true, hint: 'CPU cores the container can use.' },
    { name: 'mem_limit', step: 'res', label: 'Memory', type: 'number', value: '1', min: 0.1, max: 9.99, inc: 0.01, unit: 'GB', required: true, half: true, hint: 'Physical memory the container can use.' },
    { name: 'pids_limit', step: 'res', label: 'Max processes', type: 'number', value: '500', min: 10, max: 1000, inc: 1, required: true, half: true, hint: 'How many processes the app can run.' }
  ];
  function runtime(id, name, run, ext, file, install) {
    return {
      name: name, desc: name + ' containerized application.', manager: 'site-runtime.html', cta: 'Create ' + name + ' app', success: ['application setup completed'],
      steps: [WHERE, { id: 'app', label: 'Application', title: 'Application', desc: 'The app runs in its own container, the domain is reverse-proxied to its port.' }, RES, REVIEW], summary: ['service_name', 'version', 'port', 'startup_file'],
      note: 'Installing at the domain root is blocked when the domain already has sites in subfolders, and the other way round.',
      fields: [
        { name: 'service_name', step: 'app', label: 'Name', required: true, validate: 'service', lower: true, strip: /[^a-z0-9]/g, half: true, mono: true, hint: 'Container name, only used on the server.', auto: (d) => d.split('.')[0].replace(/[^a-z0-9]/g, '').replace(/^[0-9]+/, '').slice(0, 20) || 'app' },
        { name: 'version', step: 'app', label: 'Version', type: 'select', versions: later(VERSIONS[id]), half: true },
        { name: 'port', step: 'app', label: 'Port', type: 'number', min: 1000, max: 65535, validate: 'port', half: true, mono: true, hint: 'The port your app listens on. Left empty, port 80 is used.' },
        { name: 'startup_file', step: 'app', label: 'Startup file', required: true, mono: true, ext: ext, auto: (d, vm) => vm.installPath() + '/' + file, followsPath: true, hint: 'The main entry point of the app, usually ' + file + '.', check: (v) => new Promise((ok) => setTimeout(() => ok(v.toLowerCase().endsWith(ext) ? (v.endsWith('/' + file) ? { tone: 'success', text: 'File exists' } : { tone: 'warning', text: 'File does not exist yet, upload it before the app starts.' }) : { tone: 'warning', text: 'Incorrect file extension, use ' + ext + '.' }), 300)) },
        { name: 'requirements', step: 'app', type: 'check', label: 'Run ' + install + ' before starting the app', value: true, hint: 'Installs the dependencies first. Skip it if the app is already built.', on: 'Yes', offLabel: 'No' },
        { name: 'custom_cmd', step: 'app', label: 'Startup command', mono: true, more: true, placeholder: run + ' <startup file>', hint: 'Runs instead of the default: ' + run + ' <startup file>' },
        { name: 'git_repo_url', step: 'app', label: 'Git repository URL', validate: 'https', mono: true, more: true, placeholder: 'https://github.com/user/repo.git', hint: 'Public repos only, for a private one put a personal access token in the URL: https://<token>@github.com/user/repo.git', onBlur: (v, vm) => { if (!v) return; vm.hints.git_repo_url = { tone: 'success', text: 'Detected the startup file from the repository, change it if needed.' }; vm.v.startup_file = vm.installPath() + '/' + file; vm.touched.startup_file = true; } }
      ].concat(resFields),
      run: RT_RUN,
      creds: [{ label: 'Container', field: 'service_name' }, { label: 'Port', field: 'port' }],
      log: (vm) => ['Checking if existing installation processes are running..', 'Validating provided data', 'Generating docker-compose entry for ' + name + " application: '" + vm.v.service_name + "'", 'Starting the actual install of ' + name + ' application', "Added '" + vm.v.service_name + "' to docker-compose.yml", 'Service added successfully', 'Starting docker container..', 'Checking if container started...', "Container '" + vm.v.service_name + "' is running successfully.", 'Detecting installed webserver', 'Editing Nginx configuration: creating reverse proxy from ' + vm.host() + ' to container ' + vm.v.service_name, 'Reloading nginx to apply the new proxy', 'Installation finished, adding the new application to SiteManager', 'New ' + name + ' application setup completed!']
    };
  }
  const RT = {
    nodejs: runtime('nodejs', 'NodeJS', 'node', '.js', 'index.js', 'npm install'),
    python: runtime('python', 'Python', 'python', '.py', 'app.py', 'pip install'),
    ruby: runtime('ruby', 'Ruby', 'ruby', '.rb', 'app.rb', 'bundle install'),
    java: runtime('java', 'Java', 'java', '.java', 'Main.java', 'mvn install')
  };
  RT.nodejs.name = 'Node.js';
  const n8n = runtime('n8n', 'n8n', '', '', '', '');
  n8n.desc = 'Workflow automation, containerized.';
  n8n.manager = 'site-docker.html';
  n8n.cta = 'Install n8n';
  n8n.steps = [WHERE, { id: 'app', label: 'Application', title: 'Application', desc: 'n8n runs in its own container on port 5678, the domain is reverse-proxied to it.' }, { id: 'owner', label: 'Owner account', title: 'Owner account', desc: 'Created for you, so you land on n8n\'s login instead of its setup wizard.' }, RES, REVIEW];
  n8n.summary = ['service_name', 'version', 'owner_email'];
  n8n.fields = n8n.fields.filter((f) => ['service_name', 'version'].includes(f.name)).concat([
    { name: 'port', step: 'app', label: 'Port', value: '5678', readonly: true, half: true, mono: true, hint: 'n8n always listens on 5678.' },
    { name: 'owner_first_name', step: 'owner', label: 'First name', value: 'Admin', required: true, half: true },
    { name: 'owner_last_name', step: 'owner', label: 'Last name', value: 'User', required: true, half: true },
    { name: 'owner_email', step: 'owner', label: 'Email', type: 'email', required: true, validate: 'email', auto: (d) => 'admin@' + d },
    { name: 'owner_password', step: 'owner', label: 'Password', type: 'password', gen: 'password', required: true, minLen: 8, maxLen: 64, hint: '8 to 64 characters, you log in to n8n with it.' }
  ], resFields);
  n8n.creds = [{ label: 'Email', field: 'owner_email' }, { label: 'Password', field: 'owner_password', secret: true }, { label: 'Container', field: 'service_name' }];
  const rtLog = n8n.log;
  n8n.log = (vm) => { const l = rtLog(vm); l.splice(9, 0, 'Creating n8n owner account', 'Waiting for n8n to finish setting up its database', 'n8n owner account created'); return l; };
  RT.n8n = n8n;

  const OTHER = {
    php: {
      name: 'PHP application', desc: 'Set up a Composer project in an existing domain.', manager: 'site-php.html', cta: 'Create PHP app', success: ['application setup completed'], php: { needs: true, block: true },
      steps: [WHERE, { id: 'project', label: 'Project', title: 'Composer project', desc: 'Optional. Leave it empty to use files already in the folder.' }, REVIEW], summary: ['initial_project'],
      fields: [
        { name: 'initial_project', step: 'project', label: 'Initial project', validate: 'project', mono: true, placeholder: 'laravel/laravel or https://example.com/project.zip', hint: 'Created when the folder is empty: a Composer package name or an https:// archive URL (.zip, .tar.gz, .tgz, .tar).', empty: 'Use existing files' },
        { name: 'autorun_composer_install', step: 'project', type: 'check', label: 'Run composer install after creating the project', hint: 'Installs the dependencies from composer.json in the PHP container.', on: 'Yes', offLabel: 'No' },
        { name: 'composer_optimize_autoloader', step: 'project', type: 'check', label: 'Optimize the autoloader (--optimize-autoloader)', hint: 'Recommended for production.', disabledWhen: (v) => !v.autorun_composer_install, on: 'Yes', offLabel: 'No' }
      ],
      run: [R.pre, R.php, { label: 'Setting up project files', triggers: ['Downloading and extracting', 'Creating Composer project'] }, { label: 'Running Composer', triggers: ['Running composer install', 'Saving application settings'] }, { label: 'Saving to Site Manager', triggers: ['Installation finished', 'application setup completed'] }],
      creds: [],
      log: (vm) => ['Checking if existing installation processes are running..', 'Validating provided data', 'Starting PHP container: php-fpm-' + vm.dom().PHPVersion].concat(vm.v.initial_project ? [(/^https:/.test(vm.v.initial_project) ? 'Downloading and extracting ' : 'Creating Composer project ') + vm.v.initial_project] : []).concat(vm.v.autorun_composer_install ? ['Running composer install'] : []).concat(['Saving application settings', 'Installation finished, adding the new application to SiteManager', 'New PHP application setup completed!'])
    },
    websitebuilder: {
      name: 'Website Builder', desc: 'Build a new HTML website with a drag-and-drop builder.', manager: 'site-builder.html', managerLabel: 'Open Website Builder', cta: 'Create website', success: ['Website creation completed'],
      note: 'Starts with an empty index.html and style.css, then opens in the builder.',
      steps: [WHERE, REVIEW], creds: [],
      run: [{ label: 'Creating files', triggers: ['Creating files'] }, { label: 'Setting file permissions', triggers: ['Setting files permissions'] }, { label: 'Setting file ownership', triggers: ['Setting files owner'] }, { label: 'Saving to Site Manager', triggers: ['Saving website information'] }],
      log: (vm) => ['Creating files in ' + vm.installPath(), 'Setting files permissions', "Setting files owner to 'testinguser'", 'Saving website information', 'Website creation completed']
    }
  };

  const ALL = Object.assign({}, CMS, RT, OTHER);
  const LOGO = { websitebuilder: 'websitebuilder', php: 'php' };

  window.opIWProtoConfig = function () {
    const q = new URLSearchParams(location.search);
    const id = ALL[q.get('app')] ? q.get('app') : 'wordpress';
    const c = ALL[id];
    const fail = q.has('fail');
    return Object.assign({}, c, {
      slug: id, title: c.title || 'Install ' + c.name, logo: 'img/apps/' + (LOGO[id] || id) + '.svg', domains: q.has('empty') ? [] : DOMAINS,
      endpoint: '/' + id + '/install', strength: 50, mysqlPrefix: '', fields: (c.fields || []).map((f) => Object.assign({ step: 'site' }, f)),
      // stands in for the NDJSON stream the real page reads
      transport: function (endpoint, fd, onEvent, vm) {
        const lines = c.log(vm);
        const failAt = Math.min(lines.length - 2, Math.max(3, Math.floor(lines.length * 0.6)));
        return new Promise((done) => {
          let i = 0;
          (function tick() {
            if (vm.phase !== 'running') return done();
            if (fail && i === failAt) {
              onEvent({ error: id === 'wordpress' || c.fields.some((f) => f.name === 'db_name') ? 'You have reached the maximum number of databases allowed on your plan. Upgrade your plan to get more.' : 'Directory ' + vm.installPath() + ' already exists and is not empty.' });
              return done();
            }
            if (i >= lines.length) return done();
            const ev = { status: lines[i] };
            if (i === lines.length - 1 && id === 'dokuwiki') { ev.admin_user = vm.v.admin_user || 'admin'; ev.admin_password = vm.v.admin_password || 'Gq7tR2mXw9LpZk4e'; }
            onEvent(ev);
            i++;
            setTimeout(tick, 350 + Math.random() * 450);
          })();
        });
      }
    });
  };
  window.opIWProtoApps = Object.keys(ALL);
})();
