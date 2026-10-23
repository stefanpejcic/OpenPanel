package fixpermissions

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAPIDirectory(t *testing.T) {
	var mp bytes.Buffer
	mw := multipart.NewWriter(&mp)
	_ = mw.WriteField("directory", "/var/www/html/mp")
	_ = mw.Close()

	cases := []struct {
		name, target, contentType, body, want string
	}{
		{"json", "/x", "application/json", `{"directory":"/var/www/html/site1"}`, "/var/www/html/site1"},
		{"json without content type", "/x", "", `{"directory":"/var/www/html/site1"}`, "/var/www/html/site1"},
		// the form body used to be drained by the json attempt, so the whole account got fixed
		{"form", "/x", "application/x-www-form-urlencoded", "directory=%2Fvar%2Fwww%2Fhtml%2Fsite1", "/var/www/html/site1"},
		{"form with charset", "/x", "application/x-www-form-urlencoded; charset=UTF-8", "directory=%2Fvar%2Fwww%2Fhtml%2Fsite1", "/var/www/html/site1"},
		{"multipart", "/x", mw.FormDataContentType(), mp.String(), "/var/www/html/mp"},
		{"query string only", "/x?directory=%2Fvar%2Fwww%2Fhtml%2Fq", "", "", "/var/www/html/q"},
		{"nothing", "/x", "", "", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, c.target, strings.NewReader(c.body))
			if c.contentType != "" {
				r.Header.Set("Content-Type", c.contentType)
			}
			if got := apiDirectory(r); got != c.want {
				t.Fatalf("got %q, want %q", got, c.want)
			}
		})
	}
}
