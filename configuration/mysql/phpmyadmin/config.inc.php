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

$cfg['AllowArbitraryServer'] = false;
$cfg['NavigationDisplayServers'] = false;

$serverIndex = 1;
foreach (glob('/home/*/sockets/mysqld/mysqld.sock') as $socketPath) {
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
