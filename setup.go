package nodecache

import (
	"fmt"
	"github.com/caddyserver/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coreos/go-iptables/iptables"
	"github.com/vishvananda/netlink"
	"net"
	"strconv"
)

var log = clog.NewWithPlugin("nodecache")

func init() {
	caddy.RegisterPlugin("nodecache", caddy.Plugin{
		ServerType: "dns",
		Action:     setup,
	})
}

type config struct {
	port          int
	ifName        string
	localIPs      []net.IP
	setupIPTables bool
}

func setup(c *caddy.Controller) error {
	cfg, err := parseConfig(c, dnsserver.GetConfig(c))

	if err != nil {
		return plugin.Error("nodecache", c.ArgErr())
	}

	nl := netlink.Handle{}

	if exists, err := EnsureDummyDevice(&nl, cfg.ifName, cfg.localIPs); err != nil {
		return plugin.Error("nodecache", fmt.Errorf("failed to create dummy interface: %s", err))
	} else if !exists {
		clog.Infof("nodecache - added interface - %s with IPs %s", cfg.ifName, cfg.localIPs)
	}

	ipt, err := iptables.New()
	if err != nil {
		return plugin.Error("nodecache", fmt.Errorf("failed to create iptables context: %s", err))
	}

	rules := iptablesRules(cfg.localIPs, cfg.port)
	for _, rule := range rules {
		if err = ipt.AppendUnique(rule.table, rule.chain, rule.rulespec...); err != nil {
			return plugin.Error("nodecache", fmt.Errorf("failed to create iptable rule: %s", err))
		}
	}

	c.OnShutdown(func() error {
		log.Info("nodecache - shutting down")

		for _, rule := range rules {
			// ensureRuleDeleted returns true, nil if a rule was actually deleted, false, nil if the action was a noop
			if deleted, err := ensureRuleDeleted(ipt, rule); err != nil {
				return err
			} else if deleted {
				clog.Infof("nodecache - deleted iptable rule %s %s: %s", rule.table, rule.chain, rule.rulespec)
			}
		}

		return EnsureDummyDeviceRemoved(nl, cfg.ifName)
	})

	return nil
}

func parseConfig(c *caddy.Controller, serverConfig *dnsserver.Config) (*config, error) {
	pluginCfg := config{
		ifName:        "nodecache",
		setupIPTables: true,
		port:          53,
		localIPs:      []net.IP{net.ParseIP("192.168.10.100")},
	}

	for c.Next() {
		args := c.RemainingArgs()
		if len(args) > 0 {
			switch {
			case args[0] == "skip_iptables":
				pluginCfg.setupIPTables = false
			}
		}
	}

	if serverConfig.Port != "" {
		if p, err := strconv.Atoi(serverConfig.Port); err == nil {
			pluginCfg.port = p
		}
	}

	if len(serverConfig.ListenHosts) > 0 && serverConfig.ListenHosts[0] != "" {
		pluginCfg.localIPs = []net.IP{}
		for _, ip := range serverConfig.ListenHosts {
			pluginCfg.localIPs = append(pluginCfg.localIPs, net.ParseIP(ip))
		}
	}

	return &pluginCfg, nil
}
