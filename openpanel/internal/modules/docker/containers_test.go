package docker

import "testing"

// TestFilterContainerServicesKeepsActiveWebserver guards against the
// openlitespeed-hides-itself regression: "litespeed" is a substring of
// "openlitespeed", so a Contains-based filter removes the active webserver
// along with the others it's meant to hide.
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
