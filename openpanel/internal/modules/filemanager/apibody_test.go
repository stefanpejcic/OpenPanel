package filemanager

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestDecodeAPIBody(t *testing.T) {
	var mp bytes.Buffer
	mw := multipart.NewWriter(&mp)
	_ = mw.WriteField("editor_content", "from multipart")
	_ = mw.Close()

	cases := []struct {
		name, target, contentType, body, want string
	}{
		{"json", "/x", "application/json", `{"content":"from json"}`, "from json"},
		{"json without content type", "/x", "", `{"content":"from json"}`, "from json"},
		// the form body used to be drained by the json attempt, so an empty file got saved
		{"form", "/x", "application/x-www-form-urlencoded", "editor_content=from+form", "from form"},
		{"form with charset", "/x", "application/x-www-form-urlencoded; charset=UTF-8", "editor_content=from+form", "from form"},
		{"multipart", "/x", mw.FormDataContentType(), mp.String(), "from multipart"},
		{"query string only", "/x?editor_content=from+query", "", "", "from query"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPut, c.target, strings.NewReader(c.body))
			if c.contentType != "" {
				r.Header.Set("Content-Type", c.contentType)
			}
			var body struct {
				Content string `json:"content"`
			}
			decodeAPIBody(r, &body, func(form url.Values) {
				body.Content = form.Get("editor_content")
			})
			if body.Content != c.want {
				t.Fatalf("content = %q, want %q", body.Content, c.want)
			}
		})
	}
}

func TestAPIContentBody(t *testing.T) {
	cases := []struct {
		name, contentType, body, want string
		ok                            bool
	}{
		{"json", "application/json", `{"content":"a"}`, "a", true},
		{"json empty content", "application/json", `{"content":""}`, "", true},
		{"json missing content", "application/json", `{"text":"a"}`, "", false},
		{"form editor_content", "application/x-www-form-urlencoded", "editor_content=a", "a", true},
		{"form content", "application/x-www-form-urlencoded", "content=a", "a", true},
		{"form empty content", "application/x-www-form-urlencoded", "content=", "", true},
		// this one used to save an empty file and report success
		{"form missing content", "application/x-www-form-urlencoded", "path=a", "", false},
		{"no body", "", "", "", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPut, "/x", strings.NewReader(c.body))
			if c.contentType != "" {
				r.Header.Set("Content-Type", c.contentType)
			}
			got, ok := apiContentBody(r)
			if got != c.want || ok != c.ok {
				t.Fatalf("got (%q, %v), want (%q, %v)", got, ok, c.want, c.ok)
			}
		})
	}
}
