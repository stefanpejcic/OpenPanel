<?php

/**
 * Copyright 2022-2024 FOSSBilling
 * Copyright 2011-2021 BoxBilling, Inc.
 * SPDX-License-Identifier: Apache-2.0.
 *
 * @copyright FOSSBilling (https://www.fossbilling.org)
 * @license   http://www.apache.org/licenses/LICENSE-2.0 Apache-2.0
 */

use Random\RandomException;
use Symfony\Contracts\HttpClient\Exception\HttpExceptionInterface;

/**
 * OpenPanel API.
 * @see https://dev.openpanel.com/api/
 */
class Server_Manager_Openpanel extends Server_Manager
{
    /**
     * Returns the form configuration for the OpenPanel server manager.
     *
     * @return array the form configuration as an associative array
     */
    public static function getForm(): array
    {
        return [
            'label' => 'OpenPanel',
            'form' => [
                'credentials' => [
                    'fields' => [
                        [
                            'name' => 'username',
                            'type' => 'text',
                            'label' => 'Username',
                            'placeholder' => 'Username used to login to OpenAdmin',
                            'required' => true,
                        ],
                        [
                            'name' => 'password',
                            'type' => 'password',
                            'label' => 'Password',
                            'placeholder' => 'Password for the OpenAdmin user',
                            'required' => true,
                        ],
                    ],
                ],
            ],
        ];
    }

    /**
     * Initializes the OpenPanel server manager.
     * Checks if the necessary configuration options are set and throws an exception if any are missing.
     *
     * @throws Server_Exception if any necessary configuration options are missing
     */
    public function init(): void
    {
        if (empty($this->_config['host'])) {
            throw new Server_Exception('The ":server_manager" server manager is not fully configured. Please configure the :missing', [':server_manager' => 'OpenPanel', ':missing' => 'hostname'], 2001);
        }

        if (empty($this->_config['username'])) {
            throw new Server_Exception('The ":server_manager" server manager is not fully configured. Please configure the :missing', [':server_manager' => 'OpenPanel', ':missing' => 'username'], 2001);
        }

        if (empty($this->_config['password']) && empty($this->_config['accesshash'])) {
            throw new Server_Exception('The ":server_manager" server manager is not fully configured. Please configure the :missing', [':server_manager' => 'OpenPanel', ':missing' => 'authentication credentials'], 2001);
        }

        // If port not set, use OpenPanel default.
        $this->_config['port'] = empty($this->_config['port']) ? '2087' : $this->_config['port'];
    }










function getAuthToken() {
    $apiProtocol = $this->_config['secure'] ? 'https://' : 'http://';
    $host = $this->_config['host'];
    $username = $this->_config['username'];
    $password = $this->_config['password'];
    $port = $this->getPort();
    $authEndpoint = "{$apiProtocol}{$host}:{$port}/api/";

    //error_log("Username: $username");

    $postData = json_encode(array(
        'username' => $username,
        'password' => $password
    ));

    $curl = curl_init();
    curl_setopt_array($curl, array(
        CURLOPT_URL => $authEndpoint,
        CURLOPT_RETURNTRANSFER => true,
        CURLOPT_SSL_VERIFYHOST => 0,
        CURLOPT_SSL_VERIFYPEER => 0,
        CURLOPT_POST => true,
        CURLOPT_POSTFIELDS => $postData,
        CURLOPT_HTTPHEADER => array(
            "Content-Type: application/json"
        ),
        CURLOPT_TIMEOUT => 10,
    ));

    $response = curl_exec($curl);
    $httpCode = curl_getinfo($curl, CURLINFO_HTTP_CODE);
    //error_log("HTTP Status Code: $httpCode");

    if (curl_errno($curl)) {
        $error = "cURL Error: " . curl_error($curl);
        error_log($error);
        $token = false;
    } else {
        //error_log("Raw Response: $response");
        $responseData = json_decode($response, true);
        //error_log("Decoded response: " . print_r($responseData, true));
        if (json_last_error() !== JSON_ERROR_NONE) {
            error_log("JSON Decode Error: " . json_last_error_msg());
            curl_close($curl);
            return false;
        }

        $token = isset($responseData['access_token']) ? $responseData['access_token'] : false;
        
        if (!$token) {
            error_log("API is working, but JWT not found in response!");
        }
        //error_log("JWT: " . $token);
    }

    curl_close($curl);
    return $token;
}

    
    

function makeApiRequest($endpoint, $data = null, $method = 'GET') {
    $apiProtocol = $this->_config['secure'] ? 'https://' : 'http://';
    $host = $this->_config['host'];
    $baseUrl = $apiProtocol . $host . ':' . $this->getPort() . '/api/';

    $url = $baseUrl . $endpoint;

    error_log("URL: $url");

    $token = $this->getAuthToken();
        
    if (!$token) {
        error_log("Failed to retrieve auth token");
        return false;
    }
  
    $curl = curl_init();
  
    curl_setopt_array($curl, array(
        CURLOPT_URL => $url,
        CURLOPT_RETURNTRANSFER => true,
        CURLOPT_ENCODING => '',
        CURLOPT_MAXREDIRS => 10,
        CURLOPT_TIMEOUT => 0,
        CURLOPT_FOLLOWLOCATION => true,
        CURLOPT_HTTP_VERSION => CURL_HTTP_VERSION_1_1,
        CURLOPT_CUSTOMREQUEST => $method,
        CURLOPT_POSTFIELDS => $data,
        //CURLOPT_RESOLVE => array("HOST_HERE:PORT_HERE:IP_HERE"),
        CURLOPT_HTTPHEADER => array(
            'Content-Type: application/json',
            'Authorization: Bearer ' . $token
        ),
    ));

    $response = curl_exec($curl);
    curl_close($curl);

    return $response;
}

      















    
    /**
     * Returns the login URL for a OpenPanel account.
     *
     * @param Server_Account|null $account The account for which to get the login URL. This parameter is currently not used.
     *
     * @return string the login URL
     */
    public function getLoginUrl(Server_Account $account = null): string
    {
        $host = $this->_config['host'];
        $protocol = $this->_config['secure'] ? 'https://' : 'http://';

        return $protocol . $host . ':2083';
    }

