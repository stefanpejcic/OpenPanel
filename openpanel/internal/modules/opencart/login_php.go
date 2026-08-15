package opencart

// openpanelLoginFileName is the login helper's filename inside the
// OpenCart docroot - deployed once at install time (install.go) and read
// by handleOpenCartLogin (cli.go) to build the link the browser opens.
const openpanelLoginFileName = "openpanel-login.php"

// openpanelLoginPHP mirrors joomla/login_php.go's approach: OpenCart core
// ships no CLI equivalent of a one-time login link, so this happens
// through an actual HTTP request instead (OpenCart's Session library only
// persists via a real request/response cycle - confirmed live: booting
// admin/config.php + system/startup.php + system/framework.php with the
// route forced to the harmless "common/login" and output buffered away
// gives a fully working $registry with 'db' and 'session' already wired
// up, without needing to touch OpenCart's router/dispatch logic at all).
//
// The redirect deliberately builds its own absolute URL from $_SERVER
// rather than relying on OpenCart's own url library with a relative path,
// for the same subdirectory-safety reason documented in
// joomla/login_php.go.
//
// The session cookie is re-issued explicitly with SameSite=Lax right
// before redirecting - confirmed live: OpenCart's installer defaults
// config_session_samesite to "Strict", and framework.php's own
// startup/session bootstrap sets the cookie with that flag. The "Login as
// Admin" button opens this script via window.open() from the OpenPanel
// dashboard, a different origin - browsers correctly treat that as a
// cross-site-initiated navigation and silently drop a Strict cookie on it
// (curl has no SameSite enforcement at all, which is why testing via curl
// missed this entirely: it worked every time there, but real browsers,
// incognito included, landed back on the login form with OpenCart's own
// "Invalid token session" warning). Lax still permits the cookie on this
// kind of top-level GET navigation, which is exactly the case here.
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

define('APPLICATION', 'Admin');
require_once(__DIR__ . '/admin/config.php');
require_once(DIR_SYSTEM . 'startup.php');

$_GET['route'] = 'common/login';
$_SERVER['REQUEST_METHOD'] = 'GET';

ob_start();
require_once(DIR_SYSTEM . 'framework.php');
ob_end_clean();

$db = $registry->get('db');
$session = $registry->get('session');

$bt = chr(96);

$query = $db->query("SELECT user_id, expires FROM " . $bt . DB_PREFIX . "openpanel_login_tokens" . $bt . " WHERE token_hash = '" . $db->escape($tokenHash) . "'");
$db->query("DELETE FROM " . $bt . DB_PREFIX . "openpanel_login_tokens" . $bt . " WHERE token_hash = '" . $db->escape($tokenHash) . "'");

if (!$query->num_rows || (int) $query->row['expires'] < time()) {
    http_response_code(403);
    die('This login link is invalid or has expired.');
}

$userId = (int) $query->row['user_id'];
$userCheck = $db->query("SELECT user_id FROM " . $bt . DB_PREFIX . "user" . $bt . " WHERE user_id = '" . $userId . "' AND status = '1'");

if (!$userCheck->num_rows) {
    http_response_code(403);
    die('This login link is invalid.');
}

$session->data['user_id'] = $userId;
$session->data['user_token'] = oc_token(32);
$session->close();

$config = $registry->get('config');
$option = [
    'expires'  => 0,
    'path'     => $config->get('session_path'),
    'domain'   => $config->get('session_domain'),
    'secure'   => !empty($_SERVER['HTTPS']) && $_SERVER['HTTPS'] !== 'off',
    'httponly' => true,
    'samesite' => 'Lax',
];
setcookie($config->get('session_name'), $session->getId(), $option);

$base = dirname($_SERVER['SCRIPT_NAME']);
if ($base === '/' || $base === '\\') {
    $base = '';
}
$scheme = (!empty($_SERVER['HTTPS']) && $_SERVER['HTTPS'] !== 'off') ? 'https' : 'http';
header('Location: ' . $scheme . '://' . $_SERVER['HTTP_HOST'] . $base . '/admin/index.php?route=common/dashboard&user_token=' . $session->data['user_token'], true, 303);
`
