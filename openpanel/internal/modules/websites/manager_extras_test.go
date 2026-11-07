package websites

import (
	"reflect"
	"testing"
)

func TestComposeStartCommand(t *testing.T) {
	cases := []struct {
		name string
		svc  map[string]any
		want string
	}{
		{"pm2 sh -c string", map[string]any{"command": "sh -c \"npm install && node /var/www/html/nodejs.tests.openpanel.org/index.js\"\n"}, "npm install && node /var/www/html/nodejs.tests.openpanel.org/index.js"},
		{"single quoted", map[string]any{"command": "sh -c 'bundle install && bundle exec ruby app.rb'"}, "bundle install && bundle exec ruby app.rb"},
		{"list form", map[string]any{"command": []any{"/bin/sh", "-c", "pip install -r requirements.txt && python app.py"}}, "pip install -r requirements.txt && python app.py"},
		{"plain list", map[string]any{"command": []any{"node", "server.js"}}, "node server.js"},
		{"entrypoint first", map[string]any{"entrypoint": []any{"tini", "--"}, "command": []any{"n8n", "start"}}, "tini -- n8n start"},
		{"no command", map[string]any{"image": "n8nio/n8n:latest"}, ""},
		{"nil service", nil, ""},
	}
	for _, c := range cases {
		if got := composeStartCommand(c.svc); got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
	}
}

func TestComposeMountsAndNetworks(t *testing.T) {
	svc := map[string]any{
		"volumes":  []any{"n8n_data:/home/node/.n8n", "html_data:/var/www/html/:ro", map[string]any{"type": "volume", "source": "extra", "target": "/data"}},
		"networks": []any{"www", "db"},
	}
	wantMounts := []ComposeMount{{"n8n_data", "/home/node/.n8n"}, {"html_data", "/var/www/html/"}, {"extra", "/data"}}
	if got := composeMounts(svc); !reflect.DeepEqual(got, wantMounts) {
		t.Errorf("mounts = %#v", got)
	}
	if got := composeNetworks(svc); !reflect.DeepEqual(got, []string{"www", "db"}) {
		t.Errorf("networks = %#v", got)
	}
	if got := composeNetworks(map[string]any{"networks": map[string]any{"www": nil, "db": map[string]any{}}}); !reflect.DeepEqual(got, []string{"db", "www"}) {
		t.Errorf("map networks = %#v", got)
	}
	if composeMounts(nil) != nil || composeNetworks(nil) != nil {
		t.Error("nil service should give nothing")
	}
	if got := envVarsOf(map[string]any{"environment": []any{"NODE_ENV=production", "TZ=Europe/Berlin"}}); got != "NODE_ENV=production\nTZ=Europe/Berlin" {
		t.Errorf("env = %q", got)
	}
}

