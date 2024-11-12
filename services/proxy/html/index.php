<?php
$rootDomain = 'openpanel.org';
$logFile = __DIR__ . '/do.log';

// logging
function logMessage($message) {
    global $logFile;
    $currentTime = date('Y-m-d H:i:s');
    $logEntry = "[$currentTime] $message\n";
    file_put_contents($logFile, $logEntry, FILE_APPEND);
}

set_error_handler(function($errno, $errstr, $errfile, $errline) {
    logMessage("Error [$errno]: $errstr in $errfile on line $errline");
    return true;
});

set_exception_handler(function($exception) {
    logMessage("Uncaught Exception: " . $exception->getMessage());
});

if (!empty($_SERVER['HTTP_X_FORWARDED_FOR'])) {
    $clientIp = $_SERVER['HTTP_X_FORWARDED_FOR'];
} elseif (!empty($_SERVER['HTTP_X_REAL_IP'])) {
    $clientIp = $_SERVER['HTTP_X_REAL_IP'];
} else {
    $clientIp = $_SERVER['REMOTE_ADDR'];
}
$clientIp = explode(',', $clientIp)[0];

logMessage("Client IP: $clientIp");



// generate subdomains
function generateRandomSubdomain($domain) {
    $randomString = substr(bin2hex(random_bytes(3)), 0, 5);
    return $randomString . '.' . $domain;
}

// redirect
function redirectToSubdomain($subdomain, $domain) {
    $url = "https://$subdomain.$domain";
    logMessage("Redirecting to: $url");
    header("Location: $url");
    exit();
}

// cp skeleton
function copyData($sourceDir, $destinationDir) {
    if (!is_dir($sourceDir)) {
        logMessage("Source directory does not exist: $sourceDir");
        return false;
    }

    $files = scandir($sourceDir);
    foreach ($files as $file) {
        if ($file === '.' || $file === '..') {
            continue;
        }
        copy("$sourceDir/$file", "$destinationDir/$file");
        logMessage("Copied file: $file to $destinationDir");
    }
    return true;
}



// create proxy
if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    logMessage("Received POST request");

    $rootDomain = $GLOBALS["rootDomain"]; 
    $domain = '';
    $rootDir = '/var/www/html/virt';

    $ip = $_POST['ip'] ?? '';
    $fake_domain = $_POST['domain'] ?? '';

    logMessage("Received POST data: " . print_r($_POST, true));

    if (empty($ip) || empty($fake_domain)) {
        $errorMsg = "IP and Domain are required.";
        logMessage($errorMsg);
        echo $errorMsg;
        exit();
    }

    $subdomain = generateRandomSubdomain($fake_domain);
    $subdomainPart = explode('.', $subdomain)[0];
    $destinationDir = "/var/www/html/domains/$subdomainPart";
    $destinationDirR = "/var/www/html/$subdomainPart";
    logMessage("Generated subdomain: $subdomain");

    logMessage("Attempting to create directory: $destinationDir");

if (!is_writable(dirname($destinationDir))) {
    logMessage("Directory is not writable: " . dirname($destinationDir));
    exit("Error: Parent directory is not writable.");
}

if (!mkdir($destinationDir, 0755, true)) {
    logMessage("Failed to create directory: $destinationDir. Permissions or path issue?");
    exit("Error: Failed to create directory.");
} else {
    logMessage("Directory created successfully: $destinationDir");
}

    if (!copyData($rootDir, $destinationDir)) {
        logMessage("Failed to copy data from $rootDir to $destinationDir");
        exit("Error: Failed to copy data.");
    }
    logMessage("Copied data from $rootDir to $destinationDir");

    // config to use for proxy
    $configFilePath = "$destinationDir/config.php";
    $configContent = "<?php\n\$domen = '$fake_domain';\n\$ip = '$ip';\n";
    if (file_put_contents($configFilePath, $configContent) === false) {
        logMessage("Failed to create config file: $configFilePath");
        exit("Error: Failed to create config file.");
    }
    logMessage("Created config file at $configFilePath with content: $configContent");
    redirectToSubdomain($subdomainPart, $rootDomain);
} else {
   // display landing page
    header("Location: index.html");
    exit();
}

?>
