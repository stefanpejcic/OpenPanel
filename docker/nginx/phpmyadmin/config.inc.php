<?php
if (!function_exists('check_file_access')) {
    function check_file_access($path)
    {
        if (is_readable($path)) {
            return true;
        } else {
            error_log(
                'phpmyadmin: Failed to load ' . $path
                . ' Check group www-data has read access and open_basedir restrictions.'
            );
            return false;
        }
    }
}

if (check_file_access('/var/lib/phpmyadmin/blowfish_secret.inc.php')) {
    require('/var/lib/phpmyadmin/blowfish_secret.inc.php');
}

$i = 0;
$i++;
if (check_file_access('/etc/phpmyadmin/config-db.php')) {
    require('/etc/phpmyadmin/config-db.php');
}

if (!empty($dbname)) {
    $serverPort = $_SERVER['SERVER_PORT'];
    // AUTOLOGIN FROM OPENPANEL UI
    if ($serverPort != 80 && $serverPort != 443) {
        error_log("Using Single Sign-On (SSO) for connections from OpenPanel: $clientIp");
        $cfg['Servers'][$i]['auth_type'] = 'signon';
        $cfg['Servers'][$i]['SignonSession'] = 'OPENPANEL_PHPMYADMIN';
        $cfg['Servers'][$i]['SignonURL'] = 'pma.php';
    // LOGIN FORM ON DOMAIN/phpmyadmin
    } else {
        error_log("Using cookie authentication for connection via domain name");
        $cfg['Servers'][$i]['auth_type'] = 'cookie';
        $cfg['Servers'][$i]['user'] = '';
        $cfg['Servers'][$i]['password'] = '';
    }

    /* Server parameters */
    if (empty($dbserver)) $dbserver = 'localhost';
    $cfg['Servers'][$i]['host'] = $dbserver;

    if (!empty($dbport) || $dbserver != 'localhost') {
        $cfg['Servers'][$i]['connect_type'] = 'tcp';
        $cfg['Servers'][$i]['port'] = $dbport;
    }
    $cfg['Servers'][$i]['extension'] = 'mysqli';
    $cfg['Servers'][$i]['controluser'] = $dbuser;
    $cfg['Servers'][$i]['controlpass'] = $dbpass;
    $cfg['Servers'][$i]['pmadb'] = $dbname;
    $cfg['Servers'][$i]['bookmarktable'] = 'pma__bookmark';
    $cfg['Servers'][$i]['relation'] = 'pma__relation';
    $cfg['Servers'][$i]['table_info'] = 'pma__table_info';
    $cfg['Servers'][$i]['table_coords'] = 'pma__table_coords';
    $cfg['Servers'][$i]['pdf_pages'] = 'pma__pdf_pages';
    $cfg['Servers'][$i]['column_info'] = 'pma__column_info';
    $cfg['Servers'][$i]['history'] = 'pma__history';
    $cfg['Servers'][$i]['table_uiprefs'] = 'pma__table_uiprefs';
    $cfg['Servers'][$i]['tracking'] = 'pma__tracking';
    $cfg['Servers'][$i]['userconfig'] = 'pma__userconfig';
    $cfg['Servers'][$i]['recent'] = 'pma__recent';
    $cfg['Servers'][$i]['favorite'] = 'pma__favorite';
    $cfg['Servers'][$i]['users'] = 'pma__users';
    $cfg['Servers'][$i]['usergroups'] = 'pma__usergroups';
    $cfg['Servers'][$i]['navigationhiding'] = 'pma__navigationhiding';
    $cfg['Servers'][$i]['savedsearches'] = 'pma__savedsearches';
    $cfg['Servers'][$i]['central_columns'] = 'pma__central_columns';
    $cfg['Servers'][$i]['designer_settings'] = 'pma__designer_settings';
    $cfg['Servers'][$i]['export_templates'] = 'pma__export_templates';
    $cfg['Servers'][$i]['hide_db'] = 'information_schema|performance_schema|mysql|sys|phpmyadmin';
    $i++;
}

$cfg['ShowChgPassword'] = false;
$cfg['ShowCreateDb'] = false;
$cfg['SuggestDBName'] = false;
$cfg['AllowUserDropDatabase'] = false;
$cfg['PmaNoRelation_DisableWarning'] = true;
$cfg['UploadDir'] = '';
$cfg['SaveDir'] = '';
$cfg['ShowDatabasesNavigationAsTree'] = false;

/* Support additional configurations */
foreach (glob('/etc/phpmyadmin/conf.d/*.php') as $filename) {
    include($filename);
}

$cfg['SendErrorReports'] = 'never';
