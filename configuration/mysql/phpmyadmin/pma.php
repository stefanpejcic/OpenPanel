<?php

require_once '/etc/phpmyadmin/config.secret.inc.php';

$providedToken = $_GET['token'] ?? '';
$username      = $_GET['user']  ?? '';

if (empty($username) || empty($providedToken)) {
    header("Location: ./index.php");
    exit;
}

if (!preg_match('/^[a-zA-Z0-9_-]+$/', $username)) {
    header("Location: ./index.php?invalid");
    exit;
}

$tokenFile = '/home/' . $username . '/pma.token';
if (!file_exists($tokenFile)) {
    header("Location: ./index.php?invalid");
    exit;
}

$fileToken = trim(file_get_contents($tokenFile));
if (strlen($fileToken) < 32 || !hash_equals($fileToken, $providedToken)) {
    header("Location: ./index.php?invalid");
    exit;
}

$socketPath = '/home/' . $username . '/sockets/mysqld/mysqld.sock';
if (!file_exists($socketPath)) {
    header("Location: ./index.php?invalid");
    exit;
}

$envFile = '/home/' . $username . '/.env';
$envVars = parse_ini_file($envFile);
$mysqlPassword = $envVars['MYSQL_ROOT_PASSWORD'] ?? '';
if (empty($mysqlPassword)) {
    header("Location: ./index.php?invalid");
    exit;
}

$serverIndex = 1;
foreach (glob('/home/*/sockets/mysqld/mysqld.sock') as $i => $sock) {
    $sockUser = explode('/', $sock)[2];
    if ($sockUser === $username) {
        $serverIndex = $i + 1;
        break;
    }
}

$sessionName = 'OPENPANEL_PMA_' . strtoupper($username);
session_set_cookie_params(0, '/', '', false, true);
session_name($sessionName);
session_start();

$_SESSION['PMA_single_signon_user']     = 'root';
$_SESSION['PMA_single_signon_password'] = $mysqlPassword;
$_SESSION['PMA_single_signon_host']     = 'localhost';

session_write_close();

header("Location: ./index.php?server=" . $serverIndex);
exit;
