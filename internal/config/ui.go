package config

import "fmt"

// UISection maps YAML block `ui` — UI rendering knobs (page size).
type UISection struct {
	PageSize int `yaml:"page_size,omitempty"` // default list page size
}

func (u UISection) WithDefaults() UISection {
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
