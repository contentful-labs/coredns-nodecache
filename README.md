# Coredns-nodecache [![CircleCI](https://circleci.com/gh/contentful-labs/coredns-nodecache.svg?style=svg)](https://circleci.com/gh/contentful-labs/coredns-nodecache)

The Kubernetes [Node-local dns](https://github.com/kubernetes/kubernetes/tree/master/cluster/addons/dns/nodelocaldns)
add-on proposes running a DNS caching server on all of a Kubernetes cluster's nodes. The suggested caching server is
[node-cache](https://github.com/kubernetes/dns/tree/master/cmd/node-cache), a thin wrapper around CoreDNS, that
handles the setup & teardown of the dummy network interface & associated IPTables rules.

**Coredns-nodecache** is an attempt to implement node-cache as a CoreDNS plugin, rather than a wrapper. The motivations
for this are:

 *  the implementation relies only on CoreDNS Plugin API, which should be backward-compatible from version to version.
    This should make using the latest version of CoreDNS easier (see
    [kubernetes/dns #306](https://github.com/kubernetes/dns/issues/306))
 *  the configuration of nodecache would be done in the CoreDNS configuration file, instead of being split between
    the Corefile and command-line parameters.

Additionnally, coredns-nodecache can provide a **high-availability** setup for Node-local.

### Plugin configuration

An image is available on DockerHub: https://hub.docker.com/r/contentful/coredns-nodecache

Configuration is done by adding `nodecache` to configuration blocks in your CoreDNS configuration file.

```
nodecache [skipteardown]
```

* **skipteardown**: skips removing the iptables and dummy network interface on shutdown. This is especially useful
for high-availability setups.

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
    nodecache
    forward . 1.1.1.1:53 {
        force_tcp
    }
}
```

### High-availability setup

CoreDNS & coredns-nodecache can provide a high-availbility setup for Kubernetes Node-local, using two separate Daemonsets & SO_REUSEPORT.
An example deployment is provided in [k8s/node-local-ha.yaml](k8s/node-local-ha.yaml).


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
