#  High physical memory usage for new OpenPanel accounts

OpenPanel uses [Docker images](#) as a base for every user. We provide and maintain two official images that both use Ubuntu24 and Nginx or Apache webserver.

These images are optimized for both speed and performance, and new idle accounts will use from 10-150mb of RAM - depending if using Debian or Ubuntu as server OS.

If you want to fine tune the performanse and lower RAM usage for idle accounts, you can create a custom Docker image that will have only the services that you offer pre-installed.

Here is an example guide on creating a custom docker image:


## How OpenPanel uses RAM

When new user is created, it has a Ubuntu24 OS that has only PHP running. A unique feature of OpenPanel is that additional services are started only when needed.

For example, Nginx/Apache are started only after user adds its first domain. MySQL only after adding database/user, phpMyAdmin and WebTerminal only when opened, etc.
This allows to minimize resources needed for each user, and allows overselling of resources.


## Memory Usage shows High usage

Unlike other panels, all usage data is real-time and shows actuall usage for user. As on other panels, physical memory usage is counted in this total usage. This means that if a service or process used ram and exits, that ram is not automatically returned to the free ram and will still show as used, but in reallity its is cahced ram that can be made available at any time oif needed.

This often confuses users, so we added to periodically safely drop that cached ram after every hour.

## New Accounts

## Accounts with data
