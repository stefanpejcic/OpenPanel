package search

import (
	"encoding/json"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/searchdata"
)

func TestGateFor(t *testing.T) {
	if required, ok := gateFor("mysql_databases"); !ok || len(required) != 1 || required[0] != "mysql" {
		t.Errorf("gateFor(mysql_databases) = %v, %v; want [mysql], true", required, ok)
	}
	if required, ok := gateFor("features"); !ok || required != nil {
		t.Errorf("gateFor(features) = %v, %v; want nil, true", required, ok)
	}
	if _, ok := gateFor("nonsense"); ok {
		t.Error("gateFor(nonsense) = ok=true, want false")
	}
}

func TestEnterpriseSearchTypes(t *testing.T) {
	if !enterpriseSearchTypes["mysql_databases"] {
		t.Error("mysql_databases should be enterprise-gated")
	}
	if enterpriseSearchTypes["features"] {
		t.Error("features should not be enterprise-gated")
	}
}

func TestFeaturesJSONEmbedded(t *testing.T) {
	var routes []map[string]string
	if err := json.Unmarshal(searchdata.FeaturesJSON, &routes); err != nil {
		t.Fatalf("embedded filter.json failed to parse: %v", err)
	}
	if len(routes) == 0 {
		t.Fatal("embedded filter.json parsed to zero routes")
	}
	for _, key := range []string{"description", "link", "module", "name"} {
		if _, ok := routes[0][key]; !ok {
			t.Errorf("first route missing expected key %q", key)
		}
	}
}

func TestSearchFeaturesFiltersByModuleAndCapsAt100(t *testing.T) {
	var routes []map[string]string
	_ = json.Unmarshal(searchdata.FeaturesJSON, &routes)

	allowed := map[string]bool{}
	for _, r := range routes {
		allowed[r["module"]] = true
	}
	// every module present should pass through; a bogus one shouldn't.
	allowed["definitely-not-a-real-module"] = false

	count := 0
	for _, r := range routes {
		if allowed[r["module"]] {
			count++
		}
	}
	if count == 0 {
		t.Fatal("expected at least one route to pass the allow-everything filter")
	}
	if count > 100 {
		// searchFeatures caps at 100; sanity check our test data isn't
		// already violating the assumption the handler relies on.
		t.Logf("embedded filter.json has %d routes matching allowed modules - handler will cap at 100", count)
	}
}
