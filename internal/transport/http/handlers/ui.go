package handlers

import (
	"net/http"

	"github.com/a-h/templ"

	"github.com/Olian04/simple-admin/internal/domain/resource"
	"github.com/Olian04/simple-admin/internal/transport/http/views"
)

// UIHandler renders the admin UI screens from the resource registry.
type UIHandler struct {
	registry *resource.Registry
}

func NewUIHandler(registry *resource.Registry) *UIHandler {
	return &UIHandler{registry: registry}
}

// Index renders the shell with a placeholder. (Resource routes + redirect to
// the first resource are added in a later step.)
func (h *UIHandler) Index(w http.ResponseWriter, r *http.Request) {
	data := h.pageData("", "", views.Placeholder())
	h.render(w, r, views.Page(data))
}

// pageData builds the shell view model, marking the active resource.
func (h *UIHandler) pageData(activeBackend, activeName string, content templ.Component) views.PageData {
	var groups []views.SidebarGroup
	for _, b := range h.registry.Backends() {
		var items []views.SidebarItem
		for _, rt := range h.registry.ResourcesFor(b.ID) {
			d := rt.Descriptor
			items = append(items, views.SidebarItem{
				BackendID: d.Backend.ID,
				Name:      d.Name,
				Label:     d.Label,
				Active:    d.Backend.ID == activeBackend && d.Name == activeName,
			})
		}
		groups = append(groups, views.SidebarGroup{Backend: b, Resources: items})
	}
	return views.PageData{Title: "simple-admin", Groups: groups, Content: content}
}

// render writes a templ component, logging render errors as a 500.
func (h *UIHandler) render(w http.ResponseWriter, r *http.Request, c templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := c.Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
