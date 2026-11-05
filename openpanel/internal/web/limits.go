package web

// MaxConfigBody caps request bodies of the config editors (php.ini, webserver conf, WAF rules), these files are a few KB
const MaxConfigBody = 1 << 20
