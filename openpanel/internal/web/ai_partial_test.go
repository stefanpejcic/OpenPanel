package web

import (
	"bytes"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
)

// pages list _header.html by hand all over the modules, the AI drawer has to come along without each of them listing it
func TestHeaderRendersAIAssistantOnlyWhenAllowed(t *testing.T) {
	page := MustLoadPage("partials/_header.html", "partials/_search.html", "partials/_service.html", "partials/_service_js.html")
	tr := i18n.NewManager(t.TempDir(), nil).Translator("en")

	for _, allowed := range []bool{true, false} {
		var b bytes.Buffer
		data := LayoutData{UserAllowed: map[string]bool{"ai_assistant": allowed}, T: tr, RequestPath: "/dashboard"}
		if err := page.tmpl.ExecuteTemplate(&b, "header", data); err != nil {
			t.Fatalf("allowed=%v: %v", allowed, err)
		}
		if got := bytes.Contains(b.Bytes(), []byte(`id="ai-assistant-btn"`)); got != allowed {
			t.Errorf("allowed=%v: button rendered = %v", allowed, got)
		}
	}
}
