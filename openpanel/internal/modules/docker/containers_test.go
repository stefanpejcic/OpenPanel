package docker

import "testing"

// guards against the openlitespeed-hides-itself regression: "litespeed" is a substring of "openlitespeed", so a Contains-based filter removes the active webserver along with the others it's meant to hide
func TestFilterContainerServicesKeepsActiveWebserver(t *testing.T) {
	allWebservers := []string{"apache", "nginx", "openresty", "openlitespeed", "litespeed"}

	for _, active := range allWebservers {
		services := map[string]any{}
		for _, ws := range allWebservers {
			services[ws] = map[string]any{}
		}

		filtered := filterContainerServices(services, active, "mariadb")

		if _, ok := filtered[active]; !ok {
			t.Errorf("active webserver %q was filtered out of its own containers list", active)
		}
		for _, ws := range allWebservers {
			if ws == active {
				continue
			}
			if _, ok := filtered[ws]; ok {
				t.Errorf("webserver=%q: inactive webserver %q should have been hidden but was present", active, ws)
			}
		}
	}
}

func TestFilterContainerServicesMySQLVariant(t *testing.T) {
	services := map[string]any{
		"apache":  map[string]any{},
		"mysql":   map[string]any{},
		"mariadb": map[string]any{},
	}

	if filtered := filterContainerServices(services, "apache", "mysql"); filtered["mariadb"] != nil {
		t.Error("mysqlType=mysql: mariadb should be hidden")
	} else if filtered["mysql"] == nil {
		t.Error("mysqlType=mysql: mysql should remain")
	}

	if filtered := filterContainerServices(services, "apache", "mariadb"); filtered["mysql"] != nil {
		t.Error("mysqlType=mariadb: mysql should be hidden")
	} else if filtered["mariadb"] == nil {
		t.Error("mysqlType=mariadb: mariadb should remain")
	}
}

func TestKeepUneditedKeys(t *testing.T) {
	old := map[string]any{
		"image":       "weblate/weblate:latest",
		"labels":      map[string]any{"openpanel.domain": "translate.example.com"},
		"extra_hosts": []any{"mail.example.com:169.254.1.2"},
		"environment": map[string]any{"OLD": "1"},
	}
	updated := map[string]any{"image": "weblate/weblate:5", "environment": map[string]any{"NEW": "1"}}
	keepUneditedKeys(old, updated)
	if hosts, _ := updated["extra_hosts"].([]any); len(hosts) != 1 || hosts[0] != "mail.example.com:169.254.1.2" {
		t.Fatalf("extra_hosts not kept: %#v", updated["extra_hosts"])
	}
	if _, ok := updated["labels"]; !ok {
		t.Fatal("labels not kept")
	}
	if updated["image"] != "weblate/weblate:5" || updated["environment"].(map[string]any)["NEW"] != "1" {
		t.Fatalf("edited keys overwritten: %#v", updated)
	}

	plain := map[string]any{}
	keepUneditedKeys(map[string]any{"image": "x"}, plain)
	if len(plain) != 0 {
		t.Fatalf("added keys the service didn't have: %#v", plain)
	}
}
