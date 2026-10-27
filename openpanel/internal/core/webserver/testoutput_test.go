package webserver

import (
	"reflect"
	"testing"
)

func TestConfigTestProblems(t *testing.T) {
	cases := []struct {
		name    string
		service string
		output  string
		want    []ConfigProblem
	}{
		{
			name:    "nginx error printed twice with a timestamp",
			service: "openresty",
			output: "2026/10/27 13:10:17 [emerg] 14#14: \"http\" directive is duplicate in /usr/local/openresty/nginx/conf/nginx.conf:83\n" +
				"nginx: [emerg] \"http\" directive is duplicate in /usr/local/openresty/nginx/conf/nginx.conf:83\n" +
				"nginx: configuration file /usr/local/openresty/nginx/conf/nginx.conf test failed",
			want: []ConfigProblem{{File: "/usr/local/openresty/nginx/conf/nginx.conf", Line: 83, Message: `"http" directive is duplicate`}},
		},
		{
			name:    "nginx lookup failure in a vhost, warnings dropped",
			service: "nginx",
			output: "2026/10/27 13:11:14 [warn] 1#1: the \"listen ... http2\" directive is deprecated in /etc/nginx/conf.d/a.com.conf:3\n" +
				"2026/10/27 13:11:14 [emerg] 1#1: host not found in upstream \"php-fpm-8.5\" in /etc/nginx/conf.d/a.com.conf:35\n" +
				"nginx: configuration file /etc/nginx/nginx.conf test failed",
			want: []ConfigProblem{{File: "/etc/nginx/conf.d/a.com.conf", Line: 35, Message: `host not found in upstream "php-fpm-8.5"`}},
		},
		{
			name:    "apache error on one line",
			service: "apache",
			output:  "httpd: Syntax error on line 167 of /usr/local/apache2/conf/httpd.conf: /usr/local/apache2/conf/httpd.conf:167: <VirtualHost> was not closed.",
			want:    []ConfigProblem{{File: "/usr/local/apache2/conf/httpd.conf", Line: 167, Message: "<VirtualHost> was not closed."}},
		},
		{
			name:    "apache reason on the next line",
			service: "apache",
			output: "AH00558: httpd: Could not reliably determine the server's fully qualified domain name\n" +
				"AH00526: Syntax error on line 3 of /usr/local/apache2/conf/httpd.conf:\n" +
				"Invalid command 'Foo', perhaps misspelled or defined by a module not included in the server configuration",
			want: []ConfigProblem{{File: "/usr/local/apache2/conf/httpd.conf", Line: 3, Message: "Invalid command 'Foo', perhaps misspelled or defined by a module not included in the server configuration"}},
		},
		{
			name:    "apache error in an included vhost",
			service: "apache",
			output:  "httpd: Syntax error on line 600 of /usr/local/apache2/conf/httpd.conf: Syntax error on line 5 of /etc/httpd/conf.d/a.com.conf: Invalid command 'Foo'",
			want:    []ConfigProblem{{File: "/etc/httpd/conf.d/a.com.conf", Line: 5, Message: "Invalid command 'Foo'"}},
		},
		{
			name:    "openlitespeed error among warnings",
			service: "openlitespeed",
			output: "2026-10-27 12:55:40.310143 [WARN] [91] [config:server:tuning] invalid value of <maxCachedFileSize>:65536, use default=4096\n" +
				"2026-10-27 13:15:39.829106 [ERROR] [config:server:basics] Invalid User Name((null)) or Group Name((null))!\n" +
				"2026-10-27 13:15:39.829106 [ERROR] [29] Path for certificate file is invalid: /x/fullchain.cer",
			want: []ConfigProblem{
				{Message: "[config:server:basics] Invalid User Name((null)) or Group Name((null))!"},
				{Message: "Path for certificate file is invalid: /x/fullchain.cer"},
			},
		},
		{
			name:    "test that printed nothing",
			service: "openlitespeed",
			output:  TestNoOutput + ": exit status 1",
			want:    []ConfigProblem{{Message: TestNoOutput + ": exit status 1"}},
		},
		{
			name:    "unrecognized output keeps its last lines",
			service: "nginx",
			output:  "a\nb\nc\nd\ne\nf\nError: no such container",
			want:    []ConfigProblem{{Message: "c"}, {Message: "d"}, {Message: "e"}, {Message: "f"}, {Message: "Error: no such container"}},
		},
	}
	for _, c := range cases {
		if got := ConfigTestProblems(c.service, c.output); !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s:\n got  %#v\n want %#v", c.name, got, c.want)
		}
	}
}
