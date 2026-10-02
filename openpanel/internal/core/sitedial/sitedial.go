// Package sitedial gives installers an HTTP client that reaches a user's site without DNS or Caddy, so browser-style install wizards work before the domain is pointed here.
package sitedial

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"sync"
	"time"

	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
)

// relayPHP pipes stdin/stdout to the tcp address in argv[1], half-closing the socket when stdin ends
const relayPHP = `$s=@stream_socket_client("tcp://".$argv[1],$en,$es,10);if(!$s){fwrite(STDERR,$es);exit(1);}
$in=STDIN;stream_set_blocking($in,false);stream_set_blocking($s,false);
while(true){$r=[$s];if($in)$r[]=$in;$w=$x=null;if(@stream_select($r,$w,$x,null)===false)exit(1);
foreach($r as $h){$d=fread($h,65536);if($d===false||$d===''){if(!feof($h))continue;if($h===$s)exit(0);stream_socket_shutdown($s,STREAM_SHUT_WR);$in=null;continue;}
$o=$h===$s?STDOUT:$s;while($d!==''){$n=fwrite($o,$d);if($n===false)exit(1);$d=substr($d,$n);}}}`

// NewClient returns a cookie-keeping client whose connections are tunneled through phpContainer to webServer:443, url host still drives SNI and the Host header
func NewClient(userContext, phpContainer, webServer string, timeout time.Duration) *http.Client {
	target := webServer + ":443"
	if phpContainer == webServer {
		target = "localhost:443"
	}
	jar, _ := cookiejar.New(nil)
	return &http.Client{
		Jar:     jar,
		Timeout: timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // backend uses a self-signed cert
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return dialViaExec(userContext, phpContainer, target)
			},
		},
	}
}

type execConn struct {
	net.Conn
	once sync.Once
	kill func()
}

func (c *execConn) Close() error {
	err := c.Conn.Close()
	c.once.Do(c.kill)
	return err
}

func dialViaExec(userContext, container, target string) (net.Conn, error) {
	argv := podmanmanager.PodmanArgv(userContext, "exec", "-i", container, "php", "-r", relayPHP, "--", target)
	// not the dial ctx, that one is done as soon as dial returns
	cmd := podmanmanager.Command(context.Background(), userContext, argv)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	client, server := net.Pipe()
	go func() {
		_, _ = io.Copy(stdin, server)
		_ = stdin.Close()
	}()
	go func() {
		_, _ = io.Copy(server, stdout)
		_ = server.Close()
		_ = cmd.Wait()
	}()
	return &execConn{Conn: client, kill: func() { _ = cmd.Process.Kill() }}, nil
}
