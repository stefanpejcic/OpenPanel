package docker

import (
	"net/url"
	"strings"
)

// templateFormValues turns /containers/new?image=..&port=.. links into form values, nil when the link has none
func templateFormValues(q url.Values, availableNetworks []string) url.Values {
	known := []string{"image", "service_name", "name", "cpu", "ram", "pids", "network", "environment", "env", "volume", "add_socket", "socket", "healthcheck", "domain", "port"}
	found := false
	for _, k := range known {
		if q.Has(k) {
			found = true
			break
		}
	}
	if !found {
		return nil
	}

	v := url.Values{}
	v.Set("image", strings.TrimSpace(q.Get("image")))
	name := q.Get("service_name")
	if name == "" {
		name = q.Get("name")
	}
	v.Set("service_name", strings.ToLower(strings.TrimSpace(name)))
	v.Set("cpu", firstNonEmpty(q.Get("cpu"), "0.5"))
	// plain numbers are GB, same as the form
	ram := strings.ToUpper(strings.TrimSpace(q.Get("ram")))
	if ram != "" && !strings.HasSuffix(ram, "M") && !strings.HasSuffix(ram, "G") {
		ram += "G"
	}
	v.Set("ram", firstNonEmpty(ram, "1G"))
	v.Set("pids", firstNonEmpty(q.Get("pids"), "500"))

	for _, raw := range q["network"] {
		for _, n := range strings.Split(raw, ",") {
			if n = strings.TrimSpace(n); containsString(availableNetworks, n) && !containsString(v["network"], n) {
				v.Add("network", n)
			}
		}
	}
	if len(v["network"]) == 0 && len(availableNetworks) > 0 {
		v.Add("network", availableNetworks[0])
	}

	env := q["env"]
	if e := q.Get("environment"); e != "" {
		env = append(env, strings.Split(e, "\n")...)
	}
	v.Set("environment", strings.Join(env, "\n"))

	// volume=name:/mount or name:/mount:ro
	for _, vol := range q["volume"] {
		parts := strings.SplitN(strings.TrimSpace(vol), ":", 3)
		if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
			continue
		}
		ro := "off"
		if len(parts) == 3 && parts[2] == "ro" {
			ro = "on"
		}
		v.Add("volume_name", parts[0])
		v.Add("volume_mount", parts[1])
		v.Add("volume_readonly", ro)
	}

	if truthy(q.Get("add_socket")) || truthy(q.Get("socket")) {
		v.Set("add_socket", "on")
	}
	v.Set("healthcheck", q.Get("healthcheck"))
	v.Set("domain", strings.ToLower(strings.TrimSpace(q.Get("domain"))))
	v.Set("port", strings.TrimSpace(q.Get("port")))
	return v
}

func firstNonEmpty(s, def string) string {
	if s = strings.TrimSpace(s); s != "" {
		return s
	}
	return def
}

func truthy(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "on", "true", "yes":
		return true
	}
	return false
}
