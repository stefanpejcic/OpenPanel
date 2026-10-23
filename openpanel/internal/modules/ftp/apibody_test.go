package ftp

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestReadFTPAPIBody(t *testing.T) {
	var mp bytes.Buffer
	mw := multipart.NewWriter(&mp)
	_ = mw.WriteField("quota", "2G")
	_ = mw.WriteField("path", "/var/www/html/mp")
	_ = mw.Close()

	cases := []struct {
		name, target, contentType, body string
		want                            ftpAPIBody
	}{
		{"json number quota", "/x", "application/json", `{"quota":500}`, ftpAPIBody{Quota: "500"}},
		{"json string quota", "/x", "application/json", `{"quota":"2G","path":"/var/www/html/a"}`, ftpAPIBody{Quota: "2G", Path: "/var/www/html/a"}},
		{"json without content type", "/x", "", `{"password":"a!b"}`, ftpAPIBody{Password: "a!b"}},
		// the form body used to be drained by the json attempt, so quota=500 removed the quota
		{"form", "/x", "application/x-www-form-urlencoded", "quota=500&path=%2Fvar%2Fwww%2Fhtml%2Fsite", ftpAPIBody{Quota: "500", Path: "/var/www/html/site"}},
		{"form with charset", "/x", "application/x-www-form-urlencoded; charset=UTF-8", "quota=2G", ftpAPIBody{Quota: "2G"}},
		{"form create", "/x", "application/x-www-form-urlencoded", "username=bob&password=p&domain=ex.com&path=%2Fvar%2Fwww%2Fhtml%2F", ftpAPIBody{Username: "bob", Password: "p", Domain: "ex.com", Path: "/var/www/html/"}},
		{"multipart", "/x", mw.FormDataContentType(), mp.String(), ftpAPIBody{Quota: "2G", Path: "/var/www/html/mp"}},
		{"query string only", "/x?quota=1G", "", "", ftpAPIBody{Quota: "1G"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPatch, c.target, strings.NewReader(c.body))
			if c.contentType != "" {
				r.Header.Set("Content-Type", c.contentType)
			}
			if got := readFTPAPIBody(r); got != c.want {
				t.Fatalf("got %+v, want %+v", got, c.want)
			}
		})
	}
}
