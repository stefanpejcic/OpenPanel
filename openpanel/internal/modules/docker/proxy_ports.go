package docker

import (
	"fmt"
	"os"
	"strings"
)

// ToggleProxyHTTPPort comments or uncomments the PROXY_HTTP_PORT line in a
// user's .env file, enabling or disabling the port varnish listens on.
func ToggleProxyHTTPPort(userContext, state string) error {
	path := "/home/" + userContext + "/.env"
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(RewriteProxyHTTPPortLines(string(content), state)), 0o644)
}

// RewriteProxyHTTPPortLines is ToggleProxyHTTPPort()'s pure line-rewriting
// half, split out for testability without touching the filesystem.
func RewriteProxyHTTPPortLines(content, state string) string {
	lines := strings.SplitAfter(content, "\n")
	var out strings.Builder
	for _, line := range lines {
		if !strings.Contains(line, "PROXY_HTTP_PORT=") {
			out.WriteString(line)
			continue
		}
		switch state {
		case "on":
			out.WriteString(strings.TrimLeft(line, "# "))
		case "off":
			if !strings.HasPrefix(strings.TrimSpace(line), "#") {
				out.WriteString("#" + line)
			} else {
				out.WriteString(line)
			}
		default:
			out.WriteString(line)
		}
	}
	return out.String()
}

// ProxyPortSwapPair picks the old/new port-variable pair to swap when
// toggling varnish on or off.
func ProxyPortSwapPair(state string) (old, replacement string, err error) {
	switch state {
	case "on":
		return "${HTTP_PORT}", "${PROXY_HTTP_PORT}", nil
	case "off":
		return "${PROXY_HTTP_PORT}", "${HTTP_PORT}", nil
	default:
		return "", "", fmt.Errorf("state must be either 'on' or 'off'")
	}
}

// SwapWebserverComposePort swaps one webserver's port-mapping variable in
// place. podman-compose can't resolve a ${VAR} nested inside another
// ${VAR:-default}, so docker-compose.yml keeps each webserver's port
// mapping flat rather than using a fallback expression - the swap is
// scoped to the block between the webserver's own `container_name:` line
// and its `HTTPS_PORT` line, so only that service's port mapping is
// touched.
//
// Callers must keep exactly one webserver's block on PROXY_HTTP_PORT
// ("on") at a time when varnish is running - both switching varnish on/off
// and switching the active webserver need to call this so the compose
// file's port mappings stay consistent with whichever webserver is
// actually the one varnish proxies to.
func SwapWebserverComposePort(userContext, webserver, state string) error {
	old, replacement, err := ProxyPortSwapPair(state)
	if err != nil {
		return err
	}

	path := "/home/" + userContext + "/docker-compose.yml"
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(RewriteComposePortBlock(string(content), webserver, old, replacement)), 0o644)
}

// RewriteComposePortBlock is SwapWebserverComposePort()'s pure block-scoped
// rewrite, split out for testability without touching the filesystem.
func RewriteComposePortBlock(content, webserver, old, replacement string) string {
	lines := strings.SplitAfter(content, "\n")

	inBlock := false
	for i, line := range lines {
		if strings.HasSuffix(strings.TrimRight(line, " \t\r\n"), "container_name: "+webserver) {
			inBlock = true
		}
		if inBlock && strings.Contains(line, old) {
			lines[i] = strings.ReplaceAll(line, old, replacement)
		}
		if inBlock && strings.Contains(line, "HTTPS_PORT") {
			inBlock = false
		}
	}
	return strings.Join(lines, "")
}
