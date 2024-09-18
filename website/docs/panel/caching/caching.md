---
sidebar_position: 1
---

# Caching

Utilizing the array of out-of-the-box features provided by OpenPanel can significantly enhance your website's performance and security. Features such as PHP-FPM, Redis, Memcached, and Opcache have the potential to dramatically boost your website's speed and efficiency.

## REDIS

Redis serves as a persistent object cache backend cache server, primarily accelerating database and website-related calls and queries by storing them in RAM memory. RAM is known for its high-speed performance, surpassing even NVMe and UFS, making Redis caching a powerful tool to optimize your website's performance. Leveraging Redis cache can significantly benefit your website.

[REDIS settings and usage](/docs/panel/caching/Redis)

## Memcached

This in-memory (RAM) object cache is designed specifically to reduce the database load, making it ideal for dynamic websites. It caches only the queries related to the database.

We don't recommend using it in conjunction with other caches like Redis. However, it can be effectively used alongside PHP OPcache for improved website performance.

[Memcached configuration and usage](/docs/panel/caching/Memcached)

## Elasticsearch

Enhance your website's search capabilities and overall performance with Elasticsearch. OpenPanel provides seamless integration with Elasticsearch, a powerful search engine that allows for efficient and quick retrieval of information.

Configure and optimize Elasticsearch settings to tailor it to your website's needs. Learn about the various features and functionalities offered by Elasticsearch to make the most out of this robust search engine.

[Elasticsearch settings and usage](/docs/panel/caching/elasticsearch)

## Opcache
OPcache is a valuable tool for enhancing PHP performance. It works by storing precompiled script code in shared memory, eliminating the need for PHP to reload and analyze scripts with each request. In simpler terms, OPcache caches previously executed PHP code, reducing CPU load and improving website performance.

**This feature is enabled by default when using PHP and requires no additional settings.**
