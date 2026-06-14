package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/a-h/templ"

	"github.com/Olian04/simple-admin/internal/domain/resource"
	"github.com/Olian04/simple-admin/internal/transport/http/views"
	"github.com/Olian04/simple-admin/internal/ui/render"
	"github.com/Olian04/simple-admin/internal/backend"
)

// UIHandler renders the admin UI screens from the resource registry.
type UIHandler struct {
	registry *resource.Registry
	widgets  *render.Registry
}

func NewUIHandler(registry *resource.Registry, widgets *render.Registry) *UIHandler {
	return &UIHandler{registry: registry, widgets: widgets}
}

// Index redirects to the first resource, or renders the placeholder shell when
// no backends are configured.
func (h *UIHandler) Index(w http.ResponseWriter, r *http.Request) {
	if rt, ok := h.registry.First(); ok {
		d := rt.Descriptor
		http.Redirect(w, r, "/ui/"+d.Backend.ID+"/"+d.Name, http.StatusSeeOther)
		return
	}
	data := h.pageData("", "", views.Placeholder())
	h.render(w, r, views.Page(data))
}

// List renders a resource's table view.
func (h *UIHandler) List(w http.ResponseWriter, r *http.Request) {
	rt, ok := h.lookup(w, r)
	if !ok {
		return
	}
	res, err := rt.Service.List(r.Context(), resource.ListParams{})
	if err != nil {
		h.handleReadError(w, r, rt.Descriptor, err)
		return
	}
	vm := buildListView(rt.Descriptor, res.Items)
	data := h.pageData(rt.Descriptor.Backend.ID, rt.Descriptor.Name, views.List(vm))
	h.render(w, r, views.Page(data))
}

// Detail renders a single record.
func (h *UIHandler) Detail(w http.ResponseWriter, r *http.Request) {
	rt, ok := h.lookup(w, r)
	if !ok {
		return
	}
	id := r.PathValue("id")
	rec, err := rt.Service.Get(r.Context(), id)
	if err != nil {
		h.handleReadError(w, r, rt.Descriptor, err)
		return
	}
	vm := buildDetailView(rt.Descriptor, id, rec)
	data := h.pageData(rt.Descriptor.Backend.ID, rt.Descriptor.Name, views.Detail(vm))
	h.render(w, r, views.Page(data))
}

// lookup resolves the {backend}/{name} runtime or writes a 404 page.
func (h *UIHandler) lookup(w http.ResponseWriter, r *http.Request) (*resource.Runtime, bool) {
	backendID := r.PathValue("backend")
	name := r.PathValue("name")
	rt, ok := h.registry.Lookup(backendID, name)
	if !ok {
		data := h.pageData("", "", views.Error(views.ErrorView{Title: "Not found", Message: "Unknown resource " + backendID + "/" + name}))
		w.WriteHeader(http.StatusNotFound)
		h.render(w, r, views.Page(data))
		return nil, false
	}
	return rt, true
}

// Credentials renders the per-backend credential entry page.
func (h *UIHandler) Credentials(w http.ResponseWriter, r *http.Request) {
	var bs []views.CredBackend
	for _, b := range h.registry.Backends() {
		auth := resource.AuthScheme{Type: "none"}
		if rts := h.registry.ResourcesFor(b.ID); len(rts) > 0 {
			auth = rts[0].Descriptor.AuthScheme
		}
		bs = append(bs, views.CredBackend{
			ID:        b.ID,
			Label:     b.Label,
			AuthType:  auth.Type,
			In:        auth.In,
			Name:      auth.Name,
			NeedsCred: auth.Type == "apikey",
		})
	}
	data := h.pageData("", "", views.Credentials(views.CredentialsView{Backends: bs}))
	h.render(w, r, views.Page(data))
}

// Options returns a JSON [{value,label}] list for relation pickers.
func (h *UIHandler) Options(w http.ResponseWriter, r *http.Request) {
	rt, ok := h.lookup(w, r)
	if !ok {
		return
	}
	res, err := rt.Service.List(r.Context(), resource.ListParams{PerPage: 200})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	labelField := firstStringField(rt.Descriptor)
	out := make([]map[string]string, 0, len(res.Items))
	for _, rec := range res.Items {
		id := render.DisplayValue(rec[rt.Descriptor.IDField])
		label := id
		if labelField != "" {
			if v := render.DisplayValue(rec[labelField]); v != "" {
				label = v
			}
		}
		out = append(out, map[string]string{"value": id, "label": label})
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// firstStringField returns a sensible label field for relation options.
func firstStringField(d resource.ResourceDescriptor) string {
	for _, f := range d.Schema.Fields {
		if f.Name == d.IDField || f.ReadOnly {
			continue
		}
		if f.Kind == resource.KindString && f.Format == "" {
			return f.Name
		}
	}
	return ""
}

// handleReadError maps upstream read errors to friendly pages.
func (h *UIHandler) handleReadError(w http.ResponseWriter, r *http.Request, d resource.ResourceDescriptor, err error) {
	switch {
	case errors.Is(err, backend.ErrUnauthorized):
		h.renderStatus(w, r, http.StatusUnauthorized, d, "Credential required",
			"The backend rejected the request. Set or update this backend's credential on the Credentials page (in the sidebar).")
	case errors.Is(err, backend.ErrNotFound):
		h.renderStatus(w, r, http.StatusNotFound, d, "Not found", "Record not found.")
	default:
		h.renderError(w, r, d, "Upstream error", err)
	}
}

func (h *UIHandler) renderError(w http.ResponseWriter, r *http.Request, d resource.ResourceDescriptor, title string, err error) {
	h.renderStatus(w, r, http.StatusBadGateway, d, title, err.Error())
}

func (h *UIHandler) renderStatus(w http.ResponseWriter, r *http.Request, status int, d resource.ResourceDescriptor, title, msg string) {
	data := h.pageData(d.Backend.ID, d.Name, views.Error(views.ErrorView{Title: title, Message: msg}))
	w.WriteHeader(status)
	h.render(w, r, views.Page(data))
}

func buildListView(d resource.ResourceDescriptor, items []resource.Record) views.ListView {
	cols := render.Columns(d.Schema.Fields)
	vcols := make([]views.Column, len(cols))
	for i, c := range cols {
		vcols[i] = views.Column{Name: c.Name, Label: c.Label}
	}
	rows := make([]views.ListRow, 0, len(items))
	for _, rec := range items {
		cells := make([]string, len(cols))
		for i, c := range cols {
			cells[i] = render.Cell(rec, c)
		}
		rows = append(rows, views.ListRow{ID: render.DisplayValue(rec[d.IDField]), Cells: cells})
	}
	return views.ListView{
		Backend:   d.Backend,
		Name:      d.Name,
		Label:     d.Label,
		Columns:   vcols,
		Rows:      rows,
		CanCreate: d.Capabilities.Create,
	}
}

func buildDetailView(d resource.ResourceDescriptor, id string, rec resource.Record) views.DetailView {
	pairs := make([]views.DetailPair, 0, len(d.Schema.Fields))
	for _, f := range d.Schema.Fields {
		pairs = append(pairs, views.DetailPair{Label: f.Label, Value: render.DisplayValue(rec[f.Name])})
	}
	return views.DetailView{
		Backend:   d.Backend,
		Name:      d.Name,
		Label:     d.Label,
		ID:        id,
		Pairs:     pairs,
		CanEdit:   d.Capabilities.Update,
		CanDelete: d.Capabilities.Delete,
	}
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
