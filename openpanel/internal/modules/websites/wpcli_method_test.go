package websites

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWordPressWPCLIWriteActionsNeedPost(t *testing.T) {
	for action := range wsWriteWPCLIActions {
		if !wsAllowedWPCLIActions[action] {
			t.Errorf("write action %q is missing from wsAllowedWPCLIActions", action)
		}
		req := httptest.NewRequest(http.MethodGet, "/wordpress/wp-cli/"+action+"?website=example.com&docroot=/var/www/html/example.com", nil)
		req.SetPathValue("action", action)
		rec := httptest.NewRecorder()
		// a nil app is fine, the method check has to happen before anything touches it
		handleWordPressWPCLI(nil, rec, req)
		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("GET %s = %d, want 405", action, rec.Code)
		}
	}
}
