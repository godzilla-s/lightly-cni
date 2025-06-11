package bridge

import (
	"fmt"
	"net"
	"os"
	"syscall"

	current "github.com/containernetworking/cni/pkg/types/100"
	"github.com/containernetworking/plugins/pkg/ip"
	"github.com/containernetworking/plugins/pkg/ns"
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

	err = netlink.AddrAdd(dev, &netlink.Addr{IPNet: gateway})
	if err != nil {
		return nil, err
	}

	err = netlink.LinkSetUp(dev)
	if err != nil {
		return nil, err
	}

	return dev, nil
}

func SetupVeth(netns ns.NetNS, br netlink.Link, mtu int, ifName string, podIP *net.IPNet, gateway net.IP) error {
	hostIface := &current.Interface{}

	err := netns.Do(func(nn ns.NetNS) error {
		hostVeth, containerVeth, err0 := ip.SetupVeth(ifName, mtu, "", nn)
		if err0 != nil {
			return err0
		}

		hostIface.Name = hostVeth.Name

		connLink, err0 := netlink.LinkByName(containerVeth.Name)
		if err0 != nil {
			return err0
		}

		err0 = netlink.AddrAdd(connLink, &netlink.Addr{IPNet: podIP})
		if err0 != nil {
			return err0
		}

		err0 = netlink.LinkSetUp(connLink)
		if err0 != nil {
			return err0
		}

		err0 = ip.AddDefaultRoute(gateway, connLink)
		if err0 != nil {
			return err0
		}
		return nil
	})
	if err != nil {
		return err
	}

	hostVeth, err := netlink.LinkByName(hostIface.Name)
	if err != nil {
		return err
	}

	if hostVeth == nil {
		return fmt.Errorf("nil hostveth")
	}

	err = netlink.LinkSetMaster(hostVeth, br)
	if err != nil {
		return fmt.Errorf("failed to connect %q to bridge %q: %v", hostVeth.Attrs().Name, br.Attrs().Name, err)
	}

	return nil
}

func DelVeth(netns ns.NetNS, ifname string) error {
	return netns.Do(func(nn ns.NetNS) error {
		l, err := netlink.LinkByName(ifname)
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		return netlink.LinkDel(l)
	})
}

func CheckVeth(netns ns.NetNS, ifname string, ip net.IP) error {
	return netns.Do(func(nn ns.NetNS) error {
		l, err := netlink.LinkByName(ifname)
		if err != nil {
			return err
		}

		ips, err := netlink.AddrList(l, netlink.FAMILY_V4)
		if err != nil {
			return err
		}

		for _, addr := range ips {
			if addr.IP.Equal(ip) {
				return nil
			}
		}
		return fmt.Errorf("failed to find IP %s for %s", ip, ifname)
	})
}
