// shared data and helpers for the database create pages; ?engine=mysql|postgresql|mongodb picks the engine
(function () {
  const q = new URLSearchParams(location.search);
  const sel = q.get('engine');
  const mysqlSnips = [
    { id: 'wordpress', label: 'WordPress', file: 'wp-config-snippet.php', code: "define( 'DB_NAME',     '${db}' );\ndefine( 'DB_USER',     '${user}' );\ndefine( 'DB_PASSWORD', '${pass}' );\ndefine( 'DB_HOST',     'localhost' );\ndefine( 'DB_CHARSET',  'utf8mb4' );\ndefine( 'DB_COLLATE',  '' );" },
    { id: 'laravel', label: 'Laravel', file: '.env', code: 'DB_CONNECTION=mysql\nDB_HOST=localhost\nDB_PORT=3306\nDB_DATABASE=${db}\nDB_USERNAME=${user}\nDB_PASSWORD=${pass}' },
    { id: 'django', label: 'Django', file: 'settings.py', code: "DATABASES = {\n    'default': {\n        'ENGINE':   'django.db.backends.mysql',\n        'NAME':     '${db}',\n        'USER':     '${user}',\n        'PASSWORD': '${pass}',\n        'HOST':     'localhost',\n        'PORT':     '3306',\n    }\n}" },
    { id: 'nodejs', label: 'Node.js', file: 'db.js', code: "import mysql from 'mysql2/promise';\n\nconst pool = mysql.createPool({\n  host:     'localhost',\n  user:     '${user}',\n  password: '${pass}',\n  database: '${db}',\n  waitForConnections: true,\n  connectionLimit:    10,\n});" },
  ];
  const pgSnips = [
    { id: 'uri', label: 'Connection URI', file: '.env', code: 'DATABASE_URL=postgresql://${euser}:${epass}@postgres:5432/${db}' },
    { id: 'laravel', label: 'Laravel', file: '.env', code: 'DB_CONNECTION=pgsql\nDB_HOST=postgres\nDB_PORT=5432\nDB_DATABASE=${db}\nDB_USERNAME=${user}\nDB_PASSWORD=${pass}' },
    { id: 'django', label: 'Django', file: 'settings.py', code: "DATABASES = {\n    'default': {\n        'ENGINE':   'django.db.backends.postgresql',\n        'NAME':     '${db}',\n        'USER':     '${user}',\n        'PASSWORD': '${pass}',\n        'HOST':     'postgres',\n        'PORT':     '5432',\n    }\n}" },
    { id: 'nodejs', label: 'Node.js', file: 'db.js', code: "import pg from 'pg';\n\nconst pool = new pg.Pool({\n  host:     'postgres',\n  port:     5432,\n  user:     '${user}',\n  password: '${pass}',\n  database: '${db}',\n});" },
    { id: 'phppdo', label: 'PHP PDO', file: 'db-connect.php', code: "<?php\n\n$dsn  = 'pgsql:host=postgres;port=5432;dbname=${db}';\n$pdo  = new PDO($dsn, '${user}', '${pass}', [\n    PDO::ATTR_ERRMODE => PDO::ERRMODE_EXCEPTION,\n]);" },
  ];
  const mongoSnips = [
    { id: 'uri', label: 'Connection URI', file: '.env', code: 'MONGODB_URI=mongodb://${euser}:${epass}@mongodb:27017/${db}?authSource=admin' },
    { id: 'nodejs', label: 'Node.js', file: 'db.js', code: "import { MongoClient } from 'mongodb';\n\nconst client = new MongoClient('mongodb://${euser}:${epass}@mongodb:27017/${db}?authSource=admin');\nawait client.connect();\nconst db = client.db('${db}');" },
    { id: 'python', label: 'Python', file: 'db.py', code: "from pymongo import MongoClient\n\nclient = MongoClient('mongodb://${euser}:${epass}@mongodb:27017/${db}?authSource=admin')\ndb = client['${db}']" },
    { id: 'php', label: 'PHP', file: 'db-connect.php', code: "<?php\n\nrequire 'vendor/autoload.php';\n\n$client = new MongoDB\\Client('mongodb://${euser}:${epass}@mongodb:27017/${db}?authSource=admin');\n$db = $client->selectDatabase('${db}');" },
    { id: 'laravel', label: 'Laravel', file: '.env', code: 'DB_CONNECTION=mongodb\nDB_URI=mongodb://${euser}:${epass}@mongodb:27017/${db}?authSource=admin\nDB_DATABASE=${db}' },
  ];
  const engines = {
    mysql: {
      id: 'mysql', name: 'MySQL', server: 'MariaDB', host: 'localhost', port: '3306', remotePort: '32770', prefix: '',
      list: 'mysql.html', users: 'mysql-users.html', wizard: 'mysql-wizard.html', import: 'mysql-import.html', remote: 'mysql-remote.html',
      tabs: [['Databases', 'mysql.html'], ['Users', 'mysql-users.html'], ['Import', 'mysql-import.html'], ['phpMyAdmin', 'mysql-phpmyadmin.html'], ['Remote Access', 'mysql-remote.html'], ['Running Queries', 'mysql-processes.html'], ['Configuration', 'mysql-config.html'], ['Root Password', 'mysql-root.html'], ['Server Type', 'mysql-server-type.html']],
      dbs: ['testinguser_bkfix', 'testinguser_wp1', 'joomla_fec8yg', 'nextcloud_kvaurv', 'matomo_6suw3z'],
      usersList: ['9496x8vqef', 'app', 'testinguser_wp1', 'uj3sfgeldx', 'vjmzwuswr5'],
      assigned: [{ database: 'joomla_fec8yg', users: '9496x8vqef' }, { database: 'matomo_6suw3z', users: 'uj3sfgeldx' }, { database: 'nextcloud_kvaurv', users: 'vjmzwuswr5' }, { database: 'testinguser_bkfix', users: 'app' }, { database: 'testinguser_wp1', users: 'testinguser_wp1' }],
      snippets: mysqlSnips,
      reserved: ['information_schema', 'mysql', 'performance_schema', 'sys'],
      createDb: (d) => 'CREATE DATABASE `' + d + '`;',
      createUser: (u) => "CREATE USER '" + u + "'@'%' IDENTIFIED BY '********';",
      grant: (d, u, p) => 'GRANT ' + (p || 'ALL PRIVILEGES') + ' ON `' + d + "`.* TO '" + u + "'@'%';",
    },
    postgresql: {
      id: 'postgresql', name: 'PostgreSQL', server: 'PostgreSQL', host: 'postgres', port: '5432', prefix: '',
      list: 'postgresql.html', users: '#', wizard: 'db-wizard.html?engine=postgresql', import: '#', remote: '#',
      tabs: [['Databases', 'postgresql.html'], ['Users', '#'], ['Import', '#'], ['Remote Access', '#'], ['Running Queries', '#'], ['Configuration', '#']],
      dbs: [], usersList: [], assigned: [],
      snippets: pgSnips,
      reserved: ['postgres', 'template0', 'template1'],
      createDb: (d) => 'CREATE DATABASE "' + d + '";',
      createUser: (u) => 'CREATE USER "' + u + "\" WITH PASSWORD '********';",
      grant: (d, u) => 'GRANT ALL PRIVILEGES ON DATABASE "' + d + '" TO "' + u + '";\nGRANT USAGE, CREATE ON SCHEMA public TO "' + u + '";',
    },
    mongodb: {
      id: 'mongodb', name: 'MongoDB', server: 'MongoDB', host: 'mongodb', port: '27017', prefix: '',
      list: '#', users: '#', wizard: 'db-wizard.html?engine=mongodb', import: '#', remote: '#',
      tabs: [['Databases', '#'], ['Users', '#'], ['Import', '#'], ['Running Queries', '#']],
      dbs: [], usersList: [], assigned: [],
      snippets: mongoSnips,
      reserved: ['admin', 'local', 'config'],
      createDb: (d) => 'db.getSiblingDB("' + d + '").createCollection("_init")',
      createUser: (u) => 'db.getSiblingDB("admin").createUser({ user: "' + u + '", pwd: "********", roles: [] })',
      grant: (d, u, role) => 'db.getSiblingDB("admin").grantRolesToUser("' + u + '", [{ role: "' + (role || 'readWrite') + '", db: "' + d + '" }])',
    },
  };
  window.DBP = {
    engine: engines[sel] || engines[window.DBP_DEFAULT || 'mysql'],
    engines, limit: 40, plan: 'Developer plus', threshold: 50,
    roles: [
      { id: 'read', text: 'Read data only.' },
      { id: 'readWrite', text: 'Read and write data. What most apps need.' },
      { id: 'dbAdmin', text: 'Indexes, stats and schema work, no data access.' },
      { id: 'dbOwner', text: 'Full control of this database.' },
    ],
    privGroups: [
      { label: 'Data', privs: ['SELECT', 'INSERT', 'UPDATE', 'DELETE'] },
      { label: 'Structure', privs: ['CREATE', 'ALTER', 'DROP', 'INDEX', 'REFERENCES', 'CREATE VIEW', 'SHOW VIEW', 'CREATE TEMPORARY TABLES', 'LOCK TABLES'] },
      { label: 'Routines and events', privs: ['CREATE ROUTINE', 'ALTER ROUTINE', 'EXECUTE', 'EVENT', 'TRIGGER'] },
    ],
    gen(n) { const cs = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789'; let s = ''; for (const x of crypto.getRandomValues(new Uint32Array(n))) s += cs[x % cs.length]; return s; },
    score(v) { if (!v) return 0; let s = 0; if (v.length >= 8) s++; if (v.length >= 12) s++; if (/[a-z]/.test(v)) s++; if (/[A-Z]/.test(v)) s++; if (/\d/.test(v)) s++; if (/[^a-zA-Z0-9]/.test(v)) s++; return Math.round(s / 6 * 100); },
    level(v) { const s = this.score(v); return !v ? '' : s <= 33 ? 'Weak' : s <= 67 ? 'Medium' : 'Strong'; },
    valid(n, max) { return /^[a-zA-Z0-9_]+$/.test(n) && n.length <= max; },
    snippet(s, db, user, pass) { return s.code.replaceAll('${db}', db).replaceAll('${euser}', encodeURIComponent(user)).replaceAll('${epass}', encodeURIComponent(pass)).replaceAll('${user}', user).replaceAll('${pass}', pass); },
  };
  // page tabs, crumb and sidebar follow the engine
  document.addEventListener('DOMContentLoaded', function () {
    const e = window.DBP.engine;
    const tabs = document.getElementById('op-engine-tabs');
    if (tabs) {
      tabs.setAttribute('aria-label', e.name);
      tabs.innerHTML = e.tabs.map(([l, h]) => '<a href="' + h + '"' + (l === (tabs.dataset.tab || 'Databases') ? ' aria-current="page"' : '') + '>' + l + '</a>').join('');
    }
    document.querySelectorAll('[data-engine-name]').forEach((el) => { el.textContent = e.name; });
    document.querySelectorAll('[data-engine-list]').forEach((el) => { el.setAttribute('href', e.list); });
    const nav = document.querySelector('nav[aria-label="Databases"]');
    if (nav) nav.querySelectorAll('a').forEach((a) => { if (a.getAttribute('href') === e.list) a.setAttribute('aria-current', 'page'); else a.removeAttribute('aria-current'); });
  });
})();
