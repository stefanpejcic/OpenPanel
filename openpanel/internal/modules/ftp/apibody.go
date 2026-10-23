package ftp

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

// ftpAPIBody is what the create, password, path and quota endpoints read
type ftpAPIBody struct {
	Username, Password, Domain, Path, Quota string
}

// readFTPAPIBody reads the JSON or form body, a JSON quota can be a number or a string
func readFTPAPIBody(r *http.Request) ftpAPIBody {
	var body struct {
		Username string          `json:"username"`
		Password string          `json:"password"`
		Domain   string          `json:"domain"`
		Path     string          `json:"path"`
		Quota    json.RawMessage `json:"quota"`
	}
	var out ftpAPIBody
	formBody := false
	decodeAPIBody(r, &body, func(form url.Values) {
		formBody = true
		out = ftpAPIBody{Username: form.Get("username"), Password: form.Get("password"), Domain: form.Get("domain"), Path: form.Get("path"), Quota: form.Get("quota")}
	})
	if !formBody {
		out = ftpAPIBody{Username: body.Username, Password: body.Password, Domain: body.Domain, Path: body.Path, Quota: rawQuotaString(body.Quota)}
	}
	return out
}
