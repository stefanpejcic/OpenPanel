package joomla

// openpanelLoginFileName is the login helper's filename inside the
// Joomla docroot - deployed once at install time (install.go) and read by
// handleJoomlaLogin (cli.go) to build the link the browser opens.
const openpanelLoginFileName = "openpanel-login.php"

// openpanelLoginPHP mirrors wordpress/wpcli.go's mu-plugin technique:
// Joomla core ships no CLI equivalent of `wp login create`/drush's
// `user:login`, so a one-time login has to happen through an actual HTTP
// request (Joomla's Session/User APIs only work once the full
// AdministratorApplication is bootstrapped, which needs a real
// request context - confirmed live, see install.go/cli.go comments).
//
// The redirect deliberately builds its own absolute URL from
// $_SERVER rather than using $app->redirect() with a relative path:
// $app->redirect() resolves relative paths through Joomla's own base-URI
// detection, which - live-tested - silently drops the docroot's
// subdirectory prefix for sites not installed at the domain root.
const openpanelLoginPHP = `<?php
/**
 * OpenPanel one-time admin login handler.
 * Managed by OpenPanel - do not edit, will be overwritten on next install.
 */
define('_JEXEC', 1);
define('JPATH_BASE', __DIR__ . '/administrator');
require_once JPATH_BASE . '/includes/defines.php';
require_once JPATH_BASE . '/includes/framework.php';

use Joomla\CMS\Factory;
use Joomla\CMS\User\User;
use Joomla\CMS\Application\AdministratorApplication;

$token = isset($_GET['op_login']) ? preg_replace('/[^a-zA-Z0-9]/', '', $_GET['op_login']) : '';
if ($token === '') {
    http_response_code(400);
    die('Invalid login token.');
}
$tokenHash = hash('sha256', $token);

$container = Factory::getContainer();
$container->alias('session.web', 'session.web.administrator')
    ->alias('session', 'session.web.administrator')
    ->alias('JSession', 'session.web.administrator')
    ->alias(\Joomla\CMS\Session\Session::class, 'session.web.administrator')
    ->alias(\Joomla\Session\Session::class, 'session.web.administrator')
    ->alias(\Joomla\Session\SessionInterface::class, 'session.web.administrator');

$app = $container->get(AdministratorApplication::class);
Factory::$application = $app;

$db = Factory::getDbo();

$query = $db->getQuery(true)
    ->select(['user_id', 'expires'])
    ->from($db->quoteName('#__openpanel_login_tokens'))
    ->where($db->quoteName('token_hash') . ' = ' . $db->quote($tokenHash));
$db->setQuery($query);
$row = $db->loadObject();

$delQuery = $db->getQuery(true)
    ->delete($db->quoteName('#__openpanel_login_tokens'))
    ->where($db->quoteName('token_hash') . ' = ' . $db->quote($tokenHash));
$db->setQuery($delQuery)->execute();

if (!$row || (int) $row->expires < time()) {
    http_response_code(403);
    die('This login link is invalid or has expired.');
}

$user = User::getInstance((int) $row->user_id);
if (!$user || (int) $user->id === 0 || $user->block) {
    http_response_code(403);
    die('This login link is invalid.');
}

$session = $app->getSession();
$session->set('user', $user);
$app->checkSession();

$base = dirname($_SERVER['SCRIPT_NAME']);
if ($base === '/' || $base === '\\') {
    $base = '';
}
$scheme = (!empty($_SERVER['HTTPS']) && $_SERVER['HTTPS'] !== 'off') ? 'https' : 'http';
header('Location: ' . $scheme . '://' . $_SERVER['HTTP_HOST'] . $base . '/administrator/index.php', true, 303);
exit;
`
