package config

import (
	"encoding/json"

	"github.com/containernetworking/cni/pkg/types"
)

type Subnet struct {
	Subnet  string `json:"subnet"`
	Gateway string `json:"gateway"`
}

type CNIConfig struct {
	types.NetConf
	Bridge  string `json:"bridge"`
	Gateway string `json:"gateway"`
	Subnet  string `json:"subnet"`
	DataDir string `json:"dataDir"`
}

func LoadConfig(stdin []byte) (*CNIConfig, error) {
	var conf CNIConfig
	err := json.Unmarshal(stdin, &conf)
	if err != nil {
		return nil, err
	}
	return &conf, nil
}
