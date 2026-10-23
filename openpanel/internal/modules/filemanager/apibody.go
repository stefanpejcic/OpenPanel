package filemanager

import (
	"encoding/json"
	"mime"
	"net/http"
	"net/url"
)

// decodeAPIBody reads the body once, as a form for form posts and as JSON otherwise
func decodeAPIBody(r *http.Request, dst any, fromForm func(url.Values)) {
	ct, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	switch ct {
	case "application/x-www-form-urlencoded":
		_ = r.ParseForm()
	case "multipart/form-data":
		_ = r.ParseMultipartForm(32 << 20)
	default:
		if json.NewDecoder(r.Body).Decode(dst) == nil {
			return
		}
		// no JSON body, fields in the query string still work
		_ = r.ParseForm()
	}
	fromForm(r.Form)
}
