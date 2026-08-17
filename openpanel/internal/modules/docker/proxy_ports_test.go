package docker

import (
	"strings"
	"testing"
)

func TestRewriteProxyHTTPPortLines(t *testing.T) {
	t.Run("on uncomments", func(t *testing.T) {
		got := RewriteProxyHTTPPortLines("FOO=bar\n#PROXY_HTTP_PORT=8080\nBAZ=qux\n", "on")
		want := "FOO=bar\nPROXY_HTTP_PORT=8080\nBAZ=qux\n"
		if got != want {
			t.Errorf("on: got %q, want %q", got, want)
		}
	})

	t.Run("off comments", func(t *testing.T) {
		got := RewriteProxyHTTPPortLines("FOO=bar\nPROXY_HTTP_PORT=8080\nBAZ=qux\n", "off")
		want := "FOO=bar\n#PROXY_HTTP_PORT=8080\nBAZ=qux\n"
		if got != want {
			t.Errorf("off: got %q, want %q", got, want)
		}
	})

	t.Run("off is idempotent when already commented", func(t *testing.T) {
		got := RewriteProxyHTTPPortLines("#PROXY_HTTP_PORT=8080\n", "off")
		want := "#PROXY_HTTP_PORT=8080\n"
		if got != want {
			t.Errorf("off idempotent: got %q, want %q", got, want)
		}
	})
}

func TestProxyPortSwapPair(t *testing.T) {
	if old, repl, err := ProxyPortSwapPair("on"); err != nil || old != "${HTTP_PORT}" || repl != "${PROXY_HTTP_PORT}" {
		t.Errorf("on: got (%q, %q, %v)", old, repl, err)
	}
	if old, repl, err := ProxyPortSwapPair("off"); err != nil || old != "${PROXY_HTTP_PORT}" || repl != "${HTTP_PORT}" {
		t.Errorf("off: got (%q, %q, %v)", old, repl, err)
	}
	if _, _, err := ProxyPortSwapPair("sideways"); err == nil {
		t.Error("expected an error for an invalid state")
	}
}

func TestRewriteComposePortBlock(t *testing.T) {
	content := "services:\n" +
		"  nginx:\n" +
		"    container_name: nginx\n" +
		"    ports:\n" +
		"      - \"${HTTP_PORT}:80\"\n" +
		"      - \"${HTTPS_PORT}:443\"\n" +
		"  mysql:\n" +
		"    container_name: mysql\n" +
		"    ports:\n" +
		"      - \"${HTTP_PORT}:3306\"\n"

	got := RewriteComposePortBlock(content, "nginx", "${HTTP_PORT}", "${PROXY_HTTP_PORT}")

	if !strings.Contains(got, `"${PROXY_HTTP_PORT}:80"`) {
		t.Errorf("expected nginx's HTTP_PORT swapped, got:\n%s", got)
	}
	if !strings.Contains(got, `"${HTTP_PORT}:3306"`) {
		t.Errorf("expected mysql's HTTP_PORT left untouched (outside nginx's block), got:\n%s", got)
	}
}
