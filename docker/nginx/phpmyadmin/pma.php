<?php
$fileToken = "IZs2cM1dmE2RluSrnUYH84kKBVjuhw";

// Retrieve the token from the URL
$providedToken = isset($_GET['token']) ? $_GET['token'] : '';

if ($providedToken === $fileToken) {
    // Provided token and file token match, create a session
    session_set_cookie_params(0, '/', '', 0);
    session_name('OPENPANEL_PHPMYADMIN');
    session_start();

    $_SESSION['PMA_single_signon_user'] = 'phpmyadmin';
    $_SESSION['PMA_single_signon_password'] = 'cao1rsVFX0zPMq';
    $_SESSION['PMA_single_signon_host'] = 'localhost';

    session_write_close();

    // Remove the 'token' parameter from the query parameters
    unset($_GET['token']);

    // Construct the query parameters string
    $query_params_str = http_build_query($_GET);

    // Append query parameters to the Location header
    header("Location: ./index.php?$query_params_str");
} else {
    // Handle invalid token (e.g., show an error message)
    echo 'Invalid token: $providedToken';
}
?>
