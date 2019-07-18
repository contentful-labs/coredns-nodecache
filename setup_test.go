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
		config    string
		expected  config
		shouldErr bool
	}{
		{
			`nodecache`, // TODO caddy enable bind plugin here
			config{
				ifName:        "nodecache",
				setupIPTables: true,
				port:          53,
				localIPs:      []net.IP{net.ParseIP("192.168.10.100")},
			},
			false,
		},
	} {
		c := caddy.NewTestController("dns", test.config)
		output, err := parseConfig(dnsserver.GetConfig(c))
		if test.shouldErr {
			if err != nil {
				t.Error("Should've returned an error")
			}
		} else {
			if !reflect.DeepEqual(output, &test.expected) {
				t.Error("Expected", &test.expected, " gots ", output)
			}
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
			},
			false,
		},
	} {
		output, err := parseConfig(&test.input)
		if test.shouldErr {
			if err != nil {
				t.Error("Should've returned an error")
			}
		} else {
			if !reflect.DeepEqual(output, &test.output) {
				t.Error("Expected", &test.output, " gots ", output)
			}
		}
	}
}
