<?php
################################################################################
# Name: OpenPanel paymenter.org Module
# Usage: https://openpanel.com/docs/articles/extensions/openpanel-and-paymenter.org/
# Source: https://github.com/stefanpejcic/openpanel-paymenter.org
# Author: Stefan Pejcic
# Created: 09.10.2024
# Last Modified: 09.10.2024
# Company: openpanel.com
# Copyright (c) Stefan Pejcic
# 
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
# 
# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.
# 
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
# THE SOFTWARE.
################################################################################

namespace App\Extensions\Servers\OpenPanel;

use App\Classes\Extensions\Server;
use App\Helpers\ExtensionHelper;
use Illuminate\Support\Facades\Http;

class OpenPanel extends Server
{
    public function getMetadata()
    {
        return [
            'display_name' => 'OpenPanel',
            'version' => '1.0.0',
            'author' => 'Paymenter',
            'website' => 'https://paymenter.org',
        ];
    }
    
    public function getConfig()
    {
        return [
            [
                'name' => 'host',
                'friendlyName' => 'Url to OpenPanel server (with port)',
                'type' => 'text',
                'required' => true,
            ],
            [
                'name' => 'username',
                'friendlyName' => 'Username',
                'type' => 'text',
                'required' => true,
            ],
            [
                'name' => 'password',
                'friendlyName' => 'Password',
                'type' => 'text',
                'required' => true,
            ]
        ];
    }

    public function getProductConfig($options)
    {
        return [
            [
                'name' => 'packageName',
                'friendlyName' => 'Package Name',
                'type' => 'text',
                'required' => true,
                'description' => 'Package Name for the OpenPanel server',
            ],
        ];
    }

    public function getUserConfig()
    {
        return [
            [
                'name' => 'domain',
                'friendlyName' => 'Domain',
                'type' => 'text',
                'required' => true,
                'description' => 'Domain for the webhost',
            ],
            [
                'name' => 'username',
                'friendlyName' => 'Username',
                'type' => 'text',
                'required' => true,
                'description' => 'Username to login to the website',
            ],
            [
                'name' => 'password',
                'friendlyName' => 'Password',
                'type' => 'text',
                'required' => true,
                'description' => 'Password to login to the website',
            ]
        ];
    }

    public function createServer($user, $params, $order, $product, $configurableOptions)
    {
        list($jwtToken, $error) = $this->getAuthToken();
        
        if (!$jwtToken) {
            ExtensionHelper::error('OpenPanel', 'Failed to get authentication token: ' . $error);
            return;
        }
        
        $createUserEndpoint = $this->getApiProtocol($params["host"]) . $params["host"] . ':2087/api/users';
        $response = Http::withToken($jwtToken)
            ->post($createUserEndpoint, [
                'username' => $params['config']['username'],
                'password' => $params['config']['password'],
                'email' => $user->email,
                'plan_name' => $params['packageName'],
            ]);

        if (!$response->successful()) {
            ExtensionHelper::error('OpenPanel', 'Failed to create server: ' . $response->body());
        }
    }

    public function suspendServer($user, $params, $order, $product, $configurableOptions)
    {
        list($jwtToken, $error) = $this->getAuthToken();

        if (!$jwtToken) {
            ExtensionHelper::error('OpenPanel', 'Failed to get authentication token: ' . $error);
            return;
        }

        $suspendUserEndpoint = $this->getApiProtocol($params["host"]) . $params["host"] . ':2087/api/users/' . $params["config"]["username"] . '/suspend';
        $response = Http::withToken($jwtToken)->patch($suspendUserEndpoint);

        if (!$response->successful()) {
            ExtensionHelper::error('OpenPanel', 'Failed to suspend server: ' . $response->body());
        }
    }

    public function unsuspendServer($user, $params, $order, $product, $configurableOptions)
    {
        list($jwtToken, $error) = $this->getAuthToken();

        if (!$jwtToken) {
            ExtensionHelper::error('OpenPanel', 'Failed to get authentication token: ' . $error);
            return;
        }

        $unsuspendUserEndpoint = $this->getApiProtocol($params["host"]) . $params["host"] . ':2087/api/users/' . $params["config"]["username"] . '/unsuspend';
        $response = Http::withToken($jwtToken)->patch($unsuspendUserEndpoint);

        if (!$response->successful()) {
            ExtensionHelper::error('OpenPanel', 'Failed to unsuspend server: ' . $response->body());
        }
    }

    public function terminateServer($user, $params, $order, $product, $configurableOptions)
    {
        list($jwtToken, $error) = $this->getAuthToken();

        if (!$jwtToken) {
            ExtensionHelper::error('OpenPanel', 'Failed to get authentication token: ' . $error);
            return;
        }

        $deleteUserEndpoint = $this->getApiProtocol($params["host"]) . $params["host"] . ':2087/api/users/' . $params["config"]["username"];
        $response = Http::withToken($jwtToken)->delete($deleteUserEndpoint);

        if (!$response->successful()) {
            ExtensionHelper::error('OpenPanel', 'Failed to terminate server: ' . $response->body());
        }
    }

    private function getApiProtocol($hostname)
    {
        return filter_var($hostname, FILTER_VALIDATE_IP) === false ? 'https://' : 'http://';
    }

    private function getAuthToken()
    {
        $authEndpoint = $this->getApiProtocol(ExtensionHelper::getConfig('OpenPanel', 'host')) . ExtensionHelper::getConfig('OpenPanel', 'host') . ':2087/api/auth';
        
        $response = Http::post($authEndpoint, [
            'username' => ExtensionHelper::getConfig('OpenPanel', 'username'),
            'password' => ExtensionHelper::getConfig('OpenPanel', 'password'),
        ]);

        if (!$response->successful()) {
            return [false, 'Failed to authenticate: ' . $response->body()];
        }

        $responseData = $response->json();
        return [$responseData['access_token'] ?? false, 'Token not found in response'];
    }
}
