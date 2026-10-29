package webserverconf

import (
	"reflect"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
)

func TestDescribeTestFailure(t *testing.T) {
	tr := translatorTr(i18n.NewManager(t.TempDir(), nil).Translator("en"))

	t.Run("error in the edited file", func(t *testing.T) {
		f := describeTestFailure(tr, lookupWebserverConf("openresty"),
			"2026/10/27 13:10:17 [emerg] 14#14: \"http\" directive is duplicate in /usr/local/openresty/nginx/conf/nginx.conf:83\n"+
				"nginx: [emerg] \"http\" directive is duplicate in /usr/local/openresty/nginx/conf/nginx.conf:83\n"+
				"nginx: configuration file /usr/local/openresty/nginx/conf/nginx.conf test failed")
		want := []string{`openresty.conf, line 83: "http" directive is duplicate`}
		if !reflect.DeepEqual(f.Problems, want) || f.Line != 83 || f.VhostDomain != "" || len(f.Hints) != 0 {
			t.Errorf("got %+v, want problems %q on line 83 and no hints", f, want)
		}
	})

	t.Run("stopped upstream in a vhost", func(t *testing.T) {
		f := describeTestFailure(tr, lookupWebserverConf("nginx"),
			"2026/10/27 13:11:14 [emerg] 1#1: host not found in upstream \"php-fpm-8.5\" in /etc/nginx/conf.d/a.com.conf:35")
		want := []string{`a.com.conf, line 35: host not found in upstream "php-fpm-8.5"`}
		if !reflect.DeepEqual(f.Problems, want) || f.Line != 0 || f.VhostDomain != "a.com" {
			t.Fatalf("got %+v, want problems %q in a.com's vhost and no line to highlight", f, want)
		}
		hints := strings.Join(f.Hints, "\n")
		if !strings.Contains(hints, "php-fpm-8.5 could not be found") || !strings.Contains(hints, "VirtualHost file of a.com, not in nginx.conf") {
			t.Errorf("hints = %q, want the missing host and the vhost explained", hints)
		}
	})

	t.Run("apache", func(t *testing.T) {
		f := describeTestFailure(tr, lookupWebserverConf("apache"),
			"httpd: Syntax error on line 167 of /usr/local/apache2/conf/httpd.conf: /usr/local/apache2/conf/httpd.conf:167: <VirtualHost> was not closed.")
		if want := []string{"httpd.conf, line 167: <VirtualHost> was not closed."}; !reflect.DeepEqual(f.Problems, want) || f.Line != 167 {
			t.Errorf("got %+v, want %q on line 167", f, want)
		}
	})

	t.Run("openlitespeed config that doesn't load", func(t *testing.T) {
		f := describeTestFailure(tr, lookupWebserverConf("openlitespeed"), webserver.TestNoOutput+": exit status 1")
		if len(f.Hints) != 1 || !strings.Contains(f.Hints[0], "could not load this configuration") {
			t.Errorf("hints = %q, want the load failure explained", f.Hints)
		}
	})

	t.Run("timeout", func(t *testing.T) {
		f := describeTestFailure(tr, lookupWebserverConf("nginx"), webserver.TestTimedOut)
		if len(f.Hints) != 1 || !strings.Contains(f.Hints[0], "image may still be downloading") {
			t.Errorf("hints = %q, want the timeout explained", f.Hints)
		}
	})
}
