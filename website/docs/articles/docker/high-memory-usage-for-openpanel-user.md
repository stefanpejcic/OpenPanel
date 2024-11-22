#  High physical memory usage for new OpenPanel account

OpenPanel uses Docker images as a base for every user. We provide and maintain four official images that use Ubuntu24 and Nginx or Apache webserver, and MySQL or MariaDB.

These images are optimized for both speed and performance, and new idle accounts will use from 10-150mb of RAM - depending if using Debian or Ubuntu as server OS.

If you want to fine tune the performanse and lower RAM usage for idle accounts, you can create a custom Docker image that will have only the services that you offer pre-installed.

[Here is an example guide on creating a custom docker image](/docs/articles/docker/building_a_docker_image_example_include_php_ioncubeloader/).


## How OpenPanel uses RAM

When new user is created, it has a Ubuntu24 OS that has no services running. A unique feature of OpenPanel is that additional services are started only when needed.

For example, PHP and Nginx/Apache are started only after user adds its first domain. MySQL only after adding database/user, phpMyAdmin and WebTerminal only when opens the pages in UI, etc.
This allows to minimize resources needed for each user, and allows overselling of resources on the server.


## Memory Usage shows High usage

Unlike other panels, all usage data is real-time and shows actuall usage for user. As on other panels, **cached memory is counted in the total memory usage**. This means that if a service or process used ram and exits, that ram is not automatically returned to the free ram and will still show as used, but in reallity it is cached ram that can be made available at any time if needed.

This means that the memory usage displayed for current processes is lower than the memory usage reported in OpenPanel dashboard page.
Bacause this often confuses users, OpenPanel will periodically safely drop all cached ram at the beginning of an hour.
