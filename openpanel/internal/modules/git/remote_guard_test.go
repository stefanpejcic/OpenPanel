package git

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRemoteToPrivateAddressIsRefused(t *testing.T) {
	ctx := context.Background()
	for _, u := range []string{"http://127.0.0.1:2083/x", "http://169.254.169.254/latest/meta-data", "ssh://git@172.20.0.2:3306/x", "git@10.0.0.1:x"} {
		if _, err := remoteEnv(ctx, "bob", Repo{RemoteURL: u}); err == nil {
			t.Errorf("accepted %s", u)
		}
	}
	old := lookupRemoteIP
	lookupRemoteIP = func(context.Context, string) ([]net.IP, error) { return []net.IP{net.ParseIP("192.168.1.5")}, nil }
	defer func() { lookupRemoteIP = old }()
	if _, err := remoteEnv(ctx, "bob", Repo{RemoteURL: "https://rebind.example.com/x.git"}); err == nil {
		t.Error("accepted a name that resolves to a private address")
	}
}

func TestRemoteEnvPinsAndDisablesRedirects(t *testing.T) {
	old := lookupRemoteIP
	lookupRemoteIP = func(context.Context, string) ([]net.IP, error) { return []net.IP{net.ParseIP("93.184.216.34")}, nil }
	defer func() { lookupRemoteIP = old }()
	env, err := remoteEnv(context.Background(), "bob", Repo{RemoteURL: "https://example.com/x.git", RemoteUser: "u", RemotePass: "p"})
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(env, "\n")
	for _, want := range []string{"GIT_CONFIG_COUNT=3", "=http.followRedirects", "GIT_CONFIG_VALUE_0=false", "=example.com:443:93.184.216.34", "=http.extraHeader"} {
		if !strings.Contains(joined, want) {
			t.Errorf("missing %q in %v", want, env)
		}
	}
	env, _ = remoteEnv(context.Background(), "bob", Repo{RemoteURL: "git@example.com:org/x.git"})
	if !strings.Contains(strings.Join(env, " "), "-o HostName=93.184.216.34") {
		t.Errorf("ssh not pinned: %v", env)
	}
}

func TestCloneDoesNotFollowRedirectToInternal(t *testing.T) {
	userContext, _ := withFakeAccount(t)
	oldOK := remoteAddrOK
	remoteAddrOK = func(net.IP) bool { return true }
	defer func() { remoteAddrOK = oldOK }()
	internal := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "INTERNAL-SECRET", http.StatusForbidden)
	}))
	defer internal.Close()
	redirector := httptest.NewServer(http.RedirectHandler(internal.URL+"/x.git/info/refs?service=git-upload-pack", http.StatusFound))
	defer redirector.Close()
	r := Repo{Name: "r", Source: SourceRemote, RemoteURL: redirector.URL + "/x.git"}
	err := initRepo(context.Background(), userContext, &r)
	if err == nil || strings.Contains(err.Error(), "INTERNAL-SECRET") {
		t.Fatalf("redirect was followed: %v", err)
	}
}
