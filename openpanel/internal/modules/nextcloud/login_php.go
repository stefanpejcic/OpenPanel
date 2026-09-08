package nextcloud

// openpanelLoginFileName is the login helper's filename inside the docroot, deployed at install time (install.go) and read by handleNextcloudLogin (cli.go)
const openpanelLoginFileName = "openpanel-login.php"

// openpanelLoginPHP mirrors joomla/opencart's one-time-token approach, but Nextcloud actually has a real public login API for it: \OCP\IUserSession::completeLogin(), the same call its SSO/SAML apps use; completeLogin()'s $loginDetails must use an empty string for 'password', never null, since PostLoginEvent's constructor requires a non-nullable string and a null throws mid-login, leaving the session half-initialized; $_COOKIE is cleared before requiring lib/base.php so Nextcloud's own CSRF cookie check can't 412 the request, same bug family as the SameSite issue fixed in OpenCart's login helper; the redirect builds its own absolute URL from $_SERVER instead of Nextcloud's URL generator, same subdirectory-safety reason as joomla/opencart's login_php.go
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

// See the comment in nextcloud/login_php.go for why this is cleared.
$_COOKIE = [];

define('OC_CONSOLE', false);
require_once __DIR__ . '/lib/base.php';

// QueryBuilder::from()/delete() auto-prepend the real configured table
// prefix (dbtableprefix) via prefixTableName() - do not prepend it here too.
$db = \OC::$server->get(\OCP\IDBConnection::class);

$qb = $db->getQueryBuilder();
$qb->select('user_id', 'expires')
    ->from('openpanel_login_tokens')
    ->where($qb->expr()->eq('token_hash', $qb->createNamedParameter($tokenHash)));
$result = $qb->executeQuery();
$row = $result->fetch();
$result->closeCursor();

$del = $db->getQueryBuilder();
$del->delete('openpanel_login_tokens')
    ->where($del->expr()->eq('token_hash', $del->createNamedParameter($tokenHash)));
$del->executeStatement();

if (!$row || (int) $row['expires'] < time()) {
    http_response_code(403);
    die('This login link is invalid or has expired.');
}

$userId = $row['user_id'];

$userManager = \OC::$server->get(\OCP\IUserManager::class);
$userSession = \OC::$server->get(\OCP\IUserSession::class);

$user = $userManager->get($userId);
if (!$user || !$user->isEnabled()) {
    http_response_code(403);
    die('This login link is invalid.');
}

try {
    $userSession->completeLogin($user, ['loginName' => $userId, 'password' => ''], true);
} catch (\Throwable $e) {
    http_response_code(500);
    die('Unable to complete login.');
}

if (!$userSession->isLoggedIn()) {
    http_response_code(500);
    die('Unable to complete login.');
}

// completeLogin() alone is NOT enough for the login to survive past this
// request - confirmed live. \OC\User\Session::getUser() calls
// validateSession() the first time it's asked for the user on a *later*
// request, which looks up the current session ID in the auth token table
// (oc_authtoken) and calls logout() if no matching row exists. A real
// password login creates that row via createSessionToken() right after
// completeLogin() succeeds; without it, the very next request silently
// logs itself back out. createSessionToken() isn't on the public
// \OCP\IUserSession interface, but the object IUserSession::class resolves
// to is the concrete \OC\User\Session which has it - fetching it via
// \OC\User\Session::class instead of the OCP interface makes the intent
// explicit here.
$concreteSession = \OC::$server->get(\OC\User\Session::class);
$request = \OC::$server->get(\OCP\IRequest::class);
$concreteSession->createSessionToken($request, $userId, $userId);

// Explicit close, mirroring the fix needed in opencart/login_php.go:
// completeLogin()'s regenerateId() reopens the session but leaves it open
// (unlike Session::set(), which reopens-then-closes around itself), so
// nothing forces the write before this script's own natural end. PHP's
// implicit session_write_close() at shutdown should cover it regardless,
// but an explicit close removes any doubt that the login state survives
// into the next request.
$session = \OC::$server->get(\OCP\ISession::class);
$session->close();

$base = dirname($_SERVER['SCRIPT_NAME']);
if ($base === '/' || $base === '\\') {
    $base = '';
}
$scheme = (!empty($_SERVER['HTTPS']) && $_SERVER['HTTPS'] !== 'off') ? 'https' : 'http';
header('Location: ' . $scheme . '://' . $_SERVER['HTTP_HOST'] . $base . '/', true, 303);
`
