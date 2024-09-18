vcl 4.1;

backend default {
    .host = "127.0.0.1";
    .port = "80";
}

sub vcl_recv {
    if (req.http.X-Forwarded-Proto == "https") {
        set req.http.X-Forwarded-Proto = "https";
    }
    set req.backend_hint = default;
}

sub vcl_backend_response {
    set beresp.ttl = 5m;
}

sub vcl_deliver {
    unset resp.http.X-Powered-By;
    unset resp.http.Server;
}
