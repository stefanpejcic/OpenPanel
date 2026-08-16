package mediawiki

// openpanelLoginFileName is the login helper's filename inside the
// MediaWiki docroot - deployed once at install time (install.go) and read
// by handleMediaWikiLogin (cli.go) to build the link the browser opens.
const openpanelLoginFileName = "openpanel-login.php"

// openpanelLoginPHP mirrors joomla/login_php.go's technique: MediaWiki core
// ships no CLI equivalent of Drupal's `drush user:login`, so a one-time
// login has to happen through an actual HTTP request bootstrapped the same
// way index.php bootstraps itself (via includes/WebStart.php), then binds
// a User to the current request's session via User::setCookies() -
// MediaWiki's own documented session-persistence API (confirmed against
// the 1.42 source: User::setCookies() pulls the session off
// $this->getRequest()->getSession() and calls $session->persist()).
const openpanelLoginPHP = `<?php
/**
 * OpenPanel one-time admin login handler.
 * Managed by OpenPanel - do not edit, will be overwritten on next install.
 */
define( 'MW_ENTRY_POINT', 'index' );
require_once dirname( __FILE__ ) . '/includes/PHPVersionCheck.php';
wfEntryPointCheck( 'html', dirname( $_SERVER['SCRIPT_NAME'] ) );
require __DIR__ . '/includes/WebStart.php';

use MediaWiki\MediaWikiServices;
use MediaWiki\User\User;

$token = isset( $_GET['op_login'] ) ? preg_replace( '/[^a-zA-Z0-9]/', '', $_GET['op_login'] ) : '';
if ( $token === '' ) {
	http_response_code( 400 );
	die( 'Invalid login token.' );
}
$tokenHash = hash( 'sha256', $token );

$dbw = MediaWikiServices::getInstance()->getConnectionProvider()->getPrimaryDatabase();

$row = $dbw->newSelectQueryBuilder()
	->select( [ 'user_id', 'expires' ] )
	->from( 'openpanel_login_tokens' )
	->where( [ 'token_hash' => $tokenHash ] )
	->caller( __METHOD__ )
	->fetchRow();

$dbw->newDeleteQueryBuilder()
	->deleteFrom( 'openpanel_login_tokens' )
	->where( [ 'token_hash' => $tokenHash ] )
	->caller( __METHOD__ )
	->execute();

if ( !$row || (int)$row->expires < time() ) {
	http_response_code( 403 );
	die( 'This login link is invalid or has expired.' );
}

$user = User::newFromId( (int)$row->user_id );
$user->load();
if ( !$user || $user->getId() === 0 || $user->getBlock() !== null ) {
	http_response_code( 403 );
	die( 'This login link is invalid.' );
}

$context = RequestContext::getMain();
$context->setUser( $user );
$user->setCookies( $context->getRequest(), null, true );

$scriptPath = dirname( $_SERVER['SCRIPT_NAME'] );
if ( $scriptPath === '/' || $scriptPath === '\\' ) {
	$scriptPath = '';
}
$scheme = ( !empty( $_SERVER['HTTPS'] ) && $_SERVER['HTTPS'] !== 'off' ) ? 'https' : 'http';
header( 'Location: ' . $scheme . '://' . $_SERVER['HTTP_HOST'] . $scriptPath . '/index.php', true, 303 );
exit;
`
