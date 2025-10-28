vcl 4.1;

backend default {
    .host = "VARNISH_BACKEND_HOST";
    .port = "80";
    .max_connections = 2048;
}

# Define an access control list to restrict cache purging.
acl purge {
    "localhost";
    "172.16.0.0/12";
}

sub vcl_hit {
    if (req.http.X-Forwarded-Proto == "https") {
        set req.http.X-Forwarded-Proto = "https";
    } else {
        set req.http.X-Forwarded-Proto = "http";
    }

    if (obj.ttl > 0s) {
        return (deliver);
    }
    if (obj.ttl + obj.grace > 0s) {
        return (deliver);
    }

    return (miss);  # Use "miss" instead of "fetch"
}

sub vcl_recv {

    # Remove empty query string parameters
    # e.g.: www.example.com/index.html?
    if (req.url ~ "\?$") {
        set req.url = regsub(req.url, "\?$", "");
    }

    # Remove port number from host header
    set req.http.Host = regsub(req.http.Host, ":[0-9]+", "");

    # Remove the proxy header to mitigate the httpoxy vulnerability
    # See https://httpoxy.org/
    unset req.http.proxy;

    # Purge logic to remove objects from the cache.
    # Tailored to the Proxy Cache Purge WordPress plugin
    # See https://wordpress.org/plugins/varnish-http-purge/
    if(req.method == "PURGE") {
        if(!client.ip ~ purge) {
            return(synth(405,"PURGE not allowed for this IP address"));
        }
        if (req.http.X-Purge-Method == "regex") {
            ban("obj.http.x-url ~ " + req.url + " && obj.http.x-host == " + req.http.host);
            return(synth(200, "Purged"));
        }
        ban("obj.http.x-url == " + req.url + " && obj.http.x-host == " + req.http.host);
        return(synth(200, "Purged"));
    }

   # Only handle relevant HTTP request methods
    if (
        req.method != "GET" &&
        req.method != "HEAD" &&
        req.method != "PUT" &&
        req.method != "POST" &&
        req.method != "PATCH" &&
        req.method != "TRACE" &&
        req.method != "OPTIONS" &&
        req.method != "DELETE"
    ) {
        return (pipe);
    }

    # Remove tracking query string parameters used by analytics tools
    if (req.url ~ "(\?|&)(_branch_match_id|_bta_[a-z]+|_bta_c|_bta_tid|_ga|_gl|_ke|_kx|campid|cof|customid|cx|dclid|dm_i|ef_id|epik|fbclid|gad_source|gbraid|gclid|gclsrc|gdffi|gdfms|gdftrk|hsa_acc|hsa_ad|hsa_cam|hsa_grp|hsa_kw|hsa_mt|hsa_net|hsa_src|hsa_tgt|hsa_ver|ie|igshid|irclickid|matomo_campaign|matomo_cid|matomo_content|matomo_group|matomo_keyword|matomo_medium|matomo_placement|matomo_source|mc_[a-z]+|mc_cid|mc_eid|mkcid|mkevt|mkrid|mkwid|msclkid|mtm_campaign|mtm_cid|mtm_content|mtm_group|mtm_keyword|mtm_medium|mtm_placement|mtm_source|nb_klid|ndclid|origin|pcrid|piwik_campaign|piwik_keyword|piwik_kwd|pk_campaign|pk_keyword|pk_kwd|redirect_log_mongo_id|redirect_mongo_id|rtid|s_kwcid|sb_referer_host|sccid|si|siteurl|sms_click|sms_source|sms_uph|srsltid|toolid|trk_contact|trk_module|trk_msg|trk_sid|ttclid|twclid|utm_[a-z]+|utm_campaign|utm_content|utm_creative_format|utm_id|utm_marketing_tactic|utm_medium|utm_source|utm_source_platform|utm_term|vmcid|wbraid|yclid|zanpid)=") {
        set req.url = regsuball(req.url, "(_branch_match_id|_bta_[a-z]+|_bta_c|_bta_tid|_ga|_gl|_ke|_kx|campid|cof|customid|cx|dclid|dm_i|ef_id|epik|fbclid|gad_source|gbraid|gclid|gclsrc|gdffi|gdfms|gdftrk|hsa_acc|hsa_ad|hsa_cam|hsa_grp|hsa_kw|hsa_mt|hsa_net|hsa_src|hsa_tgt|hsa_ver|ie|igshid|irclickid|matomo_campaign|matomo_cid|matomo_content|matomo_group|matomo_keyword|matomo_medium|matomo_placement|matomo_source|mc_[a-z]+|mc_cid|mc_eid|mkcid|mkevt|mkrid|mkwid|msclkid|mtm_campaign|mtm_cid|mtm_content|mtm_group|mtm_keyword|mtm_medium|mtm_placement|mtm_source|nb_klid|ndclid|origin|pcrid|piwik_campaign|piwik_keyword|piwik_kwd|pk_campaign|pk_keyword|pk_kwd|redirect_log_mongo_id|redirect_mongo_id|rtid|s_kwcid|sb_referer_host|sccid|si|siteurl|sms_click|sms_source|sms_uph|srsltid|toolid|trk_contact|trk_module|trk_msg|trk_sid|ttclid|twclid|utm_[a-z]+|utm_campaign|utm_content|utm_creative_format|utm_id|utm_marketing_tactic|utm_medium|utm_source|utm_source_platform|utm_term|vmcid|wbraid|yclid|zanpid)=[-_A-z0-9+(){}%.*]+&?", "");
        set req.url = regsub(req.url, "[?|&]+$", "");
    }

    # Only cache GET and HEAD requests
    if (req.method != "GET" && req.method != "HEAD") {
        set req.http.X-Cacheable = "NO:REQUEST-METHOD";
        return(pass);
    }

    # Mark static files with the X-Static-File header, and remove any cookies
    # X-Static-File is also used in vcl_backend_response to identify static files
    if (req.url ~ "^[^?]*\.(7z|avi|bmp|bz2|css|csv|doc|docx|eot|flac|flv|gif|gz|ico|jpeg|jpg|js|less|mka|mkv|mov|mp3|mp4|mpeg|mpg|odt|ogg|ogm|opus|otf|pdf|png|ppt|pptx|rar|rtf|svg|svgz|swf|tar|tbz|tgz|ttf|txt|txz|wav|webm|webp|woff|woff2|xls|xlsx|xml|xz|zip)(\?.*)?$") {
        set req.http.X-Static-File = "true";
        unset req.http.Cookie;
        return(hash);
    }

    # No caching of special URLs, logged in users and some plugins
    if (
        req.http.Cookie ~ "wordpress_(?!test_)[a-zA-Z0-9_]+|wp-postpass|comment_author_[a-zA-Z0-9_]+|woocommerce_cart_hash|woocommerce_items_in_cart|wp_woocommerce_session_[a-zA-Z0-9]+|wordpress_logged_in_|comment_author|PHPSESSID" ||
        req.http.Authorization ||
        req.url ~ "add_to_cart" ||
        req.url ~ "edd_action" ||
        req.url ~ "nocache" ||
        req.url ~ "^/addons" ||
        req.url ~ "^/bb-admin" ||
        req.url ~ "^/bb-login.php" ||
        req.url ~ "^/bb-reset-password.php" ||
        req.url ~ "^/cart" ||
        req.url ~ "^/checkout" ||
        req.url ~ "^/control.php" ||
        req.url ~ "^/login" ||
        req.url ~ "^/logout" ||
        req.url ~ "^/lost-password" ||
        req.url ~ "^/my-account" ||
        req.url ~ "^/product" ||
        req.url ~ "^/register" ||
        req.url ~ "^/register.php" ||
        req.url ~ "^/server-status" ||
        req.url ~ "^/signin" ||
        req.url ~ "^/signup" ||
        req.url ~ "^/stats" ||
        req.url ~ "^/wc-api" ||
        req.url ~ "^/wp-admin" ||
        req.url ~ "^/wp-comments-post.php" ||
        req.url ~ "^/wp-cron.php" ||
        req.url ~ "^/wp-login.php" ||
        req.url ~ "^/wp-activate.php" ||
        req.url ~ "^/wp-mail.php" ||
        req.url ~ "^/wp-login.php" ||
        req.url ~ "^\?add-to-cart=" ||
        req.url ~ "^\?wc-api=" ||
        req.url ~ "^/preview=" ||
        req.url ~ "^/\.well-known/acme-challenge/"
    ) {
	     set req.http.X-Cacheable = "NO:Logged in/Got Sessions";
	     if(req.http.X-Requested-With == "XMLHttpRequest") {
		     set req.http.X-Cacheable = "NO:Ajax";
	     }
        return(pass);
    }

    # Remove any cookies left
    unset req.http.Cookie;
    return(hash);
}


