<?php

// allow access from openpanel servers
header("Access-Control-Allow-Origin: https://preview.openpanel.org");
header("Access-Control-Allow-Methods: GET, POST, OPTIONS, PUT");
header("Access-Control-Allow-Headers: Content-Type, Authorization");
header("X-Robots-Tag: noindex, nofollow", true);

$requestUri = $_SERVER['REQUEST_URI'];

include 'config.php';

// Basic validation
if (empty($ip) || empty($domen)) {
    header("Location: https://preview.openpanel.org/#expired");
    exit;
} else {
    $domainOnly = strtok($domen, '/'); # pcx3.com/blog and pcx3.com
}


if (!filter_var($ip, FILTER_VALIDATE_IP)) {
    http_response_code(400); // Bad Request
    die('Invalid IP address format: ' . htmlspecialchars($ip));
}

$scheme = (!empty($_SERVER['HTTPS']) ? 'https://' : 'http://');
$targetUrl = $scheme . $domen . $requestUri; 

$ch = curl_init($targetUrl);

curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
curl_setopt($ch, CURLOPT_FOLLOWLOCATION, true); // -L

//$proxy = 'http://IP_HERE:3128';  // proxy site!
//curl_setopt($ch, CURLOPT_PROXY, $proxy);

curl_setopt($ch, CURLOPT_HEADER, false);
curl_setopt($ch, CURLOPT_HTTPHEADER, ['X-Forwarded-For: ' . $ip]);
curl_setopt($ch, CURLOPT_RESOLVE, ["$domainOnly:80:$ip"]); // :443 to require ssl
curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, false); // -k

$response = curl_exec($ch);

if ($response === false) {
    $errorMessage = curl_error($ch);
    echo "Error connecting to <code>$ip</code> on port <code>80</code> - make sure that domain <b> $domen </b> exists on the server <code>$ip</code> and that web server is running.<br>";
    if (strpos($errorMessage, 'SSL') !== false) {
        curl_close($ch);
        echo "response: <pre>" . $errorMessage . "</pre>";
    } else {
        http_response_code(500);
        curl_close($ch);
        echo "response: <pre>" . curl_error($ch) . "</pre>";
        exit;
    }
}


// Return proper types for extensions
$filePath = parse_url($targetUrl, PHP_URL_PATH);
$fileExtension = pathinfo($filePath, PATHINFO_EXTENSION);
if (substr($fileExtension, -4) === '.jpg') {
    $fileExtension = 'jpg'; 
}

switch (strtolower($fileExtension)) {
    case 'svg':
        header("Content-Type: image/svg+xml");
        break;
    case 'png':
        header("Content-Type: image/png");
        break;
    case 'jpg':
    case 'jpeg':
        header("Content-Type: image/jpeg");
        break;
    case 'gif':
        header("Content-Type: image/gif");
        break;
    case 'bmp':
        header("Content-Type: image/bmp");
        break;
    case 'webp':
        header("Content-Type: image/webp");
        break;
    case 'css':
        header("Content-Type: text/css");
        break;
    case 'js':
        header("Content-Type: application/javascript");
        break;
    case 'txt':
        header("Content-Type: text/plain");
        break;
    case 'ttf':
        header("Content-Type: font/ttf");
        break;
    default:
        header("Content-Type: text/html");
}


// matches: /domain     http://domain   https://domain      https:\/\/domain    http://www.domain   https://www.domain
$visitingDomain = $_SERVER['HTTP_HOST'];
$response = str_replace(
    ["/$domen", "http://$domen", "https://$domen", "http://www.$domen", "https://www.$domen"],
    ["/$visitingDomain", "$scheme$visitingDomain", "$scheme$visitingDomain", "$scheme$visitingDomain", "$scheme$visitingDomain"],
    $response
);


header("Access-Control-Allow-Origin: *");
header("Access-Control-Allow-Headers: Content-Type");

echo $response;
curl_close($ch);
?>
