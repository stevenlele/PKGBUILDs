package main

import (
	"bytes"
	"encoding/json"

	"github.com/BurntSushi/toml"
	"github.com/caddyserver/caddy/v2/caddyconfig"
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
	caddycmd.Main()
}
