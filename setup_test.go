package nodecache

import (
	"net"
	"reflect"
	"testing"

	"github.com/caddyserver/caddy"
	"github.com/coredns/coredns/core/dnsserver"
)

func TestSetupParse(t *testing.T) {
	for _, test := range []struct {
		config   string
		expected config
	}{
		{
			`nodecache`, // TODO caddy enable bind plugin here
			config{
				ifName:        "nodecache",
				setupIPTables: true,
				port:          53,
				localIPs:      []net.IP{net.ParseIP("192.168.10.100")},
				skipTeardown:  false,
			},
		},
		{
			`nodecache skipteardown`,
			config{
				ifName:        "nodecache",
				setupIPTables: true,
				port:          53,
				localIPs:      []net.IP{net.ParseIP("192.168.10.100")},
				skipTeardown:  true,
			},
		},
	} {
		c := caddy.NewTestController("dns", test.config)
		cfg := getDefaultCfg()
		if cfg.parsePlgConfig(c) != nil {
			t.Error("Error while trying to parse plugin config")
		}
		if cfg.parseSrvConfig(dnsserver.GetConfig(c)) != nil {
			t.Error("Error while trying to parse server config")
		}
		if !reflect.DeepEqual(&cfg, &test.expected) {
			t.Error("Expected", &test.expected, " gots ", &cfg)
		}
		// TODO ensure interface
	}
}

func TestConfigParse(t *testing.T) {
	for _, test := range []struct {
		input     dnsserver.Config
		output    config
		shouldErr bool
	}{
		{
			dnsserver.Config{
				Port:        "53",
				ListenHosts: []string{"168.255.20.10"},
			},
			config{
				ifName:        "nodecache",
				setupIPTables: true,
				port:          53,
				localIPs:      []net.IP{net.ParseIP("168.255.20.10")},
				skipTeardown:  false,
			},
			false,
		},
	} {
		cfg := getDefaultCfg()
		err := cfg.parseSrvConfig(&test.input)
		if test.shouldErr {
			if err != nil {
				t.Error("Should've returned an error")
			}
		} else {
			if !reflect.DeepEqual(&cfg, &test.output) {
				t.Error("Expected", &test.output, " gots ", &cfg)
			}
		}
	}
}