sub vcl_hash {
    if(req.http.X-Forwarded-Proto) {
        # Create cache variations depending on the request protocol
        hash_data(req.http.X-Forwarded-Proto);
    }
}

sub vcl_pipe {
  if (req.http.upgrade) {
    set bereq.http.upgrade = req.http.upgrade;
  }
  return (pipe);
}

sub vcl_backend_response {
    # Inject URL & Host header into the object for asynchronous banning purposes
    set beresp.http.x-url = bereq.url;
    set beresp.http.x-host = bereq.http.host;

    # https://support.torproject.org/onionservices/onion-location/
    set beresp.http.Onion-Location = beresp.http.Onion-Location;

    # If we dont get a Cache-Control header from the backend
    # we default to 1h cache for all objects
    if (!beresp.http.Cache-Control) {
        set beresp.ttl = 1h;
        set beresp.http.X-Cacheable = "YES:Forced";
    }

    # If the file is marked as static we cache it for 1 day
    if (bereq.http.X-Static-File == "true") {
        unset beresp.http.Set-Cookie;
        set beresp.http.X-Cacheable = "YES:Forced";
        set beresp.ttl = 1d;
    }

    # Remove the Set-Cookie header when a specific Wordfence cookie is set
    if (beresp.http.Set-Cookie ~ "wfvt_|wordfence_verifiedHuman") {
	    unset beresp.http.Set-Cookie;
    }

    if (beresp.http.Set-Cookie) {
        set beresp.http.X-Cacheable = "NO:Got Cookies";
    } elseif(beresp.http.Cache-Control ~ "private") {
        set beresp.http.X-Cacheable = "NO:Cache-Control=private";
    }
}


sub vcl_deliver {
    # Debug header
    if(req.http.X-Cacheable) {
        set resp.http.X-Cacheable = req.http.X-Cacheable;
    } elseif(obj.uncacheable) {
        if(!resp.http.X-Cacheable) {
            set resp.http.X-Cacheable = "NO:UNCACHEABLE";
        }
    } elseif(!resp.http.X-Cacheable) {
        set resp.http.X-Cacheable = "YES";
    }

### uncomment to remove from responses ###
#  unset resp.http.X-Powered-By;
#  unset resp.http.Server;
#  unset resp.http.server;
#  unset resp.http.via;
#  unset resp.http.x-powered-by;
#  unset resp.http.x-runtime;
#  unset resp.http.x-varnish;
    unset resp.http.x-url;
    unset resp.http.x-host;

  return (deliver);
}
