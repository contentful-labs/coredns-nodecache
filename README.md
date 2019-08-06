# Coredns-nodecache

**Warning: This is an early release. Please do not use in production**

The Kubernetes [Node-local dns](https://github.com/kubernetes/kubernetes/tree/master/cluster/addons/dns/nodelocaldns)
add-on proposes running a DNS caching server on all of a Kubernetes cluster's nodes. The suggested caching server is
[node-cache](https://github.com/kubernetes/dns/tree/master/cmd/node-cache), a thin wrapper around CoreDNS, that
handles the setup & teardown of the dummy network interface & associated IPTables rules.

Coredns-nodecache is an attempt to implement node-cache as a CoreDNS plugin, rather than a wrapper. The motivations
for this are:

 *  the implementation relies only on CoreDNS Plugin API, which should be backward-compatible from version to version.
    This should make using the latest version of CoreDNS easier (see
    [kubernetes/dns #306](https://github.com/kubernetes/dns/issues/306))
 *  the configuration of nodecache would be done in the CoreDNS configuration file, instead of being split between
    the Corefile and command-line parameters.

### Configuration

Configuration is done by adding "nodecache [skip_iptables]" to configuration blocks in your CoreDNS configuration file.

As the following example shows, you can use the directive in several blocks. For each block, coredns-nodecache will
add the "bind" address to the dummy interface, and create iptable rules for the IP:PORT.

```
.:5300 {
    bind 168.255.10.20
    nodecache
    forward . 1.1.1.1:53 {
        force_tcp
    }
}

.:5301 {
    bind 168.255.10.25
    nodecache skip_iptables
    forward . 1.1.1.1:53 {
        force_tcp
    }
}
```

### Development

Checkout this repository & CoreDNS in your GOPATH. in the CoreDNS repository, in plugin.cfg, add the following line:

  ```nodecache:github.com/contentful-labs/coredns-nodecache```

Then at the end of go.mod:

  ```replace github.com/contentful-labs/coredns-nodecache => ../../contentful-labs/coredns-nodecache```

*make* should build CoreDNS and include the coredns-nodecache plugin.

   ```
   ./coredns -plugins
   [...]
   dns.metadata
   dns.nodecache
   dns.nsid
   [...]
   ```
