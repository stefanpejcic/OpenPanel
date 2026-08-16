package matomo

// openpanelLoginFileName is the login helper's filename inside the Matomo
// docroot - deployed once at install time (install.go) and used by
// handleMatomoLogin (cli.go) to build the link the browser opens.
const openpanelLoginFileName = "openpanel-login.php"

// buildOpenpanelLoginPHP renders the login helper script with the admin
// login/password/secret token baked in as literals.
//
// Matomo ships no CLI/DB-level one-time-login primitive the way Drush's
// `user:login` or a plain token table (as used by joomla/opencart/
// nextcloud/prestashop) can drive - its Auth flow expects a real
// login+password (bcrypt-verified) or DI-container access to internals that
// aren't stable to hand-roll across versions. Rather than reimplement
// Matomo's internal session bootstrap, this script instead replays the
// exact same public login form flow (nonce fetch, then POST to
// module=Login) confirmed live against a real 5.12.0 install - but does it
// server-side via a loopback cURL call, then re-issues whatever session
// cookie that exchange produced as a genuine Set-Cookie on its own
// response, so the actual browser ends up holding a real Matomo session
// for this origin. Protected by a random 64-hex-char token baked in at
// install time (regenerated whenever the file is redeployed) - same trust
// boundary as any other CMS module's config file already sitting in the
// docroot with a plaintext DB password.
//
// The loopback request targets http://<webServerContainer>/... with an
// explicit Host header, NOT the site's own public HTTPS URL - confirmed
// live that php-fpm containers here have no outbound route back to the
// server's own public IP/port 443 (connection refused, presumably a
// firewall/network-namespace restriction), but a plain HTTP request to the
// webserver container's own name (e.g. "apache") on the shared podman
// network reaches it directly and is what Caddy/Apache/nginx itself would
// see for that Host regardless.
func buildOpenpanelLoginPHP(login, password, token, webServerContainer string) string {
	return `<?php
/**
 * OpenPanel one-time admin login handler.
 * Managed by OpenPanel - do not edit, will be overwritten on next install.
 */
error_reporting(E_ALL);
ini_set('display_errors', '0');

$expectedToken = ` + phpStringLiteral(token) + `;
$adminLogin = ` + phpStringLiteral(login) + `;
$adminPassword = ` + phpStringLiteral(password) + `;
$webServerContainer = ` + phpStringLiteral(webServerContainer) + `;

$token = isset($_GET['op_login']) ? $_GET['op_login'] : '';
if ($token === '' || !hash_equals($expectedToken, $token)) {
    http_response_code(403);
    die('Invalid login token.');
}

$scheme = (!empty($_SERVER['HTTPS']) && $_SERVER['HTTPS'] !== 'off') ? 'https' : 'http';
$hostHeader = $_SERVER['HTTP_HOST'];
$path = dirname($_SERVER['SCRIPT_NAME']);
$loopbackBase = rtrim('http://' . $webServerContainer . $path, '/');

$cookieFile = tempnam(sys_get_temp_dir(), 'mtlogin');

function op_matomo_curl($url, $hostHeader, $post, $cookieFile) {
    $ch = curl_init($url);
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
    curl_setopt($ch, CURLOPT_COOKIEJAR, $cookieFile);
    curl_setopt($ch, CURLOPT_COOKIEFILE, $cookieFile);
    curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, false);
    curl_setopt($ch, CURLOPT_SSL_VERIFYHOST, false);
    curl_setopt($ch, CURLOPT_FOLLOWLOCATION, true);
    curl_setopt($ch, CURLOPT_HTTPHEADER, array('Host: ' . $hostHeader));
    if ($post !== null) {
        curl_setopt($ch, CURLOPT_POST, true);
        curl_setopt($ch, CURLOPT_POSTFIELDS, http_build_query($post));
    }
    return curl_exec($ch);
}

$loginPage = op_matomo_curl($loopbackBase . '/index.php', $hostHeader, null, $cookieFile);
if (!preg_match('/id="login_form_nonce"\s+value="([a-f0-9]+)"/', (string) $loginPage, $m)) {
    @unlink($cookieFile);
    http_response_code(500);
    die('Could not read the Matomo login nonce.');
}
$nonce = $m[1];

op_matomo_curl($loopbackBase . '/index.php?module=Login', $hostHeader, array(
    'form_login'    => $adminLogin,
    'form_password' => $adminPassword,
    'form_nonce'    => $nonce,
    'form_redirect' => '',
), $cookieFile);

// The loopback requests above happened entirely server-side against a
// throwaway cookie jar file, so whatever Matomo session cookie they
// produced has to be re-issued as a genuine Set-Cookie on THIS response
// for the real browser to actually pick it up.
$cookieJarContents = @file_get_contents($cookieFile);
@unlink($cookieFile);
$sessionCookieSet = false;
if ($cookieJarContents) {
    foreach (explode("\n", $cookieJarContents) as $line) {
        $line = trim($line);
        if ($line === '') {
            continue;
        }
        // curl prefixes HttpOnly cookie lines with "#HttpOnly_" rather than
        // a true comment marker - strip that prefix before the plain "#"
        // comment-line check, or every HttpOnly session cookie (which is
        // exactly what Matomo's login sets) gets skipped as a comment.
        if (strpos($line, '#HttpOnly_') === 0) {
            $line = substr($line, strlen('#HttpOnly_'));
        } elseif ($line[0] === '#') {
            continue;
        }
        $parts = preg_split('/\t/', $line);
        if (count($parts) < 7) {
            continue;
        }
        $cookiePath = $parts[2];
        $cookieName = $parts[5];
        $cookieValue = $parts[6];
        if (
            stripos($cookieName, 'MATOMO_SESSID') !== false ||
            stripos($cookieName, 'PIWIK_SESSID') !== false ||
            stripos($cookieName, 'PHPSESSID') !== false
        ) {
            setcookie($cookieName, $cookieValue, 0, $cookiePath !== '' ? $cookiePath : '/', '', $scheme === 'https', true);
            $sessionCookieSet = true;
        }
    }
}
if (!$sessionCookieSet) {
    http_response_code(500);
    die('Matomo login did not return a session cookie.');
}

$publicBase = rtrim($scheme . '://' . $hostHeader . $path, '/');
header('Location: ' . $publicBase . '/index.php', true, 303);
`
}

// phpStringLiteral renders s as a single-quoted PHP string literal, escaping
// backslashes and single quotes (the only two characters that matter inside
// PHP single-quoted strings).
func phpStringLiteral(s string) string {
	escaped := ""
	for _, c := range s {
		if c == '\\' || c == '\'' {
			escaped += `\`
		}
		escaped += string(c)
	}
	return "'" + escaped + "'"
}
