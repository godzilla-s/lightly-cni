package store

import "github.com/alexflint/go-filemutex"

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
	dir  string
	data *data
}

func New() (*Store, error) {
	return nil, nil
}
