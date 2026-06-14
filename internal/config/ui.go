package config

import "fmt"

// UISection maps YAML block `ui` — UI rendering knobs (widget CDNs, page size).
type UISection struct {
	MonacoCDN  string `yaml:"monaco_cdn,omitempty"`  // base URL for the Monaco editor assets (code widget)
	EasyMDECDN string `yaml:"easymde_cdn,omitempty"` // base URL for EasyMDE assets (markdown widget)
	PageSize   int    `yaml:"page_size,omitempty"`   // default list page size
}

func (u UISection) WithDefaults() UISection {
	if u.MonacoCDN == "" {
		u.MonacoCDN = DefaultMonacoCDN
	}
	if u.EasyMDECDN == "" {
		u.EasyMDECDN = DefaultEasyMDECDN
	}
	if u.PageSize == 0 {
		u.PageSize = DefaultPageSize
	}
	return u
}

func (u UISection) Validate() error {
	if u.PageSize < 1 {
		return fmt.Errorf("ui.page_size must be >= 1")
	}
	return nil
}
