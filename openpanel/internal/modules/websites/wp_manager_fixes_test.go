package websites

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/modules/domains"
)

func pngOf(t *testing.T, img image.Image) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestBlankScreenshot(t *testing.T) {
	white := image.NewRGBA(image.Rect(0, 0, 1280, 720))
	for i := range white.Pix {
		white.Pix[i] = 0xff
	}
	if !blankScreenshot(pngOf(t, white)) {
		t.Error("an all-white capture should count as blank")
	}

	// a stray dot doesn't make a page
	white.Set(10, 10, color.Black)
	if !blankScreenshot(pngOf(t, white)) {
		t.Error("one dark pixel should still count as blank")
	}

	page := image.NewRGBA(image.Rect(0, 0, 1280, 720))
	for y := 0; y < 720; y++ {
		for x := 0; x < 1280; x++ {
			c := color.RGBA{0xff, 0xff, 0xff, 0xff}
			if y < 80 || (y > 200 && y < 230 && x > 100 && x < 900) {
				c = color.RGBA{0x21, 0x75, 0x9b, 0xff}
			}
			page.Set(x, y, c)
		}
	}
	if blankScreenshot(pngOf(t, page)) {
		t.Error("a page with a header and text isn't blank")
	}
	if blankScreenshot([]byte("not an image")) {
		t.Error("undecodable data is left to the browser, not called blank")
	}
}

// the Options tab saves from the POST body, the API's query string form still works
func TestSiteInfoOptionsFromBody(t *testing.T) {
	body := url.Values{"website": {"example.com"}, "blogname": {"My site"}, "admin_email": {"a@example.com"}, "blog_public": {"0"}}
	post := httptest.NewRequest(http.MethodPost, "/wordpress/wp-cli/update_site_information", strings.NewReader(body.Encode()))
	post.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	got := map[string]string{}
	for _, kv := range siteInfoOptionsFrom(post) {
		got[kv[0]] = kv[1]
	}
	if len(got) != 3 || got["blogname"] != "My site" || got["admin_email"] != "a@example.com" || got["blog_public"] != "0" {
		t.Errorf("options from body = %v", got)
	}

	get := httptest.NewRequest(http.MethodPost, "/api/sites/wp-cli/update_site_information?blogdescription=Hi&website=example.com", nil)
	opts := siteInfoOptionsFrom(get)
	if len(opts) != 1 || opts[0] != [2]string{"blogdescription", "Hi"} {
		t.Errorf("options from query = %v", opts)
	}
}

func TestSSLState(t *testing.T) {
	cases := []struct {
		https string
		cert  domains.SSLCertInfo
		want  string
	}{
		{"Unknown", domains.SSLCertInfo{}, ""},
		{"Automatic", domains.SSLCertInfo{Expires: "2027-01-01"}, "valid"},
		{"Custom", domains.SSLCertInfo{Problem: "expiring"}, "expiring"},
		{"Automatic", domains.SSLCertInfo{Problem: "expired"}, "expired"},
		{"Automatic", domains.SSLCertInfo{Problem: "missing"}, "missing"},
		{"Automatic", domains.SSLCertInfo{Problem: "unreadable"}, "missing"},
	}
	for _, c := range cases {
		if got := sslState(c.https, c.cert); got != c.want {
			t.Errorf("sslState(%q, %+v) = %q, want %q", c.https, c.cert, got, c.want)
		}
	}
}
