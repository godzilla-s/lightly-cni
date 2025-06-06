package main

import (
	"lightly-cni/pkg/config"
	"lightly-cni/pkg/ipam"
	"lightly-cni/pkg/store"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/version"
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

	mtu := 1500

	return nil
}

func cmdDel(args *skel.CmdArgs) error {
	return nil
}

func cmdCheck(args *skel.CmdArgs) error {
	return nil
}
