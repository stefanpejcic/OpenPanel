package filemanager

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRedirectPathsEscapeNames(t *testing.T) {
	// unescaped, # started a fragment, ? a query and % a broken escape, so the redirect opened the wrong file
	cases := map[string]string{
		"notes#1.txt":     "notes%231.txt",
		"100%.txt":        "100%25.txt",
		"what?.txt":       "what%3F.txt",
		"my site/a b.txt": "my%20site/a%20b.txt",
		"c++/a&b.txt":     "c++/a&b.txt",
	}
	mux := http.NewServeMux()
	var got string
	mux.HandleFunc("/file-manager/edit-file/{file_path...}", func(w http.ResponseWriter, r *http.Request) {
		got = r.PathValue("file_path")
	})
	for name, want := range cases {
		escaped := escapeURLPath(name)
		if escaped != want {
			t.Errorf("escapeURLPath(%q) = %q, want %q", name, escaped, want)
		}
		got = ""
		mux.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/file-manager/edit-file/"+escaped, nil))
		if got != name {
			t.Errorf("redirect for %q opens %q", name, got)
		}
	}
	if p := filesRedirectPath("blog/#drafts"); p != "/files/blog/%23drafts" {
		t.Errorf("filesRedirectPath = %q", p)
	}
	if p := filesRedirectPath(""); p != "/files" {
		t.Errorf("filesRedirectPath(\"\") = %q", p)
	}
}
