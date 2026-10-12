package dockertemplates

import (
	"net/url"
	"strings"
	"testing"
)

const sample = `
applications:
  - id: uptime-kuma
    name: Uptime Kuma
    image: louislam/uptime-kuma:2
    service: uptimekuma
    ram: "0.5"
    port: 3001
    networks: [www, db]
    volumes: [uptimekuma_data:/app/data]
    env:
      - TZ=UTC
      - URL=https://{domain}/
    logo: http://insecure.example/logo.svg
  - id: uptime-kuma
    name: duplicate
    image: x:1
    service: dup
  - id: bad
    name: Bad image
    image: "x; rm -rf /"
    service: bad
  - id: bad2
    name: Bad service
    image: nginx
    service: No
`

func TestParse(t *testing.T) {
	list, err := Parse([]byte(sample))
	if err != nil || len(list) != 1 {
		t.Fatalf("got %v %v", list, err)
	}
	if list[0].Logo != "" {
		t.Error("http logo should be dropped")
	}
	if _, err := Parse([]byte("applications: []")); err == nil {
		t.Error("empty list should error so the cache isn't filled with nothing")
	}
}

func TestInstallURL(t *testing.T) {
	list, _ := Parse([]byte(sample))
	u, err := url.Parse(list[0].InstallURL())
	if err != nil || u.Path != "/containers/new" {
		t.Fatalf("bad url %v", u)
	}
	q := u.Query()
	if q.Get("image") != "louislam/uptime-kuma:2" || q.Get("name") != "uptimekuma" || q.Get("port") != "3001" ||
		q.Get("network") != "www,db" || q.Get("volume") != "uptimekuma_data:/app/data" || len(q["env"]) != 2 || q.Get("ram") != "0.5" {
		t.Fatalf("got %v", q)
	}
	if !strings.Contains(q["env"][1], "{domain}") {
		t.Error("placeholder lost")
	}
}
