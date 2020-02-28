package nodecache

import (
	"fmt"
	"net"

	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

// Code derived from https://github.com/kubernetes/dns/blob/master/pkg/netif/netif.go

type netlinkIf interface {
	LinkByName(string) (netlink.Link, error)
	AddrList(netlink.Link, int) ([]netlink.Addr, error)
	LinkAdd(netlink.Link) error
	LinkDel(netlink.Link) error
}

type addrAdder = func(netlink.Link, *netlink.Addr) error

// EnsureDummyDevice checks for the presence of the given dummy device and creates one if it does not exist.
// Returns a boolean to indicate if this device was found and error if any.
func EnsureDummyDevice(nl netlinkIf, ifName string, ensureIPs []net.IP, addrAdd addrAdder) (bool, error) {
	l, err := nl.LinkByName(ifName)
	linkAlreadyPresent := (err == nil)

	if !linkAlreadyPresent {
		dummy := &netlink.Dummy{
			LinkAttrs: netlink.LinkAttrs{Name: ifName},
		}
		if err = nl.LinkAdd(dummy); err != nil {
			return linkAlreadyPresent, err
		}
		if l, err = nl.LinkByName(ifName); err != nil {
			return linkAlreadyPresent, fmt.Errorf("failed getting link after creating it")
		}
	}

	linkAddrs, err := nl.AddrList(l, unix.AF_INET)
	if err != nil {
		return linkAlreadyPresent, err
	}

	// found dummy device, make sure all required ips are present on the interface.
	// If not, try to add them.
	for _, ensureIP := range ensureIPs {
		addrPresent := false
		for _, linkAddr := range linkAddrs {
			if linkAddr.IP.Equal(ensureIP) {
				addrPresent = true
				break
			}
		}

		if !addrPresent {
			if err := addrAdd(l, &netlink.Addr{IPNet: netlink.NewIPNet(ensureIP)}); err != nil {
				return linkAlreadyPresent, fmt.Errorf("failed adding ip %s to interface %s: %s", ensureIP, ifName, err)
			}
		}
	}

	return linkAlreadyPresent, nil
}

// RemoveDummyDevice deletes the dummy device with the given name.
func EnsureDummyDeviceRemoved(nl netlinkIf, ifName string) error {
	link, err := nl.LinkByName(ifName)
	if err != nil {
		// link does not exist, we do nothing
		return nil
	}
	return nl.LinkDel(link)
}