    /**
     * Returns the login URL for a OpenAdmin reseller account.
     *
     * @param Server_Account|null $account The account for which to get the login URL. This parameter is currently not used.
     *
     * @return string the login URL
     */
    public function getResellerLoginUrl(Server_Account $account = null): string
    {
        $host = $this->_config['host'];
        $protocol = $this->_config['secure'] ? 'https://' : 'http://';
        $url = $protocol . $this->_config['host'] . ':' . $this->getPort() . '/api/';

        return $protocol . $host . ':2087';
    }

    # OpenAdmin can use custom port
    public function getPort(): int|string
    {
        $port = $this->_config['port'];

        if (filter_var($port, FILTER_VALIDATE_INT) !== false && $port >= 0 && $port <= 65535) {
            return $this->_config['port'];
        } else {
            return 2087;
        }
    }



    
    /**
     * Tests the connection to the OpenPanel server.
     * Sends a request to the OpenPanel server to check if api is working
     *
     * @return true if the connection was successful
     *
     * @throws Server_Exception if an error occurs during the request
     */
public function testConnection(): bool
{
    // Construct the request URL directly
    $apiProtocol = $this->_config['secure'] ? 'https://' : 'http://';
    $host = $this->_config['host'];
    $baseUrl = $apiProtocol . $host . ':' . $this->getPort() . '/api/';
    $url = $baseUrl;  // As no endpoint is provided, it's just the base URL

    // Make the API request
    $response = $this->makeApiRequest(null);
        
    if (!$response) {
        // Log the error and include the URL in the exception message
        throw new Server_Exception("Can't connect to $url Possible invalid credentials or unreachable host.");
    }

    $decoded = json_decode($response);
    if (json_last_error() !== JSON_ERROR_NONE) {
        throw new Server_Exception("Can't connect to the server: Invalid JSON - " . json_last_error_msg() . ". Response was: " . $response);
    }

    if (isset($decoded->message) && $decoded->message === "API is working!") {
        return true;
    }

    $errorMessage = isset($decoded->message) ? $decoded->message : 'Unexpected API response';
    throw new Server_Exception("Can't connect to the server: {$errorMessage}. Full response: " . json_encode($decoded));
}