func TestParsePackageJSON(t *testing.T) {
	pkg := []byte(`{
  "name": "nodejs-tests",
  "scripts": {"start": "node index.js", "dev": "node --watch index.js"},
  "dependencies": {"express": "^4.21.2", "morgan": "^1.10.0"},
  "devDependencies": {"eslint": "^9.0.0"}
}`)
	lock := []byte(`{"lockfileVersion": 3, "packages": {"": {}, "node_modules/express": {"version": "4.22.3"}, "node_modules/morgan": {"version": "1.10.1"}, "node_modules/foo/node_modules/express": {"version": "3.0.0"}}}`)
	got := parsePackageJSON(pkg, lock)
	want := []ManifestGroup{
		{Title: "Scripts", ValueLabel: "Command", Rows: []ManifestRow{{"dev", "node --watch index.js", ""}, {"start", "node index.js", ""}}},
		{Title: "Dependencies", ValueLabel: "Constraint", HasInstalled: true, Rows: []ManifestRow{{"express", "^4.21.2", "4.22.3"}, {"morgan", "^1.10.0", "1.10.1"}}},
		{Title: "Dev dependencies", ValueLabel: "Constraint", HasInstalled: true, Rows: []ManifestRow{{"eslint", "^9.0.0", ""}}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v", got)
	}
	if g := parsePackageJSON(pkg, nil); g[1].HasInstalled {
		t.Error("no lock file should mean no installed column")
	}
	if parsePackageJSON([]byte("{broken"), nil) != nil {
		t.Error("broken json should give nil")
	}
}

func TestParseRequirements(t *testing.T) {
	data := []byte("flask==3.1.0\npython-dotenv==1.0.1\n# comment\n-r base.txt\nrequests[socks]>=2.31,<3 ; python_version > '3.8'\ngunicorn\ngit+https://github.com/x/y.git\nuvicorn==0.30.0  # pinned\n")
	got := parseRequirements(data)
	want := []ManifestGroup{{Title: "Pinned in requirements.txt", ValueLabel: "Version", Rows: []ManifestRow{
		{"flask", "3.1.0", ""}, {"python-dotenv", "1.0.1", ""}, {"requests", ">=2.31,<3", ""}, {"gunicorn", "any", ""}, {"uvicorn", "0.30.0", ""},
	}}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v", got)
	}
	if parseRequirements([]byte("# nothing\n")) != nil {
		t.Error("empty file should give nil")
	}
}

func TestParseGemfile(t *testing.T) {
	gemfile := []byte("source \"https://rubygems.org\"\n\ngem \"sinatra\", \"~> 4.1\"\ngem \"rackup\", \"~> 2.2\"\ngem 'puma', '>= 6.0', '< 7'\ngem \"rake\"\n")
	lock := []byte("GEM\n  remote: https://rubygems.org/\n  specs:\n    base64 (0.3.0)\n    puma (6.6.1)\n      nio4r (~> 2.0)\n    nio4r (2.7.5)\n    rackup (2.3.1)\n      rack (>= 3)\n    sinatra (4.2.1)\n\nPLATFORMS\n  ruby\n\nDEPENDENCIES\n  puma (~> 6.5)\n")
	got := parseGemfile(gemfile, lock)
	want := []ManifestGroup{
		{Title: "Gems in the Gemfile", ValueLabel: "Constraint", HasInstalled: true, Rows: []ManifestRow{
			{"sinatra", "~> 4.1", "4.2.1"}, {"rackup", "~> 2.2", "2.3.1"}, {"puma", ">= 6.0, < 7", "6.6.1"}, {"rake", "any", ""},
		}},
		{Title: "Locked in Gemfile.lock", ValueLabel: "Version", Rows: []ManifestRow{{"base64", "0.3.0", ""}, {"nio4r", "2.7.5", ""}}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v", got)
	}
}

func TestParsePom(t *testing.T) {
	pom := []byte(`<project xmlns="http://maven.apache.org/POM/4.0.0">
  <modelVersion>4.0.0</modelVersion>
  <groupId>org.openpanel.tests</groupId>
  <artifactId>java-tests</artifactId>
  <version>1.0.0</version>
  <properties>
    <maven.compiler.release>21</maven.compiler.release>
  </properties>
  <dependencies>
    <dependency><groupId>junit</groupId><artifactId>junit</artifactId><version>4.13.2</version><scope>test</scope></dependency>
    <dependency><groupId>org.slf4j</groupId><artifactId>slf4j-api</artifactId></dependency>
  </dependencies>
</project>`)
	got := parsePom(pom)
	want := []ManifestGroup{
		{Title: "Project", ValueLabel: "Value", Rows: []ManifestRow{{"groupId", "org.openpanel.tests", ""}, {"artifactId", "java-tests", ""}, {"version", "1.0.0", ""}, {"Java release", "21", ""}}},
		{Title: "Dependencies", ValueLabel: "Version", Rows: []ManifestRow{{"junit:junit", "4.13.2 (test)", ""}, {"org.slf4j:slf4j-api", "managed", ""}}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v", got)
	}
	if parsePom([]byte("<project><groupId>")) != nil {
		t.Error("broken xml should give nil")
	}
}

func TestParseComposer(t *testing.T) {
	composer := []byte(`{"require": {"php": "^7.4 || ^8.0", "slim/slim": "^4.10", "monolog/monolog": "^2.8"}, "require-dev": {"phpunit/phpunit": "^9.5"}}`)
	lock := []byte(`{"packages": [{"name": "slim/slim", "version": "4.11.0"}, {"name": "Monolog/Monolog", "version": "2.9.1"}], "packages-dev": [{"name": "phpunit/phpunit", "version": "9.6.13"}]}`)
	got := parseComposer(composer, lock)
	want := []ComposerRow{
		{"monolog/monolog", "^2.8", "2.9.1", false},
		{"php", "^7.4 || ^8.0", "", false},
		{"slim/slim", "^4.10", "4.11.0", false},
		{"phpunit/phpunit", "^9.5", "9.6.13", true},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v", got)
	}
	if parseComposer([]byte("nope"), nil) != nil {
		t.Error("broken composer.json should give nil")
	}
}

func TestReadSiteFileRejectsOutsideDocroot(t *testing.T) {
	for _, p := range []string{"/etc/passwd", "/var/www/html/../../etc/passwd", "relative/file"} {
		if _, ok := readSiteFile("someuser", p, 1024); ok {
			t.Errorf("%s should be refused", p)
		}
	}
}

func TestManifestFile(t *testing.T) {
	for typ, want := range map[string]string{"nodejs": "package.json", "python": "requirements.txt", "ruby": "Gemfile", "java": "pom.xml"} {
		if got := manifestFile(typ); got != want {
			t.Errorf("%s: %s", typ, got)
		}
	}
}
