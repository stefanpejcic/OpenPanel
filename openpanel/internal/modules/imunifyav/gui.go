package imunifyav

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"
)

// guiKeyPath is written by opencli imunify start, the gui on the gateway refuses requests without it
var guiKeyPath = "/etc/openpanel/imunify/panel.key"

var guiPort = "9000"

// gatewayRun is the container's default gateway, the host side of the panel network where opencli binds the gui
var gatewayRun = func() string {
	f, err := os.Open("/proc/net/route")
	if err != nil {
		return ""
	}
	defer f.Close()
	return defaultGateway(f)
}

// defaultGateway reads /proc/net/route, where addresses are little endian hex
func defaultGateway(r io.Reader) string {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) < 3 || fields[1] != "00000000" {
			continue
		}
		b, err := hex.DecodeString(fields[2])
		if err != nil || len(b) != 4 {
			continue
		}
		return net.IPv4(b[3], b[2], b[1], b[0]).String()
	}
	return ""
}

func guiKey() string {
	raw, err := os.ReadFile(guiKeyPath)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(raw))
}

func guiHost() string {
	gw := gatewayRun()
	if gw == "" {
		return ""
	}
	return net.JoinHostPort(gw, guiPort)
}

// guiReady is false when opencli hasn't set the gui up for the panel yet, the page then falls back to the summary
var guiReady = func() bool {
	host := guiHost()
	if host == "" || guiKey() == "" {
		return false
	}
	conn, err := net.DialTimeout("tcp", host, time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// loginToken gets a token for the user's own context, the gui then only shows that user's data
var loginToken = func(userContext string) (string, error) {
	data, err := call([]string{"login", "get"}, map[string]any{"username": userContext})
	if err != nil {
		return "", err
	}
	var token string
	if err := items(data, &token); err != nil {
		return "", err
	}
	if token == "" {
		return "", fmt.Errorf("empty token")
	}
	return token, nil
}

func newGUIProxy() http.Handler {
	return &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			host := guiHost()
			pr.SetURL(&url.URL{Scheme: "http", Host: host})
			pr.Out.Host = host
			// the panel's session and csrf cookies are none of imunify's business
			pr.Out.Header.Del("Cookie")
			pr.Out.Header.Del("Authorization")
			pr.Out.Header.Set("X-OpenPanel-Key", guiKey())
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, "ImunifyAV is not reachable right now.", http.StatusBadGateway)
		},
	}
}
