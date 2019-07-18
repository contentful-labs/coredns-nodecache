package nodecache

import (
	"fmt"
	"net"

	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

// Code derived from https://github.com/kubernetes/dns/blob/master/pkg/netif/netif.go

// EnsureDummyDevice checks for the presence of the given dummy device and creates one if it does not exist.
// Returns a boolean to indicate if this device was found and error if any.
func EnsureDummyDevice(nlHandle netlink.Handle, ifName string, ensureIPs []net.IP) (bool, error) {
	l, err := nlHandle.LinkByName(ifName)
	if err == nil {
		linkAddrs, err := nlHandle.AddrList(l, unix.AF_INET)
		if err != nil {
			return true, err
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

			if addrPresent == false {
				if err := netlink.AddrAdd(l, &netlink.Addr{IPNet: netlink.NewIPNet(ensureIP)}); err != nil {
					return true, fmt.Errorf("failed adding ip %s to interface %s: %s", ensureIP, ifName, err)
				}
				log.Infof("adding IP %s to interface %s", ensureIP, ifName)
			}
		}
		return true, nil
	}

	return false, AddDummyDevice(nlHandle, ifName, ensureIPs)
}

// AddDummyDevice creates a dummy device with the given name. It also binds the ip address of the NetifManager instance
// to this device. This function returns an error if the device exists or if address binding fails.
func AddDummyDevice(nl netlink.Handle, ifName string, addrs []net.IP) error {
	var err error

	if _, err = nl.LinkByName(ifName); err == nil {
		return fmt.Errorf("Link %s exists", ifName)
	}

	dummy := &netlink.Dummy{
		LinkAttrs: netlink.LinkAttrs{Name: ifName},
	}
	if err = nl.LinkAdd(dummy); err != nil {
		return err
	}

	l, err := nl.LinkByName(ifName)
	if err != nil {
		return fmt.Errorf("failed getting link %s after creating it", ifName)
	}

	for _, addr := range addrs {
		if err := nl.AddrAdd(l, &netlink.Addr{IPNet: netlink.NewIPNet(addr)}); err != nil {
			return err
		}
	}

	return nil
}

// RemoveDummyDevice deletes the dummy device with the given name.
func EnsureDummyDeviceRemoved(nl netlink.Handle, ifName string) error {
	link, err := nl.LinkByName(ifName)
	if err != nil {
		// link does not exist, we do nothing
		return nil
	}
	return nl.LinkDel(link)
}
