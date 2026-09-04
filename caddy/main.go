package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"net"

	"github.com/BurntSushi/toml"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	caddycmd "github.com/caddyserver/caddy/v2/cmd"

	// plug in Caddy modules here
	_ "github.com/caddy-dns/duckdns"
	_ "github.com/caddyserver/caddy/v2/modules/standard"
)

type tomlAdapter struct{}

func (tomlAdapter) Adapt(body []byte, _ map[string]any) ([]byte, []caddyconfig.Warning, error) {
	config := make(map[string]any)
	if _, err := toml.NewDecoder(bytes.NewReader(body)).Decode(&config); err != nil {
		return nil, nil, err
	}

	adapted, err := json.Marshal(config)
	return adapted, nil, err
}

func main() {
	caddyconfig.RegisterAdapter("toml", tomlAdapter{})
	caddy.RegisterModule(abortHTTPListenerWrapper{})
	caddycmd.Main()
}

type abortHTTPListenerWrapper struct{}

func (abortHTTPListenerWrapper) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID: "caddy.listeners.abort_http",
		New: func() caddy.Module {
			return new(abortHTTPListenerWrapper)
		},
	}
}

func (*abortHTTPListenerWrapper) UnmarshalCaddyfile(_ *caddyfile.Dispenser) error {
	return nil
}

func (*abortHTTPListenerWrapper) WrapListener(l net.Listener) net.Listener {
	return &abortHTTPListener{Listener: l}
}

type abortHTTPListener struct {
	net.Listener
}

func (l *abortHTTPListener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return &abortHTTPConn{
		Conn: conn,
		r:    bufio.NewReader(conn),
	}, nil
}

type abortHTTPConn struct {
	net.Conn
	r       *bufio.Reader
	checked bool
}

func (c *abortHTTPConn) Read(p []byte) (int, error) {
	if !c.checked {
		c.checked = true

		first, err := c.r.Peek(5)
		if err != nil {
			return 0, err
		}
		if looksLikeHTTP(first) {
			_ = c.Conn.Close()
			return 0, net.ErrClosed
		}
	}
	return c.r.Read(p)
}

func looksLikeHTTP(hdr []byte) bool {
	switch string(hdr[:5]) {
	case "GET /", "HEAD ", "POST ", "PUT /", "OPTIO":
		return true
	default:
		return false
	}
}
