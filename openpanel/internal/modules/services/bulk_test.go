package services

import (
	"net/http/httptest"
	"testing"
)

func TestRunServiceBulkActionRejects(t *testing.T) {
	r := httptest.NewRequest("POST", "/services/bulk", nil)
	if res := runServiceBulkAction(r, "restart", "apache", "testuser", []string{"nginx"}); res.OK || res.Message != "Service not found." {
		t.Errorf("service outside the list = %+v, want not found", res)
	}
	if res := runServiceBulkAction(r, "delete", "nginx", "testuser", []string{"nginx"}); res.OK || res.Message != "Unknown bulk action." {
		t.Errorf("unknown action = %+v", res)
	}
}
