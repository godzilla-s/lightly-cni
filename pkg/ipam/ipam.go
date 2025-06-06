package ipam

import (
	"errors"
	"lightly-cni/pkg/config"
	"lightly-cni/pkg/store"
	"net"

	cip "github.com/containernetworking/plugins/pkg/ip"
)

var (
	ErrIPOverFlow = errors.New("no more ips available")
)

type IPAM struct {
	subnet  *net.IPNet
	gateway net.IP
	store   *store.Store
}

func New(conf *config.CNIConfig, s *store.Store) (*IPAM, error) {
	_, ipnet, err := net.ParseCIDR(conf.Subnet)
	if err != nil {
		return nil, err
	}

	ipam := &IPAM{
		subnet: ipnet,
		store:  s,
	}

	return ipam, nil
}

func (i *IPAM) Mask() net.IPMask {
	return i.subnet.Mask
}

func (i *IPAM) Gateway() net.IP {
	return i.gateway
}

func (i *IPAM) NextIP(ip net.IP) (net.IP, error) {
	nextIP := cip.NextIP(ip)
	if !i.subnet.Contains(nextIP) {
		return nil, ErrIPOverFlow
	}
	return nextIP, nil
}

// 这里的id为容器的id
func (i *IPAM) Allocate(id, ifname string) (net.IP, error) {
	i.store.Lock()
	defer i.store.Unlock()

	if err := i.store.LoadData(); err != nil {
		return nil, err
	}

	// 如果已存在，则直接返回对应的IP
	ip, _ := i.store.GetIPById(id)
	if len(ip) > 0 {
		return ip, nil
	}

	last := i.store.Last()
	start := make(net.IP, len(last))
	copy(start, last)

	for {
		next, err := i.NextIP(start)
		if err == ErrIPOverFlow && !last.Equal(i.gateway) {
			// 从头开始
			start = i.gateway
			continue
		} else if err != nil {
			return nil, err
		}

		if err := i.store.Add(next, id, ifname); err != nil {
			break
		}
	}

	return nil, ErrIPOverFlow
}

func (i *IPAM) ReleaseIP(id string) error {
	i.store.Lock()
	defer i.store.Unlock()

	if err := i.store.LoadData(); err != nil {
		return err
	}
	return i.store.Remove(id)
}

func (i *IPAM) CheckIP(id string) (net.IP, error) {
	i.store.Lock()
	defer i.store.Unlock()

	if err := i.store.LoadData(); err != nil {
		return nil, err
	}

	ip, ok := i.store.GetIPById(id)
	if !ok {
		return nil, errors.New("ip not found")
	}

	return ip, nil
}
