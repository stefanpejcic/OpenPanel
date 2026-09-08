package opencart

// openpanelLoginFileName is the login helper's filename inside the docroot, deployed at install time (install.go) and read by handleOpenCartLogin (cli.go)
const openpanelLoginFileName = "openpanel-login.php"

// openpanelLoginPHP mirrors joomla/login_php.go's approach: booting admin/config.php + startup.php + framework.php with the route forced to "common/login" gives a working $registry with db/session wired up, since OpenCart's Session only persists via a real request/response cycle; the redirect builds its own absolute URL from $_SERVER instead of OpenCart's url library, same subdirectory-safety reason as joomla/login_php.go; the session cookie is re-issued with SameSite=Lax before redirecting since OpenCart defaults to Strict, which browsers silently drop on the cross-origin window.open() the "Login as Admin" button uses (curl doesn't enforce SameSite so this was missed in curl testing)
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