    /**
     * Generates a username for a new account on the OpenPanel server.
     * The username is generated based on the domain name, with some modifications to comply with OpenPanel's username restrictions.
     *
     * @param string $domain the domain name for which to generate a username
     *
     * @return string the generated username
     *
     * @throws RandomException if an error occurs during the generation of a random number
     */
    public function generateUsername(string $domain): string
    {
        $processedDomain = strtolower(preg_replace('/[^A-Za-z0-9]/', '', $domain));
        $username = substr($processedDomain, 0, 7) . random_int(0, 9);

        // OpenPanel doesn't allow usernames to start with "test", so replace it with a random string if it does (test3456 would then become something like a62f93456).
        if (str_starts_with($username, 'test')) {
            $username = substr_replace($username, 'a' . bin2hex(random_bytes(2)), 0, 5);
        }

        return $username;
    }

    /**
     * Synchronizes an account with the OpenPanel server.
     * Sends a request to the OpenPanel server to get the account's details and updates the Server_Account object accordingly.
     *
     * @param Server_Account $account the account to be synchronized
     *
     * @return Server_Account the updated account
     *
     * @throws Server_Exception if an error occurs during the request, or if the account does not exist on the OpenPanel server
     */
    public function synchronizeAccount(Server_Account $account)
    {
       return false;
    }

    /**
     * Creates a new account on the OpenPanel server.
     * Sends a request to the OpenPanel server to create a new account with the details provided in the Server_Account object.
     * If the account is a reseller account, it also sets up the reseller and assigns the appropriate ACL list.
     *
     * @param Server_Account $account The account to be created. This object should contain all the necessary details for the new account.
     *
     * @return bool returns true if the account was successfully created, false otherwise
     *
     * @throws Server_Exception if an error occurs during the request, or if the response from the OpenPanel server indicates an error
     */
    public function createAccount(Server_Account $account)
    {
        $client = $account->getClient();
        $package = $account->getPackage();
        $this->getLog()->info('Creating account ' . $client->getUsername());
        $data = json_encode(array(
            "email" => $client->getEmail(),
            'username' => $account->getUsername(),
            'password' => $account->getPassword(),
            "plan_name" => $package->getName()

        ));
    

        $rawResponse = $this->makeApiRequest("users", $data, 'POST');
        $response = json_decode($rawResponse);
    
        if (is_object($response) && !empty($response->success)) {
            return true;
        }
    
        // https://github.com/stefanpejcic/FOSSBilling-OpenPanel/issues/2
        if (strpos($rawResponse, 'Successfully added user') !== false) {
            return true;
        }
    
        $errorMsg = is_object($response) && isset($response->error) ? $response->error : $rawResponse;
        throw new Server_Exception('Error when creating ' . $client->getUsername() . ': ' . $errorMsg);
        
    }
        
        /**
     * Suspends an account on the OpenPanel server.
     *
     * @param Server_Account $account the account to be suspended
     *
     * @return bool returns true if the account was successfully suspended
     *
     * @throws Server_Exception if an error occurs during the request
     */
    public function suspendAccount(Server_Account $account): bool
    {
        // Log the suspension
        $this->getLog()->info('Suspending account ' . $account->getUsername());

        $client = $account->getClient();

        $data = json_encode(array("action" => "suspend"));
        $response = $this->makeApiRequest("users/" . $account->getUsername() , $data, 'PATCH');
        $response = json_decode($response);

        if ($response->success == 1 || $response->success ==  true ) {
            return true;    
            
        }
        
        throw new Server_Exception('Error when suspending ' . $client->getUsername() . ': ' . json_encode($response));

    }

    /**
     * Unsuspends an account on the OpenPanel server.
     *
     * @param Server_Account $account the account to be unsuspended
     *
     * @return bool returns true if the account was successfully unsuspended
     *
     * @throws Server_Exception if an error occurs during the request
     */
    public function unsuspendAccount(Server_Account $account): bool
    {
        // Log the unsuspension
        $this->getLog()->info('Activating account ' . $account->getUsername());

        $client = $account->getClient();

        $data = json_encode(array("action" => "unsuspend"));
        $response = $this->makeApiRequest("users/". $account->getUsername() , $data, 'PATCH');
        $response = json_decode($response);

        if ($response->success == 1 || $response->success ==  true ) {
            return true;    
        
        }


        throw new Server_Exception('Failed to  unsuspend ' . $client->getUsername() . ': ' . $response->error);

    }

