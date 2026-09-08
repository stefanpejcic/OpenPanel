package prestashop

// openpanelLoginFileName is the login helper's filename inside PrestaShop's randomly-named admin directory (see findAdminDir), deployed at install time (install.go) and read by handlePrestashopLogin (cli.go)
const openpanelLoginFileName = "openpanel-login.php"

// openpanelLoginPHP mirrors joomla/opencart/nextcloud's one-time-token approach - PrestaShop's employee auth is a self-contained encrypted cookie (classes/Cookie.php) rather than a server-side session, so this replicates AdminLoginController::processLogin()'s post-password-check sequence: load the Employee, populate the cookie's fields, registerSession(), then write() - it never touches password verification, just binds an already-known-good employee id to a session
// remote_addr must be built from Tools::getRemoteAddr(), not $_SERVER['REMOTE_ADDR'] directly, since Employee::isLoggedBack() compares against ip2long(Tools::getRemoteAddr()) which prefers X-Forwarded-For behind a proxy like Caddy - raw REMOTE_ADDR silently fails isLoggedBack() on later requests
// the redirect builds its own absolute URL from $_SERVER instead of Context::getContext()->link->getAdminLink(), which doubles the subdirectory on a subdirectory install (404) - the admin token is still generated for real via Tools::getAdminTokenLite('AdminDashboard'), which only needs $context->employee->id
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
