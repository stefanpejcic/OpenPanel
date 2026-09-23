# Limiting Connections with CSF

CSF allows you to restrict the number of simultaneous connections to specific ports on your server.

For example, you can configure CSF to limit port 80 to a maximum of 5 concurrent connections, and port 443 to 20 concurrent connections by adding the following to your configuration file:

```
CONNLIMIT = "80;5,443;20"
```

**How CONNLIMIT Works:**

The `CONNLIMIT` setting uses a comma-separated list of `port;limit` pairs.

For example:
```
CONNLIMIT = "22;5,80;20"
```

means:

1. Each IP address can have up to 5 concurrent new connections to port 22.
2. Each IP address can have up to 20 concurrent new connections to port 80.


