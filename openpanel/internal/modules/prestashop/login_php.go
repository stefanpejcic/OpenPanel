package prestashop

// openpanelLoginFileName is the login helper's filename inside PrestaShop's
// (randomly named, see prestashop.findAdminDir) admin directory - deployed
// once at install time (install.go) and read by handlePrestashopLogin
// (cli.go) to build the link the browser opens.
const openpanelLoginFileName = "openpanel-login.php"

// openpanelLoginPHP mirrors joomla/opencart/nextcloud's one-time-token
// approach: PrestaShop core ships no CLI equivalent of a one-time login
// link either. Its employee auth is a self-contained encrypted cookie (see
// classes/Cookie.php), not a server-side PHP session, and
// controllers/admin/AdminLoginController.php's processLogin() shows the
// exact legitimate post-password-check sequence to replicate: load the
// Employee, populate Context::getContext()->cookie's id_employee/email/
// profile/passwd/remote_addr fields, register an EmployeeSession row via
// $cookie->registerSession(), then $cookie->write(). This script does only
// that sequence - it never touches password hashing/verification at all,
// it just binds an already-known-good employee id (from our own one-time
// token) to a session the same way a successful password login would.
//
// Two real bugs were found and fixed by testing this live end-to-end
// (round-tripping the resulting cookie through a second request) rather
// than trusting that "the code compiles and doesn't error" meant it worked:
//
//  1. remote_addr MUST be built from Tools::getRemoteAddr(), not
//     $_SERVER['REMOTE_ADDR'] directly. Employee::isLoggedBack() compares
//     cookie->remote_addr against ip2long(Tools::getRemoteAddr()) on every
//     later request, and Tools::getRemoteAddr() prefers
//     X-Forwarded-For over REMOTE_ADDR whenever REMOTE_ADDR looks like a
//     private/proxy address (exactly OpenPanel's setup, behind Caddy) - so
//     storing raw REMOTE_ADDR here produced a value that could never match
//     what later requests compute, silently failing isLoggedBack() with no
//     error anywhere.
//  2. The redirect builds its own absolute URL from $_SERVER rather than
//     trusting PrestaShop's own Context::getContext()->link->getAdminLink()
//     - confirmed live: getAdminLink() DOUBLES the subdirectory for a
//     subdirectory install (produced
//     ".../psreal/psreal/admin.../index.php?...", a 404), the same class of
//     bug already found and worked around in joomla/opencart/nextcloud's
//     login helpers, just tripped by PrestaShop's Link class instead of a
//     redirect() call. The per-controller admin token PrestaShop's routing
//     requires is still generated the real way, via
//     Tools::getAdminTokenLite('AdminDashboard') - that only needs
//     $context->employee->id (already set above) and doesn't touch the
//     broken URL-building path at all.
const openpanelLoginPHP = `<?php
/**
 * OpenPanel one-time admin login handler.
 * Managed by OpenPanel - do not edit, will be overwritten on next install.
 */
error_reporting(E_ALL);
ini_set('display_errors', '0');

$token = isset($_GET['op_login']) ? preg_replace('/[^a-zA-Z0-9]/', '', $_GET['op_login']) : '';
if ($token === '') {
    http_response_code(400);
    die('Invalid login token.');
}
$tokenHash = hash('sha256', $token);

if (!defined('_PS_ADMIN_DIR_')) {
    define('_PS_ADMIN_DIR_', __DIR__);
}
require _PS_ADMIN_DIR_ . '/../config/config.inc.php';

$prefix = _DB_PREFIX_;
$bt = chr(96);

Db::getInstance()->execute(
    'CREATE TABLE IF NOT EXISTS ' . $bt . $prefix . 'openpanel_login_tokens' . $bt . ' (' .
    'token_hash CHAR(64) PRIMARY KEY, user_id INT UNSIGNED NOT NULL, expires INT UNSIGNED NOT NULL' .
    ') ENGINE=InnoDB'
);

$row = Db::getInstance()->getRow(
    'SELECT user_id, expires FROM ' . $bt . $prefix . 'openpanel_login_tokens' . $bt .
    " WHERE token_hash = '" . pSQL($tokenHash) . "'"
);
Db::getInstance()->execute(
    'DELETE FROM ' . $bt . $prefix . 'openpanel_login_tokens' . $bt .
    " WHERE token_hash = '" . pSQL($tokenHash) . "'"
);

if (!$row || (int) $row['expires'] < time()) {
    http_response_code(403);
    die('This login link is invalid or has expired.');
}

$employeeId = (int) $row['user_id'];
$employee = new Employee($employeeId);
if (!$employee->id || !$employee->active) {
    http_response_code(403);
    die('This login link is invalid.');
}

$context = Context::getContext();
$context->employee = $employee;

$cookie = $context->cookie;
$cookie->id_employee = $employee->id;
$cookie->email = $employee->email;
$cookie->profile = $employee->id_profile;
$cookie->passwd = $employee->passwd;
$cookie->remote_addr = (int) ip2long(Tools::getRemoteAddr());
$cookie->registerSession(new EmployeeSession());
$cookie->last_activity = time();
$cookie->write();

$adminToken = Tools::getAdminTokenLite('AdminDashboard', $context);

$base = dirname($_SERVER['SCRIPT_NAME']);
if ($base === '/' || $base === '\\') {
    $base = '';
}
$scheme = (!empty($_SERVER['HTTPS']) && $_SERVER['HTTPS'] !== 'off') ? 'https' : 'http';
$url = $scheme . '://' . $_SERVER['HTTP_HOST'] . $base . '/index.php?controller=AdminDashboard&token=' . $adminToken;
header('Location: ' . $url, true, 303);
`
