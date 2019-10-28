package nodecache

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/caddyserver/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coreos/go-iptables/iptables"
	"github.com/vishvananda/netlink"
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
	skipTeardown  bool
}

func setup(c *caddy.Controller) error {

	cfg := getDefaultCfg()
	if shouldSkipTearDown(c) {
		cfg.skipTeardown = true
	}

	if cfg.parseSrvConfig(dnsserver.GetConfig(c)) != nil {
		log.Errorf("Error while parsing server config")
		return plugin.Error("nodecache", c.ArgErr())
	}

	nl := netlink.Handle{}

	if exists, err := EnsureDummyDevice(&nl, cfg.ifName, cfg.localIPs, netlink.AddrAdd); err != nil {
		return plugin.Error("nodecache", fmt.Errorf("failed to create dummy interface: %s", err))
	} else if !exists {
		log.Infof("Added interface - %s", cfg.ifName)
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
		log.Info("nodecache shutting down")

		if !cfg.skipTeardown {
			for _, rule := range rules {
				// ensureRuleDeleted returns true, nil if a rule was actually deleted, false, nil if the action was a noop
				if deleted, err := ensureRuleDeleted(ipt, rule); err != nil {
					return err
				} else if deleted {
					log.Infof("deleted iptable rule %s %s: %s", rule.table, rule.chain, rule.rulespec)
				}
			}

			return EnsureDummyDeviceRemoved(&nl, cfg.ifName)
		}
		log.Info("skipping teardown")
		return nil
	})

	return nil
}

func getDefaultCfg() config {
	return config{
		ifName:        "nodecache",
		setupIPTables: true,
		port:          53,
		localIPs:      []net.IP{net.ParseIP("192.168.10.100")},
		skipTeardown:  false,
	}
}

func shouldSkipTearDown(c *caddy.Controller) bool{
	for c.Next() {
		for c.NextArg() {
			if strings.ToLower(c.Val()) == "skipteardown" {
				return true
			}
		}
	}

	return false
}

func (cfg *config) parseSrvConfig(serverConfig *dnsserver.Config) error {
	if serverConfig.Port != "" {
		if p, err := strconv.Atoi(serverConfig.Port); err == nil {
			cfg.port = p
		} else {
			return err
		}
	}

	if len(serverConfig.ListenHosts) > 0 && serverConfig.ListenHosts[0] != "" {
		cfg.localIPs = []net.IP{}
		for _, ip := range serverConfig.ListenHosts {
			cfg.localIPs = append(cfg.localIPs, net.ParseIP(ip))
		}
	}

	return nil
}
