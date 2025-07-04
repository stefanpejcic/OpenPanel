<?php
$fileToken = "IZs2cM1dmE2RluSrnUYH84kKBVjuhw";

require_once '/etc/phpmyadmin/config.secret.inc.php';
require_once '/etc/phpmyadmin/helpers.php';

/* Ensure we got the environment */
$vars = [
    'PMA_HOST',
    'MYSQL_ROOT_PASSWORD'
];

foreach ($vars as $var) {
    $env = getenv($var);
    if (!isset($_ENV[$var]) && $env !== false) {
        $_ENV[$var] = $env;
    }
}

// Retrieve the token from the URL
$providedToken = isset($_GET['token']) ? $_GET['token'] : '';

if ($providedToken === $fileToken) {
    // Provided token and file token match, create a session
    session_set_cookie_params(0, '/', '', 0);
    session_name('OPENPANEL_PHPMYADMIN');
    session_start();

    if (isset($_ENV['MYSQL_ROOT_PASSWORD'])) {
        $_SESSION['PMA_single_signon_user'] = 'root';
        $_SESSION['PMA_single_signon_password'] = isset($_ENV['MYSQL_ROOT_PASSWORD']) ? $_ENV['MYSQL_ROOT_PASSWORD'] : '';
        $_SESSION['PMA_single_signon_host'] = isset($_ENV['PMA_HOST']) ? $_ENV['PMA_HOST'] : 'mysql';
    } else {
        #echo "No root password.";
        header("Location: ./index.php?invalid");
    }

    session_write_close();

    // Remove the 'token' parameter from the query parameters
    unset($_GET['token']);
    $query_params_str = http_build_query($_GET);

    // Append query parameters to the Location header
    header("Location: ./index.php?$query_params_str");
} else {
    // for dev
    #echo "Invalid token: $providedToken VS $fileToken";
    // for prod
    header("Location: ./index.php?loginform=true");
}
?>