    /**
     * Cancels an account on the OpenPanel server.
     *
     * @param Server_Account $account the account to be cancelled
     *
     * @return bool returns true if the account was successfully cancelled
     *
     * @throws Server_Exception if an error occurs during the request
     */
    public function cancelAccount(Server_Account $account): bool
    {
        // Log the cancellation
        $this->getLog()->info('Canceling account ' . $account->getUsername());

        $response = $this->makeApiRequest(endpoint: "users/". $account->getUsername(), method: 'DELETE');
        $response = json_decode($response);

        if ($response->success) {
            return true;    
        
        }
        $client = $account->getClient();


        throw new Server_Exception('Failed to  canceling ' . $client->getUsername() . ': ' . $response->error);

    }

    /**
     * Changes the package of an account on the OpenPanel server.
     *
     * @param Server_Account $account the account for which to change the package
     * @param Server_Package $package the new package
     *
     * @return bool returns true if the package was successfully changed
     *
     * @throws Server_Exception if an error occurs during the request
     */
    public function changeAccountPackage(Server_Account $account, Server_Package $package)
    {
        
        // Log the package change
        $this->getLog()->info('Changing account ' . $account->getUsername() . ' package');
        $data = json_encode(array("plan_name" => $package->getName()));

        $response = $this->makeApiRequest("users/". $account->getUsername(),$data   , 'PUT');
        $response = json_decode($response);

        if ($response->success) {
            return true;    
        
        }
        $client = $account->getClient();


        throw new Server_Exception('Failed to change package for user ' . $client->getUsername() . ' | Error: ' . $response->error);

       

       
    }

    /**
     * Changes the password of an account on the OpenPanel server.
     *
     * @param Server_Account $account     the account for which to change the password
     * @param string         $newPassword the new password
     *
     * @return bool returns true if the password was successfully changed
     *
     * @throws Server_Exception if an error occurs during the request
     */
    public function changeAccountPassword(Server_Account $account, string $newPassword)
    {
        // Log the password change
        $this->getLog()->info('Changing account ' . $account->getUsername() . ' password');

        $data = json_encode(array("password" => $newPassword));

        $response = $this->makeApiRequest("users/". $account->getUsername(),$data   , 'PATCH');
        $response = json_decode($response);

        if ($response->success) {
            return true;    
        
        }
        $client = $account->getClient();


        throw new Server_Exception('Failed to change package for user ' . $client->getUsername() . ' | Error: ' . $response->error);
    }

    /**
     * Changes the username of an account on the OpenPanel server.
     *
     * @param Server_Account $account     the account for which to change the username
     * @param string         $newUsername the new username
     *
     * @return bool returns true if the username was successfully changed
     *
     * @throws Server_Exception if an error occurs during the request
     */
    public function changeAccountUsername(Server_Account $account, string $newUsername): bool
    {

        throw new Server_Exception('Error: Changing username is not enabled on OpenPanel server.');
    }

    /**
     * Changes the domain of an account on the OpenPanel server.
     *
     * @param Server_Account $account   the account for which to change the domain
     * @param string         $newDomain the new domain
     *
     * @return bool returns true if the domain was successfully changed
     *
     * @throws Server_Exception if an error occurs during the request
     */
    public function changeAccountDomain(Server_Account $account, string $newDomain): bool
    {
        throw new Server_Exception('Error: OpenPanel does not have a concept of primary domains.');

    }

    /**
     * Changes the IP of an account on the OpenPanel server.
     *
     * @param Server_Account $account the account for which to change the IP
     * @param string         $newIp   the new IP
     *
     * @return bool returns true if the IP was successfully changed
     *
     * @throws Server_Exception if an error occurs during the request
     */
    public function changeAccountIp(Server_Account $account, string $newIp): bool
    {
        throw new Server_Exception('OpenPanel does not supporting change account IP');

    }



  

   
}
