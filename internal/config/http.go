package config

import (
	"fmt"
	"strings"
)

// HTTPSection maps YAML block `http`.
type HTTPSection struct {
	ListenAddr string `yaml:"listen_addr,omitempty"`
}

func (h HTTPSection) WithDefaults() HTTPSection {
	if h.ListenAddr == "" {
		h.ListenAddr = DefaultHTTPListenAddr
	}
	return h
}

func (h HTTPSection) Validate() error {
	if strings.TrimSpace(h.ListenAddr) == "" {
		return fmt.Errorf("http.listen_addr must be non-empty")
	}
	return nil
}
