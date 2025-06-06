package bridge

import (
	"net"
	"syscall"

	"github.com/vishvananda/netlink"
)

func Create(bridgeName string, mtu int, gateway *net.IPNet) (netlink.Link, error) {
	if bridgeName == "" {
		bridgeName = "cni0" // set default bridge name
	}

	link, _ := netlink.LinkByName(bridgeName)
	if link != nil {
		return link, nil
	}

	br := &netlink.Bridge{
		LinkAttrs: netlink.LinkAttrs{
			Name:   bridgeName,
			MTU:    mtu,
			TxQLen: -1,
		},
	}

	err := netlink.LinkAdd(br)
	if err != nil && err != syscall.EEXIST {
		return nil, err
	}

	dev, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return nil, err
	}

	err = netlink.AddrAdd(dev, &netlink.Addr{})
}
