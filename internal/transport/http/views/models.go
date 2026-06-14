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

// Column is a list-view table column header.
type Column struct {
	Name  string
	Label string
}

// ListRow is one pre-rendered table row.
type ListRow struct {
	ID    string
	Cells []string
}

// ListView is the data for a resource list/table screen.
type ListView struct {
	Backend   resource.Backend
	Name      string
	Label     string
	Columns   []Column
	Rows      []ListRow
	CanCreate bool
}

// DetailPair is one field label/value in the detail view.
type DetailPair struct {
	Label string
	Value string
}

// DetailView is the data for a single-record detail screen.
type DetailView struct {
	Backend   resource.Backend
	Name      string
	Label     string
	ID        string
	Pairs     []DetailPair
	CanEdit   bool
	CanDelete bool
}

// ErrorView renders an error message in the main pane.
type ErrorView struct {
	Title   string
	Message string
}

// FormFieldView is one input in a create/edit form.
type FormFieldView struct {
	Name      string
	Label     string
	Kind      string // string|number|integer|boolean|enum|array|object
	InputType string // HTML input type for text-ish kinds
	Value     string
	Checked   bool // for boolean
	Enum      []string
	Required  bool
	Error     string
}

// DeleteView is the data for the delete-confirm modal.
type DeleteView struct {
	Backend resource.Backend
	Name    string
	Label   string
	ID      string
	Error   string
}

// FormView is the data for a create/edit form.
type FormView struct {
	Backend    resource.Backend
	Name       string
	Label      string
	Mode       string // create|edit
	ID         string
	Action     string // admin route the form submits to
	UsePut     bool   // edit submits via hx-put; create via hx-post
	CancelURL  string
	FormError  string // non-field-level error
	Fields     []FormFieldView
}
