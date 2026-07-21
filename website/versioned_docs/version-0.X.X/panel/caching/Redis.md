---
sidebar_position: 2
---

# REDIS

![redis_disabled.png](/img/panel/v1/caching/redis_disabled.png)

REDIS stands as a robust and persistent object cache solution, purpose-built to efficiently retain frequently accessed website data within the RAM memory.

With its capability to store and serve data from RAM, it minimizes the need for time-consuming database queries, significantly enhancing website performance.

## Enable / Disable

You have the option to enable or disable the REDIS service as necessary. Choosing to disable it will promptly clear (delete) all existing Redis data from memory.

Enabling the REDIS service will initiate the Redis service on the default port, which is _6379_.

![redis_enabled.png](/img/panel/v1/caching/redis_enabled.png)

## Set Memory Limits

By default, Redis has no memory limits set and can utilize all available RAM memory, as indicated by the ∞ symbol. To avoid Redis from consuming all available memory, it is advisable to establish a maximum memory limit.
You can set this limit to a maximum of 2GB for the Redis service.

:::info
Modifying the memory limit will require the service to reload to apply the new restrictions, resulting in the removal of all existing data from the cache.
:::

## Connect to REDIS

To establish a connection to your REDIS instance, use the following details:

- Server address: **127.0.0.1**
- Port: **6379** (the default REDIS port)

For testing the connection to the REDIS server, you can use the 'telnet' command in your terminal or refer to example scripts bellow.

### Test connection with PHP

1. Using FileManager enter the directory of your website
2. Create a new file named _redis-test.php_
3. Add the following code in the newly created file and save:
```php
<?php 
   //Connecting to Redis server on localhost 
   $redis = new Redis(); 
   $redis->connect('127.0.0.1', 6379); 
   echo "Connection to server sucessfully"; 
   //check whether server is running or not 
   echo "Server is running: ".$redis->ping(); 
?>
```

4. In your browser navigate to your website and add /redis-test.php for example, if your website is example.com you should open example.com/redis-test.php

You should see the _Server is running message.._ indicating that the REDIS service is active and connection is established.

### Test connection using telnet

To connect to the Redis service, you can run the telnet command and specify the hostname and port of the Redis service:

```bash
telnet localhost 6379
```

The command always prints the first line of the diagnostic message, stating that it’s connecting to the hostname. Then, if the connection is successful, you will see two additional lines, the first additional line confirms that connection is established:

```bash
$ telnet localhost 6379
Trying 127.0.0.1...
Connected to localhost.
Escape character is '^]'.
```

### WordPress plugins

To incorporate REDIS caching into your WordPress website, a WordPress plugin is required. Below are some of the WordPress plugins we evaluated for REDIS caching:

- [Redis Object Cache](https://wordpress.org/plugins/redis-cache/)
- [Use Redis](https://plugins.club/premium-wordpress-plugins/use-redis/)
   
## View Logs

You have the option to view the REDIS service logs. By doing so, you can identify any service errors or, for instance, determine whether memory limits have been reached.

![redis_log.png](/img/panel/v1/caching/redis_log.png)
