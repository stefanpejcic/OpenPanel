vcl 4.1;

backend default {
    .host = "localhost";
    .port = "8080";
}

sub vcl_recv {
    if (req.http.X-Forwarded-Proto == "https") {
        set req.http.X-Forwarded-Proto = "https";
    } else {
        set req.http.X-Forwarded-Proto = "http";
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
