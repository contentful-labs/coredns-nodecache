package nodecache

import (
	"github.com/coreos/go-iptables/iptables"
	"net"
	"strconv"
)

type rule struct {
	table    string
	chain    string
	rulespec []string
}

func NewRule(table, chain string, rulespec ...string) rule {
	return rule{
		table:    table,
		chain:    chain,
		rulespec: rulespec,
	}
}

func iptablesRules(localIPs []net.IP, localPort int) []rule {
	r := make([]rule, 0)
	slocalPort := strconv.Itoa(localPort)

	for _, localIP := range localIPs {
		slocalIP := localIP.String()
		r = append(r, NewRule("raw", "PREROUTING", "-p", "tcp", "-d", slocalIP, "--dport", slocalPort, "-j", "NOTRACK", "-w"))
		r = append(r, NewRule("raw", "PREROUTING", "-p", "udp", "-d", slocalIP, "--dport", slocalPort, "-j", "NOTRACK", "-w"))
		r = append(r, NewRule("filter", "INPUT", "-p", "tcp", "-d", slocalIP, "--dport", slocalPort, "-j", "ACCEPT", "-w"))
		r = append(r, NewRule("filter", "INPUT", "-p", "udp", "-d", slocalIP, "--dport", slocalPort, "-j", "ACCEPT", "-w"))
		r = append(r, NewRule("raw", "OUTPUT", "-p", "tcp", "-s", slocalIP, "--sport", slocalPort, "-j", "NOTRACK", "-w"))
		r = append(r, NewRule("raw", "OUTPUT", "-p", "udp", "-s", slocalIP, "--sport", slocalPort, "-j", "NOTRACK", "-w"))
		r = append(r, NewRule("filter", "OUTPUT", "-p", "tcp", "-s", slocalIP, "--sport", slocalPort, "-j", "ACCEPT", "-w"))
		r = append(r, NewRule("filter", "OUTPUT", "-p", "udp", "-s", slocalIP, "--sport", slocalPort, "-j", "ACCEPT", "-w"))
	}

	return r
}

func ensureRuleDeleted(ipt *iptables.IPTables, r rule) (bool, error) {
	exists, err := ipt.Exists(r.table, r.chain, r.rulespec...)
	if err != nil {
		return false, err
	}

	if exists {
		return true, ipt.Delete(r.table, r.chain, r.rulespec...)
	}

	return false, nil
}
