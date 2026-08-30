package ojs

// openpanelLoginFileName is the login helper's filename inside the OJS
// approot - deployed once at install time (install.go) and read by
// handleOJSLogin (login.go) to build the link the browser opens.
const openpanelLoginFileName = "openpanel-login.php"

// openpanelLoginPHP mirrors joomla/login_php.go's technique: OJS core (like
// Joomla) ships no CLI equivalent of Drupal's `drush user:login`/WP-CLI's
// `wp login create`, so a one-time login has to happen through an actual
// bootstrapped OJS application object. Unlike Joomla though, OJS's own
// index.php bootstrap ('./lib/pkp/includes/bootstrap.php', which
// constructs \APP\core\Application() - see install.go's fixUpOJSConfig
// comment and the package doc comment for how that was confirmed by
// reading PKPApplication's constructor) initializes everything a normal
// request needs (Laravel container, session guard, DB) on its own, with no
// separate "administrator" sub-application the way Joomla has - so this is
// simpler than Joomla's version: no Session/User class aliasing dance, just
// Application::get()->getRequest() and PKP\security\Validation's own
// registerUserSession() helper, which does exactly what a normal
// Validation::login() call does short of checking the password.
const openpanelLoginPHP = `<?php
/**
 * OpenPanel one-time admin login handler.
 * Managed by OpenPanel - do not edit, will be overwritten on next install.
 */
define('INDEX_FILE_LOCATION', __DIR__ . '/index.php');
require_once './lib/pkp/includes/bootstrap.php';

use APP\core\Application;
use APP\facades\Repo;
use PKP\security\Validation;

$token = isset($_GET['op_login']) ? preg_replace('/[^a-zA-Z0-9]/', '', $_GET['op_login']) : '';
if ($token === '') {
    http_response_code(400);
    die('Invalid login token.');
}
$tokenHash = hash('sha256', $token);

$conn = \Illuminate\Support\Facades\DB::connection();

$row = $conn->selectOne(
    'SELECT user_id, expires FROM openpanel_login_tokens WHERE token_hash = ?',
    [$tokenHash]
);
$conn->delete('DELETE FROM openpanel_login_tokens WHERE token_hash = ?', [$tokenHash]);

if (!$row || (int) $row->expires < time()) {
    http_response_code(403);
    die('This login link is invalid or has expired.');
}

$user = Repo::user()->get((int) $row->user_id, true);
if (!$user || $user->getDisabled()) {
    http_response_code(403);
    die('This login link is invalid.');
}

$request = Application::get()->getRequest();
$reason = null;
if (!Validation::registerUserSession($user, $reason)) {
    http_response_code(403);
    die('Unable to start session for this account.');
}

$base = dirname($_SERVER['SCRIPT_NAME']);
if ($base === '/' || $base === '\\') {
    $base = '';
}
$scheme = (!empty($_SERVER['HTTPS']) && $_SERVER['HTTPS'] !== 'off') ? 'https' : 'http';

// A raw header('Location: ...') here (as this file used to do) skips the
// session cookie entirely: PKPSessionGuard::updateSession() only queues
// the encrypted OJSSID cookie onto Laravel's internal Response singleton
// - it's never actually flushed via a real header() call unless something
// explicitly triggers that. A normal request gets this via PKPRouter's own
// dispatch cycle; a bare script like this one bypasses that entirely, so
// the browser was never receiving a cookie tied to the session that was
// actually just saved to the sessions table (confirmed live: every
// subsequent request silently started a brand new session instead).
// PKPRequest::redirectUrl() is PKP's own helper for exactly this - it
// calls $request->getSessionGuard()->sendCookies() immediately before the
// Location header, so use it instead of hand-rolling the redirect.
$request->redirectUrl($scheme . '://' . $_SERVER['HTTP_HOST'] . $base . '/index.php/index');
`
