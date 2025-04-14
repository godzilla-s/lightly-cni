package store

import (
	"encoding/json"
	"net"
	"os"
	"path/filepath"

	"github.com/alexflint/go-filemutex"
)

type containerNetInfo struct {
	ID     string `json:"id"`
	IfName string `json:"ifName"`
}

type data struct {
	containers map[string]containerNetInfo
	Last       string
}
type Store struct {
	*filemutex.FileMutex
	dir      string
	cache    *data
	metaFile string
}

func New(dataDir, network string) (*Store, error) {
	if dataDir == "" {
		dataDir = "/var/lib/cni"
	}

	dir := filepath.Join(dataDir, network)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
	}
	fl, err := newFileLock(dir)
	if err != nil {
		return nil, err
	}

	cache := &data{make(map[string]containerNetInfo), ""}
	metaFile := filepath.Join(dir, network+".json")
	return &Store{fl, dir, cache, metaFile}, nil
}

func newFileLock(path string) (*filemutex.FileMutex, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if fi.IsDir() {
		path = filepath.Join(path, "lock")
	}

	fl, err := filemutex.New(path)
	if err != nil {
		return nil, err
	}
	return fl, nil
}

func (s *Store) LoadData() error {
	cache := &data{}
	raw, err := os.ReadFile(s.metaFile)
	if err != nil {
		if os.IsNotExist(err) {
			f, err := os.Create(s.metaFile)
			if err != nil {
				return err
			}
			defer f.Close()
			_, err = f.WriteString("{}")
			if err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		if err = json.Unmarshal(raw, cache); err != nil {
			return err
		}
	}

	if cache.containers == nil {
		cache.containers = make(map[string]containerNetInfo)
	}
	s.cache = cache
	return nil
}

func (s *Store) Last() net.IP {
	return net.ParseIP(s.cache.Last)
}

func (s *Store) GetIPById(id string) (net.IP, bool) {
	for ip, info := range s.cache.containers {
		if info.ID == id {
			return net.ParseIP(ip), true
		}
	}
	return nil, false
}

func (s *Store) Add(ip net.IP, id, ifName string) error {
	if len(ip) > 0 {
		s.cache.containers[ip.String()] = containerNetInfo{id, ifName}
		s.cache.Last = ip.String()
		return s.Save()
	}
	return nil
}

func (s *Store) Remove(id string) error {
	for ip, info := range s.cache.containers {
		if info.ID == id {
			delete(s.cache.containers, ip)
			return s.Save()
		}
	}
	return nil
}

func (s *Store) Save() error {
	raw, err := json.Marshal(s.cache)
	if err != nil {
		return err
	}
	return os.WriteFile(s.metaFile, raw, 0644)
}
