---
sidebar_position: 2
---

# REDIS

![redis_disabled.png](/img/panel/v2/redismain.png)

REDIS stands as a robust and persistent object cache solution, purpose-built to efficiently retain frequently accessed website data within the RAM memory.

With its capability to store and serve data from RAM, it minimizes the need for time-consuming database queries, significantly enhancing website performance.

## Enable / Disable

You have the option to enable or disable the REDIS service container as necessary. Choosing to disable it will promptly clear (delete) all existing Redis data from memory.

Enabling the REDIS service container will initiate the Redis container service on the default port, which is _6379_.

![redis_enabled.png](/img/panel/v2/redisenabled.png)

## Set Memory Limits

Upon initialization the Redis container has default memory limits set, it is advisable to set memory limits appropriate to your use case and needs.

You can set these limits on the /containers interface which is accessible through the user panel navigation on Docker/containers .

![redis_limits.png](/img/panel/v2/redislimits.png)

:::info
Modifying the memory limit will require the Redis container to be restarted to apply the new restrictions, resulting in the removal of all existing data from the cache.
:::

## Connect to REDIS

To establish a connection to your REDIS instance, use the following details:

- Server address: **redis** (not 127.0.0.1)
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
   $redis->connect('redis', 6379); 
   echo "Connection to server successfully"; 
   //check whether server is running or not 
   echo "Server is running: ".$redis->ping(); 
?>
```

4. In your browser navigate to your website and add /redis-test.php for example, if your website is example.com you should open example.com/redis-test.php

You should see the _Server is running message.._ indicating that the REDIS service is active and connection is established.


### WordPress plugins

To incorporate REDIS caching into your WordPress website, a WordPress plugin is required. Below are some of the WordPress plugins we evaluated for REDIS caching:

- [Redis Object Cache](https://wordpress.org/plugins/redis-cache/)
- [Use Redis](https://plugins.club/premium-wordpress-plugins/use-redis/)
   
## View Logs

You have the option to view the REDIS service logs. By doing so, you can identify any service errors or, for instance, determine whether memory limits have been reached.

![redis_log.png](/img/panel/v2/redislogs.png)
