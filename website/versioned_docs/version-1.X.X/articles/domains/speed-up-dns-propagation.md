# Speed up DNS propagation

DNS propagation cannot be speed up.

To minimize the effects of DNS propagation, lower the `TTL` value for the DNS record you need to change at least 24 hours before the record change.

Setting the TTL to 300 seconds causes most DNS resolvers to update their cache for the DNS record every 5 minutes.
