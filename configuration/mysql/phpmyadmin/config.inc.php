<?php
require_once '/etc/phpmyadmin/config.secret.inc.php';
require_once '/etc/phpmyadmin/helpers.php';
if (isset($_ENV['MAX_EXECUTION_TIME'])) {
    $cfg['ExecTimeLimit'] = (int) $_ENV['MAX_EXECUTION_TIME'];
}
if (isset($_ENV['MEMORY_LIMIT'])) {
    $cfg['MemoryLimit'] = $_ENV['MEMORY_LIMIT'];
}
if (isset($_ENV['PMA_ABSOLUTE_URI'])) {
    $cfg['PmaAbsoluteUri'] = trim($_ENV['PMA_ABSOLUTE_URI']);
}
$cfg['DisplayServersList'] = false;
$cfg['AllowArbitraryServer'] = false;
$cfg['NavigationDisplayServers'] = false;

$sockets = glob('/home/*/sockets/mysqld/mysqld.sock');

// manual login: index.php?manual=USERNAME  (kept in a cookie so it survives the login POST)
$manualUser = $_GET['manual'] ?? (isset($_GET['server']) ? '' : ($_COOKIE['pma_manual_user'] ?? ''));
$manualValid = ($manualUser !== ''
    && preg_match('/^[a-zA-Z0-9_-]+$/', $manualUser)
    && in_array('/home/' . $manualUser . '/sockets/mysqld/mysqld.sock', $sockets, true));

if ($manualValid) {
    if (isset($_GET['manual'])) {
        setcookie('pma_manual_user', $manualUser, 0, '/', '', false, true);
    }
    $cfg['Servers'][1]['verbose']         = $manualUser;
    $cfg['Servers'][1]['host']            = 'localhost';
    $cfg['Servers'][1]['socket']          = '/home/' . $manualUser . '/sockets/mysqld/mysqld.sock';
    $cfg['Servers'][1]['auth_type']       = 'cookie';
    $cfg['Servers'][1]['compress']        = false;
    $cfg['Servers'][1]['AllowNoPassword'] = false;
    $cfg['ServerDefault']                 = 1;
} else {
    // auto login from OpenPanel UI: pma.php?user=USERNAME&token=..
    $serverIndex = 1;
    foreach ($sockets as $socketPath) {
        $username = explode('/', $socketPath)[2];
        $cfg['Servers'][$serverIndex]['verbose']         = $username;
        $cfg['Servers'][$serverIndex]['host']            = 'localhost';
        $cfg['Servers'][$serverIndex]['socket']          = $socketPath;
        $cfg['Servers'][$serverIndex]['auth_type']       = 'signon';
        $cfg['Servers'][$serverIndex]['SignonSession']   = 'OPENPANEL_PMA_' . strtoupper($username);
        $cfg['Servers'][$serverIndex]['SignonURL']       = 'pma.php';
        $cfg['Servers'][$serverIndex]['compress']        = false;
        $cfg['Servers'][$serverIndex]['AllowNoPassword'] = true;
        $serverIndex++;
    }
}

if (isset($_ENV['PMA_UPLOADDIR'])) { $cfg['UploadDir'] = $_ENV['PMA_UPLOADDIR']; }
if (isset($_ENV['PMA_SAVEDIR']))   { $cfg['SaveDir']   = $_ENV['PMA_SAVEDIR'];   }
if (file_exists('/etc/phpmyadmin/config.user.inc.php')) {
    include '/etc/phpmyadmin/config.user.inc.php';
}
if (is_dir('/etc/phpmyadmin/conf.d/')) {
    foreach (glob('/etc/phpmyadmin/conf.d/*.php') as $filename) {
        include $filename;
    }
}
