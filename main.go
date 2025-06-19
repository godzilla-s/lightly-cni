package main

import (
	"fmt"
	"lightly-cni/pkg/bridge"
	"lightly-cni/pkg/config"
	"lightly-cni/pkg/ipam"
	"lightly-cni/pkg/store"
	"net"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	current "github.com/containernetworking/cni/pkg/types/100"
	"github.com/containernetworking/cni/pkg/version"
	"github.com/containernetworking/plugins/pkg/ns"
)

const (
	pluginName = "lightly-cni"
)

func main() {
	skel.PluginMainFuncs(skel.CNIFuncs{
		Add:   cmdAdd,
		Del:   cmdDel,
		Check: cmdCheck,
	}, version.All, pluginName)
}

func cmdAdd(args *skel.CmdArgs) error {
	conf, err := config.LoadConfig(args.StdinData)
	if err != nil {
		return err
	}

	s, err := store.New(conf.DataDir, conf.Name)
	if err != nil {
		return err
	}

	defer s.Close()

	ipam, err := ipam.New(conf, s)
	if err != nil {
		return err
	}

	gateway := ipam.Gateway()

	ip, err := ipam.Allocate(args.ContainerID, args.IfName)
	if err != nil {
		return err
	}

	// 创建网桥，虚拟设备，并绑定到网桥
	mtu := 1500
	br, err := bridge.Create(conf.Bridge, mtu, ipam.IPNet(gateway))
	if err != nil {
		return fmt.Errorf("create bridge failed: %v", err)
	}

	netns, err := ns.GetNS(args.Netns)
	if err != nil {
		return fmt.Errorf("get ns failed: %v", err)
	}

	defer netns.Close()

	err = bridge.SetupVeth(netns, br, mtu, args.IfName, ipam.IPNet(ip), gateway)
	if err != nil {
		return fmt.Errorf("setup veth failed: %v", err)
	}

	result := &current.Result{
		CNIVersion: current.ImplementedSpecVersion,
		IPs: []*current.IPConfig{
			{
				Address: net.IPNet{IP: ip},
				Gateway: gateway,
			},
		},
	}
	return types.PrintResult(result, conf.CNIVersion)
}

func cmdDel(args *skel.CmdArgs) error {
	conf, err := config.LoadConfig(args.StdinData)
	if err != nil {
		return fmt.Errorf("load config failed: %v", err)
	}

	s, err := store.New(conf.DataDir, conf.Name)
	if err != nil {
		return fmt.Errorf("new store failed: %v", err)
	}
	defer s.Close()

	ipam, err := ipam.New(conf, s)
	if err != nil {
		return fmt.Errorf("new ipam failed: %v", err)
	}

	err = ipam.ReleaseIP(args.ContainerID)
	if err != nil {
		return fmt.Errorf("release ip failed: %v", err)
	}

	netns, err := ns.GetNS(args.Netns)
	if err != nil {
		return fmt.Errorf("get ns failed: %v", err)
	}

	defer netns.Close()

	return bridge.DelVeth(netns, args.IfName)
}

func cmdCheck(args *skel.CmdArgs) error {
	conf, err := config.LoadConfig(args.StdinData)
	if err != nil {
		return fmt.Errorf("load config failed: %v", err)
	}

	s, err := store.New(conf.DataDir, conf.Name)
	if err != nil {
		return fmt.Errorf("new store failed: %v", err)
	}
	defer s.Close()

	ipam, err := ipam.New(conf, s)
	if err != nil {
		return fmt.Errorf("new ipam failed: %v", err)
	}

	ip, err := ipam.CheckIP(args.ContainerID)
	if err != nil {
		return fmt.Errorf("check ip failed: %v", err)
	}

	netns, err := ns.GetNS(args.Netns)
	if err != nil {
		return fmt.Errorf("get ns failed: %v", err)
	}
	defer netns.Close()
	return bridge.CheckVeth(netns, args.IfName, ip)
}
