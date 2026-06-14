// Package views holds the Templ components (compiled to *_templ.go) and their
// view models for the admin UI shell and resource screens.
package views

import (
	"github.com/a-h/templ"

	"github.com/Olian04/simple-admin/internal/domain/resource"
)

// SidebarItem is one resource link in the sidebar.
type SidebarItem struct {
	BackendID string
	Name      string
	Label     string
	Active    bool
}

// SidebarGroup groups a backend's resources under its label.
type SidebarGroup struct {
	Backend   resource.Backend
	Resources []SidebarItem
}

// PageData is the data for the full-page shell.
type PageData struct {
	Title   string
	Groups  []SidebarGroup
	Content templ.Component
}
