package nodecache

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
	"net"
	"testing"
)

type linkMock struct {
	name  string
	Addrs []netlink.Addr
}

func newLinkMock(name string) *linkMock {
	return &linkMock{
		name:  name,
		Addrs: make([]netlink.Addr, 0),
	}
}

func (l *linkMock) Attrs() *netlink.LinkAttrs {
	return &netlink.LinkAttrs{
		Name: l.name,
	}
}

func (*linkMock) Type() string {
	return ""
}

type nlMock struct {
	ifs map[string]*linkMock
}

func newNlMock() *nlMock {
	return &nlMock{ifs: make(map[string]*linkMock)}
}

func (n *nlMock) LinkByName(name string) (netlink.Link, error) {
	if l, ok := n.ifs[name]; !ok {
		return nil, fmt.Errorf("if %s not found", name)
	} else {
		return l, nil
	}
}

func (n *nlMock) AddrList(nl netlink.Link, i int) ([]netlink.Addr, error) {
	lMock := nl.(*linkMock)
	return lMock.Addrs, nil
}

func (n *nlMock) LinkAdd(nl netlink.Link) error {
	n.ifs[nl.Attrs().Name] = newLinkMock("nodecache")
	return nil
}

func (n *nlMock) LinkDel(nl netlink.Link) error {
	delete(n.ifs, nl.Attrs().Name)
	return nil
}

func mockAddrAdder(nl netlink.Link, addr *netlink.Addr) error {
	lMock := nl.(*linkMock)
	lMock.Addrs = append(lMock.Addrs, *addr)

	return nil
}

func TestEnsureDummyDevice(t *testing.T) {
	nl := newNlMock()
	ifName := "mockIf"

	if _, err := EnsureDummyDevice(nl, ifName, []net.IP{net.IPv4(8, 8, 8, 8), net.IPv4(8, 8, 4, 4)}, mockAddrAdder); err != nil {
		t.Errorf("failed adding dummy device: %s", err)
	}

	l, err := nl.LinkByName(ifName)
	if err != nil {
		t.Errorf("failed creating test interface")
	}

	addrs, err := nl.AddrList(l, unix.AF_INET)
	if len(addrs) != 2 || err != nil {
		t.Errorf("failed assigning IP address to interface")
	}
}

func TestEnsureDummyDeviceWithExistingIPs(t *testing.T) {
	var err error
	var alreadyExists bool
	ifName := "mockIf"

	nl := newNlMock()
	_ = nl.LinkAdd(&netlink.Dummy{ LinkAttrs: netlink.LinkAttrs{Name: ifName}})
	l, _ := nl.LinkByName(ifName)
	_ = mockAddrAdder(l, &netlink.Addr{IPNet: netlink.NewIPNet(net.IP{8, 8, 8, 8})} )

	if alreadyExists, err = EnsureDummyDevice(nl, ifName, []net.IP{net.IPv4(8, 8, 8, 8), net.IPv4(8, 8, 4, 4)}, mockAddrAdder); err != nil {
		t.Errorf("failed adding dummy device: %s", err)
	}

	if alreadyExists != true {
		t.Errorf("test interface should already exist")
	}

	addrs, err := nl.AddrList(l, unix.AF_INET)
	if len(addrs) != 2 || err != nil {
		t.Errorf("failed assigning IP address to interface")
	}
}
